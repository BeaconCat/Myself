package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

var historyIDRE = regexp.MustCompile(`^[0-9a-f]{32}$`)

type HistoryEntry struct {
	ID          string `json:"id"`
	Build       Build  `json:"build"`
	CreatedAt   string `json:"createdAt"`
	Size        int64  `json:"size"`
	SHA256      string `json:"sha256"`
	Installable bool   `json:"installable"`
	Reason      string `json:"reason,omitempty"`
}

type HistoryPage struct {
	Items    []HistoryEntry `json:"items"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
	Total    int            `json:"total"`
	HasNext  bool           `json:"hasNext"`
}

// Reject symlinks at every managed level; metadata never supplies a file path.
func historyDir(stateDir, id string) (string, error) {
	if !historyIDRE.MatchString(id) {
		return "", errors.New("invalid_history_id")
	}
	dir := filepath.Join(stateDir, "history")
	for _, path := range []string{dir, filepath.Join(dir, id)} {
		info, err := os.Lstat(path)
		if err != nil {
			return "", err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("invalid_history_path")
		}
	}
	return filepath.Join(dir, id), nil
}

func readHistory(stateDir, id string) (HistoryEntry, error) {
	var entry HistoryEntry
	dir, err := historyDir(stateDir, id)
	if err != nil {
		return entry, err
	}
	info, err := os.Lstat(filepath.Join(dir, "entry.json"))
	if err != nil || !info.Mode().IsRegular() || info.Size() > 16384 {
		return entry, errors.New("invalid_history_entry")
	}
	data, err := os.ReadFile(filepath.Join(dir, "entry.json"))
	if err != nil {
		return entry, err
	}
	if json.Unmarshal(data, &entry) != nil || entry.ID != id || !hashRE.MatchString(entry.SHA256) || entry.Size < 1 || entry.Size > maxBinarySize || entry.Build.Version == "" {
		return entry, errors.New("invalid_history_entry")
	}
	if _, err := time.Parse(time.RFC3339Nano, entry.CreatedAt); err != nil {
		return entry, errors.New("invalid_history_entry")
	}
	info, err = os.Lstat(filepath.Join(dir, "program.bak"))
	if err != nil || !info.Mode().IsRegular() || info.Size() != entry.Size {
		return entry, errors.New("invalid_history_binary")
	}
	return entry, nil
}

func historyEntries(stateDir string) ([]HistoryEntry, error) {
	entries := []HistoryEntry{}
	dir := filepath.Join(stateDir, "history")
	info, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return entries, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("invalid_history_path")
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() || !historyIDRE.MatchString(file.Name()) {
			continue
		}
		if entry, err := readHistory(stateDir, file.Name()); err == nil {
			entries = append(entries, entry)
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		a, _ := time.Parse(time.RFC3339Nano, entries[i].CreatedAt)
		b, _ := time.Parse(time.RFC3339Nano, entries[j].CreatedAt)
		if a.Equal(b) {
			return entries[i].ID > entries[j].ID
		}
		return a.After(b)
	})
	return entries, nil
}

func archivePrevious(p Plan) error {
	stateDir := filepath.Join(p.Root, ".updates")
	if !historyIDRE.MatchString(p.Nonce) {
		return errors.New("invalid_history_id")
	}
	// An earlier helper or a status request may already have archived this plan.
	if entry, err := readHistory(stateDir, p.Nonce); err == nil {
		if entry.Build != p.Current {
			return errors.New("history_entry_conflict")
		}
		in, err := os.Open(filepath.Join(stateDir, "history", p.Nonce, "program.bak"))
		if err != nil {
			return err
		}
		hash := sha256.New()
		n, err := io.Copy(hash, io.LimitReader(in, entry.Size+1))
		in.Close()
		if err != nil {
			return err
		}
		if n != entry.Size || hex.EncodeToString(hash.Sum(nil)) != entry.SHA256 {
			return errors.New("history_binary_changed")
		}
		return nil
	}
	info, err := os.Lstat(p.Previous)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maxBinarySize {
		return errors.New("invalid_history_binary")
	}
	dir := filepath.Join(stateDir, "history")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	infoDir, err := os.Lstat(dir)
	if err != nil || !infoDir.IsDir() || infoDir.Mode()&os.ModeSymlink != 0 {
		return errors.New("invalid_history_path")
	}
	tmp, err := os.MkdirTemp(dir, ".archive-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	in, err := os.Open(p.Previous)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(filepath.Join(tmp, "program.bak"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	hash := sha256.New()
	n, err := io.Copy(io.MultiWriter(out, hash), io.LimitReader(in, maxBinarySize+1))
	syncErr, closeErr := out.Sync(), out.Close()
	if err != nil {
		return err
	}
	if syncErr != nil {
		return syncErr
	}
	if closeErr != nil {
		return closeErr
	}
	if n != info.Size() {
		return errors.New("history_binary_changed")
	}
	entry := HistoryEntry{ID: p.Nonce, Build: p.Current, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Size: n, SHA256: hex.EncodeToString(hash.Sum(nil))}
	if err := writeJSONFile(filepath.Join(tmp, "entry.json"), entry); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, p.Nonce))
}

func pruneHistory(stateDir string, limit int, protect string) error {
	if limit < 1 || limit > 20 {
		return errors.New("invalid_history_limit")
	}
	entries, err := historyEntries(stateDir)
	if err != nil {
		return err
	}
	kept := 0
	for _, entry := range entries {
		if entry.ID == protect {
			kept++
		}
	}
	for _, entry := range entries {
		if entry.ID == protect {
			continue
		}
		if kept < limit {
			kept++
			continue
		}
		dir, err := historyDir(stateDir, entry.ID)
		if err != nil {
			return err
		}
		if rel, err := filepath.Rel(filepath.Join(stateDir, "history"), dir); err != nil || rel != entry.ID {
			return errors.New("invalid_history_path")
		}
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) historyEligibility(entry *HistoryEntry) {
	entry.Reason = ""
	if entry.Build.OS != m.opts.Build.OS || entry.Build.Arch != m.opts.Build.Arch {
		entry.Reason = "platform_not_available"
	}
	if entry.Build.UpdateProtocol != Protocol {
		entry.Reason = "unsupported_release_manifest"
	}
	entry.Installable = entry.Reason == ""
}

func (m *Manager) History(page, size int) (HistoryPage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if page < 1 {
		page = 1
	}
	if page > 100000 {
		page = 100000
	}
	if size < 1 || size > 20 {
		size = 10
	}
	entries, err := historyEntries(m.stateDir)
	result := HistoryPage{Items: []HistoryEntry{}, Page: page, PageSize: size, Total: len(entries)}
	if err != nil {
		return result, err
	}
	start := (page - 1) * size
	if start >= len(entries) {
		return result, nil
	}
	end := min(start+size, len(entries))
	result.Items, result.HasNext = entries[start:end], end < len(entries)
	for i := range result.Items {
		m.historyEligibility(&result.Items[i])
	}
	return result, nil
}

func (m *Manager) historyRelease(id string) (*Release, error) {
	entry, err := readHistory(m.stateDir, id)
	if err != nil {
		return nil, errors.New("history_unavailable")
	}
	m.historyEligibility(&entry)
	if !entry.Installable {
		return nil, errors.New(entry.Reason)
	}
	return &Release{Version: entry.Build.Version, Codename: entry.Build.Codename, Name: "Myself " + entry.Build.Version, Size: entry.Size, Installable: true, commit: entry.Build.Commit, artifact: Artifact{SHA256: entry.SHA256}}, nil
}

func (m *Manager) StartHistory(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if reason := m.installDisabledReason(); reason != "" {
		return errors.New(reason)
	}
	if m.busy || activePhase(m.state.Phase) {
		return errors.New("update_in_progress")
	}
	release, err := m.historyRelease(id)
	if err != nil {
		return err
	}
	p := m.prefs
	p.AutoUpdate = false
	if err := m.savePreferences(p); err != nil {
		return err
	}
	m.prefs = p
	m.busy = true
	m.state.Target, m.state.Phase, m.state.Error = release, "verifying", ""
	m.state.Downloaded, m.state.Total = 0, release.Size
	if err := m.save(); err != nil {
		m.busy = false
		m.state.Phase = "failed"
		return err
	}
	go m.installFrom(release.Version, 0, id)
	return nil
}
