package httpapi

import (
	"bytes"
	"errors"
	"image"
	_ "image/gif"  // 头像上传支持 GIF（取首帧）
	_ "image/jpeg" // 头像上传支持 JPEG
	_ "image/png"  // 头像上传支持 PNG
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/config"
	"myself/server/internal/imaging"
	"myself/server/internal/store"
)

/*
 * 个人资料：头像（上传 + 裁切由前端完成，服务端统一缩成 256px WebP；非站长需站长审核）、
 * 登录名（每 30 天可改一次）、邮箱（改后需重新验证）。
 * 站长未单独设置头像时，全站展示「身份」里的头像。
 */

const (
	avatarSide     = 256
	avatarMaxBytes = 5 << 20
	avatarPrefix   = "avatar-"
	// loginCooldown 登录名两次修改的最短间隔
	loginCooldown = 30 * 24 * time.Hour
)

// loginRe 登录名：字母开头，3–24 位字母、数字、下划线或短横线。
var loginRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{2,23}$`)

// builtinLogo 内置 logo（与前端 BUILTIN_LOGO 一致）。
const builtinLogo = "/favicon-256.png"

// ownerAvatar 站长未单独设置头像时展示的头像：身份头像 → 站点 logo → 内置 logo（与前台身份展示同一条回退链）。
func (s *Server) ownerAvatar() string {
	cfg := s.Config.Get()
	if v, _ := config.Sub(cfg, "about")["avatar"].(string); v != "" {
		return v
	}
	if v, _ := config.Sub(cfg, "site")["logo"].(string); v != "" {
		return v
	}
	return builtinLogo
}

// avatarOf 用户对外展示的头像：站长没有自己的头像时用身份头像。
func (s *Server) avatarOf(role, avatar string) string {
	if avatar == "" && role == store.RoleAdmin {
		return s.ownerAvatar()
	}
	return avatar
}

// publicUser 返回给本人的资料。avatarDefault = 站长正在使用身份头像。
func (s *Server) publicUser(u *store.User) map[string]any {
	next := ""
	if t, err := time.Parse("2006-01-02 15:04:05", u.LoginChangedAt); err == nil {
		if n := t.Add(loginCooldown); n.After(time.Now().UTC()) {
			next = n.Format("2006-01-02 15:04:05")
		}
	}
	return map[string]any{
		"id": u.ID, "login": u.Login, "email": u.Email, "name": u.Name, "role": u.Role,
		"avatar": s.avatarOf(u.Role, u.Avatar), "avatarDefault": u.Role == store.RoleAdmin && u.Avatar == "",
		"avatarPending": u.AvatarPending, "loginNextChange": next, "emailPending": s.pendingEmail(u.ID),
		"hasPassword": u.PasswordHash != "", "github": u.GitHubID != 0, "emailVerified": u.EmailVerified,
		"mustChange": u.MustChange,
	}
}

// removeAvatarFile 删除本站生成的头像文件（外链头像、身份头像不动）。
func (s *Server) removeAvatarFile(url string) {
	name := strings.TrimPrefix(url, "/uploads/")
	if name == url || !strings.HasPrefix(name, avatarPrefix) || safeName(name) == "" {
		return
	}
	_ = os.Remove(filepath.Join(s.UploadDir, name))
}

// POST /me/avatar（multipart: file）上传头像：站长直接生效，其他人进入待审。
func (s *Server) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	u := userOf(r)
	r.Body = http.MaxBytesReader(w, r.Body, avatarMaxBytes+1<<20)
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_file")
		return
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, avatarMaxBytes+1))
	if err != nil || len(raw) > avatarMaxBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "file_too_large")
		return
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_image")
		return
	}
	// 前端已裁成正方形；这里再居中取正方形兜底，统一缩放与格式
	b := img.Bounds()
	side := min(b.Dx(), b.Dy())
	img = imaging.Crop(img, imaging.Rect{Left: (b.Dx() - side) / 2, Top: (b.Dy() - side) / 2, Width: side, Height: side})
	img = imaging.Fit(img, avatarSide)
	data, err := imaging.EncodeBytes(img, ".webp", 86)
	if err != nil {
		fail(w, err)
		return
	}
	name := avatarPrefix + randomFilename(".webp")
	if err := os.WriteFile(filepath.Join(s.UploadDir, name), data, 0o644); err != nil {
		fail(w, err)
		return
	}
	url := "/uploads/" + name
	if u.Role == store.RoleAdmin {
		_, err = s.DB.Exec(`UPDATE users SET avatar = ?, avatar_pending = '' WHERE id = ?`, url, u.ID)
		s.removeAvatarFile(u.Avatar)
	} else {
		_, err = s.DB.Exec(`UPDATE users SET avatar_pending = ? WHERE id = ?`, url, u.ID)
	}
	if err != nil {
		fail(w, err)
		return
	}
	s.removeAvatarFile(u.AvatarPending)
	s.writeMe(w, u.ID)
}

// DELETE /me/avatar 移除头像（连同待审的）；站长恢复为身份头像。?pending=1 只撤回待审的那张。
func (s *Server) deleteAvatar(w http.ResponseWriter, r *http.Request) {
	u := userOf(r)
	if r.URL.Query().Get("pending") == "1" {
		if _, err := s.DB.Exec(`UPDATE users SET avatar_pending = '' WHERE id = ?`, u.ID); err != nil {
			fail(w, err)
			return
		}
		s.removeAvatarFile(u.AvatarPending)
		s.writeMe(w, u.ID)
		return
	}
	if _, err := s.DB.Exec(`UPDATE users SET avatar = '', avatar_pending = '' WHERE id = ?`, u.ID); err != nil {
		fail(w, err)
		return
	}
	s.removeAvatarFile(u.Avatar)
	s.removeAvatarFile(u.AvatarPending)
	s.writeMe(w, u.ID)
}

// PUT /me/login {login} 修改登录名：格式校验、唯一、每 30 天一次。
func (s *Server) changeLogin(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	u := userOf(r)
	login := strings.TrimSpace(b.strOr("login"))
	if !loginRe.MatchString(login) {
		writeError(w, http.StatusBadRequest, "invalid_login")
		return
	}
	if strings.EqualFold(login, u.Login) {
		s.writeMe(w, u.ID)
		return
	}
	if t, err := time.Parse("2006-01-02 15:04:05", u.LoginChangedAt); err == nil && time.Since(t) < loginCooldown {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "login_cooldown", "next": t.Add(loginCooldown).Format("2006-01-02 15:04:05")})
		return
	}
	if _, err := s.DB.Exec(`UPDATE users SET login = ?, login_changed_at = datetime('now') WHERE id = ?`, login, u.ID); err != nil {
		if store.IsUniqueErr(err) {
			writeError(w, http.StatusConflict, "login_taken")
			return
		}
		fail(w, err)
		return
	}
	s.writeMe(w, u.ID)
}

// emailChangeTTL 换绑邮箱确认链接的有效期。
const emailChangeTTL = 24 * time.Hour

// pendingEmail 用户待确认的新邮箱（最近一条未使用、未过期的换绑令牌）。
func (s *Server) pendingEmail(uid int64) string {
	var email string
	_ = s.DB.QueryRow(`SELECT email FROM user_tokens WHERE kind = 'email' AND user_id = ? AND used_at IS NULL
		AND expires_at > datetime('now') ORDER BY id DESC LIMIT 1`, uid).Scan(&email)
	return email
}

// PUT /me/email {email, password} 添加 / 更换邮箱：已设密码的账号需验证当前密码；
// 向新邮箱发确认邮件，点开链接后才真正换绑（并标记已验证）。站点未配置发信时无法添加。
func (s *Server) changeEmail(w http.ResponseWriter, r *http.Request) {
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
	u := userOf(r)
	email := strings.ToLower(strings.TrimSpace(b.strOr("email")))
	if !emailRe.MatchString(email) {
		writeError(w, http.StatusBadRequest, "invalid_email")
		return
	}
	if u.PasswordHash != "" {
		var ok bool
		withScryptSlot(func() { ok = auth.VerifyPassword(b.strOr("password"), u.PasswordHash) })
		if !ok {
			s.limiter.fail(ip)
			writeError(w, http.StatusUnauthorized, "bad_credentials")
			return
		}
	}
	if strings.EqualFold(email, u.Email) {
		s.writeMe(w, u.ID)
		return
	}
	if other, err := s.DB.UserBy(`email = ?`, email); err != nil {
		fail(w, err)
		return
	} else if other != nil {
		writeError(w, http.StatusConflict, "email_taken")
		return
	}
	cfg := s.Config.Typed()
	if !cfg.Mail.Ready() {
		writeError(w, http.StatusServiceUnavailable, "mail_unavailable")
		return
	}
	// 新的换绑请求作废旧的
	if _, err := s.DB.Exec(`DELETE FROM user_tokens WHERE kind = 'email' AND user_id = ? AND used_at IS NULL`, u.ID); err != nil {
		fail(w, err)
		return
	}
	raw, err := s.Auth.CreateToken("email", u.ID, "", email, "", emailChangeTTL)
	if err != nil {
		fail(w, err)
		return
	}
	base := s.siteBase(r)
	site := cfg.Site.Title
	title, line := "确认你的新邮箱", "你正在为「"+site+"」账号绑定这个邮箱。点下面的按钮确认，之后就可以用它登录和找回密码。"
	if u.Email != "" {
		title, line = "确认更换邮箱", "你正在把「"+site+"」账号的邮箱从 "+u.Email+" 更换为这个邮箱。点下面的按钮确认，确认前原邮箱仍然有效。"
	}
	if err := s.sendLetter(email, letter{
		Subject:   title + " · " + site,
		Preheader: "点按钮确认，这个邮箱才会绑定到你的账号。",
		Title:     title,
		Greeting:  u.Name + "，你好：",
		Lines:     []string{line},
		Action:    &mailAction{Label: "确认邮箱", URL: base + "/account/verify?token=" + raw},
		Expire:    "链接 24 小时内有效，只能使用一次。",
		Note:      "如果这不是你本人的操作，忽略这封邮件即可，账号不会有任何变化。",
	}, base); err != nil {
		s.logMail(err)
		_, _ = s.DB.Exec(`DELETE FROM user_tokens WHERE kind = 'email' AND user_id = ? AND used_at IS NULL`, u.ID)
		writeError(w, http.StatusBadGateway, "mail_failed")
		return
	}
	fresh, err := s.DB.UserByID(u.ID)
	if err != nil || fresh == nil {
		fail(w, errors.New("user vanished"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": s.publicUser(fresh)})
}

// DELETE /me/email/pending 取消待确认的换绑。
func (s *Server) cancelEmailChange(w http.ResponseWriter, r *http.Request) {
	u := userOf(r)
	if _, err := s.DB.Exec(`DELETE FROM user_tokens WHERE kind = 'email' AND user_id = ? AND used_at IS NULL`, u.ID); err != nil {
		fail(w, err)
		return
	}
	s.writeMe(w, u.ID)
}

// PUT /admin/users/{id}/avatar {action: approve|reject} 审核待审头像。
func (s *Server) adminReviewAvatar(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	u, err := s.DB.UserByID(pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	if u == nil || u.AvatarPending == "" {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	switch b.strOr("action") {
	case "approve":
		if _, err := s.DB.Exec(`UPDATE users SET avatar = avatar_pending, avatar_pending = '' WHERE id = ?`, u.ID); err != nil {
			fail(w, err)
			return
		}
		s.removeAvatarFile(u.Avatar)
	case "reject":
		if _, err := s.DB.Exec(`UPDATE users SET avatar_pending = '' WHERE id = ?`, u.ID); err != nil {
			fail(w, err)
			return
		}
		s.removeAvatarFile(u.AvatarPending)
	default:
		writeError(w, http.StatusBadRequest, "invalid_action")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// writeMe 返回本人最新资料。
func (s *Server) writeMe(w http.ResponseWriter, id int64) {
	u, err := s.DB.UserByID(id)
	if err != nil || u == nil {
		fail(w, errors.New("user vanished"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": s.publicUser(u)})
}
