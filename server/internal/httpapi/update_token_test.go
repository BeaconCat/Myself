package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"myself/server/internal/updater"
)

func TestUpdateTokenAPIWriteOnly(t *testing.T) {
	t.Setenv("MYSELF_UPDATE_TOKEN", "")
	e := newEnv(t)
	m, err := updater.New(updater.Options{Root: e.root, Build: updater.Build{Version: "dev"}})
	if err != nil {
		t.Fatal(err)
	}
	e.server.Updates = m
	secret := "fake-update-secret-for-api-test"
	body := map[string]any{"repository": "BeaconCat/Myself", "token": secret}
	var response map[string]any
	e.call(http.MethodPut, "/api/v1/admin/system/preferences", body, &response, http.StatusOK)
	check := func(response map[string]any) {
		t.Helper()
		encoded, _ := json.Marshal(response)
		if strings.Contains(string(encoded), secret) {
			t.Fatal("API returned token content")
		}
	}
	check(response)
	if response["tokenConfigured"] != true || response["tokenSource"] != "settings" {
		t.Fatal("Missing saved token status")
	}
	e.call(http.MethodGet, "/api/v1/admin/system", nil, &response, http.StatusOK)
	check(response)
	e.call(http.MethodGet, "/api/v1/site-config", nil, &response, http.StatusOK)
	check(response)
	e.call(http.MethodGet, "/api/v1/admin/settings", nil, &response, http.StatusOK)
	check(response)
	e.call(http.MethodPut, "/api/v1/admin/system/preferences", map[string]any{"repository": "BeaconCat/Myself"}, nil, http.StatusOK)
	if !m.Status().TokenConfigured {
		t.Fatal("Omitted token cleared secret")
	}
	body["token"] = ""
	e.call(http.MethodPut, "/api/v1/admin/system/preferences", body, &response, http.StatusOK)
	if m.Status().TokenConfigured {
		t.Fatal("Explicit clear ignored")
	}
	e.token = ""
	body["token"] = secret
	e.call(http.MethodPut, "/api/v1/admin/system/preferences", body, nil, http.StatusUnauthorized)
	if m.Status().TokenConfigured {
		t.Fatal("Unauthenticated request changed credential")
	}
}
