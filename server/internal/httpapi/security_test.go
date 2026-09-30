package httpapi

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image/color"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

// bombPNG 只有文件头、声明巨大尺寸的 PNG（解码时会一次性分配数 GB）
func bombPNG(w, h uint32) []byte {
	var buf bytes.Buffer
	buf.Write([]byte("\x89PNG\r\n\x1a\n"))
	chunk := func(typ string, data []byte) {
		_ = binary.Write(&buf, binary.BigEndian, uint32(len(data)))
		buf.WriteString(typ)
		buf.Write(data)
		c := crc32.NewIEEE()
		c.Write([]byte(typ))
		c.Write(data)
		_ = binary.Write(&buf, binary.BigEndian, c.Sum32())
	}
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:], w)
	binary.BigEndian.PutUint32(ihdr[4:], h)
	ihdr[8], ihdr[9] = 8, 6 // 8 位 RGBA
	chunk("IHDR", ihdr)
	chunk("IDAT", nil)
	chunk("IEND", nil)
	return buf.Bytes()
}

// 头像解压炸弹：解码前按文件头拒绝，不分配巨量内存
func TestAvatarDecodeBomb(t *testing.T) {
	e := newEnv(t)
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("file", "a.png")
	part.Write(bombPNG(20000, 20000))
	mw.Close()
	res := e.do(http.MethodPost, "/api/v1/me/avatar", &body, map[string]string{"Content-Type": mw.FormDataContentType(), "Authorization": "Bearer " + e.token})
	res.Body.Close()
	if res.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("bomb avatar: %d", res.StatusCode)
	}
}

// 压缩包目录体量：尾部记录声明百万条目的包在打开前就被拒绝
func TestZipDirectoryGuard(t *testing.T) {
	dir := t.TempDir()
	eocd := make([]byte, 22)
	binary.LittleEndian.PutUint32(eocd, 0x06054b50)
	binary.LittleEndian.PutUint16(eocd[8:], 0xffff-1)
	binary.LittleEndian.PutUint16(eocd[10:], 0xffff-1)
	binary.LittleEndian.PutUint32(eocd[12:], 64<<20)
	p := filepath.Join(dir, "bomb.zip")
	if err := os.WriteFile(p, eocd, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := checkZipDirectory(p); err != errZipTooBig {
		t.Fatalf("oversized directory accepted: %v", err)
	}
	var ok bytes.Buffer
	zw := zip.NewWriter(&ok)
	f, _ := zw.Create("a.txt")
	f.Write([]byte("x"))
	zw.Close()
	os.WriteFile(p, ok.Bytes(), 0o644)
	if err := checkZipDirectory(p); err != nil {
		t.Fatalf("normal zip rejected: %v", err)
	}
}

func TestSafeNameRejectsReserved(t *testing.T) {
	for _, n := range []string{"NUL.png", "con", "com1.txt", "lpt9.zip", "../x.png", "a/b.png", ".pending"} {
		if safeName(n) != "" && n != ".pending" {
			t.Errorf("safeName(%q) accepted", n)
		}
	}
	if safeName("nullable.png") == "" || safeName("a1b2-c3.webp") == "" {
		t.Error("normal names rejected")
	}
}

// 压缩记录里被篡改的原文件名（如恶意备份写入 ../）不被当成路径使用
func TestCompressionRecordSanitized(t *testing.T) {
	e := newEnv(t)
	img := e.uploadFiles(pngOf(color.RGBA{R: 9, A: 255}))[0]
	if _, err := e.server.DB.Exec(`UPDATE media SET compressed_from = '../../data/myself.db', compressed_before = 1 WHERE name = ?`, img.Name); err != nil {
		t.Fatal(err)
	}
	if c := e.server.compressionOf(img.Name); c != nil {
		t.Fatalf("unsafe compressed_from used: %+v", c)
	}
}
