package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	"myself/server/internal/config"
	"myself/server/internal/updater"
)

type updateRoundTrip func(*http.Request) (*http.Response, error)

func (f updateRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNightlyAutomaticUpdateStartsOnlyOnce(t *testing.T) {
	for _, tc := range []struct {
		name, channel       string
		prerelease, install bool
	}{
		{"stable release", "stable", false, true}, {"stable excludes preview", "stable", true, false},
		{"preview release", "preview", true, true}, {"preview includes stable", "preview", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(t)
			transport := updateRoundTrip(func(r *http.Request) (*http.Response, error) {
				var value any
				entry := map[string]any{"id": 42, "tag_name": "tag-with-any-name", "name": "Myself Astra v0.1.0.9", "prerelease": tc.prerelease, "assets": []map[string]any{{"id": 1, "name": "release-manifest.json", "size": 100}, {"id": 2, "name": "binary", "size": 3}}}
				switch r.URL.Path {
				case "/repos/BeaconCat/Myself/releases":
					value = []any{entry}
				case "/repos/BeaconCat/Myself/releases/42":
					value = entry
				case "/repos/BeaconCat/Myself/releases/assets/1":
					value = map[string]any{"schemaVersion": 1, "updateProtocol": 1, "version": "v0.1.0.9", "commit": strings.Repeat("a", 40), "artifacts": []map[string]any{{"name": "binary", "os": runtime.GOOS, "arch": runtime.GOARCH, "kind": "binary", "size": 3, "sha256": strings.Repeat("0", 64)}}}
				case "/repos/BeaconCat/Myself/releases/assets/2":
					return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("bad"))}, nil
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				body, _ := json.Marshal(value)
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
			})
			m, err := updater.New(updater.Options{Root: e.root, Build: updater.Build{Version: "v0.0.1", OS: runtime.GOOS, Arch: runtime.GOARCH}, Transport: transport,
				Prepare: func() (string, error) { t.Error("unverified program reached backup"); return "", nil }, Ready: func(string) error { t.Error("unverified program reached installer"); return nil }})
			if err != nil {
				t.Fatal(err)
			}
			e.server.Updates = m
			if err := m.Configure(updater.Preferences{Repository: "BeaconCat/Myself", Channel: tc.channel, AutoUpdate: true}, "2026-09-01"); err != nil {
				t.Fatal(err)
			}
			now := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
			if err := e.server.runUpdateCycle(context.Background(), now, m.Check); err != nil {
				t.Fatal(err)
			}
			if tc.install {
				deadline := time.Now().Add(5 * time.Second)
				for m.Status().Phase != "failed" && time.Now().Before(deadline) {
					time.Sleep(10 * time.Millisecond)
				}
				if status := m.Status(); status.Target == nil || status.Target.Version != "v0.1.0.9" || status.Error != "checksum_mismatch" {
					t.Fatal(status)
				}
			} else if status := m.Status(); status.Target != nil || status.Available != nil || status.Phase != "no_releases" {
				t.Fatal("stable channel attempted preview update", status)
			}
			if err := e.server.runUpdateCycle(context.Background(), now.Add(time.Minute), func(context.Context) (updater.Status, error) {
				t.Fatal("scanned twice in one day")
				return updater.Status{}, nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUpdateScheduleSMTPGateAndOnceOnlyMail(t *testing.T) {
	e := newEnv(t)
	options := updater.Options{Root: e.root, Build: updater.Build{Version: "v0.0.1"}, Disabled: true}
	m, err := updater.New(options)
	if err != nil {
		t.Fatal(err)
	}
	e.server.Updates = m
	p := updater.Preferences{Repository: "BeaconCat/Myself", Subscribe: true, Email: "owner@example.com"}
	e.call(http.MethodPut, "/api/v1/admin/system/preferences", p, nil, http.StatusBadRequest)
	if _, err := e.server.Config.Save(config.Map{"timezone": "Asia/Hong_Kong", "mail": config.Map{"enabled": true, "host": "smtp.example.com", "port": 25, "from": "Myself <site@example.com>", "security": "none"}}); err != nil {
		t.Fatal(err)
	}
	if err := m.Configure(p, "2026-10-01"); err != nil {
		t.Fatal(err)
	}
	checks, sent := 0, 0
	e.server.mailer = func(_ config.Mail, to, subject, text, html string) error {
		sent++
		if to != p.Email {
			t.Error(to)
		}
		return nil
	}
	check := func(context.Context) (updater.Status, error) {
		checks++
		return updater.Status{Current: options.Build, Available: &updater.Release{Version: "v0.1.0.9", Name: "Myself Astra", URL: "https://github.com/BeaconCat/Myself/releases/tag/anything"}}, nil
	}
	before := time.Date(2026, 10, 1, 15, 59, 0, 0, time.UTC)
	if err := e.server.runUpdateCycle(context.Background(), before, check); err != nil {
		t.Fatal(err)
	}
	if checks != 0 {
		t.Fatal("checked before local midnight")
	}
	if err := e.server.runUpdateCycle(context.Background(), before.Add(time.Minute), check); err != nil {
		t.Fatal(err)
	}
	if checks != 1 || sent != 1 {
		t.Fatalf("checks=%d mail=%d", checks, sent)
	}
	restarted, err := updater.New(options)
	if err != nil {
		t.Fatal(err)
	}
	e.server.Updates = restarted
	if err := e.server.runUpdateCycle(context.Background(), before.Add(2*time.Minute), check); err != nil {
		t.Fatal(err)
	}
	if err := e.server.runUpdateCycle(context.Background(), before.Add(25*time.Hour), check); err != nil {
		t.Fatal(err)
	}
	if checks != 2 || sent != 1 {
		t.Fatalf("duplicate after restart/next day: checks=%d mail=%d", checks, sent)
	}
	if restarted.LastNotification().Status != "sent" {
		t.Fatal(restarted.LastNotification())
	}
}

func TestFailedNotificationIsNotRetried(t *testing.T) {
	e := newEnv(t)
	m, err := updater.New(updater.Options{Root: e.root, Disabled: true})
	if err != nil {
		t.Fatal(err)
	}
	e.server.Updates = m
	_, err = e.server.Config.Save(config.Map{"mail": config.Map{"enabled": true, "host": "smtp.example.com", "port": 25, "from": "site@example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Configure(updater.Preferences{Repository: "BeaconCat/Myself", Subscribe: true, Email: "owner@example.com"}, "2026-09-01"); err != nil {
		t.Fatal(err)
	}
	calls := 0
	e.server.mailer = func(config.Mail, string, string, string, string) error {
		calls++
		return errors.New("ambiguous SMTP acknowledgement")
	}
	check := func(context.Context) (updater.Status, error) {
		return updater.Status{Available: &updater.Release{Version: "v1.2.3", Name: "Free title"}}, nil
	}
	for _, day := range []int{2, 3} {
		if err := e.server.runUpdateCycle(context.Background(), time.Date(2026, 9, day, 0, 0, 0, 0, time.UTC), check); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 || m.LastNotification().Status != "failed" {
		t.Fatalf("attempts=%d notification=%+v", calls, m.LastNotification())
	}
}

func TestMidnightScheduleHonorsDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 3, 8, 0, 0, 0, 0, loc)
	if next := nextUpdateCheck(start, loc); next.Sub(start) != 23*time.Hour || next.Hour() != 0 {
		t.Fatal(next)
	}
}
