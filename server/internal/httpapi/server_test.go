package httpapi

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/config"
	"myself/server/internal/store"
)

type env struct {
	t     *testing.T
	srv   *httptest.Server
	token string
	root  string
	auth  *auth.Service
	// server 直接查库断言用
	server *Server
}

func newEnv(t *testing.T) *env {
	t.Helper()
	root := t.TempDir()
	db, err := store.Open(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	a := auth.New(db)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	s := New(Deps{
		DB: db, Auth: a, Config: config.New(db),
		UploadDir: filepath.Join(root, "uploads"),
		BackupDir: filepath.Join(root, "backups"),
		DataDir:   filepath.Join(root, "data"),
		// 内嵌默认封面的替身：初始化选 Demo 时从这里导入素材库
		DemoCovers: demoCoversFS(),
	})
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	e := &env{t: t, srv: ts, root: root, auth: a, server: s}
	// 首次启动初始化：初始化码 + 管理员 + Demo 数据；成功即写入会话 Cookie
	if tok := e.sessionFrom(http.MethodPost, "/api/v1/setup", map[string]any{
		"code": a.SetupCode(), "username": "admin", "password": testPassword, "demo": true,
		"site": map[string]any{"title": "Myself"},
	}); tok == "" {
		t.Fatal("setup did not set a session cookie")
	}
	e.token = e.sessionFrom(http.MethodPost, "/api/v1/auth/login", map[string]string{"username": "admin", "password": testPassword})
	if e.token == "" {
		t.Fatal("login did not set a session cookie")
	}
	return e
}

// sessionFrom 发 JSON 请求，断言 200，返回 Set-Cookie 里的会话令牌；并确认响应体不含令牌。
func (e *env) sessionFrom(method, path string, body any) string {
	e.t.Helper()
	raw, _ := json.Marshal(body)
	headers := map[string]string{"Content-Type": "application/json"}
	if e.token != "" {
		headers["Authorization"] = "Bearer " + e.token
	}
	res := e.do(method, path, bytes.NewReader(raw), headers)
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		e.t.Fatalf("%s %s: %d %s", method, path, res.StatusCode, data)
	}
	if strings.Contains(string(data), "token") {
		e.t.Fatalf("%s %s leaks token in body: %s", method, path, data)
	}
	for _, c := range res.Cookies() {
		if c.Name == "myself_session" {
			if !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
				e.t.Fatalf("session cookie not HttpOnly/Strict: %+v", c)
			}
			return c.Value
		}
	}
	return ""
}

const testPassword = "correct-horse-9"

func (e *env) do(method, path string, body io.Reader, headers map[string]string) *http.Response {
	e.t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, body)
	if err != nil {
		e.t.Fatal(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	return res
}

// call 发 JSON 请求（管理员 token 自动附带），校验状态码并解析响应。
func (e *env) call(method, path string, body any, out any, wantStatus int) {
	e.t.Helper()
	var rd io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		rd = bytes.NewReader(raw)
	}
	headers := map[string]string{"Content-Type": "application/json"}
	if e.token != "" {
		headers["Authorization"] = "Bearer " + e.token
	}
	res := e.do(method, path, rd, headers)
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode != wantStatus {
		e.t.Fatalf("%s %s: status %d, want %d, body %s", method, path, res.StatusCode, wantStatus, data)
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			e.t.Fatalf("%s %s: bad json %s: %v", method, path, data, err)
		}
	}
}

func TestPublicEndpoints(t *testing.T) {
	e := newEnv(t)
	var list pagedPosts
	e.call(http.MethodGet, "/api/v1/posts?pageSize=2", nil, &list, http.StatusOK)
	if list.Total != 6 || len(list.Items) != 2 {
		t.Fatalf("posts: total=%d items=%d", list.Total, len(list.Items))
	}
	var tags []store.TagCount
	e.call(http.MethodGet, "/api/v1/tags", nil, &tags, http.StatusOK)
	if len(tags) == 0 || tags[0].Name != "指南" || tags[0].Count != 4 {
		t.Fatalf("tags: %+v", tags)
	}
	var hero struct {
		IntervalMs int          `json:"intervalMs"`
		Items      []store.Post `json:"items"`
	}
	e.call(http.MethodGet, "/api/v1/hero", nil, &hero, http.StatusOK)
	if hero.IntervalMs != 3000 || len(hero.Items) != 4 {
		t.Fatalf("hero: %+v", hero)
	}
	e.call(http.MethodGet, "/api/v1/posts/nope", nil, nil, http.StatusNotFound)
	res := e.do(http.MethodGet, "/api/v1/admin/posts", nil, nil)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("admin without token: %d", res.StatusCode)
	}
}

func TestPostLifecycle(t *testing.T) {
	e := newEnv(t)
	var created struct {
		ID int64 `json:"id"`
	}
	draft := map[string]any{"slug": "t-1", "title": " T ", "contentMd": "# x", "tags": []string{"a"}, "status": "draft", "pinned": true}
	e.call(http.MethodPost, "/api/v1/admin/posts", draft, &created, http.StatusCreated)
	e.call(http.MethodPost, "/api/v1/admin/posts", draft, nil, http.StatusConflict)
	e.call(http.MethodGet, "/api/v1/posts/t-1", nil, nil, http.StatusNotFound) // 草稿不公开

	var got store.Post
	e.call(http.MethodGet, "/api/v1/admin/posts/"+itoa(created.ID), nil, &got, http.StatusOK)
	if got.Title != "T" || !got.Pinned || got.Status != "draft" || *got.ContentMd != "# x" {
		t.Fatalf("admin get: %+v", got)
	}
	draft["status"] = "published"
	e.call(http.MethodPut, "/api/v1/admin/posts/"+itoa(created.ID), draft, nil, http.StatusOK)
	var public store.Post
	e.call(http.MethodGet, "/api/v1/posts/t-1", nil, &public, http.StatusOK)
	if public.Status != "" || public.ContentMd == nil {
		t.Fatalf("public get leaks status or lacks content: %+v", public)
	}
	e.call(http.MethodDelete, "/api/v1/admin/posts/"+itoa(created.ID), nil, nil, http.StatusOK)
	e.call(http.MethodDelete, "/api/v1/admin/posts/"+itoa(created.ID), nil, nil, http.StatusNotFound)
}

func TestExternalChannel(t *testing.T) {
	e := newEnv(t)
	var key struct {
		ID  int64  `json:"id"`
		Key string `json:"key"`
	}
	e.call(http.MethodPost, "/api/v1/admin/apikeys", map[string]string{"name": "bot", "scope": "full"}, &key, http.StatusCreated)
	if !strings.HasPrefix(key.Key, "myk_") {
		t.Fatalf("key: %q", key.Key)
	}
	hdr := map[string]string{"Content-Type": "application/json", "X-Api-Key": key.Key}
	body := `{"slug":"ext-1","title":"E","contentMd":"c","covers":["1","2","3","4"]}`
	res := e.do(http.MethodPost, "/api/v1/ext/posts", strings.NewReader(body), hdr)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("ext create: %d", res.StatusCode)
	}
	res = e.do(http.MethodGet, "/api/v1/ext/posts?status=draft", nil, hdr)
	var posts []store.Post
	json.NewDecoder(res.Body).Decode(&posts)
	if len(posts) != 1 || len(posts[0].Covers) != 3 || posts[0].Status != "draft" {
		t.Fatalf("ext list: %+v", posts)
	}
	res = e.do(http.MethodGet, "/api/v1/ext/posts", nil, map[string]string{"X-Api-Key": "bad"})
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("bad key: %d", res.StatusCode)
	}
	e.call(http.MethodDelete, "/api/v1/admin/apikeys/"+itoa(key.ID), nil, nil, http.StatusOK)
	res = e.do(http.MethodGet, "/api/v1/ext/posts", nil, hdr)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("revoked key still works: %d", res.StatusCode)
	}
}

func TestMediaUploadCropThumb(t *testing.T) {
	e := newEnv(t)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, _ := mw.CreateFormFile("files", "pic.png")
	img := image.NewNRGBA(image.Rect(0, 0, 1200, 800))
	for i := range img.Pix {
		img.Pix[i] = 200
	}
	img.Set(0, 0, color.NRGBA{255, 0, 0, 128})
	png.Encode(part, img)
	txt, _ := mw.CreateFormFile("files", "x.txt")
	txt.Write([]byte("nope"))
	mw.Close()

	res := e.do(http.MethodPost, "/api/v1/admin/media", &buf, map[string]string{
		"Content-Type": mw.FormDataContentType(), "Authorization": "Bearer " + e.token,
	})
	var items []mediaItem
	json.NewDecoder(res.Body).Decode(&items)
	if res.StatusCode != http.StatusCreated || len(items) != 1 || !strings.HasSuffix(items[0].Name, ".png") {
		t.Fatalf("upload: %d %+v", res.StatusCode, items)
	}
	name := items[0].Name

	res = e.do(http.MethodGet, items[0].Thumb, nil, nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("thumb: %d", res.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(e.root, "uploads", "thumbs", name+".webp")); err != nil {
		t.Fatal("thumb file not generated")
	}

	var cropped mediaItem
	e.call(http.MethodPost, "/api/v1/admin/media/"+name+"/crop", map[string]int{"left": 100, "top": 50, "width": 9999, "height": 300}, &cropped, http.StatusOK)
	if cropped.Crop == nil || cropped.Crop.Width != 1100 || cropped.Crop.Height != 300 || !cropped.HasOriginal {
		t.Fatalf("crop: %+v", cropped)
	}
	f, _ := os.Open(filepath.Join(e.root, "uploads", name))
	cfg, err := png.DecodeConfig(f)
	f.Close()
	if err != nil || cfg.Width != 1100 || cfg.Height != 300 {
		t.Fatalf("cropped file: %+v %v", cfg, err)
	}

	// 后台压缩任务：202 → 轮询到完成，png 变 webp
	var accepted struct {
		ID    string `json:"id"`
		Total int    `json:"total"`
	}
	e.call(http.MethodPost, "/api/v1/admin/quality/compress", map[string]any{"names": []string{name, "../x.png"}, "quality": 70}, &accepted, http.StatusAccepted)
	if accepted.Total != 1 || accepted.ID == "" {
		t.Fatalf("compress accepted: %+v", accepted)
	}
	var job compressJob
	for i := 0; i < 200; i++ {
		e.call(http.MethodGet, "/api/v1/admin/quality/jobs/"+accepted.ID, nil, &job, http.StatusOK)
		if !job.Running {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if job.Running || job.Done != 1 || len(job.Results) != 1 || !strings.HasSuffix(job.Results[0].NewName, ".webp") {
		t.Fatalf("compress job: %+v", job)
	}
	if fileExists(filepath.Join(e.root, "uploads", name)) {
		t.Fatal("png should be replaced by webp")
	}
	name = job.Results[0].NewName
	e.call(http.MethodGet, "/api/v1/admin/quality/jobs/nope", nil, nil, http.StatusNotFound)

	res = e.do(http.MethodGet, "/uploads/.originals/"+name, nil, nil)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("dot dir must be hidden: %d", res.StatusCode)
	}
	e.call(http.MethodDelete, "/api/v1/admin/media/"+name, nil, nil, http.StatusOK)
	if _, err := os.Stat(filepath.Join(e.root, "uploads", "thumbs", name+".webp")); err == nil {
		t.Fatal("thumb should be removed with source")
	}
}

func TestSettingsAndSiteConfig(t *testing.T) {
	e := newEnv(t)
	var merged config.Map
	e.call(http.MethodPut, "/api/v1/admin/settings", map[string]any{
		"site": map[string]any{"title": "X"}, "github": map[string]any{"token": "secret"},
	}, &merged, http.StatusOK)
	if config.Sub(merged, "site")["title"] != "X" || config.Sub(merged, "site")["subtitle"] != "个人博客" {
		t.Fatalf("deep merge broke: %+v", merged["site"])
	}
	var public config.Map
	e.call(http.MethodGet, "/api/v1/site-config", nil, &public, http.StatusOK)
	if _, leaked := config.Sub(public, "github")["token"]; leaked {
		t.Fatal("token leaked into public config")
	}
}

func TestFeedAndBackup(t *testing.T) {
	e := newEnv(t)
	res := e.do(http.MethodGet, "/feed", nil, nil)
	if res.StatusCode != http.StatusOK || !strings.HasPrefix(res.Header.Get("Content-Type"), "application/rss+xml") {
		t.Fatalf("feed: %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
	var doc struct {
		Channel struct {
			Title string `xml:"title"`
			Items []struct {
				Title string `xml:"title"`
				Link  string `xml:"link"`
			} `xml:"item"`
		} `xml:"channel"`
	}
	if err := xml.NewDecoder(res.Body).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	if doc.Channel.Title != "Myself" || len(doc.Channel.Items) != 6 || !strings.Contains(doc.Channel.Items[0].Link, "/articles/") {
		t.Fatalf("feed content: %+v", doc.Channel)
	}

	var b struct {
		Name string `json:"name"`
	}
	e.call(http.MethodPost, "/api/v1/admin/backups", nil, &b, http.StatusCreated)
	var list []backupInfo
	e.call(http.MethodGet, "/api/v1/admin/backups", nil, &list, http.StatusOK)
	if len(list) != 1 || list[0].Name != b.Name || list[0].Size == 0 {
		t.Fatalf("backups: %+v", list)
	}
	e.call(http.MethodGet, "/api/v1/admin/backups/evil.zip", nil, nil, http.StatusBadRequest)
	e.call(http.MethodDelete, "/api/v1/admin/backups/"+b.Name, nil, nil, http.StatusOK)
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

func TestSetupOnlyOnce(t *testing.T) {
	e := newEnv(t)
	var st struct {
		NeedsSetup bool `json:"needsSetup"`
	}
	e.call(http.MethodGet, "/api/v1/setup", nil, &st, http.StatusOK)
	if st.NeedsSetup {
		t.Fatal("still needs setup after setup")
	}
	e.call(http.MethodPost, "/api/v1/setup", map[string]any{"code": "X", "username": "evil", "password": "whatever-123"}, nil, http.StatusConflict)
}

func TestSetupRequiresCode(t *testing.T) {
	root := t.TempDir()
	db, err := store.Open(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	a := auth.New(db)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	s := New(Deps{DB: db, Auth: a, Config: config.New(db), UploadDir: filepath.Join(root, "u"), BackupDir: filepath.Join(root, "b"), DataDir: filepath.Join(root, "data")})
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	e := &env{t: t, srv: ts, root: root}
	e.call(http.MethodPost, "/api/v1/setup/verify", map[string]string{"code": a.SetupCode()}, nil, http.StatusOK)
	body := map[string]any{"code": "WRONG000", "username": "admin", "password": "long-enough-1"}
	for i := 0; i < 5; i++ {
		e.call(http.MethodPost, "/api/v1/setup", body, nil, http.StatusForbidden)
	}
	// 连续失败后按 IP 锁定，正确的码也要等锁定结束
	body["code"] = a.SetupCode()
	e.call(http.MethodPost, "/api/v1/setup", body, nil, http.StatusTooManyRequests)
	var public config.Map
	e.call(http.MethodGet, "/api/v1/site-config", nil, &public, http.StatusOK)
	if public["needsSetup"] != true {
		t.Fatal("site-config should report needsSetup")
	}
}

func TestLoginLockout(t *testing.T) {
	e := newEnv(t)
	e.token = ""
	for i := 0; i < 5; i++ {
		e.call(http.MethodPost, "/api/v1/auth/login", map[string]string{"username": "admin", "password": "nope-nope"}, nil, http.StatusUnauthorized)
	}
	e.call(http.MethodPost, "/api/v1/auth/login", map[string]string{"username": "admin", "password": testPassword}, nil, http.StatusTooManyRequests)
}

func TestPasswordChangeRevokesTokens(t *testing.T) {
	e := newEnv(t)
	old := e.token
	fresh := e.sessionFrom(http.MethodPut, "/api/v1/auth/password", map[string]string{"oldPassword": testPassword, "newPassword": "another-pass-2"})
	if fresh == "" || fresh == old {
		t.Fatal("password change should set a fresh session cookie")
	}
	e.call(http.MethodGet, "/api/v1/admin/settings", nil, nil, http.StatusUnauthorized) // 旧令牌失效
	e.token = fresh
	e.call(http.MethodGet, "/api/v1/admin/settings", nil, nil, http.StatusOK)
}

func TestCookieSessionCSRF(t *testing.T) {
	e := newEnv(t)
	cookie := map[string]string{"Cookie": "myself_session=" + e.token, "Content-Type": "application/json"}
	// 读操作只凭 Cookie 即可
	if res := e.do(http.MethodGet, "/api/v1/admin/settings", nil, cookie); res.StatusCode != http.StatusOK {
		t.Fatalf("cookie GET: %d", res.StatusCode)
	}
	// 写操作缺自定义头：拒绝
	if res := e.do(http.MethodPut, "/api/v1/admin/settings", strings.NewReader(`{}`), cookie); res.StatusCode != http.StatusForbidden {
		t.Fatalf("cookie PUT without header: %d", res.StatusCode)
	}
	// 跨站 Origin：拒绝
	cross := map[string]string{"Cookie": cookie["Cookie"], "Content-Type": "application/json", "X-Requested-With": "myself", "Origin": "https://evil.example"}
	if res := e.do(http.MethodPut, "/api/v1/admin/settings", strings.NewReader(`{}`), cross); res.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin PUT: %d", res.StatusCode)
	}
	// 同源 + 自定义头：放行
	ok := map[string]string{"Cookie": cookie["Cookie"], "Content-Type": "application/json", "X-Requested-With": "myself"}
	if res := e.do(http.MethodPut, "/api/v1/admin/settings", strings.NewReader(`{}`), ok); res.StatusCode != http.StatusOK {
		t.Fatalf("same-origin PUT: %d", res.StatusCode)
	}
	// 注销清除 Cookie
	res := e.do(http.MethodPost, "/api/v1/auth/logout", nil, nil)
	var cleared bool
	for _, c := range res.Cookies() {
		cleared = cleared || (c.Name == "myself_session" && c.MaxAge < 0)
	}
	if !cleared {
		t.Fatal("logout should clear the session cookie")
	}
}

func TestContribKeyScope(t *testing.T) {
	e := newEnv(t)
	var key struct {
		Key   string `json:"key"`
		Scope string `json:"scope"`
	}
	e.call(http.MethodPost, "/api/v1/admin/apikeys", map[string]string{"name": "writer"}, &key, http.StatusCreated)
	if key.Scope != "contrib" {
		t.Fatalf("default scope: %q", key.Scope)
	}
	hdr := map[string]string{"Content-Type": "application/json", "X-Api-Key": key.Key}
	// 试图直接发布：被强制为草稿
	res := e.do(http.MethodPost, "/api/v1/ext/posts", strings.NewReader(`{"slug":"c-1","title":"C","contentMd":"x","status":"published"}`), hdr)
	var created struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	json.NewDecoder(res.Body).Decode(&created)
	if res.StatusCode != http.StatusCreated || created.Status != "draft" {
		t.Fatalf("contrib create: %d %+v", res.StatusCode, created)
	}
	// 只能看到自己的草稿，看不到 Demo 文章
	res = e.do(http.MethodGet, "/api/v1/ext/posts", nil, hdr)
	var posts []store.Post
	json.NewDecoder(res.Body).Decode(&posts)
	if len(posts) != 1 {
		t.Fatalf("contrib sees %d posts", len(posts))
	}
	// 不能改、删别人的文章
	var all []store.Post
	e.call(http.MethodGet, "/api/v1/admin/posts", nil, &all, http.StatusOK)
	for _, p := range all {
		if p.Slug == "welcome-to-myself" {
			res = e.do(http.MethodDelete, "/api/v1/ext/posts/"+itoa(p.ID), nil, hdr)
			if res.StatusCode != http.StatusNotFound {
				t.Fatalf("contrib deleted others: %d", res.StatusCode)
			}
		}
	}
	// 随想只开放给全托管
	res = e.do(http.MethodPost, "/api/v1/ext/notes", strings.NewReader(`{"contentMd":"hi"}`), hdr)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("contrib note: %d", res.StatusCode)
	}
}

func TestSecurityHeadersAndUploads(t *testing.T) {
	e := newEnv(t)
	res := e.do(http.MethodGet, "/api/v1/site-config", nil, nil)
	if res.Header.Get("X-Content-Type-Options") != "nosniff" || res.Header.Get("X-Frame-Options") != "DENY" {
		t.Fatalf("missing security headers: %v", res.Header)
	}
	for _, p := range []string{"/uploads/", "/uploads/thumbs/", "/uploads/.originals/x.png", "/uploads/a/b.png", "/uploads/x.html"} {
		if res := e.do(http.MethodGet, p, nil, nil); res.StatusCode != http.StatusNotFound {
			t.Fatalf("%s: %d", p, res.StatusCode)
		}
	}
}

func TestUploadRejectsFakeImage(t *testing.T) {
	e := newEnv(t)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, _ := mw.CreateFormFile("files", "evil.png")
	part.Write([]byte("<html><script>alert(1)</script></html>"))
	mw.Close()
	res := e.do(http.MethodPost, "/api/v1/admin/media", &buf, map[string]string{"Content-Type": mw.FormDataContentType(), "Authorization": "Bearer " + e.token})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("fake image accepted: %d", res.StatusCode)
	}
}

// demoCoversFS 12 张默认封面的替身（内容无关紧要，只校验导入与地址改写）。
func demoCoversFS() fstest.MapFS {
	fsys := fstest.MapFS{}
	for i := 1; i <= 12; i++ {
		fsys[fmt.Sprintf("%02d.webp", i)] = &fstest.MapFile{Data: []byte("RIFF-demo")}
	}
	return fsys
}

// 初始化选 Demo：示例用到的默认封面进素材库，文章 / 随想 / 关于页模块都改用素材库地址。
func TestDemoCoversImportedIntoMedia(t *testing.T) {
	e := newEnv(t)
	var media []struct{ URL string }
	e.call(http.MethodGet, "/api/v1/admin/media", nil, &media, http.StatusOK)
	inLib := map[string]bool{}
	for _, m := range media {
		inLib[m.URL] = true
	}
	if len(inLib) == 0 {
		t.Fatal("demo covers were not imported into the media library")
	}
	for url := range inLib {
		if !strings.HasPrefix(url, "/uploads/demo-cover-") {
			t.Fatalf("unexpected media %q", url)
		}
	}

	var posts []struct{ Covers []string }
	e.call(http.MethodGet, "/api/v1/admin/posts", nil, &posts, http.StatusOK)
	var notes struct{ Items []struct{ Images []string } }
	e.call(http.MethodGet, "/api/v1/notes?pageSize=50", nil, &notes, http.StatusOK)
	var refs []string
	for _, p := range posts {
		refs = append(refs, p.Covers...)
	}
	for _, n := range notes.Items {
		refs = append(refs, n.Images...)
	}
	if len(refs) == 0 {
		t.Fatal("demo content has no images")
	}
	for _, u := range refs {
		if !inLib[u] {
			t.Fatalf("demo image %q is not in the media library", u)
		}
	}

	var cfg struct{ About struct{ Modules []any } }
	e.call(http.MethodGet, "/api/v1/site-config", nil, &cfg, http.StatusOK)
	raw, _ := json.Marshal(cfg.About.Modules)
	if strings.Contains(string(raw), "/covers/") {
		t.Fatal("demo about modules still reference /covers/")
	}
}
