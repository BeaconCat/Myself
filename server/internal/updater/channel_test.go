package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

func channelServer(t *testing.T) (*httptest.Server, *atomic.Bool) {
	t.Helper()
	changed := &atomic.Bool{}
	versions := map[int64]string{42: "v3.0.0", 43: "v2.0.0", 44: "v9.0.0", 45: "v2.0.0", 46: "v1.5.0"}
	remote := func(id int64) remoteRelease {
		return remoteRelease{ID: id, Tag: versions[id], Name: "Free release title", Draft: id == 44, Prerelease: id == 42 || id == 45,
			Assets: []remoteAsset{{ID: id * 10, Name: "release-manifest.json", Size: 100}, {ID: id*10 + 1, Name: "binary", Size: 3}}}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/BeaconCat/Myself/releases" {
			if r.URL.Query().Get("page") == "1" {
				w.Header().Set("Link", `<https://api.github.com/next>; rel="next"`)
				json.NewEncoder(w).Encode([]remoteRelease{remote(42), remote(43), remote(44), remote(45)})
			} else {
				json.NewEncoder(w).Encode([]remoteRelease{remote(46)})
			}
			return
		}
		var asset int64
		if _, err := fmt.Sscanf(r.URL.Path, "/repos/BeaconCat/Myself/releases/assets/%d", &asset); err == nil {
			if asset%10 == 1 {
				w.Write([]byte("bad"))
				return
			}
			json.NewEncoder(w).Encode(manifest{SchemaVersion: 1, UpdateProtocol: Protocol, Version: versions[asset/10], Commit: strings.Repeat("a", 40), Artifacts: []Artifact{{Name: "binary", OS: runtime.GOOS, Arch: runtime.GOARCH, Kind: "binary", Size: 3, SHA256: strings.Repeat("0", 64)}}})
			return
		}
		var id int64
		if _, err := fmt.Sscanf(r.URL.Path, "/repos/BeaconCat/Myself/releases/%d", &id); err == nil {
			item := remote(id)
			if id == 43 && changed.Load() {
				item.Prerelease = true
			}
			json.NewEncoder(w).Encode(item)
			return
		}
		http.NotFound(w, r)
	}))
	return server, changed
}

func TestChannelFilteringPaginationAndLatest(t *testing.T) {
	m := testManager(t)
	server, _ := channelServer(t)
	defer server.Close()
	m.apiBase = server.URL
	if m.Preferences().Channel != "stable" {
		t.Fatal("new installations must default to stable")
	}
	first, err := m.List(context.Background(), 1, 1)
	if err != nil || first.Channel != "stable" || len(first.Items) != 1 || first.Items[0].ID != 43 || !first.HasNext {
		t.Fatal(first, err)
	}
	second, err := m.List(context.Background(), 2, 1)
	if err != nil || len(second.Items) != 1 || second.Items[0].ID != 46 || second.HasNext {
		t.Fatal(second, err)
	}
	status, err := m.Check(context.Background())
	if err != nil || status.Available == nil || status.Available.ID != 43 || !status.CanApply {
		t.Fatal(status, err)
	}
	if err := m.StartSelected(context.Background(), m.opts.Repository, 42); err == nil || err.Error() != "release_not_in_channel" {
		t.Fatal("stable accepted preview", err)
	}
	p := m.Preferences()
	p.Channel = "preview"
	if err := m.Configure(p, "2026-10-02"); err != nil {
		t.Fatal(err)
	}
	if s := m.Status(); s.Available != nil || s.Phase != "idle" || s.CanApply {
		t.Fatal("stale selection survived channel change", s)
	}
	page, err := m.List(context.Background(), 1, 2)
	if err != nil || page.Channel != "preview" || len(page.Items) != 2 || page.Items[0].ID != 42 || page.Items[1].ID != 43 || !page.HasNext {
		t.Fatal(page, err)
	}
	status, err = m.Check(context.Background())
	if err != nil || status.Available == nil || status.Available.ID != 42 || !status.CanApply {
		t.Fatal(status, err)
	}
	// Automatic installation must use the selected channel too, reaching the hash gate.
	if err := m.Start(status.Available.Version); err != nil {
		t.Fatal(err)
	}
	waitFailed(t, m, "checksum_mismatch")
}

func TestAutomaticInstallRechecksReleaseChannel(t *testing.T) {
	m := testManager(t)
	server, changed := channelServer(t)
	defer server.Close()
	m.apiBase = server.URL
	s, err := m.Check(context.Background())
	if err != nil || s.Available == nil || s.Available.ID != 43 {
		t.Fatal(s, err)
	}
	changed.Store(true) // The publisher reclassifies it after our check.
	if err := m.Start(s.Available.Version); err != nil {
		t.Fatal(err)
	}
	waitFailed(t, m, "release_not_in_channel")
}

func TestChannelPreferenceMigrationAndValidation(t *testing.T) {
	m := testManager(t)
	p := m.Preferences()
	p.Channel = "preview"
	if err := m.Configure(p, "2026-10-02"); err != nil {
		t.Fatal(err)
	}
	restarted, err := New(m.opts)
	if err != nil || restarted.Preferences().Channel != "preview" {
		t.Fatal(err, restarted.Preferences())
	}
	p.Channel = "" // Older clients must preserve the explicit choice.
	if err := m.Configure(p, ""); err != nil || m.Preferences().Channel != "preview" {
		t.Fatal(err, m.Preferences())
	}
	p.Channel = "nightly"
	if err := m.Configure(p, ""); err == nil || err.Error() != "invalid_update_channel" {
		t.Fatal(err)
	}
	if m.Preferences().Channel != "preview" {
		t.Fatal("invalid request changed channel")
	}
	if err := writeJSONFile(filepath.Join(m.stateDir, "preferences.json"), map[string]any{"repository": m.opts.Repository, "autoUpdate": true}); err != nil {
		t.Fatal(err)
	}
	legacy, err := New(m.opts)
	if err != nil || legacy.Preferences().Channel != "preview" || !legacy.Preferences().AutoUpdate {
		t.Fatal("legacy update scope changed", err)
	}
}

func TestPreviewChannelPrefersStableAtEqualVersion(t *testing.T) {
	m := testManager(t)
	p := m.Preferences()
	p.Channel = "preview"
	if err := m.Configure(p, ""); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]remoteRelease{{ID: 1, Tag: "v2.0.0", Prerelease: true}, {ID: 2, Tag: "v2.0.0", Prerelease: false}})
	}))
	defer server.Close()
	m.apiBase = server.URL
	s, err := m.Check(context.Background())
	if err != nil || s.Available == nil || s.Available.ID != 2 || s.Available.Prerelease {
		t.Fatal(s, err)
	}
}

func TestChannelHidesLegacyCandidateDuringHandoff(t *testing.T) {
	m := testManager(t)
	m.pending = true
	m.state.Phase = "installed"
	m.state.Available = &Release{Version: "v9.0.0", Prerelease: true, Installable: true}
	if s := m.Status(); s.Available != nil || s.CanApply || s.Phase != "installed" {
		t.Fatal(s)
	}
}
