package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/config"
)

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9_.-]{3,32}$`)

// GET /setup 是否尚未初始化（前台据此进入初始化流程）
func (s *Server) setupStatus(w http.ResponseWriter, _ *http.Request) {
	need, err := s.Auth.NeedsSetup()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"needsSetup": need})
}

// POST /setup/verify 校验初始化码（失败计入 IP 锁定）
func (s *Server) setupVerify(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if d := s.limiter.blocked(ip); d > 0 {
		tooMany(w, d)
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	switch err := s.Auth.CheckSetupCode(b.strOr("code")); {
	case errors.Is(err, auth.ErrAlreadySetup):
		writeError(w, http.StatusConflict, "already_setup")
	case errors.Is(err, auth.ErrBadSetupCode):
		s.limiter.fail(ip)
		writeError(w, http.StatusForbidden, "bad_setup_code")
	default:
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// POST /setup 首次启动初始化：
// {code, username, password, site:{title, subtitle, url}, identity:{name, hello, tagline, bio, motto, avatar}, demo}
// 初始化码只打印在服务端启动日志里；失败计入与登录相同的 IP 锁定。成功后写入会话 Cookie（即已登录）。
func (s *Server) setup(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if d := s.limiter.blocked(ip); d > 0 {
		tooMany(w, d)
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	username := strings.TrimSpace(b.strOr("username"))
	password := b.strOr("password")
	if !usernameRe.MatchString(username) {
		writeError(w, http.StatusBadRequest, "invalid_username")
		return
	}
	if len([]rune(password)) < auth.MinPasswordLen || len(password) > 256 {
		writeError(w, http.StatusBadRequest, "weak_password")
		return
	}
	site, _ := b["site"].(map[string]any)
	siteURL := strings.TrimRight(strings.TrimSpace(str(site["url"])), "/")
	if siteURL != "" {
		if u, err := url.Parse(siteURL); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			writeError(w, http.StatusBadRequest, "invalid_site_url")
			return
		}
	}

	var (
		adminID int64
		err     error
	)
	id, _ := b["identity"].(map[string]any)
	withScryptSlot(func() {
		adminID, err = s.Auth.Setup(b.strOr("code"), username, password, limitRunes(strings.TrimSpace(str(id["name"])), 40))
	})
	switch {
	case errors.Is(err, auth.ErrAlreadySetup):
		writeError(w, http.StatusConflict, "already_setup")
		return
	case errors.Is(err, auth.ErrBadSetupCode):
		s.limiter.fail(ip)
		writeError(w, http.StatusForbidden, "bad_setup_code")
		return
	case err != nil:
		fail(w, err)
		return
	}
	s.limiter.success(ip)

	// 站点与身份：只取白名单字段，空值不覆盖默认
	patch := config.Map{}
	siteP := config.Map{}
	for _, k := range []string{"title", "subtitle"} {
		if v := strings.TrimSpace(str(site[k])); v != "" {
			siteP[k] = limitRunes(v, 80)
		}
	}
	if siteURL != "" {
		siteP["url"] = siteURL
	}
	if len(siteP) > 0 {
		patch["site"] = siteP
	}
	about := config.Map{}
	for _, k := range []string{"name", "alias", "hello", "tagline", "bio", "motto", "avatar"} {
		if v := strings.TrimSpace(str(id[k])); v != "" {
			about[k] = limitRunes(v, 2000)
		}
	}
	// 建站日期 = 初始化当天（站点时区），关于页「第 N 天」从这里起算
	about["foundedAt"] = time.Now().In(s.siteLocation()).Format("2006-01-02")
	patch["about"] = about
	if len(patch) > 0 {
		if _, err := s.Config.Save(patch); err != nil {
			fail(w, err)
			return
		}
	}
	if b.truthy("demo") {
		if err := s.DB.SeedDemo(); err != nil {
			fail(w, err)
			return
		}
	}
	admin, err := s.DB.UserByID(adminID)
	if err != nil || admin == nil {
		fail(w, err)
		return
	}
	if err := s.startSession(w, r, admin); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": publicUser(admin)})
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func limitRunes(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

// siteLocation 站点时区（配置 timezone），无效时回落本地时区。
func (s *Server) siteLocation() *time.Location {
	if name, _ := s.Config.Get()["timezone"].(string); name != "" {
		if loc, err := time.LoadLocation(name); err == nil {
			return loc
		}
	}
	return time.Local
}
