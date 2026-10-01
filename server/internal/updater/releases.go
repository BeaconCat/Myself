// Package updater checks a fixed GitHub release source and installs verified
// native binaries through a separate helper process after graceful shutdown.
package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const Protocol = 1
const maxBinarySize int64 = 512 << 20

type Build struct {
	Version        string `json:"version"`
	Commit         string `json:"commit"`
	Date           string `json:"date"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
	UpdateProtocol int    `json:"updateProtocol"`
}

type Artifact struct {
	Name   string `json:"name"`
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	Kind   string `json:"kind"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type manifest struct {
	SchemaVersion  int        `json:"schemaVersion"`
	UpdateProtocol int        `json:"updateProtocol"`
	Version        string     `json:"version"`
	Commit         string     `json:"commit"`
	Artifacts      []Artifact `json:"artifacts"`
}

type Release struct {
	Version     string `json:"version"`
	Name        string `json:"name"`
	Notes       string `json:"notes"`
	URL         string `json:"url"`
	PublishedAt string `json:"publishedAt"`
	Size        int64  `json:"size"`
	artifact    Artifact
	assetID     int64
	commit      string
}

var versionRE = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z][0-9A-Za-z.-]*))?$`)
var repoRE = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
var hashRE = regexp.MustCompile(`^[0-9a-f]{64}$`)
var commitRE = regexp.MustCompile(`^[0-9a-f]{40}$`)

func newer(stable, current string) bool {
	a, b := versionRE.FindStringSubmatch(stable), versionRE.FindStringSubmatch(current)
	if a == nil || b == nil || a[4] != "" {
		return false
	}
	for i := 1; i <= 3; i++ {
		x, e1 := strconv.ParseUint(a[i], 10, 64)
		y, e2 := strconv.ParseUint(b[i], 10, 64)
		if e1 != nil || e2 != nil {
			return false
		}
		if x != y {
			return x > y
		}
	}
	return b[4] != ""
}

func newClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		req.Header.Del("Authorization")
		host := req.URL.Hostname()
		if req.URL.Scheme != "https" || (host != "github.com" && host != "api.github.com" && !strings.HasSuffix(host, ".githubusercontent.com")) {
			return errors.New("untrusted release redirect")
		}
		return nil
	}}
}

func (m *Manager) request(ctx context.Context, endpoint string, binary bool) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.apiBase+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Myself/"+m.opts.Build.Version)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/vnd.github+json")
	if binary {
		req.Header.Set("Accept", "application/octet-stream")
	}
	if m.token != "" {
		req.Header.Set("Authorization", "Bearer "+m.token)
	}
	response, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		if response.StatusCode == http.StatusNotFound {
			return nil, errors.New("no_release")
		}
		return nil, fmt.Errorf("release server returned HTTP %d", response.StatusCode)
	}
	return response, nil
}

func readJSON(response *http.Response, target any) error {
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil {
		return err
	}
	if len(data) > 1<<20 {
		return errors.New("release metadata too large")
	}
	return json.Unmarshal(data, target)
}

func (m *Manager) latest(ctx context.Context) (*Release, error) {
	var remote struct {
		Tag         string `json:"tag_name"`
		Name        string `json:"name"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		Draft       bool   `json:"draft"`
		Prerelease  bool   `json:"prerelease"`
		Assets      []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"assets"`
	}
	response, err := m.request(ctx, "/repos/"+m.opts.Repository+"/releases/latest", false)
	if err != nil {
		return nil, err
	}
	if err := readJSON(response, &remote); err != nil {
		return nil, err
	}
	match := versionRE.FindStringSubmatch(remote.Tag)
	if remote.Draft || remote.Prerelease || match == nil || match[4] != "" {
		return nil, errors.New("invalid stable release")
	}
	var metadataID int64
	for _, asset := range remote.Assets {
		if asset.Name == "release-manifest.json" {
			metadataID = asset.ID
		}
	}
	if metadataID <= 0 {
		return nil, errors.New("release_manifest_missing")
	}
	response, err = m.request(ctx, m.assetPath(metadataID), true)
	if err != nil {
		return nil, err
	}
	var metadata manifest
	if err := readJSON(response, &metadata); err != nil {
		return nil, err
	}
	if metadata.SchemaVersion != 1 || metadata.UpdateProtocol != Protocol || metadata.Version != remote.Tag || !commitRE.MatchString(metadata.Commit) {
		return nil, errors.New("unsupported_release_manifest")
	}
	for _, item := range metadata.Artifacts {
		if item.Kind != "binary" || item.OS != m.opts.Build.OS || item.Arch != m.opts.Build.Arch {
			continue
		}
		if path.Base(item.Name) != item.Name || strings.ContainsAny(item.Name, `\`) || !hashRE.MatchString(item.SHA256) || item.Size <= 0 || item.Size > maxBinarySize {
			return nil, errors.New("invalid release artifact")
		}
		for _, asset := range remote.Assets {
			if asset.Name == item.Name && asset.ID > 0 && asset.Size == item.Size {
				return &Release{Version: remote.Tag, Name: remote.Name, Notes: remote.Body, PublishedAt: remote.PublishedAt,
					URL:  "https://github.com/" + m.opts.Repository + "/releases/tag/" + url.PathEscape(remote.Tag),
					Size: item.Size, artifact: item, assetID: asset.ID, commit: metadata.Commit}, nil
			}
		}
	}
	return nil, errors.New("platform_not_available")
}

func (m *Manager) assetPath(id int64) string {
	return "/repos/" + m.opts.Repository + "/releases/assets/" + strconv.FormatInt(id, 10)
}
