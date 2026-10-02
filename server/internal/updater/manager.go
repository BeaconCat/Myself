package updater

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Options struct {
	Build      Build
	Repository string
	Root       string
	Port       string
	Args       []string
	Disabled   bool
	Transport  http.RoundTripper
	// Prepare quiesces the site and returns a complete safety backup.
	Prepare func() (string, error)
	Resume  func()
	Ready   func(string) error
}

type Status struct {
	Current         Build       `json:"current"`
	Repository      string      `json:"repository"`
	Phase           string      `json:"phase"`
	CheckedAt       string      `json:"checkedAt,omitempty"`
	Available       *Release    `json:"available,omitempty"`
	Target          *Release    `json:"target,omitempty"`
	Preferences     Preferences `json:"preferences"`
	TokenConfigured bool        `json:"tokenConfigured"`
	TokenSource     string      `json:"tokenSource"`
	CanApply        bool        `json:"canApply"`
	Busy            bool        `json:"busy"`
	Reason          string      `json:"reason,omitempty"`
	AutomaticReason string      `json:"automaticReason,omitempty"`
	HistoryError    string      `json:"historyError,omitempty"`
	Downloaded      int64       `json:"downloaded"`
	Total           int64       `json:"total"`
	Error           string      `json:"error,omitempty"`
	Backup          string      `json:"backup,omitempty"`
}

type Manager struct {
	mu         sync.Mutex
	opts       Options
	state      Status
	busy       bool
	pending    bool
	client     *http.Client
	apiBase    string
	token      string
	savedToken string
	envToken   string
	executable string
	stateDir   string
	prefs      Preferences
	catalog    map[string]manifest
}

func New(opts Options) (*Manager, error) {
	if opts.Repository == "" {
		opts.Repository = opts.Build.Repository
		if opts.Repository == "" {
			opts.Repository = "BeaconCat/Myself"
		}
	}
	if !repoRE.MatchString(opts.Repository) {
		return nil, errors.New("invalid update repository")
	}
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(opts.Root, ".updates")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	m := &Manager{opts: opts, executable: exe, stateDir: dir, apiBase: "https://api.github.com", client: newClient(), token: os.Getenv("MYSELF_UPDATE_TOKEN"), pending: os.Getenv("MYSELF_UPDATE_NONCE") != ""}
	m.envToken = m.token
	if opts.Transport != nil {
		m.client.Transport = opts.Transport
	}
	m.state = Status{Current: opts.Build, Repository: opts.Repository, Phase: "idle"}
	m.prefs.Repository = opts.Repository
	m.prefs.Channel = "stable"
	if data, err := os.ReadFile(filepath.Join(dir, "preferences.json")); err == nil {
		var stored storedPreferences
		if json.Unmarshal(data, &stored) == nil {
			if repo, err := NormalizeRepository(stored.Repository); err == nil {
				m.prefs = stored.Preferences
				// Before channels existed, saved preferences included prereleases.
				if m.prefs.Channel == "" {
					m.prefs.Channel = "preview"
				}
				if m.prefs.Channel != "preview" && m.prefs.Channel != "stable" {
					m.prefs.Channel = "stable"
				}
				m.prefs.Repository = repo
				m.prefs.LastScanDay = stored.LastScanDay
				m.opts.Repository = repo
				m.savedToken = stored.Token
				if m.savedToken != "" {
					m.token = m.savedToken
				}
			}
		}
	}
	if m.prefs.HistoryLimit < 1 || m.prefs.HistoryLimit > 20 {
		m.prefs.HistoryLimit = 3
	}
	m.loadCatalog()
	if data, err := os.ReadFile(filepath.Join(dir, "status.json")); err == nil {
		_ = json.Unmarshal(data, &m.state)
		if m.state.Repository != m.opts.Repository && !m.pending {
			m.state.Available, m.state.Target = nil, nil
			m.state.Phase, m.state.Error = "idle", ""
		}
	}
	m.state.Current, m.state.Repository = opts.Build, m.opts.Repository
	if !m.pending && m.state.Available != nil && !channelAllows(m.prefs.Channel, m.state.Available.Prerelease) {
		m.state.Available = nil
		if m.state.Phase == "available" {
			m.state.Phase = "idle"
		}
	}
	if !m.pending && activePhase(m.state.Phase) {
		data, _ := os.ReadFile(exe + ".update-lock")
		pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		if !processRunning(pid) {
			m.state.Phase, m.state.Error = "failed", "previous_update_interrupted"
			_ = os.Remove(exe + ".update-lock")
			_ = m.save()
		}
	}
	return m, nil
}

func (m *Manager) disabledReason() string {
	if reason := m.installDisabledReason(); reason != "" {
		return reason
	}
	if !versionRE.MatchString(m.opts.Build.Version) {
		return "development_build"
	}
	return ""
}

func (m *Manager) installDisabledReason() string {
	if m.opts.Disabled {
		return "disabled_by_operator"
	}
	// go run places its program under go-build<nonce>/b<id>/exe/.
	// Test executables also use go-build directories, without the exe parent.
	if filepath.Base(filepath.Dir(m.executable)) == "exe" {
		for _, part := range strings.FieldsFunc(m.executable, func(r rune) bool { return r == '/' || r == '\\' }) {
			if strings.HasPrefix(part, "go-build") {
				if _, err := strconv.ParseUint(strings.TrimPrefix(part, "go-build"), 10, 64); err == nil {
					return "temporary_executable"
				}
			}
		}
	}
	if m.opts.Prepare == nil || m.opts.Ready == nil {
		return "not_configured"
	}
	return ""
}

func activePhase(phase string) bool {
	return phase == "downloading" || phase == "verifying" || phase == "backing_up" || phase == "restarting" || phase == "rolling_back"
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pending || m.state.Phase == "restarting" || m.state.Phase == "rolling_back" {
		if data, err := os.ReadFile(filepath.Join(m.stateDir, "status.json")); err == nil {
			_ = json.Unmarshal(data, &m.state)
		}
	}
	s := m.state
	if s.Available != nil && !channelAllows(m.prefs.Channel, s.Available.Prerelease) {
		s.Available = nil
		if s.Phase == "available" {
			s.Phase = "idle"
		}
	}
	if s.Phase == "installed" || s.Phase == "rolled_back" {
		if plan, err := ReadPlan(filepath.Join(m.stateDir, "plan.json")); err == nil && plan.Executable == m.executable {
			_ = os.Remove(plan.Helper)
			if s.Phase == "installed" {
				if _, err := os.Lstat(plan.Previous); err == nil {
					if err := archivePrevious(plan); err != nil {
						s.HistoryError = "history_archive_failed"
					} else {
						_ = os.Remove(plan.Previous)
					}
				}
				if err := pruneHistory(m.stateDir, m.prefs.HistoryLimit, plan.Nonce); err != nil {
					s.HistoryError = "history_cleanup_failed"
				}
			}
		}
	}
	s.Current, s.Repository = m.opts.Build, m.opts.Repository
	s.Preferences = m.prefs
	s.TokenConfigured = m.token != ""
	s.TokenSource = "none"
	if m.savedToken != "" {
		s.TokenSource = "settings"
	} else if m.envToken != "" {
		s.TokenSource = "environment"
	}
	s.Busy = m.busy || activePhase(s.Phase)
	s.Reason, s.AutomaticReason = m.installDisabledReason(), m.disabledReason()
	_, currentKnown := NumericVersion(s.Current.Version)
	s.CanApply = s.Reason == "" && !m.busy && !activePhase(s.Phase) && s.Available != nil && s.Available.Installable && channelAllows(m.prefs.Channel, s.Available.Prerelease) && (!currentKnown || newer(s.Available.Version, s.Current.Version))
	return s
}

func writeJSONFile(filename string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(filename), ".update-json-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filename)
}

func (m *Manager) save() error {
	return writeJSONFile(filepath.Join(m.stateDir, "status.json"), m.state)
}

func (m *Manager) Check(ctx context.Context) (Status, error) {
	m.mu.Lock()
	if m.busy || activePhase(m.state.Phase) {
		m.mu.Unlock()
		return m.Status(), errors.New("update_in_progress")
	}
	m.busy = true
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, catalogueTimeout)
	defer cancel()
	release, err := m.latest(ctx)
	m.mu.Lock()
	m.busy = false
	m.state.CheckedAt = time.Now().UTC().Format(time.RFC3339)
	m.state.Error = ""
	m.state.Available = nil
	if err != nil {
		m.state.Phase, m.state.Error = "check_failed", err.Error()
		if err.Error() == "no_releases" || err.Error() == "no_compatible_releases" {
			m.state.Phase, m.state.Error = err.Error(), ""
			err = nil
		}
	} else {
		m.state.Phase = "up_to_date"
		_, currentKnown := NumericVersion(m.opts.Build.Version)
		if !currentKnown || newer(release.Version, m.opts.Build.Version) {
			m.state.Phase, m.state.Available = "available", release
		}
	}
	_ = m.save()
	m.mu.Unlock()
	return m.Status(), err
}

// Start rechecks release metadata so a stale UI cannot install another version.
func (m *Manager) Start(version string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if reason := m.disabledReason(); reason != "" {
		return errors.New(reason)
	}
	if m.busy || activePhase(m.state.Phase) {
		return errors.New("update_in_progress")
	}
	if m.state.Available == nil || !m.state.Available.Installable || !channelAllows(m.prefs.Channel, m.state.Available.Prerelease) || m.state.Available.Version != version || !newer(version, m.opts.Build.Version) {
		return errors.New("check_update_first")
	}
	m.busy = true
	m.state.Phase, m.state.Error, m.state.Downloaded = "downloading", "", 0
	m.state.Total = m.state.Available.Size
	m.state.Target = m.state.Available
	if err := m.save(); err != nil {
		m.busy = false
		m.state.Phase = "failed"
		return err
	}
	go m.install(version, m.state.Available.ID)
	return nil
}

func (m *Manager) phase(phase string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Phase = phase
	return m.save()
}

func (m *Manager) install(version string, id int64) {
	m.installFrom(version, id, "")
}

func (m *Manager) installFrom(version string, id int64, historyID string) {
	var installErr error
	var staged, helper, lockPath string
	prepared, dispatched := false, false
	defer func() {
		if dispatched {
			return
		}
		if prepared && m.opts.Resume != nil {
			m.opts.Resume()
		}
		for _, file := range []string{staged, helper, lockPath} {
			if file != "" {
				_ = os.Remove(file)
			}
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		m.busy = false
		m.state.Phase = "failed"
		if installErr != nil {
			m.state.Error = installErr.Error()
		}
		_ = m.save()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	var release *Release
	var err error
	if historyID == "" {
		release, err = m.selected(ctx, id)
	} else {
		release, err = m.historyRelease(historyID)
	}
	if err != nil {
		installErr = err
		return
	}
	if release.Version != version {
		installErr = errors.New("release_changed_check_again")
		return
	}
	lock := m.executable + ".update-lock"
	file, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		installErr = fmt.Errorf("cannot lock executable directory: %w", err)
		return
	}
	file.Close()
	lockPath = lock
	if err := os.WriteFile(lock, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		installErr = err
		return
	}
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		installErr = err
		return
	}
	nonce := hex.EncodeToString(nonceBytes)
	staged = m.executable + ".next-" + nonce
	if filepath.Ext(m.executable) == ".exe" {
		staged += ".exe"
	}
	var source io.ReadCloser
	if historyID == "" {
		response, requestErr := m.request(ctx, m.assetPath(release.assetID), true)
		err = requestErr
		if err == nil {
			source = response.Body
		}
	} else {
		source, err = os.Open(filepath.Join(m.stateDir, "history", historyID, "program.bak"))
	}
	if err != nil {
		installErr = err
		return
	}
	out, err := os.OpenFile(staged, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		source.Close()
		installErr = err
		return
	}
	hash := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(out, hash, &progressWriter{m: m}), io.LimitReader(source, release.Size+1))
	source.Close()
	syncErr := out.Sync()
	closeErr := out.Close()
	if copyErr != nil {
		installErr = copyErr
		return
	}
	if syncErr != nil {
		installErr = syncErr
		return
	}
	if closeErr != nil {
		installErr = closeErr
		return
	}
	if n != release.Size || hex.EncodeToString(hash.Sum(nil)) != release.artifact.SHA256 {
		installErr = errors.New("checksum_mismatch")
		return
	}
	if err := m.phase("verifying"); err != nil {
		installErr = err
		return
	}
	probeCtx, stop := context.WithTimeout(ctx, 15*time.Second)
	cmd := exec.CommandContext(probeCtx, staged, "--version-json")
	hideWindow(cmd)
	metadata, err := cmd.Output()
	stop()
	var build Build
	decodeErr := json.Unmarshal(metadata, &build)
	versionOrder, versionValid := CompareVersions(build.Version, version)
	if historyID != "" {
		versionValid, versionOrder = build.Version == version, 0
	}
	if err != nil || decodeErr != nil || !versionValid || versionOrder != 0 || build.Commit != release.commit || build.OS != m.opts.Build.OS || build.Arch != m.opts.Build.Arch || build.UpdateProtocol != Protocol || (release.Codename != "" && build.Codename != release.Codename) {
		installErr = errors.New("binary_verification_failed")
		return
	}
	helper = m.executable + ".updater-" + nonce
	if filepath.Ext(m.executable) == ".exe" {
		helper += ".exe"
	}
	if err := copyExecutable(m.executable, helper); err != nil {
		installErr = err
		return
	}
	if err := m.phase("backing_up"); err != nil {
		installErr = err
		return
	}
	prepared = true
	backup, err := m.opts.Prepare()
	if err != nil {
		installErr = err
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		installErr = err
		return
	}
	plan := Plan{Executable: m.executable, Staged: staged, Previous: m.executable + ".previous-" + nonce,
		Helper: helper, Root: m.opts.Root, Backup: backup, Args: m.opts.Args, CWD: cwd, ParentPID: os.Getpid(),
		Port: m.opts.Port, Version: build.Version, Commit: release.commit, Nonce: nonce, SHA256: release.artifact.SHA256, Current: m.opts.Build}
	planPath := filepath.Join(m.stateDir, "plan.json")
	if err := writeJSONFile(planPath, plan); err != nil {
		installErr = err
		return
	}
	m.mu.Lock()
	m.state.Backup, m.state.Phase = filepath.Base(backup), "restarting"
	err = m.save()
	m.mu.Unlock()
	if err != nil {
		installErr = err
		return
	}
	if err := m.opts.Ready(planPath); err != nil {
		installErr = err
		return
	}
	dispatched = true
}

type progressWriter struct {
	m    *Manager
	last time.Time
}

func (p *progressWriter) Write(data []byte) (int, error) {
	p.m.mu.Lock()
	defer p.m.mu.Unlock()
	p.m.state.Downloaded += int64(len(data))
	if time.Since(p.last) > time.Second {
		_ = p.m.save()
		p.last = time.Now()
	}
	return len(data), nil
}
