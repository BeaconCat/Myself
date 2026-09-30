package httpapi

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// uploadAvatarAs 以指定令牌上传一张 w×h 的 PNG 头像，返回状态码与响应。
func (e *env) uploadAvatarAs(token string, w, h int) (int, map[string]any) {
	e.t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: 90, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, _ := mw.CreateFormFile("file", "a.png")
	png.Encode(part, img)
	mw.Close()
	h2 := map[string]string{"Content-Type": mw.FormDataContentType(), "Authorization": "Bearer " + token}
	res := e.do(http.MethodPost, "/api/v1/me/avatar", &buf, h2)
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	var out map[string]any
	_ = json.Unmarshal(data, &out)
	return res.StatusCode, out
}

func userField(out map[string]any, key string) string {
	u, _ := out["user"].(map[string]any)
	v, _ := u[key].(string)
	return v
}

// 头像：读者上传进入待审（展示仍是旧头像），站长通过后生效、驳回则删文件；头像文件不进素材库。
func TestAvatarReview(t *testing.T) {
	e := newEnv(t)
	e.enableUsers(map[string]any{"enabled": true, "readers": map[string]any{"enabled": true, "signup": "open"}})
	tok := e.signup(map[string]any{"email": "r@example.com", "name": "读者", "password": "password123"})

	code, out := e.uploadAvatarAs(tok, 400, 300)
	if code != http.StatusOK {
		t.Fatalf("upload: %d %v", code, out)
	}
	pending := userField(out, "avatarPending")
	if pending == "" || userField(out, "avatar") != "" {
		t.Fatalf("reader avatar should be pending: %v", out)
	}
	file := filepath.Join(e.root, "uploads", strings.TrimPrefix(pending, "/uploads/"))
	cfg, _, err := func() (image.Config, string, error) {
		f, err := os.Open(file)
		if err != nil {
			return image.Config{}, "", err
		}
		defer f.Close()
		return image.DecodeConfig(f)
	}()
	if err != nil || cfg.Width != 256 || cfg.Height != 256 {
		t.Fatalf("avatar should be a 256px square: %+v %v", cfg, err)
	}

	// 不进素材库
	var media []struct{ URL string }
	e.call(http.MethodGet, "/api/v1/admin/media", nil, &media, http.StatusOK)
	for _, m := range media {
		if strings.Contains(m.URL, avatarPrefix) {
			t.Fatal("avatar leaked into media library")
		}
	}

	// 管理员列表里能看到待审头像，通过后生效
	var list struct {
		Items []struct {
			ID            int64
			Email         string
			AvatarPending string
		}
		Stats map[string]int
	}
	e.call(http.MethodGet, "/api/v1/admin/users", nil, &list, http.StatusOK)
	var id int64
	for _, u := range list.Items {
		if u.Email == "r@example.com" {
			id = u.ID
			if u.AvatarPending != pending {
				t.Fatalf("admin list pending = %q", u.AvatarPending)
			}
		}
	}
	if list.Stats["avatars"] != 1 {
		t.Fatalf("avatars stat = %d", list.Stats["avatars"])
	}
	e.call(http.MethodPut, "/api/v1/admin/users/"+itoa(id)+"/avatar", map[string]string{"action": "approve"}, nil, http.StatusOK)
	_, me := e.as(tok, http.MethodGet, "/api/v1/auth/session", nil)
	if userField(me, "avatar") != pending || userField(me, "avatarPending") != "" {
		t.Fatalf("after approve: %v", me)
	}

	// 再传一张并驳回：文件删除、展示不变
	_, out = e.uploadAvatarAs(tok, 64, 64)
	second := userField(out, "avatarPending")
	e.call(http.MethodPut, "/api/v1/admin/users/"+itoa(id)+"/avatar", map[string]string{"action": "reject"}, nil, http.StatusOK)
	if _, err := os.Stat(filepath.Join(e.root, "uploads", strings.TrimPrefix(second, "/uploads/"))); !os.IsNotExist(err) {
		t.Fatal("rejected avatar file should be removed")
	}
	_, me = e.as(tok, http.MethodGet, "/api/v1/auth/session", nil)
	if userField(me, "avatar") != pending {
		t.Fatalf("reject changed avatar: %v", me)
	}

	// 撤回待审：只清待审那张，当前头像不变
	_, out = e.uploadAvatarAs(tok, 64, 64)
	third := userField(out, "avatarPending")
	if c, out := e.as(tok, http.MethodDelete, "/api/v1/me/avatar?pending=1", nil); c != http.StatusOK || userField(out, "avatarPending") != "" || userField(out, "avatar") != pending {
		t.Fatalf("withdraw pending: %d %v", c, out)
	}
	if _, err := os.Stat(filepath.Join(e.root, "uploads", strings.TrimPrefix(third, "/uploads/"))); !os.IsNotExist(err) {
		t.Fatal("withdrawn avatar file should be removed")
	}

	// PUT /me 不再接受外链头像
	e.as(tok, http.MethodPut, "/api/v1/me", map[string]string{"name": "读者", "avatar": "https://evil.example/x.png"})
	_, me = e.as(tok, http.MethodGet, "/api/v1/auth/session", nil)
	if userField(me, "avatar") != pending {
		t.Fatal("PUT /me must not set the avatar")
	}
}

// 站长：没有自己的头像时用身份头像；上传直接生效；删除后回到身份头像。
func TestOwnerAvatar(t *testing.T) {
	e := newEnv(t)
	e.call(http.MethodPut, "/api/v1/admin/settings", map[string]any{"about": map[string]any{"avatar": "/uploads/identity.webp"}}, nil, http.StatusOK)
	_, me := e.as(e.token, http.MethodGet, "/api/v1/auth/session", nil)
	if userField(me, "avatar") != "/uploads/identity.webp" {
		t.Fatalf("owner should default to identity avatar: %v", me)
	}
	code, out := e.uploadAvatarAs(e.token, 128, 128)
	if code != http.StatusOK || !strings.HasPrefix(userField(out, "avatar"), "/uploads/"+avatarPrefix) || userField(out, "avatarPending") != "" {
		t.Fatalf("owner upload should apply directly: %d %v", code, out)
	}
	code, out = e.as(e.token, http.MethodDelete, "/api/v1/me/avatar", nil)
	if code != http.StatusOK || userField(out, "avatar") != "/uploads/identity.webp" {
		t.Fatalf("delete should fall back to identity: %v", out)
	}
}

// 登录名：格式、唯一、每 30 天一次；邮箱：需当前密码、唯一、改后待验证；新旧登录名 / 邮箱都能登录新账号。
func TestChangeLoginAndEmail(t *testing.T) {
	e := newEnv(t)
	e.enableUsers(map[string]any{"enabled": true, "readers": map[string]any{"enabled": true, "signup": "open"}})
	tok := e.signup(map[string]any{"email": "a@example.com", "name": "甲", "password": "password123"})
	e.signup(map[string]any{"email": "b@example.com", "name": "乙", "password": "password123"})

	if c, out := e.as(tok, http.MethodPut, "/api/v1/me/login", map[string]string{"login": "1bad"}); c != http.StatusBadRequest {
		t.Fatalf("invalid login accepted: %d %v", c, out)
	}
	if c, out := e.as(tok, http.MethodPut, "/api/v1/me/login", map[string]string{"login": "admin"}); c != http.StatusConflict {
		t.Fatalf("taken login accepted: %d %v", c, out)
	}
	c, out := e.as(tok, http.MethodPut, "/api/v1/me/login", map[string]string{"login": "jia"})
	if c != http.StatusOK || userField(out, "login") != "jia" || userField(out, "loginNextChange") == "" {
		t.Fatalf("change login: %d %v", c, out)
	}
	if c, out := e.as(tok, http.MethodPut, "/api/v1/me/login", map[string]string{"login": "jia2"}); c != http.StatusTooManyRequests || out["next"] == nil {
		t.Fatalf("cooldown not enforced: %d %v", c, out)
	}

	if c, _ := e.as(tok, http.MethodPut, "/api/v1/me/email", map[string]string{"email": "new@example.com", "password": "wrong-pass"}); c != http.StatusUnauthorized {
		t.Fatalf("wrong password accepted: %d", c)
	}
	if c, _ := e.as(tok, http.MethodPut, "/api/v1/me/email", map[string]string{"email": "b@example.com", "password": "password123"}); c != http.StatusConflict {
		t.Fatalf("taken email accepted: %d", c)
	}
	c, out = e.as(tok, http.MethodPut, "/api/v1/me/email", map[string]string{"email": "New@Example.com", "password": "password123"})
	u, _ := out["user"].(map[string]any)
	if c != http.StatusOK || u["email"] != "new@example.com" || u["emailVerified"] != false {
		t.Fatalf("change email: %d %v", c, out)
	}
	for _, id := range []string{"jia", "new@example.com"} {
		if c, _ := e.as("", http.MethodPost, "/api/v1/auth/login", map[string]string{"username": id, "password": "password123"}); c != http.StatusOK {
			t.Fatalf("login with %s: %d", id, c)
		}
	}
}
