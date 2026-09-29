package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/config"
	"myself/server/internal/store"
)

// GitHub 登录（OAuth App）。state 存在服务端（10 分钟），并以 HttpOnly Cookie 绑定发起登录的浏览器；
// 回调后：已绑定的账号直接登录；新账号在「开放注册」或持有邀请时创建；mode=link 为当前登录用户绑定 GitHub。

const oauthCookie = "myself_oauth"

type oauthState struct {
	mode   string // login | link
	invite string
	next   string
	userID int64 // link 模式：发起绑定的用户
	exp    time.Time
}

type oauthStates struct {
	mu sync.Mutex
	m  map[string]oauthState
}

func (o *oauthStates) put(key string, st oauthState) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.m == nil {
		o.m = map[string]oauthState{}
	}
	now := time.Now()
	for k, v := range o.m {
		if now.After(v.exp) {
			delete(o.m, k)
		}
	}
	o.m[key] = st
}

func (o *oauthStates) take(key string) (oauthState, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	st, ok := o.m[key]
	delete(o.m, key)
	return st, ok && time.Now().Before(st.exp)
}

// safeNext 只允许站内相对路径作为登录后的落点（防开放重定向）。
func safeNext(p string) string {
	if strings.HasPrefix(p, "/") && !strings.HasPrefix(p, "//") && !strings.HasPrefix(p, "/\\") {
		return p
	}
	return "/"
}

func (s *Server) oauthRedirectURI(r *http.Request) string {
	return s.siteBase(r) + "/api/v1/auth/github/callback"
}

// failOAuth 回到前台账号页并带上错误码。
func failOAuth(w http.ResponseWriter, r *http.Request, code string) {
	http.Redirect(w, r, "/account/login?error="+url.QueryEscape(code), http.StatusFound)
}

// GET /auth/github/start?mode=login|link&invite=&next=
func (s *Server) githubStart(w http.ResponseWriter, r *http.Request) {
	cfg := s.Config.Typed()
	if !cfg.GitHubLoginReady() {
		failOAuth(w, r, "github_off")
		return
	}
	q := r.URL.Query()
	st := oauthState{mode: "login", invite: q.Get("invite"), next: safeNext(q.Get("next")), exp: time.Now().Add(10 * time.Minute)}
	if q.Get("mode") == "link" {
		u, _ := s.currentUser(r)
		if u == nil {
			failOAuth(w, r, "login_required")
			return
		}
		st.mode, st.userID = "link", u.ID
	}
	key, err := auth.RandomHex(20)
	if err != nil {
		fail(w, err)
		return
	}
	s.oauth.put(key, st)
	http.SetCookie(w, &http.Cookie{
		Name: oauthCookie, Value: key, Path: "/api/v1/auth/github", MaxAge: 600,
		HttpOnly: true, Secure: secureRequest(r), SameSite: http.SameSiteLaxMode, // 回调是 GitHub 发起的顶层跳转，需 Lax
	})
	v := url.Values{
		"client_id":    {cfg.OAuth.GitHub.ClientID},
		"redirect_uri": {s.oauthRedirectURI(r)},
		"scope":        {"read:user user:email"},
		"state":        {key},
		"allow_signup": {"true"},
	}
	http.Redirect(w, r, "https://github.com/login/oauth/authorize?"+v.Encode(), http.StatusFound)
}

type ghProfile struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// GET /auth/github/callback?code=&state=
func (s *Server) githubCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	c, err := r.Cookie(oauthCookie)
	http.SetCookie(w, &http.Cookie{Name: oauthCookie, Value: "", Path: "/api/v1/auth/github", MaxAge: -1, HttpOnly: true})
	if err != nil || c.Value == "" || c.Value != q.Get("state") {
		failOAuth(w, r, "state_mismatch")
		return
	}
	st, ok := s.oauth.take(c.Value)
	if !ok {
		failOAuth(w, r, "state_expired")
		return
	}
	if q.Get("error") != "" || q.Get("code") == "" {
		failOAuth(w, r, "github_denied")
		return
	}
	cfg := s.Config.Typed()
	if !cfg.GitHubLoginReady() {
		failOAuth(w, r, "github_off")
		return
	}
	client := githubClient(cfg.GitHub)
	profile, email, err := fetchGitHubIdentity(client, cfg, q.Get("code"), s.oauthRedirectURI(r))
	if err != nil {
		s.logSync(false, "GitHub 登录失败："+err.Error())
		failOAuth(w, r, "github_failed")
		return
	}

	// 绑定到当前用户
	if st.mode == "link" {
		if _, err := s.DB.Exec(`UPDATE users SET github_id = ? WHERE id = ?`, profile.ID, st.userID); err != nil {
			if store.IsUniqueErr(err) {
				failOAuth(w, r, "github_taken")
				return
			}
			fail(w, err)
			return
		}
		http.Redirect(w, r, st.next, http.StatusFound)
		return
	}

	u, err := s.DB.UserBy(`github_id = ?`, profile.ID)
	if err != nil {
		fail(w, err)
		return
	}
	if u == nil {
		u, err = s.createGitHubUser(cfg, st, profile, email)
		if err != nil {
			var code oauthErr
			if errors.As(err, &code) {
				failOAuth(w, r, string(code))
				return
			}
			fail(w, err)
			return
		}
	}
	switch {
	case u.Status == store.StatusDisabled:
		failOAuth(w, r, "account_disabled")
		return
	case !roleOn(cfg, u.Role):
		failOAuth(w, r, "role_closed")
		return
	}
	if u.Status == store.StatusPending { // GitHub 已验证过邮箱：直接激活
		_, _ = s.DB.Exec(`UPDATE users SET status = 'active' WHERE id = ?`, u.ID)
	}
	if err := s.startSession(w, r, u); err != nil {
		fail(w, err)
		return
	}
	s.DB.TouchUser(u.ID)
	http.Redirect(w, r, st.next, http.StatusFound)
}

type oauthErr string

func (e oauthErr) Error() string { return string(e) }

// createGitHubUser 首次用 GitHub 登录：需开放注册或持有有效邀请。
func (s *Server) createGitHubUser(cfg config.Typed, st oauthState, p ghProfile, email string) (*store.User, error) {
	role := store.RoleReader
	var invite *auth.Token
	if st.invite != "" {
		t, err := s.Auth.PeekToken("invite", st.invite)
		if err != nil || !roleOn(cfg, t.Role) {
			return nil, oauthErr("invalid_invite")
		}
		invite, role = t, t.Role
	} else if !cfg.Users.ReadersOn() || cfg.Users.Readers.Signup != "open" {
		return nil, oauthErr("signup_closed")
	}
	if email != "" {
		if taken, _ := s.DB.UserBy(`email = ?`, email); taken != nil {
			email = "" // 邮箱已被其它账号使用：不自动合并，GitHub 账号单独成户
		}
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		name = p.Login
	}
	login := "gh_" + p.Login
	if taken, _ := s.DB.UserByLogin(login); taken != nil {
		login += "_" + strconv.FormatInt(p.ID, 10)
	}
	avatar := ""
	if strings.HasPrefix(p.AvatarURL, "https://") {
		avatar = p.AvatarURL
	}
	id, err := s.DB.CreateUser(store.NewUser{
		Login: login, Email: email, Name: limitRunes(name, 40), Avatar: avatar, Role: role,
		Status: store.StatusActive, GitHubID: p.ID, EmailVerified: email != "",
	})
	if err != nil {
		return nil, err
	}
	if invite != nil {
		if err := s.Auth.ConsumeToken(invite, id); err != nil {
			_, _ = s.DB.Exec(`DELETE FROM users WHERE id = ?`, id)
			return nil, oauthErr("invalid_invite")
		}
	}
	return s.DB.UserByID(id)
}

// fetchGitHubIdentity 用授权码换令牌，再取用户资料与已验证的主邮箱。
func fetchGitHubIdentity(client *http.Client, cfg config.Typed, code, redirectURI string) (ghProfile, string, error) {
	var p ghProfile
	form := url.Values{
		"client_id": {cfg.OAuth.GitHub.ClientID}, "client_secret": {cfg.OAuth.GitHub.ClientSecret},
		"code": {code}, "redirect_uri": {redirectURI},
	}
	req, _ := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	var tok struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := doJSON(client, req, &tok); err != nil {
		return p, "", err
	}
	if tok.AccessToken == "" {
		return p, "", fmt.Errorf("token exchange: %s", tok.Error)
	}
	get := func(path string, out any) error {
		req, _ := http.NewRequest(http.MethodGet, "https://api.github.com"+path, nil)
		req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
		req.Header.Set("Accept", "application/vnd.github+json")
		return doJSON(client, req, out)
	}
	if err := get("/user", &p); err != nil {
		return p, "", err
	}
	if p.ID == 0 {
		return p, "", errors.New("empty github profile")
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	email := ""
	if err := get("/user/emails", &emails); err == nil {
		for _, e := range emails {
			if e.Primary && e.Verified {
				email = strings.ToLower(e.Email)
			}
		}
	}
	return p, email, nil
}

func doJSON(client *http.Client, req *http.Request, out any) error {
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("%s %s: %d", req.Method, req.URL.Path, res.StatusCode)
	}
	return json.Unmarshal(data, out)
}
