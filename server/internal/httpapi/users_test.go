package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"myself/server/internal/store"
)

// enableUsers 以管理员身份写入用户系统开关。
func (e *env) enableUsers(users map[string]any) {
	e.t.Helper()
	e.call(http.MethodPut, "/api/v1/admin/settings", map[string]any{"users": users}, nil, http.StatusOK)
}

// as 以指定令牌发 JSON 请求，返回状态码与响应体。
func (e *env) as(token, method, path string, body any) (int, map[string]any) {
	e.t.Helper()
	var rd io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		rd = bytes.NewReader(raw)
	}
	h := map[string]string{"Content-Type": "application/json"}
	if token != "" {
		h["Authorization"] = "Bearer " + token
	}
	res := e.do(method, path, rd, h)
	defer res.Body.Close()
	var out map[string]any
	data, _ := io.ReadAll(res.Body)
	_ = json.Unmarshal(data, &out)
	return res.StatusCode, out
}

// signup 注册并返回会话令牌。
func (e *env) signup(body map[string]any) string {
	e.t.Helper()
	saved := e.token
	e.token = ""
	tok := e.sessionFrom(http.MethodPost, "/api/v1/auth/register", body)
	e.token = saved
	return tok
}

func TestSignupFollowsSwitches(t *testing.T) {
	e := newEnv(t)
	reader := map[string]any{"email": "a@example.com", "name": "阿澈", "password": "reader-pass-1"}
	// 默认总开关关闭：不能注册
	if code, _ := e.as("", http.MethodPost, "/api/v1/auth/register", reader); code != http.StatusForbidden {
		t.Fatalf("signup with users off: %d", code)
	}
	e.enableUsers(map[string]any{"enabled": true, "readers": map[string]any{"enabled": true, "signup": "open"}})
	tok := e.signup(reader)
	if tok == "" {
		t.Fatal("open signup should log in")
	}
	code, sess := e.as(tok, http.MethodGet, "/api/v1/auth/session", nil)
	if code != http.StatusOK || sess["loggedIn"] != true {
		t.Fatalf("session: %d %v", code, sess)
	}
	// 读者不能进管理接口
	if code, _ := e.as(tok, http.MethodGet, "/api/v1/admin/posts", nil); code != http.StatusForbidden {
		t.Fatalf("reader on admin posts: %d", code)
	}
	// 同一邮箱不能重复注册
	if code, _ := e.as("", http.MethodPost, "/api/v1/auth/register", reader); code != http.StatusConflict {
		t.Fatalf("duplicate email: %d", code)
	}
	// 关掉读者：已登录的读者会话立即失效
	e.enableUsers(map[string]any{"readers": map[string]any{"enabled": false}})
	if code, _ := e.as(tok, http.MethodPut, "/api/v1/me", map[string]string{"name": "x"}); code != http.StatusUnauthorized {
		t.Fatalf("reader after readers off: %d", code)
	}
	// 仅邀请：公开注册被拒
	e.enableUsers(map[string]any{"readers": map[string]any{"enabled": true, "signup": "invite"}})
	if code, _ := e.as("", http.MethodPost, "/api/v1/auth/register", map[string]any{"email": "b@example.com", "name": "B", "password": "reader-pass-2"}); code != http.StatusForbidden {
		t.Fatalf("invite-only public signup: %d", code)
	}
}

func TestAuthorInviteAndScope(t *testing.T) {
	e := newEnv(t)
	e.enableUsers(map[string]any{"enabled": true, "authors": map[string]any{"enabled": true, "directPublish": false}})
	var inv struct {
		URL string `json:"url"`
	}
	e.call(http.MethodPost, "/api/v1/admin/invites", map[string]any{"role": "author", "email": "w@example.com"}, &inv, http.StatusCreated)
	code := inv.URL[strings.Index(inv.URL, "code=")+5:]
	// 邀请指定了邮箱：别的邮箱不能用
	if c, _ := e.as("", http.MethodPost, "/api/v1/auth/register", map[string]any{"email": "x@example.com", "name": "X", "password": "author-pass-1", "invite": code}); c != http.StatusForbidden {
		t.Fatalf("invite email mismatch: %d", c)
	}
	tok := e.signup(map[string]any{"email": "w@example.com", "name": "林间", "password": "author-pass-1", "invite": code})
	if tok == "" {
		t.Fatal("invited author should be logged in")
	}
	// 邀请只能用一次
	if c, _ := e.as("", http.MethodPost, "/api/v1/auth/register", map[string]any{"email": "w@example.com", "name": "X", "password": "author-pass-1", "invite": code}); c != http.StatusForbidden {
		t.Fatalf("invite reuse: %d", c)
	}
	// 作者看不到 Demo 文章，只看到自己的
	c, _ := e.as(tok, http.MethodGet, "/api/v1/admin/posts", nil)
	if c != http.StatusOK {
		t.Fatalf("author list: %d", c)
	}
	// 投稿：要求发布 + 置顶，被改成草稿、不置顶
	c, created := e.as(tok, http.MethodPost, "/api/v1/admin/posts", map[string]any{"slug": "by-author", "title": "作者稿", "contentMd": "x", "status": "published", "pinned": true})
	if c != http.StatusCreated {
		t.Fatalf("author create: %d", c)
	}
	var got store.Post
	e.call(http.MethodGet, "/api/v1/admin/posts/"+itoa(int64(created["id"].(float64))), nil, &got, http.StatusOK)
	if got.Status != "draft" || got.Pinned || got.Author == nil || got.Author.Name != "林间" {
		t.Fatalf("author post: %+v author=%+v", got, got.Author)
	}
	// 作者不能改 / 删别人的文章
	var all []store.Post
	e.call(http.MethodGet, "/api/v1/admin/posts", nil, &all, http.StatusOK)
	for _, p := range all {
		if p.Slug == "welcome-to-myself" {
			if c, _ := e.as(tok, http.MethodDelete, "/api/v1/admin/posts/"+itoa(p.ID), nil); c != http.StatusNotFound {
				t.Fatalf("author deleted others: %d", c)
			}
		}
	}
	// 作者不能进其它管理接口
	if c, _ := e.as(tok, http.MethodGet, "/api/v1/admin/settings", nil); c != http.StatusForbidden {
		t.Fatalf("author on settings: %d", c)
	}
	// 开启直接发布后可以发布
	e.enableUsers(map[string]any{"authors": map[string]any{"directPublish": true}})
	e.as(tok, http.MethodPut, "/api/v1/admin/posts/"+itoa(got.ID), map[string]any{"slug": "by-author", "title": "作者稿", "contentMd": "x", "status": "published"})
	var pub store.Post
	e.call(http.MethodGet, "/api/v1/posts/by-author", nil, &pub, http.StatusOK)
	if pub.Author == nil {
		t.Fatal("public post should carry author byline")
	}
}

func TestCommentModeration(t *testing.T) {
	e := newEnv(t)
	post := map[string]any{"target": "post", "key": "welcome-to-myself", "body": "写得真好"}
	// 评论未开放
	if c, _ := e.as("", http.MethodPost, "/api/v1/comments", post); c != http.StatusForbidden {
		t.Fatalf("comments closed: %d", c)
	}
	e.enableUsers(map[string]any{"enabled": true, "readers": map[string]any{"enabled": true, "signup": "open"},
		"comments": map[string]any{"enabled": true, "anonymous": false, "moderation": "first"}})
	if c, _ := e.as("", http.MethodPost, "/api/v1/comments", post); c != http.StatusUnauthorized {
		t.Fatalf("anonymous when not allowed: %d", c)
	}
	tok := e.signup(map[string]any{"email": "r@example.com", "name": "小满", "password": "reader-pass-3"})
	c, first := e.as(tok, http.MethodPost, "/api/v1/comments", post)
	if c != http.StatusCreated || first["pending"] != true {
		t.Fatalf("first comment should be pending: %d %v", c, first)
	}
	// 公开列表看不到待审，本人看得到
	_, pub := e.as("", http.MethodGet, "/api/v1/comments?target=post&key=welcome-to-myself", nil)
	if n := len(pub["items"].([]any)); n != 0 {
		t.Fatalf("pending visible to public: %d", n)
	}
	_, mine := e.as(tok, http.MethodGet, "/api/v1/comments?target=post&key=welcome-to-myself", nil)
	if n := len(mine["items"].([]any)); n != 1 {
		t.Fatalf("own pending not visible: %d", n)
	}
	// 管理员通过后，第二条自动公开
	e.call(http.MethodPut, "/api/v1/admin/comments/"+itoa(int64(first["id"].(float64))), map[string]string{"status": "approved"}, nil, http.StatusOK)
	_, second := e.as(tok, http.MethodPost, "/api/v1/comments", post)
	if second["pending"] != false && second["pending"] != nil {
		t.Fatalf("second comment should auto-approve: %v", second)
	}
	// 匿名：开启后可发，但一律待审
	e.enableUsers(map[string]any{"comments": map[string]any{"anonymous": true}})
	c, anon := e.as("", http.MethodPost, "/api/v1/comments", map[string]any{"target": "guestbook", "body": "路过", "guestName": "访客"})
	if c != http.StatusCreated || anon["pending"] != true {
		t.Fatalf("anonymous comment: %d %v", c, anon)
	}
	var queue []map[string]any
	e.call(http.MethodGet, "/api/v1/admin/comments?status=pending", nil, &queue, http.StatusOK)
	if len(queue) != 1 {
		t.Fatalf("pending queue: %d", len(queue))
	}
}

func TestAdminUserManagement(t *testing.T) {
	e := newEnv(t)
	e.enableUsers(map[string]any{"enabled": true, "readers": map[string]any{"enabled": true, "signup": "open"}})
	tok := e.signup(map[string]any{"email": "=cmd@example.com", "name": "=HYPERLINK(1)", "password": "reader-pass-4"})
	var list struct {
		Items []adminUser      `json:"items"`
		Stats map[string]int `json:"stats"`
	}
	e.call(http.MethodGet, "/api/v1/admin/users", nil, &list, http.StatusOK)
	if len(list.Items) != 2 || list.Stats["total"] != 2 {
		t.Fatalf("users: %+v", list)
	}
	var reader adminUser
	for _, u := range list.Items {
		if u.Role == "reader" {
			reader = u
		}
		if u.Role == "admin" && !u.Self {
			t.Fatal("admin row should be marked self")
		}
	}
	// 管理员不能停用自己
	for _, u := range list.Items {
		if u.Self {
			e.call(http.MethodPut, "/api/v1/admin/users/"+itoa(u.ID), map[string]string{"status": "disabled"}, nil, http.StatusBadRequest)
		}
	}
	// 停用：会话立即失效，也不能再登录
	e.call(http.MethodPut, "/api/v1/admin/users/"+itoa(reader.ID), map[string]string{"status": "disabled"}, nil, http.StatusOK)
	if c, _ := e.as(tok, http.MethodPut, "/api/v1/me", map[string]string{"name": "y"}); c != http.StatusUnauthorized {
		t.Fatalf("disabled session still valid: %d", c)
	}
	if c, _ := e.as("", http.MethodPost, "/api/v1/auth/login", map[string]string{"username": "=cmd@example.com", "password": "reader-pass-4"}); c != http.StatusForbidden {
		t.Fatalf("disabled login: %d", c)
	}
	// 重置链接：恢复后用新密码登录
	e.call(http.MethodPut, "/api/v1/admin/users/"+itoa(reader.ID), map[string]string{"status": "active"}, nil, http.StatusOK)
	var link struct {
		URL string `json:"url"`
	}
	e.call(http.MethodPost, "/api/v1/admin/users/"+itoa(reader.ID)+"/reset", map[string]any{}, &link, http.StatusOK)
	token := link.URL[strings.Index(link.URL, "token=")+6:]
	if c, _ := e.as("", http.MethodPost, "/api/v1/auth/reset", map[string]string{"token": token, "password": "brand-new-pass-5"}); c != http.StatusOK {
		t.Fatalf("reset: %d", c)
	}
	if c, _ := e.as("", http.MethodPost, "/api/v1/auth/reset", map[string]string{"token": token, "password": "again-pass-6"}); c != http.StatusBadRequest {
		t.Fatalf("reset token reuse: %d", c)
	}
	// 导出：公式注入被转义
	res := e.do(http.MethodGet, "/api/v1/admin/users/export", nil, map[string]string{"Authorization": "Bearer " + e.token})
	data, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(data), "'=HYPERLINK") {
		t.Fatalf("csv not escaped: %s", data)
	}
}
