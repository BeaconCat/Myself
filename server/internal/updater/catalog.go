package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type remoteAsset struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	UpdatedAt string `json:"updated_at"`
}
type remoteRelease struct {
	ID          int64         `json:"id"`
	Tag         string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	PublishedAt string        `json:"published_at"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	Assets      []remoteAsset `json:"assets"`
}

type ReleasePage struct {
	Items      []Release `json:"items"`
	Page       int       `json:"page"`
	PageSize   int       `json:"pageSize"`
	HasNext    bool      `json:"hasNext"`
	Repository string    `json:"repository"`
	Channel    string    `json:"channel"`
}

func (m *Manager) remotePage(ctx context.Context, page, size int) ([]remoteRelease, bool, error) {
	response, err := m.request(ctx, fmt.Sprintf("/repos/%s/releases?per_page=%d&page=%d", m.opts.Repository, size, page), false)
	if err != nil {
		return nil, false, err
	}
	next := strings.Contains(response.Header.Get("Link"), `rel="next"`)
	var releases []remoteRelease
	err = readJSON(response, &releases)
	return releases, next, err
}

func (m *Manager) resolve(ctx context.Context, remote remoteRelease, fresh bool) (*Release, error) {
	r := &Release{ID: remote.ID, Tag: remote.Tag, Name: remote.Name, Notes: remote.Body,
		PublishedAt: remote.PublishedAt, Prerelease: remote.Prerelease,
		URL: "https://github.com/" + m.opts.Repository + "/releases/tag/" + url.PathEscape(remote.Tag)}
	if r.Name == "" {
		r.Name = r.Tag
	}
	if _, ok := NumericVersion(remote.Tag); ok {
		r.Version = remote.Tag
	}
	if remote.Draft {
		r.Reason = "draft_release"
		return r, nil
	}
	var asset remoteAsset
	for _, candidate := range remote.Assets {
		if candidate.Name == "release-manifest.json" {
			asset = candidate
		}
	}
	if asset.ID <= 0 {
		r.Reason = "release_manifest_missing"
		return r, nil
	}
	key := fmt.Sprintf("%s/%d/%s/%d", strings.ToLower(m.opts.Repository), asset.ID, asset.UpdatedAt, asset.Size)
	metadata, cached := m.catalog[key]
	if fresh || !cached {
		response, err := m.request(ctx, m.assetPath(asset.ID), true)
		if err != nil {
			return nil, err
		}
		if err := readJSON(response, &metadata); err != nil {
			r.Reason = "invalid_release_manifest"
			return r, nil
		}
		if len(m.catalog) > 2000 {
			m.catalog = map[string]manifest{}
		}
		m.catalog[key] = metadata
	}
	if metadata.SchemaVersion != 1 || metadata.UpdateProtocol != Protocol || !commitRE.MatchString(metadata.Commit) {
		r.Reason = "unsupported_release_manifest"
		return r, nil
	}
	if _, ok := NumericVersion(metadata.Version); !ok {
		r.Reason = "invalid_release_version"
		return r, nil
	}
	r.Version, r.Codename, r.commit = metadata.Version, metadata.Codename, metadata.Commit
	if order, ok := CompareVersions(r.Version, m.opts.Build.Version); ok {
		r.Relation = "current"
		if order < 0 {
			r.Relation = "older"
		}
		if order > 0 {
			r.Relation = "newer"
		}
	}
	for _, item := range metadata.Artifacts {
		if item.Kind != "binary" || item.OS != m.opts.Build.OS || item.Arch != m.opts.Build.Arch {
			continue
		}
		if path.Base(item.Name) != item.Name || strings.Contains(item.Name, `\`) || !hashRE.MatchString(item.SHA256) || item.Size <= 0 || item.Size > maxBinarySize {
			r.Reason = "invalid_release_artifact"
			return r, nil
		}
		for _, candidate := range remote.Assets {
			if candidate.Name == item.Name && candidate.ID > 0 && candidate.Size == item.Size {
				r.Size, r.artifact, r.assetID, r.Installable = item.Size, item, candidate.ID, true
				return r, nil
			}
		}
	}
	r.Reason = "platform_not_available"
	return r, nil
}

func (m *Manager) saveCatalog() {
	_ = writeJSONFile(filepath.Join(m.stateDir, "catalog.json"), m.catalog)
}

func channelAllows(channel string, prerelease bool) bool {
	return !prerelease || channel == "preview"
}

// Preview includes published prereleases and stable releases; stable excludes prereleases.
// Never report a partial scan as "latest" if the catalogue exceeds the bound.
func (m *Manager) latest(ctx context.Context) (*Release, error) {
	defer m.saveCatalog()
	var best *Release
	total := 0
	for page := 1; page <= 20; page++ {
		items, next, err := m.remotePage(ctx, page, 100)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item.Draft || !channelAllows(m.prefs.Channel, item.Prerelease) {
				continue
			}
			total++
			r, err := m.resolve(ctx, item, false)
			if err != nil {
				return nil, err
			}
			order, comparable := 0, false
			if best != nil {
				order, comparable = CompareVersions(r.Version, best.Version)
			}
			if _, valid := NumericVersion(r.Version); valid && (best == nil || newer(r.Version, best.Version) || (comparable && order == 0 && best.Prerelease && !r.Prerelease)) {
				best = r
			}
		}
		if !next {
			if total == 0 {
				return nil, errors.New("no_releases")
			}
			if best == nil {
				return nil, errors.New("no_compatible_releases")
			}
			return best, nil
		}
	}
	return nil, errors.New("catalogue_too_large")
}

func (m *Manager) selected(ctx context.Context, id int64) (*Release, error) {
	if id <= 0 {
		return nil, errors.New("invalid_release_id")
	}
	response, err := m.request(ctx, fmt.Sprintf("/repos/%s/releases/%d", m.opts.Repository, id), false)
	if err != nil {
		return nil, err
	}
	var item remoteRelease
	if err := readJSON(response, &item); err != nil {
		return nil, err
	}
	if item.ID != id {
		return nil, errors.New("release_changed_check_again")
	}
	if !channelAllows(m.prefs.Channel, item.Prerelease) {
		return nil, errors.New("release_not_in_channel")
	}
	r, err := m.resolve(ctx, item, true)
	if err == nil && !r.Installable {
		err = errors.New(r.Reason)
	}
	return r, err
}

func (m *Manager) List(ctx context.Context, page, size int) (ReleasePage, error) {
	page, size = max(1, min(100000, page)), max(1, min(20, size))
	m.mu.Lock()
	if m.busy || activePhase(m.state.Phase) {
		m.mu.Unlock()
		return ReleasePage{}, errors.New("update_in_progress")
	}
	m.busy = true
	repo := m.opts.Repository
	channel := m.prefs.Channel
	m.mu.Unlock()
	defer func() { m.mu.Lock(); m.busy = false; m.mu.Unlock() }()
	ctx, cancel := context.WithTimeout(ctx, catalogueTimeout)
	defer cancel()
	out := ReleasePage{Items: []Release{}, Page: page, PageSize: size, Repository: repo, Channel: channel}
	defer m.saveCatalog()
	// Paginate after filtering, so a page of prereleases cannot hide later stable releases.
	seen, offset := 0, (page-1)*size
	for remotePage := 1; remotePage <= 20; remotePage++ {
		items, next, err := m.remotePage(ctx, remotePage, 100)
		if err != nil {
			return out, err
		}
		for _, item := range items {
			if item.Draft || !channelAllows(channel, item.Prerelease) {
				continue
			}
			if seen < offset {
				seen++
				continue
			}
			if len(out.Items) == size {
				out.HasNext = true
				return out, nil
			}
			r, err := m.resolve(ctx, item, false)
			if err != nil {
				return out, err
			}
			out.Items = append(out.Items, *r)
			seen++
		}
		if !next {
			return out, nil
		}
	}
	return out, errors.New("catalogue_too_large")
}

func (m *Manager) loadCatalog() {
	m.catalog = map[string]manifest{}
	if data, err := os.ReadFile(filepath.Join(m.stateDir, "catalog.json")); err == nil && len(data) <= 32<<20 {
		_ = json.Unmarshal(data, &m.catalog)
	}
	if m.catalog == nil {
		m.catalog = map[string]manifest{}
	}
}
