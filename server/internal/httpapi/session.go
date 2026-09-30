package httpapi

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 会话 Cookie：管理员令牌只放在 HttpOnly Cookie 里，页面脚本读不到，XSS 也就偷不走令牌。
// SameSite=Strict 挡住跨站请求携带 Cookie；另要求写操作带自定义头并校验 Origin（CSRF 纵深防御）。
const (
	sessionCookie = "myself_session"
	// csrfHeader 前端所有写请求都带 X-Requested-With: myself；跨站表单无法设置自定义头。
	csrfHeader = "X-Requested-With"
	csrfValue  = "myself"
)

func secureRequest(r *http.Request) bool {
	return r.TLS != nil || (trustProxy && strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"))
}

// setSession 写入会话 Cookie（仅 /api/ 路径携带）。
// maxCookieAge 浏览器对 Cookie 有效期的上限（Chrome 约 400 天）；更长的登录靠访问时续期。
const maxCookieAge = 400 * 24 * time.Hour

// setSession 写入会话 Cookie，有效期跟随后台「登录保持时长」。
func (s *Server) setSession(w http.ResponseWriter, r *http.Request, token string) {
	age := min(s.Config.Typed().SessionTTL(), maxCookieAge)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/api/",
		MaxAge:   int(age.Seconds()),
		HttpOnly: true,
		Secure:   secureRequest(r),
		SameSite: http.SameSiteStrictMode,
	})
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/api/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secureRequest(r),
		SameSite: http.SameSiteStrictMode,
	})
}

// sessionToken 取令牌：优先会话 Cookie（浏览器），其次 Authorization: Bearer（脚本 / 测试）。
// fromCookie 用于决定是否需要 CSRF 校验。
func sessionToken(r *http.Request) (token string, fromCookie bool) {
	if c, err := r.Cookie(sessionCookie); err == nil && c.Value != "" {
		return c.Value, true
	}
	if raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return raw, false
	}
	return "", false
}

// csrfOK 写操作（非 GET/HEAD）经 Cookie 认证时：必须带自定义头；若有 Origin 则须与 Host 同源。
func csrfOK(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	if r.Header.Get(csrfHeader) != csrfValue {
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" && !sameOrigin(origin, r.Host) {
		return false
	}
	return true
}

// sameOrigin Origin 头的主机是否与请求 Host 一致
func sameOrigin(origin, host string) bool {
	u, err := url.Parse(origin)
	return err == nil && u.Host != "" && strings.EqualFold(u.Host, host)
}

// withCSRF 公开的写接口（登录 / 注册 / 找回密码等）同样要求自定义头与同源 Origin：
// 防止跨站表单把受害者登录进攻击者账号（登录 CSRF）、强制登出或代发邮件。
func withCSRF(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !csrfOK(r) {
			writeError(w, http.StatusForbidden, "csrf")
			return
		}
		h(w, r)
	}
}

// POST /auth/logout 清除会话 Cookie（无需登录，幂等）
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	clearSession(w, r)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
