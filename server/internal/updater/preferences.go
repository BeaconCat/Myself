package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const catalogueTimeout = 90 * time.Second

type Preferences struct {
	Repository  string `json:"repository"`
	AutoUpdate  bool   `json:"autoUpdate"`
	Subscribe   bool   `json:"subscribe"`
	Email       string `json:"email"`
	LastScanDay string `json:"-"`
}

type storedPreferences struct {
	Preferences
	LastScanDay string `json:"lastScanDay"`
	Token       string `json:"token,omitempty"`
}

func NormalizeRepository(value string) (string, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "https://github.com/")
	value = strings.TrimSuffix(strings.TrimSuffix(value, "/"), ".git")
	if !repoRE.MatchString(value) {
		return "", errors.New("invalid_repository")
	}
	return value, nil
}

func (m *Manager) Preferences() Preferences { m.mu.Lock(); defer m.mu.Unlock(); return m.prefs }

func (m *Manager) savePreferences(p Preferences) error {
	return m.savePreferencesWithToken(p, m.savedToken)
}

func (m *Manager) savePreferencesWithToken(p Preferences, token string) error {
	return writeJSONFile(filepath.Join(m.stateDir, "preferences.json"), storedPreferences{Preferences: p, LastScanDay: p.LastScanDay, Token: token})
}

func (m *Manager) Configure(p Preferences, day string) error {
	return m.ConfigureWithToken(p, day, nil)
}

// A nil token preserves the saved secret; an empty token clears the UI override.
// Environment credentials remain available as the deployment-level fallback.
func (m *Manager) ConfigureWithToken(p Preferences, day string, token *string) error {
	var replacement string
	if token != nil {
		replacement = strings.TrimSpace(*token)
		if len(replacement) > 4096 {
			return errors.New("invalid_update_token")
		}
		for _, ch := range replacement {
			if ch < 33 || ch > 126 {
				return errors.New("invalid_update_token")
			}
		}
	}
	repo, err := NormalizeRepository(p.Repository)
	if err != nil {
		return err
	}
	p.Repository = repo
	p.Email = strings.TrimSpace(p.Email)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy || activePhase(m.state.Phase) {
		return errors.New("update_in_progress")
	}
	if !strings.EqualFold(p.Repository, m.prefs.Repository) {
		p.AutoUpdate = false
	}
	if p.AutoUpdate && m.disabledReason() != "" {
		return errors.New(m.disabledReason())
	}
	p.LastScanDay = m.prefs.LastScanDay
	if p.Repository != m.prefs.Repository || ((!m.prefs.AutoUpdate && !m.prefs.Subscribe) && (p.AutoUpdate || p.Subscribe)) {
		p.LastScanDay = day
	}
	savedToken := m.savedToken
	if token != nil {
		savedToken = replacement
	}
	if err := m.savePreferencesWithToken(p, savedToken); err != nil {
		return err
	}
	if savedToken != m.savedToken {
		m.catalog = map[string]manifest{}
	}
	if p.Repository != m.opts.Repository || savedToken != m.savedToken {
		m.state.Available, m.state.Target = nil, nil
		m.state.Phase, m.state.Error, m.state.CheckedAt = "idle", "", ""
	}
	m.savedToken = savedToken
	m.token = savedToken
	if m.token == "" {
		m.token = m.envToken
	}
	m.prefs, m.opts.Repository = p, p.Repository
	m.state.Repository = p.Repository
	return m.save()
}

// ClaimDay persists before network activity so process restarts do not repeat a scan.
func (m *Manager) ClaimDay(day string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy || activePhase(m.state.Phase) || (!m.prefs.AutoUpdate && !m.prefs.Subscribe) || m.prefs.LastScanDay == day {
		return false, nil
	}
	p := m.prefs
	p.LastScanDay = day
	if err := m.savePreferences(p); err != nil {
		return false, err
	}
	m.prefs = p
	return true, nil
}

func (m *Manager) ReleaseDay(day string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.prefs.LastScanDay != day {
		return
	}
	p := m.prefs
	p.LastScanDay = ""
	if m.savePreferences(p) == nil {
		m.prefs = p
	}
}

type Notification struct {
	Repository string `json:"repository"`
	Version    string `json:"version"`
	Email      string `json:"email"`
	Status     string `json:"status"`
	At         string `json:"at"`
	Error      string `json:"error,omitempty"`
}

func notificationKey(repo, version string) string {
	numeric, _ := NumericVersion(version)
	sum := sha256.Sum256([]byte(strings.ToLower(repo) + "\x00" + numeric))
	return hex.EncodeToString(sum[:])
}

// Reserve before contacting SMTP: an ambiguous SMTP acknowledgement must never
// produce duplicate automatic mail. Interrupted/failed attempts remain visible.
func (m *Manager) ReserveNotification(n Notification) (bool, error) {
	dir := filepath.Join(m.stateDir, "notifications")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return false, err
	}
	file, err := os.OpenFile(filepath.Join(dir, notificationKey(n.Repository, n.Version)+".json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if os.IsExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	data, err := json.Marshal(n)
	if err == nil {
		_, err = file.Write(data)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return false, err
	}
	return true, writeJSONFile(filepath.Join(m.stateDir, "notification-last.json"), n)
}

func (m *Manager) FinishNotification(n Notification) error {
	if err := writeJSONFile(filepath.Join(m.stateDir, "notifications", notificationKey(n.Repository, n.Version)+".json"), n); err != nil {
		return err
	}
	return writeJSONFile(filepath.Join(m.stateDir, "notification-last.json"), n)
}

func (m *Manager) LastNotification() *Notification {
	data, err := os.ReadFile(filepath.Join(m.stateDir, "notification-last.json"))
	if err != nil {
		return nil
	}
	var n Notification
	if json.Unmarshal(data, &n) != nil {
		return nil
	}
	return &n
}
