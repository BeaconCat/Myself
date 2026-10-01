package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestConfiguredUpdateTokenLifecycle(t *testing.T) {
	t.Setenv("MYSELF_UPDATE_TOKEN", "environment-test-secret")
	m := testManager(t)
	if status := m.Status(); !status.TokenConfigured || status.TokenSource != "environment" {
		t.Fatal("Missing environment credential status")
	}
	token := "ui-test-secret"
	if err := m.ConfigureWithToken(m.Preferences(), "2026-10-01", &token); err != nil {
		t.Fatal(err)
	}
	if m.token != token || m.Status().TokenSource != "settings" {
		t.Fatal("UI token was not applied immediately")
	}
	encoded, _ := json.Marshal(m.Status())
	if strings.Contains(string(encoded), token) || strings.Contains(string(encoded), m.envToken) {
		t.Fatal("Status exposed a credential")
	}
	filename := filepath.Join(m.stateDir, "preferences.json")
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("Secret file permissions: %v", info.Mode().Perm())
	}
	// Other preferences and daily reservations must preserve the credential.
	p := m.Preferences()
	p.Subscribe = true
	p.Email = "owner@example.com"
	if err := m.Configure(p, "2026-10-01"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.ClaimDay("2026-10-02"); err != nil {
		t.Fatal(err)
	}
	restarted, err := New(m.opts)
	if err != nil {
		t.Fatal(err)
	}
	if restarted.token != token || restarted.Status().TokenSource != "settings" {
		t.Fatal("Restart lost saved token")
	}
	replacement := "replacement-test-secret"
	if err := restarted.ConfigureWithToken(restarted.Preferences(), "2026-10-02", &replacement); err != nil {
		t.Fatal(err)
	}
	var header string
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header = r.Header.Get("Authorization")
		w.Write([]byte("[]"))
	}))
	defer remote.Close()
	restarted.apiBase = remote.URL
	if _, err := restarted.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	if header != "Bearer "+replacement {
		t.Fatal("Release requests used stale credentials")
	}
	empty := ""
	if err := restarted.ConfigureWithToken(restarted.Preferences(), "2026-10-02", &empty); err != nil {
		t.Fatal(err)
	}
	if restarted.token != "environment-test-secret" || restarted.Status().TokenSource != "environment" {
		t.Fatal("Clearing did not restore environment fallback")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), replacement) || strings.Contains(string(data), token) || strings.Contains(string(data), "environment-test-secret") {
		t.Fatal("Cleared or environment credentials were persisted")
	}
	t.Setenv("MYSELF_UPDATE_TOKEN", "")
	withoutEnv, err := New(m.opts)
	if err != nil {
		t.Fatal(err)
	}
	if withoutEnv.Status().TokenConfigured || withoutEnv.token != "" {
		t.Fatal("Cleared token reappeared after restart")
	}
}

func TestUpdateTokenValidationAndBusyProtection(t *testing.T) {
	t.Setenv("MYSELF_UPDATE_TOKEN", "")
	m := testManager(t)
	for _, invalid := range []string{"bad token", "bad\r\nHeader:value", strings.Repeat("a", 4097), "非ASCII"} {
		if err := m.ConfigureWithToken(m.Preferences(), "2026-10-01", &invalid); err == nil || err.Error() != "invalid_update_token" {
			t.Fatal("Invalid token accepted")
		}
	}
	m.busy = true
	token := "new-test-secret"
	if err := m.ConfigureWithToken(m.Preferences(), "2026-10-01", &token); err == nil || err.Error() != "update_in_progress" {
		t.Fatal("Token changed during an update")
	}
	if m.token != "" {
		t.Fatal("Rejected change mutated credentials")
	}
}
