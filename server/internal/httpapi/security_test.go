package httpapi

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"myself/server/internal/config"
)

// Host 头投毒：伪造 Host 发起找回密码，邮件链接只能指向配置的站点地址；未配置时不发信。
func TestResetLinkIgnoresHostHeader(t *testing.T) {
	e := newEnv(t)
	if _, err := e.server.DB.Exec(`UPDATE users SET email = 'owner@example.com' WHERE role = 'admin'`); err != nil {
		t.Fatal(err)
	}
	if _, err := e.server.Config.Save(config.Map{"mail": config.Map{"enabled": true, "host": "smtp.example", "port": 587, "from": "Site <bot@example.com>"}}); err != nil {
		t.Fatal(err)
	}
	sent := make(chan string, 4)
	e.server.mailer = func(_ config.Mail, _, _, text, _ string) error {
		sent <- text
		return nil
	}
	forgot := func() {
		req, _ := http.NewRequest(http.MethodPost, e.srv.URL+"/api/v1/auth/forgot", strings.NewReader(`{"email":"owner@example.com"}`))
		req.Host = "evil.example"
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Requested-With", "myself")
		req.Header.Set("Origin", "http://evil.example")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("forgot: %d", res.StatusCode)
		}
	}

	// 未配置站点地址：不发
	forgot()
	select {
	case txt := <-sent:
		t.Fatalf("mail sent without site url: %s", txt)
	case <-time.After(300 * time.Millisecond):
	}

	// 配置后：链接只用配置的地址
	if _, err := e.server.Config.Save(config.Map{"site": config.Map{"url": "https://blog.example"}}); err != nil {
		t.Fatal(err)
	}
	forgot()
	select {
	case txt := <-sent:
		if strings.Contains(txt, "evil.example") || !strings.Contains(txt, "https://blog.example/account/reset?token=") {
			t.Fatalf("reset link: %s", txt)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("reset mail not sent")
	}
}

func TestSafeNext(t *testing.T) {
	for in, want := range map[string]string{
		"/articles/x?y=1":      "/articles/x?y=1",
		"/\t/evil.com":         "/",
		"/%09/evil.com":        "/%09/evil.com", // 仍是站内路径（未解码的 %09 不会被浏览器剔除）
		"//evil.com":           "/",
		"/\\evil.com":          "/",
		"https://evil.com":     "/",
		"javascript:alert(1)":  "/",
		"/ok\r\nSet-Cookie: x": "/",
		"":                     "/",
	} {
		if got := safeNext(in); got != want {
			t.Errorf("safeNext(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClientIPAndLimitKey(t *testing.T) {
	old := trustProxy
	trustProxy = true
	defer func() { trustProxy = old }()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	r.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.9")
	if got := clientIP(r); got != "203.0.113.9" {
		t.Fatalf("clientIP took spoofable left-most entry: %s", got)
	}
	if limitKey("2001:db8::1") != limitKey("2001:db8::ffff") || limitKey("2001:db8::1") == limitKey("2001:db9::1") {
		t.Fatal("IPv6 not aggregated per /64")
	}
	if limitKey("203.0.113.9") != "203.0.113.9" {
		t.Fatal("IPv4 key changed")
	}
}

// 公开写接口同样要 CSRF 头；按账号的失败锁定独立于 IP
func TestLoginCSRFAndAccountLockout(t *testing.T) {
	e := newEnv(t)
	res := e.do(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"x"}`),
		map[string]string{"Content-Type": "text/plain", "X-Requested-With": ""})
	res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("login without csrf header: %d", res.StatusCode)
	}
	for i := 0; i < accountFailLimit; i++ {
		e.server.limiter.success("127.0.0.1") // 模拟来自不同 IP：只让账号维度计数
		res := e.do(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"wrong"}`), map[string]string{"Content-Type": "application/json"})
		res.Body.Close()
	}
	e.server.limiter.success("127.0.0.1")
	res = e.do(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"`+testPassword+`"}`), map[string]string{"Content-Type": "application/json"})
	res.Body.Close()
	if res.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("account lockout not applied: %d", res.StatusCode)
	}
}
