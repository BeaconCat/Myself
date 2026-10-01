package httpapi

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"myself/server/internal/updater"
)

func (s *Server) systemStatus(w http.ResponseWriter, r *http.Request) {
	s.writeSystemStatus(w, r, http.StatusOK)
}

func (s *Server) writeSystemStatus(w http.ResponseWriter, r *http.Request, code int) {
	if s.Updates == nil {
		writeJSON(w, http.StatusOK, map[string]any{"current": s.Build, "phase": "unavailable", "reason": "updates_unavailable", "canApply": false, "databaseDriver": s.DB.Driver()})
		return
	}
	status := s.Updates.Status()
	if s.updateScheduleRunning.Load() {
		status.Busy = true
		status.CanApply = false
	}
	writeJSON(w, code, struct {
		updater.Status
		DatabaseDriver string                `json:"databaseDriver"`
		SMTPReady      bool                  `json:"smtpReady"`
		AdminEmail     string                `json:"adminEmail"`
		NextCheck      string                `json:"nextCheck"`
		Timezone       string                `json:"timezone"`
		Notification   *updater.Notification `json:"notification,omitempty"`
	}{status, s.DB.Driver(), s.Config.Typed().Mail.Ready(), userOf(r).Email,
		nextUpdateCheck(time.Now(), s.siteLocation()).Format(time.RFC3339), s.siteLocation().String(), s.Updates.LastNotification()})
}

func (s *Server) checkUpdate(w http.ResponseWriter, r *http.Request) {
	if s.Updates == nil {
		writeError(w, http.StatusServiceUnavailable, "updates_unavailable")
		return
	}
	if !s.updateMu.TryLock() {
		writeError(w, http.StatusConflict, "update_in_progress")
		return
	}
	defer s.updateMu.Unlock()
	_, _ = s.Updates.Check(r.Context())
	s.systemStatus(w, r)
}

func (s *Server) applyUpdate(w http.ResponseWriter, r *http.Request) {
	if s.Updates == nil {
		writeError(w, http.StatusServiceUnavailable, "updates_unavailable")
		return
	}
	if !s.updateMu.TryLock() {
		writeError(w, http.StatusConflict, "update_in_progress")
		return
	}
	defer s.updateMu.Unlock()
	var b struct {
		Version    string `json:"version"`
		ReleaseID  int64  `json:"releaseId"`
		Repository string `json:"repository"`
	}
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	var err error
	if b.ReleaseID != 0 {
		err = s.Updates.StartSelected(r.Context(), b.Repository, b.ReleaseID)
	} else {
		err = s.Updates.Start(b.Version)
	}
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	s.writeSystemStatus(w, r, http.StatusAccepted)
}

func (s *Server) listReleases(w http.ResponseWriter, r *http.Request) {
	if s.Updates == nil {
		writeError(w, http.StatusServiceUnavailable, "updates_unavailable")
		return
	}
	page, err := s.Updates.List(r.Context(), queryInt(r, "page", 1, 1, 100000), queryInt(r, "pageSize", 10, 1, 20))
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (s *Server) updatePreferences(w http.ResponseWriter, r *http.Request) {
	if s.Updates == nil {
		writeError(w, http.StatusServiceUnavailable, "updates_unavailable")
		return
	}
	if !s.updateMu.TryLock() {
		writeError(w, http.StatusConflict, "update_in_progress")
		return
	}
	defer s.updateMu.Unlock()
	var p updater.Preferences
	if err := readJSON(w, r, &p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if p.Subscribe {
		if !s.Config.Typed().Mail.Ready() {
			writeError(w, http.StatusBadRequest, "smtp_required")
			return
		}
		if p.Email == "" {
			p.Email = userOf(r).Email
		}
		if !emailRe.MatchString(p.Email) {
			writeError(w, http.StatusBadRequest, "notification_email_required")
			return
		}
	}
	if err := s.Updates.Configure(p, time.Now().In(s.siteLocation()).Format("2006-01-02")); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	s.systemStatus(w, r)
}

func (s *Server) systemHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()
	if err := s.DB.PingContext(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database_unavailable")
		return
	}
	nonce := os.Getenv("MYSELF_UPDATE_NONCE")
	if nonce != "" && subtle.ConstantTimeCompare([]byte(nonce), []byte(r.Header.Get("X-Myself-Update-Probe"))) == 1 {
		writeJSON(w, http.StatusOK, map[string]any{"version": s.Build.Version, "commit": s.Build.Commit, "nonce": nonce, "pid": os.Getpid()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// PrepareUpdate waits for in-flight requests and refuses to interrupt media jobs.
func (s *Server) PrepareUpdate() (string, error) {
	if !s.jobs.pause() {
		return "", errors.New("job_running")
	}
	s.SetMaintenance()
	s.requests.Lock()
	s.requests.Unlock()
	name, err := s.createBackupWith(false)
	if err != nil {
		s.ResumeAfterUpdate()
		return "", err
	}
	return filepath.Join(s.BackupDir, name), nil
}

func (s *Server) ResumeAfterUpdate() {
	s.maintenance.Store(false)
	s.jobs.resume()
}

// RecoverUpdate is used only by the local updater helper while HTTP is stopped.
func (s *Server) RecoverUpdate(backup string) error {
	s.backupMu.Lock()
	defer s.backupMu.Unlock()
	return s.restoreBackup(backup)
}
