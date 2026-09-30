package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"myself/server/internal/store"
)

type userPage struct {
	Items          []adminUser    `json:"items"`
	Page           int            `json:"page"`
	PageSize       int            `json:"pageSize"`
	Total          int            `json:"total"`
	Counts         map[string]int `json:"counts"`
	PendingAvatars []adminUser    `json:"pendingAvatars"`
	Stats          map[string]int `json:"stats"`
}

func TestAdminUsersPagination(t *testing.T) {
	e := newEnv(t)
	db := e.server.DB
	// 1 位站长 + 3 位作者 + 20 位读者（其中 2 位停用，1 位有待审头像）
	for i := 0; i < 23; i++ {
		role, status := store.RoleReader, store.StatusActive
		if i < 3 {
			role = store.RoleAuthor
		}
		if i == 10 || i == 11 {
			status = store.StatusDisabled
		}
		if _, err := db.CreateUser(store.NewUser{
			Login: fmt.Sprintf("user%02d", i), Email: fmt.Sprintf("u%02d@example.com", i),
			Name: fmt.Sprintf("用户%02d", i), Role: role, Status: status,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`UPDATE users SET avatar_pending = '/uploads/avatars/x.webp' WHERE login = 'user15'`); err != nil {
		t.Fatal(err)
	}

	// 兼容：不带 page 返回全部
	var legacy userPage
	e.call(http.MethodGet, "/api/v1/admin/users", nil, &legacy, http.StatusOK)
	if len(legacy.Items) != 24 || legacy.Total != 24 || legacy.Stats["total"] != 24 {
		t.Fatalf("legacy list: %d items, total %d", len(legacy.Items), legacy.Total)
	}

	var p1 userPage
	e.call(http.MethodGet, "/api/v1/admin/users?page=1&pageSize=10", nil, &p1, http.StatusOK)
	if p1.Total != 24 || len(p1.Items) != 10 || p1.PageSize != 10 || p1.Items[0].Role != "admin" {
		t.Fatalf("page 1: total %d items %d", p1.Total, len(p1.Items))
	}
	want := map[string]int{"all": 24, "admin": 1, "author": 3, "reader": 20, "disabled": 2}
	for k, v := range want {
		if p1.Counts[k] != v {
			t.Fatalf("counts[%s] = %d, want %d", k, p1.Counts[k], v)
		}
	}
	if len(p1.PendingAvatars) != 1 || p1.PendingAvatars[0].Login != "user15" {
		t.Fatalf("pending avatars: %+v", p1.PendingAvatars)
	}
	var p3 userPage
	e.call(http.MethodGet, "/api/v1/admin/users?page=3&pageSize=10", nil, &p3, http.StatusOK)
	if len(p3.Items) != 4 || p3.Items[3].Login != "user22" {
		t.Fatalf("page 3: %+v", p3.Items)
	}

	// 角色分段
	var authors userPage
	e.call(http.MethodGet, "/api/v1/admin/users?page=1&pageSize=10&role=author", nil, &authors, http.StatusOK)
	if authors.Total != 3 || len(authors.Items) != 3 {
		t.Fatalf("authors: %+v", authors)
	}
	var disabled userPage
	e.call(http.MethodGet, "/api/v1/admin/users?page=1&role=disabled", nil, &disabled, http.StatusOK)
	if disabled.Total != 2 || disabled.Items[0].Status != "disabled" {
		t.Fatalf("disabled: %+v", disabled)
	}

	// 搜索：名字 / 账号 / 邮箱，分段计数随搜索收窄；LIKE 通配符按字面匹配
	var found userPage
	e.call(http.MethodGet, "/api/v1/admin/users?page=1&q=u1", nil, &found, http.StatusOK)
	if found.Total != 10 || found.Counts["disabled"] != 2 || found.Counts["author"] != 0 {
		t.Fatalf("search u1: total %d counts %v", found.Total, found.Counts)
	}
	e.call(http.MethodGet, "/api/v1/admin/users?page=1&q="+url.QueryEscape("用户2"), nil, &found, http.StatusOK)
	if found.Total != 3 {
		t.Fatalf("search by name: %d", found.Total)
	}
	e.call(http.MethodGet, "/api/v1/admin/users?page=1&q=%25", nil, &found, http.StatusOK)
	if found.Total != 0 {
		t.Fatalf("wildcard search: %d", found.Total)
	}
	e.call(http.MethodGet, "/api/v1/admin/users?page=1&role=reader&q=user2", nil, &found, http.StatusOK)
	if found.Total != 3 || found.Counts["all"] != 3 {
		t.Fatalf("reader + search: %+v", found)
	}
}

func TestAdminCommentsPagination(t *testing.T) {
	e := newEnv(t)
	var postID int64
	if err := e.server.DB.QueryRow(`SELECT id FROM posts WHERE slug = 'welcome-to-myself'`).Scan(&postID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 25; i++ {
		status := "pending"
		if i >= 20 {
			status = "spam"
		}
		if _, err := e.server.DB.Exec(`INSERT INTO comments (target, target_id, guest_name, body, status) VALUES ('post', ?, '访客', ?, ?)`,
			postID, fmt.Sprintf("评论 %d", i), status); err != nil {
			t.Fatal(err)
		}
	}
	var page struct {
		Items    []adminComment `json:"items"`
		Page     int            `json:"page"`
		PageSize int            `json:"pageSize"`
		Total    int            `json:"total"`
		Counts   map[string]int `json:"counts"`
	}
	e.call(http.MethodGet, "/api/v1/admin/comments?status=pending&page=2&pageSize=8", nil, &page, http.StatusOK)
	if page.Total != 20 || len(page.Items) != 8 || page.Page != 2 || page.Counts["spam"] != 5 || page.Counts["approved"] == 0 {
		t.Fatalf("comments page: total %d items %d counts %v", page.Total, len(page.Items), page.Counts)
	}
	if page.Items[0].Body != "评论 11" {
		t.Fatalf("page 2 starts at %q", page.Items[0].Body)
	}
	// 兼容：不带 page 仍回数组
	var legacy []adminComment
	e.call(http.MethodGet, "/api/v1/admin/comments?status=spam", nil, &legacy, http.StatusOK)
	if len(legacy) != 5 {
		t.Fatalf("legacy comments: %d", len(legacy))
	}
}
