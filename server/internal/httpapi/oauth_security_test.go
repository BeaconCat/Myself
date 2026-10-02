package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"myself/server/internal/config"
)

func enableTestOAuth(t *testing.T, e *env, proxy string) {
	t.Helper()
	_, err := e.server.Config.Save(config.Map{
		"users":  config.Map{"enabled": true, "login": config.Map{"github": true}},
		"oauth":  config.Map{"github": config.Map{"clientId": "test-client", "clientSecret": "test-secret"}},
		"github": config.Map{"proxy": proxy},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOAuthLinkRejectsForcedPasswordChange(t *testing.T) {
	e := newEnv(t)
	enableTestOAuth(t, e, "")
	if _, err := e.server.DB.Exec(`UPDATE users SET must_change = 1 WHERE login = 'admin'`); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/github/start?mode=link", nil)
	r.Header.Set("Authorization", "Bearer "+e.token)
	w := httptest.NewRecorder()
	e.server.githubStart(w, r)
	if !strings.HasPrefix(w.Header().Get("Location"), "/account/login?error=") {
		t.Fatal("account awaiting mandatory password change can start GitHub binding")
	}
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == oauthCookie && cookie.Value != "" {
			t.Fatal("blocked account received a binding state cookie")
		}
	}
}

func TestOAuthLinkRejectsRevokedFlowBeforeNetwork(t *testing.T) {
	for _, mutation := range []string{
		`UPDATE users SET token_version = token_version + 1 WHERE login = 'admin'`,
		`UPDATE users SET status = 'disabled' WHERE login = 'admin'`,
		`UPDATE users SET must_change = 1 WHERE login = 'admin'`,
	} {
		t.Run(mutation, func(t *testing.T) {
			e := newEnv(t)
			var networkCalls atomic.Int32
			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				networkCalls.Add(1)
				w.WriteHeader(http.StatusBadGateway)
			}))
			defer proxy.Close()
			enableTestOAuth(t, e, proxy.URL)
			u, err := e.server.DB.UserByLogin("admin")
			if err != nil || u == nil {
				t.Fatal("missing fixture owner")
			}
			const state = "test-binding-state"
			e.server.oauth.put(state, oauthState{mode: "link", userID: u.ID, exp: time.Now().Add(time.Minute)})
			if _, err := e.server.DB.Exec(mutation); err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/github/callback?state="+state+"&code=test-code", nil)
			r.AddCookie(&http.Cookie{Name: oauthCookie, Value: state})
			w := httptest.NewRecorder()
			e.server.githubCallback(w, r)
			if networkCalls.Load() != 0 || !strings.Contains(w.Header().Get("Location"), "error=login_required") {
				t.Fatalf("revoked binding was allowed to exchange a code: calls=%d redirect=%s", networkCalls.Load(), w.Header().Get("Location"))
			}
		})
	}
}

func TestOAuthClientAlwaysVerifiesTLS(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("HTTP_PROXY", "")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client := githubOAuthClient(config.GitHub{InsecureTLS: true})
	defer client.CloseIdleConnections()
	response, err := client.Get(server.URL)
	if err == nil {
		response.Body.Close()
		t.Fatal("OAuth accepted an untrusted certificate")
	}
	if !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("TLS request failed for an unrelated reason: %v", err)
	}
}

func TestOAuthBindingStateTracksCurrentAccount(t *testing.T) {
	e := newEnv(t)
	enableTestOAuth(t, e, "")
	u, err := e.server.DB.UserByLogin("admin")
	if err != nil || u == nil {
		t.Fatal("missing fixture owner")
	}
	if err := e.auth.Revoke(u.ID); err != nil {
		t.Fatal(err)
	}
	u, _ = e.server.DB.UserByID(u.ID)
	token, err := e.auth.Issue(u)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/github/start?mode=link", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	e.server.githubStart(w, r)
	var key string
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == oauthCookie {
			key = cookie.Value
		}
	}
	st, ok := e.server.oauth.take(key)
	if !ok || st.tokenVersion != u.TokenVersion || !e.server.oauthLinkAllowed(st) {
		t.Fatal("fresh authorized binding was not retained")
	}
	if err := e.auth.Revoke(u.ID); err != nil {
		t.Fatal(err)
	}
	if e.server.oauthLinkAllowed(st) {
		t.Fatal("binding remained valid after session revocation")
	}
}
