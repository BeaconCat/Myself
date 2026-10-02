package httpapi

import (
	"context"
	"fmt"
	"log"
	"time"

	"myself/server/internal/config"
	"myself/server/internal/updater"
)

func nextUpdateCheck(now time.Time, location *time.Location) time.Time {
	local := now.In(location)
	return time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, location)
}

// Check on the first tick after local midnight; a missed day is caught up after
// restart. The persisted day and notification reservations survive upgrades.
func (s *Server) StartUpdateSchedule(ctx context.Context) {
	if s.Updates == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			if err := s.runUpdateCycle(ctx, time.Now(), s.Updates.Check); err != nil {
				log.Printf("[update schedule] %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (s *Server) runUpdateCycle(ctx context.Context, now time.Time, check func(context.Context) (updater.Status, error)) error {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()
	if s.Updates == nil || s.maintenance.Load() || s.Updates.Pending() || ctx.Err() != nil {
		return nil
	}
	p := s.Updates.Preferences()
	if !p.AutoUpdate && (!p.Subscribe || !s.Config.Typed().Mail.Ready()) {
		return nil
	}
	day := now.In(s.siteLocation()).Format("2006-01-02")
	claimed, err := s.Updates.ClaimDay(day)
	if err != nil || !claimed {
		return err
	}
	s.updateScheduleRunning.Store(true)
	defer s.updateScheduleRunning.Store(false)
	status, err := check(ctx)
	if err != nil {
		if err.Error() == "update_in_progress" {
			s.Updates.ReleaseDay(day)
		}
		return err
	}
	if status.Available == nil {
		return nil
	}
	r := status.Available
	if p.Channel == "stable" && r.Prerelease {
		return nil
	}
	if p.Subscribe && s.Config.Typed().Mail.Ready() && emailRe.MatchString(p.Email) {
		n := updater.Notification{Repository: p.Repository, Version: r.Version, Email: p.Email, Status: "attempting", At: now.UTC().Format(time.RFC3339)}
		reserved, err := s.Updates.ReserveNotification(n)
		if err != nil {
			return err
		}
		if reserved {
			err := s.sendLetter(p.Email, letter{
				Subject: "Myself 有新版本：" + r.Version, Title: "发现新版本", Greeting: "站长，你好：",
				Lines:  []string{fmt.Sprintf("%s 发布了 %s（数字版本 %s）。", p.Repository, r.Name, r.Version), "当前运行版本：" + status.Current.Version, "同一仓库的同一数字版本只自动通知一次。"},
				Action: &mailAction{Label: "查看版本详情", URL: r.URL},
			}, str(config.Sub(s.Config.Get(), "site")["url"]))
			n.Status = "sent"
			if err != nil {
				n.Status, n.Error = "failed", err.Error()
			}
			if err := s.Updates.FinishNotification(n); err != nil {
				return err
			}
		}
	}
	if p.AutoUpdate && status.CanApply && status.AutomaticReason == "" {
		return s.Updates.Start(r.Version)
	}
	return nil
}
