package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"myself/server/internal/store"
)

// publicationInput accepts RFC3339 or the schedule picker's wall time in the site timezone.
// Missing fields on updates preserve a pending schedule; explicit draft/published cancels it.
func (s *Server) publicationInput(b body, previousStatus, previousAt string) (string, string, error) {
	status := previousStatus
	if _, exists := b["status"]; exists {
		var ok bool
		status, ok = b.str("status")
		if !ok {
			return "", "", errors.New("invalid_status")
		}
	}
	if status != "draft" && status != "published" && status != "scheduled" {
		return "", "", errors.New("invalid_status")
	}
	if status != "scheduled" {
		return status, "", nil
	}
	value := previousAt
	if _, exists := b["publishAt"]; exists {
		value = b.strOr("publishAt")
	}
	when, ok := s.parsePublishTime(value)
	if !ok {
		return "", "", errors.New("invalid_publish_time")
	}
	when = when.UTC()
	if when.Nanosecond() != 0 {
		when = when.Truncate(time.Second).Add(time.Second)
	}
	old, same := s.parsePublishTime(previousAt)
	if !when.After(time.Now()) && !(previousStatus == "scheduled" && same && old.Equal(when)) {
		return "", "", errors.New("publish_time_must_be_future")
	}
	return status, when.UTC().Format("2006-01-02 15:04:05"), nil
}

func (s *Server) parsePublishTime(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}
	if when, err := time.Parse(time.RFC3339, value); err == nil {
		return when, true
	}
	// Stored UTC timestamps are accepted when preserving an existing row.
	if when, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return when, true
	}
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02T15:04:05"} {
		when, err := time.ParseInLocation(layout, value, s.siteLocation())
		// Reject wall times skipped by a daylight-saving transition.
		if err == nil && when.In(s.siteLocation()).Format(layout) == value {
			return when, true
		}
	}
	return time.Time{}, false
}

// StartPublicationSchedule checks each second and catches up after restarts.
// State transitions are atomic and retry-safe.
func (s *Server) StartPublicationSchedule(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			if err := s.publishDue(ctx, time.Now()); err != nil && ctx.Err() == nil {
				log.Printf("[publication] %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (s *Server) publishDue(ctx context.Context, now time.Time) error {
	s.requests.RLock()
	defer s.requests.RUnlock()
	if ctx.Err() != nil || s.maintenance.Load() || (s.Updates != nil && s.Updates.Pending()) {
		return nil
	}
	s.publicationMu.Lock()
	defer s.publicationMu.Unlock()
	direct := boolInt(s.Config.Typed().Users.Authors.DirectPublish)
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stamp := now.UTC().Format("2006-01-02 15:04:05")
	// A revoked author's queued publication must not bypass the current review policy.
	allowed := `(author_id IS NULL OR EXISTS (SELECT 1 FROM users WHERE users.id = posts.author_id AND users.status = 'active' AND (users.role = 'admin' OR (users.role = 'author' AND ? = 1))))`
	if _, err = tx.ExecContext(ctx, `UPDATE posts SET status = 'draft', publish_at = '', updated_at = ? WHERE status = 'scheduled' AND publish_at != '' AND publish_at <= ? AND NOT `+allowed, stamp, stamp, direct); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE posts SET status = 'published', created_at = publish_at, publish_at = '', updated_at = ? WHERE status = 'scheduled' AND publish_at != '' AND publish_at <= ?`, stamp, stamp); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE notes SET status = 'published', created_at = publish_at, publish_at = '' WHERE status = 'scheduled' AND publish_at != '' AND publish_at <= ?`, stamp); err != nil {
		return err
	}
	return tx.Commit()
}

// recordNoteView counts public detail visits, not list renders or editor requests.
// A signed visitor/user identity is deduplicated for 30 minutes in this process.
func (s *Server) recordNoteView(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var views int64
	if err := s.DB.QueryRow(`SELECT views FROM notes WHERE id = ? AND `+store.PublicNote, id).Scan(&views); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found")
		} else {
			fail(w, err)
		}
		return
	}
	voter := s.voterOf(w, r, true)
	if voter == "" {
		writeError(w, http.StatusServiceUnavailable, "view_unavailable")
		return
	}
	key := strconv.FormatInt(id, 10) + ":" + voter
	now := time.Now()
	s.viewMu.Lock()
	defer s.viewMu.Unlock()
	if s.viewed == nil {
		s.viewed = map[string]time.Time{}
	}
	if at, exists := s.viewed[key]; !exists || now.Sub(at) >= 30*time.Minute {
		if len(s.viewed) >= 10000 {
			oldestKey, oldest := "", now
			for k, at := range s.viewed {
				if now.Sub(at) >= 30*time.Minute {
					delete(s.viewed, k)
				} else if at.Before(oldest) {
					oldestKey, oldest = k, at
				}
			}
			if len(s.viewed) >= 10000 {
				delete(s.viewed, oldestKey)
			}
		}
		res, err := s.DB.Exec(`UPDATE notes SET views = views + 1 WHERE id = ? AND `+store.PublicNote, id)
		if err != nil {
			fail(w, err)
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		s.viewed[key] = now
	}
	if err := s.DB.QueryRow(`SELECT views FROM notes WHERE id = ? AND `+store.PublicNote, id).Scan(&views); err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]int64{"views": views})
}

func publicationSaved(w http.ResponseWriter, result sql.Result, status, at string) {
	affected, _ := result.RowsAffected()
	if affected == 0 {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": status, "publishAt": store.PublishTime(at)})
}
