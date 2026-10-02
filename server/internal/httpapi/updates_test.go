package httpapi

import (
	"net/http"
	"os"
	"testing"

	"myself/server/internal/store"
	"myself/server/internal/updater"
)

func TestSystemUpdatePermissionsAndMaintenance(t *testing.T) {
	e := newEnv(t)
	manager, err := updater.New(updater.Options{Root: e.root, Build: updater.Build{Version: "v1.0.0"}, Disabled: true})
	if err != nil {
		t.Fatal(err)
	}
	e.server.Updates = manager
	var status map[string]any
	e.call(http.MethodGet, "/api/v1/admin/system", nil, &status, http.StatusOK)
	if status["databaseDriver"] != e.server.DB.Driver() {
		t.Fatal(status)
	}
	for _, endpoint := range []string{"/api/v1/admin/system", "/api/v1/admin/system/check-update", "/api/v1/admin/system/update", "/api/v1/admin/system/history", "/api/v1/admin/system/history/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/restore"} {
		method := http.MethodPost
		if endpoint == "/api/v1/admin/system" || endpoint == "/api/v1/admin/system/history" {
			method = http.MethodGet
		}
		response := e.do(method, endpoint, nil, map[string]string{})
		response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("anonymous %s: %d", endpoint, response.StatusCode)
		}
	}
	e.enableUsers(map[string]any{"enabled": true, "readers": map[string]any{"enabled": true, "signup": "open"}})
	reader := e.signup(map[string]any{"email": "reader@example.com", "name": "Reader", "password": "reader-password-1"})
	response := e.do(http.MethodGet, "/api/v1/admin/system", nil, map[string]string{"Authorization": "Bearer " + reader})
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("reader system access: %d", response.StatusCode)
	}
	for _, endpoint := range []string{"/api/v1/admin/system/history", "/api/v1/admin/system/history/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/restore"} {
		method := http.MethodGet
		if endpoint != "/api/v1/admin/system/history" {
			method = http.MethodPost
		}
		response := e.do(method, endpoint, nil, map[string]string{"Authorization": "Bearer " + reader})
		response.Body.Close()
		if response.StatusCode != http.StatusForbidden {
			t.Fatalf("reader history access: %d", response.StatusCode)
		}
	}
	backup, err := e.server.PrepareUpdate()
	if err != nil {
		t.Fatal(err)
	}
	defer e.server.ResumeAfterUpdate()
	if _, err := os.Stat(backup); err != nil {
		t.Fatal(err)
	}
	e.call(http.MethodGet, "/api/v1/posts", nil, nil, http.StatusServiceUnavailable)
	e.call(http.MethodGet, "/api/v1/system/health", nil, nil, http.StatusOK)
	e.call(http.MethodGet, "/api/v1/admin/system", nil, nil, http.StatusOK)
}

func TestInitializedSiteCannotReconfigureDatabase(t *testing.T) {
	e := newEnv(t)
	called := false
	e.server.ConfigureDatabase = func(store.DatabaseConfig, string) error { called = true; return nil }
	e.call(http.MethodPost, "/api/v1/setup/database", map[string]any{"code": "anything", "database": map[string]string{"driver": "sqlite"}}, nil, http.StatusForbidden)
	if called {
		t.Fatal("initialized site reconfigured its database")
	}
}
