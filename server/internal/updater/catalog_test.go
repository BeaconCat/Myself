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
	"time"
)

func TestNumericVersionsWithArbitraryComponents(t *testing.T) {
	for _, tc := range []struct {
		a, b  string
		order int
	}{
		{"v0.1.0.9", "v0.1.0.8", 1}, {"0.1.0.10", "0.1.0.9", 1},
		{"v0.1", "0.1.0.0", 0}, {"0.00.001", "0.0.1", 0},
		{"v0.0.2-beta", "v0.0.1-alpha", 1}, {"1.0.0", "1.0.0-rc.2", 0},
		{"0.1.0.9", "0.2", -1},
	} {
		if order, ok := CompareVersions(tc.a, tc.b); !ok || order != tc.order {
			t.Errorf("%s vs %s: %d %v", tc.a, tc.b, order, ok)
		}
	}
	for _, invalid := range []string{"Myself Astra v0.1.0.9", "dev", "1..2", "-1.2", "1.2/evil"} {
		if _, ok := NumericVersion(invalid); ok {
			t.Errorf("accepted %q", invalid)
		}
	}
}

func TestCataloguePaginationNamesAndGreatestVersion(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/BeaconCat/Myself/releases":
			id, tag, name := int64(2), "random-tag", "Myself Beta with any title"
			if r.URL.Query().Get("page") == "1" {
				id, tag, name = 1, "unrelated-tag", "Myself Astra"
				w.Header().Set("Link", `<https://api.github.com/next>; rel="next"`)
			}
			json.NewEncoder(w).Encode([]remoteRelease{{ID: id, Tag: tag, Name: name, Prerelease: true, Assets: []remoteAsset{{ID: id, Name: "release-manifest.json", Size: 100}, {ID: id + 10, Name: "binary", Size: 3}}}})
		case strings.HasPrefix(r.URL.Path, "/repos/BeaconCat/Myself/releases/assets/"):
			calls++
			version := "0.0.2"
			if strings.HasSuffix(r.URL.Path, "/1") {
				version = "0.1.0.9"
			}
			json.NewEncoder(w).Encode(manifest{SchemaVersion: 1, UpdateProtocol: 1, Version: version, Codename: "Astra", Commit: strings.Repeat("a", 40), Artifacts: []Artifact{{Name: "binary", Kind: "binary", OS: runtime.GOOS, Arch: runtime.GOARCH, Size: 3, SHA256: strings.Repeat("a", 64)}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	m := testManager(t)
	m.opts.Build.Version = "0.0.1"
	p := m.Preferences()
	p.Channel = "preview"
	if err := m.Configure(p, ""); err != nil {
		t.Fatal(err)
	}
	m.apiBase = server.URL
	page, err := m.List(context.Background(), 1, 1)
	if err != nil || !page.HasNext || len(page.Items) != 1 || page.Items[0].Name != "Myself Astra" || page.Items[0].Version != "0.1.0.9" {
		t.Fatalf("page: %+v %v", page, err)
	}
	state, err := m.Check(context.Background())
	if err != nil || state.Available == nil || state.Available.Version != "0.1.0.9" || !state.CanApply {
		t.Fatalf("latest: %+v %v", state, err)
	}
	if calls != 2 {
		t.Fatalf("manifest cache not reused: %d", calls)
	}
}

func TestEmptyRepositoryIsNotAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("[]")) }))
	defer server.Close()
	m := testManager(t)
	m.apiBase = server.URL
	status, err := m.Check(context.Background())
	if err != nil || status.Phase != "no_releases" || status.Error != "" {
		t.Fatalf("%+v %v", status, err)
	}
	page, err := m.List(context.Background(), 1, 10)
	if err != nil || len(page.Items) != 0 || page.HasNext {
		t.Fatalf("%+v %v", page, err)
	}
}

func TestPreferencesAndNotificationsSurviveRestart(t *testing.T) {
	m := testManager(t)
	p := Preferences{Repository: "https://github.com/fork/Myself.git", AutoUpdate: true, Subscribe: true, Email: "owner@example.com"}
	if err := m.Configure(p, "2026-10-01"); err != nil {
		t.Fatal(err)
	}
	if m.Preferences().AutoUpdate {
		t.Fatal("changing source did not pause automation")
	}
	if m.Preferences().Repository != "fork/Myself" {
		t.Fatal(m.Preferences())
	}
	n := Notification{Repository: "fork/Myself", Version: "v0.1.0.0", Status: "attempting"}
	if ok, err := m.ReserveNotification(n); !ok || err != nil {
		t.Fatalf("%v %v", ok, err)
	}
	restarted, err := New(m.opts)
	if err != nil {
		t.Fatal(err)
	}
	if restarted.Preferences().Email != p.Email || !restarted.Preferences().Subscribe {
		t.Fatal("preferences lost")
	}
	n.Repository, n.Version = "Fork/myself", "0.1"
	if ok, err := restarted.ReserveNotification(n); ok || err != nil {
		t.Fatalf("duplicate after restart: %v %v", ok, err)
	}
	if claimed, _ := restarted.ClaimDay("2026-10-01"); claimed {
		t.Fatal("repeated day")
	}
	if claimed, _ := restarted.ClaimDay("2026-10-02"); !claimed {
		t.Fatal("missed next day")
	}
	if _, err := NormalizeRepository("https://evil.example/owner/repo"); err == nil {
		t.Fatal("untrusted source accepted")
	}
}

func TestExplicitDowngradePausesAutomaticUpdate(t *testing.T) {
	server := releaseServer(t, strings.Repeat("0", 64), 1, []byte("bad"))
	defer server.Close()
	m := testManager(t)
	m.opts.Build.Version = "v2.0.0"
	m.apiBase = server.URL
	p := m.Preferences()
	p.AutoUpdate = true
	if err := m.Configure(p, "2026-10-01"); err != nil {
		t.Fatal(err)
	}
	if err := m.StartSelected(context.Background(), p.Repository, 42); err != nil {
		t.Fatal(err)
	}
	if m.Preferences().AutoUpdate {
		t.Fatal("downgrade did not pause automatic upgrades")
	}
	deadline := time.Now().Add(5 * time.Second)
	for m.Status().Phase != "failed" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if m.Status().Error != "checksum_mismatch" {
		t.Fatal(m.Status())
	}
	data, err := os.ReadFile(m.executable)
	if err != nil || string(data) != "existing program" {
		t.Fatal("failed downgrade changed executable")
	}
	if _, err := os.Stat(filepath.Join(m.stateDir, "preferences.json")); err != nil {
		t.Fatal(err)
	}
}
