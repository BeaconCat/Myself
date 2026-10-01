package updater

import (
	"context"
	"errors"
)

// StartSelected allows an explicitly selected compatible version in either
// direction. A downgrade pauses unattended upgrades before accepting the job.
func (m *Manager) StartSelected(ctx context.Context, repository string, id int64) error {
	m.mu.Lock()
	if reason := m.disabledReason(); reason != "" {
		m.mu.Unlock()
		return errors.New(reason)
	}
	if m.busy || activePhase(m.state.Phase) {
		m.mu.Unlock()
		return errors.New("update_in_progress")
	}
	if repository != m.opts.Repository {
		m.mu.Unlock()
		return errors.New("repository_changed_reload")
	}
	m.busy = true
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, catalogueTimeout)
	defer cancel()
	release, err := m.selected(ctx, id)
	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil {
		m.busy = false
		return err
	}
	if cmp, ok := CompareVersions(release.Version, m.opts.Build.Version); ok && cmp < 0 && m.prefs.AutoUpdate {
		p := m.prefs
		p.AutoUpdate = false
		if err := m.savePreferences(p); err != nil {
			m.busy = false
			return err
		}
		m.prefs = p
	}
	m.state.Target, m.state.Phase, m.state.Error = release, "downloading", ""
	m.state.Downloaded, m.state.Total = 0, release.Size
	if err := m.save(); err != nil {
		m.busy = false
		m.state.Phase = "failed"
		return err
	}
	go m.install(release.Version, release.ID)
	return nil
}
