package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

func TestStableVersionSelection(t *testing.T) {
	for _, tc := range []struct {
		latest, current string
		newer           bool
	}{
		{"v1.2.0", "v1.1.9", true}, {"v1.0.0", "v1.0.0-rc.1", false},
		{"v1.0.0", "v1.0.0", false}, {"v1.0.0", "v2.0.0", false},
		{"v2.0.0-beta", "v1.0.0", true}, {"v1.0.0", "dev", false},
		{"v999999999999999999999999.0.0", "v1.0.0", true},
	} {
		if got := newer(tc.latest, tc.current); got != tc.newer {
			t.Errorf("%s > %s = %v", tc.latest, tc.current, got)
		}
	}
}

func releaseServer(t *testing.T, hash string, protocol int, binary []byte) *httptest.Server {
	t.Helper()
	name := "myself-v1.1.0-" + runtime.GOOS + "-" + runtime.GOARCH
	metadata := manifest{SchemaVersion: 1, UpdateProtocol: protocol, Version: "v1.1.0", Commit: strings.Repeat("a", 40),
		Artifacts: []Artifact{{Name: name, OS: runtime.GOOS, Arch: runtime.GOARCH, Kind: "binary", Size: int64(len(binary)), SHA256: hash}}}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/BeaconCat/Myself/releases", "/repos/BeaconCat/Myself/releases/42":
			release := map[string]any{"id": 42, "tag_name": "v1.1.0", "name": "New release", "assets": []map[string]any{
				{"id": 1, "name": "release-manifest.json", "size": 100}, {"id": 2, "name": name, "size": len(binary)},
			}}
			if strings.HasSuffix(r.URL.Path, "/42") {
				json.NewEncoder(w).Encode(release)
			} else {
				json.NewEncoder(w).Encode([]any{release})
			}
		case "/repos/BeaconCat/Myself/releases/assets/1":
			json.NewEncoder(w).Encode(metadata)
		case "/repos/BeaconCat/Myself/releases/assets/2":
			w.Write(binary)
		default:
			http.NotFound(w, r)
		}
	}))
}

func testManager(t *testing.T) *Manager {
	t.Helper()
	m, err := New(Options{Root: t.TempDir(), Build: Build{Version: "v1.0.0", OS: runtime.GOOS, Arch: runtime.GOARCH},
		Prepare: func() (string, error) { t.Error("unsafe binary reached backup"); return "", nil },
		Ready:   func(string) error { t.Error("unsafe binary reached installer"); return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	m.executable = filepath.Join(t.TempDir(), "myself")
	if err := os.WriteFile(m.executable, []byte("existing program"), 0700); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestReleaseCheckAndProtocolGate(t *testing.T) {
	binary := []byte("a binary")
	sum := sha256.Sum256(binary)
	for _, protocol := range []int{Protocol, 999} {
		server := releaseServer(t, hex.EncodeToString(sum[:]), protocol, binary)
		m := testManager(t)
		m.apiBase = server.URL
		state, err := m.Check(context.Background())
		if protocol == Protocol {
			if err != nil || state.Available == nil || !state.CanApply {
				t.Fatalf("valid release: %+v, %v", state, err)
			}
			if err := m.Start("v9.9.9"); err == nil {
				t.Fatal("stale/unselected version accepted")
			}
		} else if state.CanApply || state.Available == nil || state.Available.Installable {
			t.Fatal("unknown protocol accepted")
		}
		server.Close()
	}
}

func TestChecksumFailureLeavesProgramAndDataUntouched(t *testing.T) {
	server := releaseServer(t, strings.Repeat("0", 64), Protocol, []byte("tampered binary"))
	defer server.Close()
	m := testManager(t)
	m.apiBase = server.URL
	if _, err := m.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := m.Start("v1.1.0"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for m.Status().Phase != "failed" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	state := m.Status()
	if state.Phase != "failed" || state.Error != "checksum_mismatch" {
		t.Fatalf("unexpected outcome: %+v", state)
	}
	data, err := os.ReadFile(m.executable)
	if err != nil || string(data) != "existing program" {
		t.Fatal("old binary changed")
	}
	if _, err := os.Stat(m.executable + ".update-lock"); !os.IsNotExist(err) {
		t.Fatal("failed download kept update lock")
	}
}

func TestPendingGateStaysOpenAfterSuccessfulUpgrade(t *testing.T) {
	m := testManager(t)
	m.pending = true
	m.state.Phase = "restarting"
	if err := writeJSONFile(filepath.Join(m.stateDir, "status.json"), Status{Phase: "installed"}); err != nil {
		t.Fatal(err)
	}
	if m.Pending() {
		t.Fatal("successful upgrade remains blocked")
	}
	m.state.Phase = "up_to_date"
	if m.Pending() {
		t.Fatal("checking updates reactivated maintenance")
	}
}

func TestRedirectRejectsUntrustedDownloadHost(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://evil.example/binary", nil)
	request.Header.Set("Authorization", "Bearer secret")
	if err := newClient().CheckRedirect(request, nil); err == nil {
		t.Fatal("HTTP/untrusted redirect accepted")
	}
	if request.Header.Get("Authorization") != "" {
		t.Fatal("release token forwarded to redirected host")
	}
}
