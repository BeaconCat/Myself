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

func (s *Server) systemStatus(w http.ResponseWriter, _ *http.Request) {
	if s.Updates == nil {
		writeError(w, http.StatusServiceUnavailable, "updates_unavailable")
		return
	}
	writeJSON(w, http.StatusOK, struct {
		updater.Status
		DatabaseDriver string `json:"databaseDriver"`
	}{s.Updates.Status(), s.DB.Driver()})
}

func (s *Server) checkUpdate(w http.ResponseWriter, r *http.Request) {
	if s.Updates == nil {
		writeError(w, http.StatusServiceUnavailable, "updates_unavailable")
		return
	}
	status, _ := s.Updates.Check(r.Context())
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) applyUpdate(w http.ResponseWriter, r *http.Request) {
	if s.Updates == nil {
		writeError(w, http.StatusServiceUnavailable, "updates_unavailable")
		return
	}
	var b struct {
		Version string `json:"version"`
	}
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if err := s.Updates.Start(b.Version); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, s.Updates.Status())
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
