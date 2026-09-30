package httpapi

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/config"
	"myself/server/internal/store"
)

/* ===== 当前用户与角色鉴权 ===== */

type userCtxKey struct{}

// userOf 取经 requireRole 认证的当前用户（未经认证的路由为 nil）。
func userOf(r *http.Request) *store.User {
	u, _ := r.Context().Value(userCtxKey{}).(*store.User)
	return u
}

// roleOn 该角色在当前开关下是否可用：管理员始终可用；作者 / 读者受用户系统总开关与各自开关控制。
func roleOn(cfg config.Typed, role string) bool {
	switch role {
	case store.RoleAdmin:
		return true
	case store.RoleAuthor:
		return cfg.Users.AuthorsOn()
	case store.RoleReader:
		return cfg.Users.ReadersOn()
	}
	return false
}

// currentUser 解析会话（Cookie 或 Bearer）；无会话或会话无效返回 nil。fromCookie 用于 CSRF 判定。
func (s *Server) currentUser(r *http.Request) (u *store.User, fromCookie bool) {
	token, fromCookie := sessionToken(r)
	if token == "" {
		return nil, fromCookie
	}
	u, err := s.Auth.VerifyToken(token)
	if err != nil || !roleOn(s.Config.Typed(), u.Role) {
		return nil, fromCookie
	}
	return u, fromCookie
}

// requireRole 认证中间件：会话有效、角色在允许列表内且该角色当前已开放；
// Cookie 会话的写操作需过 CSRF 校验；仍需强制改密的账号只放行改密接口。
func (s *Server) requireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			u, fromCookie := s.currentUser(r)
			if u == nil {
				if fromCookie {
					clearSession(w, r)
				}
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if fromCookie && !csrfOK(r) {
				writeError(w, http.StatusForbidden, "csrf_rejected")
				return
			}
			allowed := false
			for _, role := range roles {
				allowed = allowed || u.Role == role
			}
			if !allowed {
				writeError(w, http.StatusForbidden, "forbidden")
				return
			}
			if u.MustChange && !(r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/auth/password")) {
				writeError(w, http.StatusForbidden, "must_change_password")
				return
			}
			s.DB.TouchUser(u.ID)
			next(w, r.WithContext(context.WithValue(r.Context(), userCtxKey{}, u)))
		}
	}
}

// publicUser 下发给前端的当前用户信息（不含口令哈希等）。

// startSession 签发令牌并写入会话 Cookie。
func (s *Server) startSession(w http.ResponseWriter, r *http.Request, u *store.User) error {
	token, err := s.Auth.Issue(u)
	if err != nil {
		return err
	}
	setSession(w, r, token)
	return nil
}

/* ===== 登录 / 会话 / 改密 ===== */

// POST /auth/login {username|email, password}：失败计数按 IP 锁定；成功写入 HttpOnly 会话 Cookie。
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
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
	id := b.strOr("username")
	if id == "" {
		id = b.strOr("email")
	}
	var (
		u   *store.User
		err error
	)
	withScryptSlot(func() { u, err = s.Auth.Authenticate(id, b.strOr("password")) })
	switch {
	case errors.Is(err, auth.ErrPending):
		writeError(w, http.StatusForbidden, "pending_verification")
		return
	case errors.Is(err, auth.ErrDisabled):
		writeError(w, http.StatusForbidden, "account_disabled")
		return
	case err != nil:
		fail(w, err)
		return
	case u == nil:
		s.limiter.fail(ip)
		writeError(w, http.StatusUnauthorized, "bad_credentials")
		return
	}
	if !roleOn(s.Config.Typed(), u.Role) {
		writeError(w, http.StatusForbidden, "role_closed")
		return
	}
	s.limiter.success(ip)
	if err := s.startSession(w, r, u); err != nil {
		fail(w, err)
		return
	}
	s.DB.TouchUser(u.ID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "mustChange": u.MustChange, "user": s.publicUser(u)})
}

// GET /auth/session 当前登录状态与用户信息
func (s *Server) session(w http.ResponseWriter, r *http.Request) {
	u, _ := s.currentUser(r)
	if u == nil {
		writeJSON(w, http.StatusOK, map[string]any{"loggedIn": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"loggedIn": true, "mustChange": u.MustChange, "user": s.publicUser(u)})
}

// PUT /auth/password 改密：该用户其它会话全部失效，当前会话换发新 Cookie
func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
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
	newPassword, ok := b.str("newPassword")
	if !validPassword(newPassword) {
		ok = false
	}
	if !ok {
		writeError(w, http.StatusBadRequest, "weak_password")
		return
	}
	var (
		token string
		err   error
	)
	withScryptSlot(func() { token, err = s.Auth.ChangePassword(userOf(r).ID, b.strOr("oldPassword"), newPassword) })
	if err != nil {
		fail(w, err)
		return
	}
	if token == "" {
		s.limiter.fail(ip)
		writeError(w, http.StatusUnauthorized, "bad_credentials")
		return
	}
	setSession(w, r, token)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// PUT /me 修改自己的昵称（头像走 /me/avatar 上传 + 审核，不接受任意外链）
func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	u := userOf(r)
	name := strings.TrimSpace(b.strOr("name"))
	if name == "" || len([]rune(name)) > 40 {
		writeError(w, http.StatusBadRequest, "invalid_name")
		return
	}
	if _, err := s.DB.Exec(`UPDATE users SET name = ? WHERE id = ?`, name, u.ID); err != nil {
		fail(w, err)
		return
	}
	s.writeMe(w, u.ID)
}

var emailRe = regexp.MustCompile(`^[^\s@]{1,64}@[^\s@]{1,190}\.[^\s@]{2,}$`)

func validPassword(p string) bool {
	return len([]rune(p)) >= auth.MinPasswordLen && len(p) <= 256
}

/* ===== 注册 / 邀请 / 邮箱验证 ===== */

// GET /auth/invite?code= 查看邀请（前台加入页据此显示角色与预填邮箱）
func (s *Server) inviteInfo(w http.ResponseWriter, r *http.Request) {
	t, err := s.Auth.PeekToken("invite", r.URL.Query().Get("code"))
	if err != nil || !roleOn(s.Config.Typed(), t.Role) {
		writeError(w, http.StatusNotFound, "invalid_invite")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"role": t.Role, "email": t.Email})
}

// POST /auth/register {email, password, name, invite?}
// 无邀请：读者开放注册时可用；有邀请：按邀请角色加入（作者 / 读者，对应开关需开启），邀请即视为邮箱已验证。
// 开放注册且要求验证邮箱（并已配置发信）时账号先处于待验证状态。
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if d := s.signupLimiter.blocked(ip); d > 0 {
		tooMany(w, d)
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	// 蜜罐：真人看不到的字段，填了就是机器
	if b.strOr("website") != "" {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "pending": true})
		return
	}
	cfg := s.Config.Typed()
	email := strings.ToLower(strings.TrimSpace(b.strOr("email")))
	name := strings.TrimSpace(b.strOr("name"))
	password := b.strOr("password")
	if !emailRe.MatchString(email) {
		writeError(w, http.StatusBadRequest, "invalid_email")
		return
	}
	if name == "" || len([]rune(name)) > 40 {
		writeError(w, http.StatusBadRequest, "invalid_name")
		return
	}
	if !validPassword(password) {
		writeError(w, http.StatusBadRequest, "weak_password")
		return
	}

	role, verified := store.RoleReader, false
	var invite *auth.Token
	if code := b.strOr("invite"); code != "" {
		t, err := s.Auth.PeekToken("invite", code)
		if err != nil || !roleOn(cfg, t.Role) || (t.Email != "" && !strings.EqualFold(t.Email, email)) {
			writeError(w, http.StatusForbidden, "invalid_invite")
			return
		}
		invite, role, verified = t, t.Role, true
	} else if !cfg.Users.ReadersOn() || cfg.Users.Readers.Signup != "open" {
		writeError(w, http.StatusForbidden, "signup_closed")
		return
	}

	status := store.StatusActive
	needVerify := !verified && cfg.Users.Readers.RequireVerify && cfg.Mail.Ready()
	if needVerify {
		status = store.StatusPending
	}
	var hash string
	var err error
	withScryptSlot(func() { hash, err = auth.HashPassword(password) })
	if err != nil {
		fail(w, err)
		return
	}
	s.signupLimiter.fail(ip) // 注册也计次：同一 IP 15 分钟内最多 5 个
	id, err := s.DB.CreateUser(store.NewUser{
		Login: email, Email: email, Name: name, Role: role, Status: status, PasswordHash: hash, EmailVerified: verified,
	})
	if store.IsUniqueErr(err) {
		writeError(w, http.StatusConflict, "email_taken")
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	if invite != nil {
		if err := s.Auth.ConsumeToken(invite, id); err != nil {
			_, _ = s.DB.Exec(`DELETE FROM users WHERE id = ?`, id)
			writeError(w, http.StatusForbidden, "invalid_invite")
			return
		}
	}
	u, err := s.DB.UserByID(id)
	if err != nil || u == nil {
		fail(w, err)
		return
	}
	if needVerify {
		if err := s.sendVerifyMail(r, u); err != nil {
			s.logMail(err)
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "pending": true})
		return
	}
	if err := s.startSession(w, r, u); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": s.publicUser(u)})
}

func (s *Server) sendVerifyMail(r *http.Request, u *store.User) error {
	raw, err := s.Auth.CreateToken("verify", u.ID, "", u.Email, "", 48*time.Hour)
	if err != nil {
		return err
	}
	link := s.siteBase(r) + "/account/verify?token=" + raw
	site := s.Config.Typed().Site.Title
	return s.sendLetter(u.Email, letter{
		Subject:   "验证你在「" + site + "」的邮箱",
		Preheader: "点一下按钮完成邮箱验证，账号即可使用。",
		Title:     "验证你的邮箱",
		Greeting:  u.Name + "，你好：",
		Lines:     []string{"欢迎加入「" + site + "」。点下面的按钮完成邮箱验证，账号就会激活并自动登录。"},
		Action:    &mailAction{Label: "验证邮箱", URL: link},
		Expire:    "链接 48 小时内有效，只能使用一次。",
		Note:      "如果这不是你本人的操作，忽略这封邮件即可，不会有任何影响。",
	}, s.siteBase(r))
}

// POST /auth/verify {token} 验证邮箱并登录
func (s *Server) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	t, err := s.Auth.PeekToken("verify", b.strOr("token"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token")
		return
	}
	if err := s.Auth.ConsumeToken(t, t.UserID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token")
		return
	}
	if _, err := s.DB.Exec(`UPDATE users SET email_verified = 1, status = CASE status WHEN 'pending' THEN 'active' ELSE status END WHERE id = ?`, t.UserID); err != nil {
		fail(w, err)
		return
	}
	u, err := s.DB.UserByID(t.UserID)
	if err != nil || u == nil || u.Status != store.StatusActive {
		writeError(w, http.StatusBadRequest, "invalid_token")
		return
	}
	if roleOn(s.Config.Typed(), u.Role) {
		if err := s.startSession(w, r, u); err != nil {
			fail(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": s.publicUser(u)})
}

/* ===== 找回密码 ===== */

// POST /auth/forgot {email}：已配置发信时寄出重置链接。无论邮箱是否存在都回同样的结果，避免探测注册邮箱。
func (s *Server) forgotPassword(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if d := s.signupLimiter.blocked(ip); d > 0 {
		tooMany(w, d)
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	cfg := s.Config.Typed()
	if !cfg.Mail.Ready() {
		writeError(w, http.StatusServiceUnavailable, "mail_unavailable")
		return
	}
	s.signupLimiter.fail(ip)
	email := strings.TrimSpace(b.strOr("email"))
	if u, err := s.DB.UserBy(`email = ?`, email); err == nil && u != nil && u.Status == store.StatusActive {
		go func(u store.User, base string) {
			raw, err := s.Auth.CreateToken("reset", u.ID, "", u.Email, "", time.Hour)
			if err != nil {
				s.logMail(err)
				return
			}
			site := cfg.Site.Title
			err = s.sendLetter(u.Email, letter{
				Subject:   "重置你在「" + site + "」的密码",
				Preheader: "有人申请重置你的密码；如果是你，点按钮设置新密码。",
				Title:     "重置密码",
				Greeting:  u.Name + "，你好：",
				Lines:     []string{"我们收到了重置「" + site + "」账号密码的申请。点下面的按钮设置一个新密码。"},
				Action:    &mailAction{Label: "设置新密码", URL: base + "/account/reset?token=" + raw},
				Expire:    "链接 1 小时内有效，只能使用一次。",
				Note:      "如果你没有申请重置密码，忽略这封邮件即可，原密码仍然有效。",
			}, base)
			if err != nil {
				s.logMail(err)
			}
		}(*u, s.siteBase(r))
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /auth/reset?token= 检查重置链接是否有效
func (s *Server) resetInfo(w http.ResponseWriter, r *http.Request) {
	t, err := s.Auth.PeekToken("reset", r.URL.Query().Get("token"))
	if err != nil {
		writeError(w, http.StatusNotFound, "invalid_token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"email": t.Email})
}

// POST /auth/reset {token, password} 设置新密码并登录（该用户其它会话全部失效）
func (s *Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	password := b.strOr("password")
	if !validPassword(password) {
		writeError(w, http.StatusBadRequest, "weak_password")
		return
	}
	t, err := s.Auth.PeekToken("reset", b.strOr("token"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token")
		return
	}
	if err := s.Auth.ConsumeToken(t, t.UserID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token")
		return
	}
	withScryptSlot(func() { err = s.Auth.SetPassword(t.UserID, password) })
	if err != nil {
		fail(w, err)
		return
	}
	u, err := s.DB.UserByID(t.UserID)
	if err != nil || u == nil {
		fail(w, err)
		return
	}
	if u.Status == store.StatusActive && roleOn(s.Config.Typed(), u.Role) {
		if err := s.startSession(w, r, u); err != nil {
			fail(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": s.publicUser(u)})
}
