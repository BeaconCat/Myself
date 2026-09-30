package httpapi

import (
	"net/http"
	"testing"
)

// 隐藏：文章 / 随想从前台列表、详情、标签中消失，后台仍可见；取消隐藏恢复；批量删除。
func TestHideAndBatch(t *testing.T) {
	e := newEnv(t)
	var post, note struct {
		ID int64 `json:"id"`
	}
	e.call(http.MethodPost, "/api/v1/admin/posts", map[string]any{"slug": "h-1", "title": "H", "contentMd": "x", "tags": []string{"only-hidden"}, "status": "published"}, &post, http.StatusCreated)
	e.call(http.MethodPost, "/api/v1/admin/notes", map[string]any{"contentMd": "secret"}, &note, http.StatusCreated)

	e.call(http.MethodPost, "/api/v1/admin/posts/batch", map[string]any{"ids": []int64{post.ID}, "action": "hide"}, nil, http.StatusOK)
	e.call(http.MethodPost, "/api/v1/admin/notes/batch", map[string]any{"ids": []int64{note.ID}, "action": "hide"}, nil, http.StatusOK)

	e.call(http.MethodGet, "/api/v1/posts/h-1", nil, nil, http.StatusNotFound)
	var list struct {
		Items []struct {
			Slug string `json:"slug"`
		} `json:"items"`
	}
	e.call(http.MethodGet, "/api/v1/posts?pageSize=50", nil, &list, http.StatusOK)
	for _, p := range list.Items {
		if p.Slug == "h-1" {
			t.Fatal("hidden post listed")
		}
	}
	var tags []map[string]any
	e.call(http.MethodGet, "/api/v1/tags", nil, &tags, http.StatusOK)
	for _, tg := range tags {
		if tg["name"] == "only-hidden" {
			t.Fatal("hidden post's tag listed")
		}
	}
	// 公开接口（不带会话）看不到隐藏随想；站长带 all=1 能看到
	res := e.do(http.MethodGet, "/api/v1/notes/"+itoa(note.ID), nil, map[string]string{})
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("hidden note public: %d", res.StatusCode)
	}
	var notes struct {
		Items []struct {
			ID     int64 `json:"id"`
			Hidden bool  `json:"hidden"`
		} `json:"items"`
	}
	e.call(http.MethodGet, "/api/v1/notes?all=1", nil, &notes, http.StatusOK)
	found := false
	for _, n := range notes.Items {
		found = found || (n.ID == note.ID && n.Hidden)
	}
	if !found {
		t.Fatalf("admin list missing hidden note: %+v", notes.Items)
	}
	var admin []map[string]any
	e.call(http.MethodGet, "/api/v1/admin/posts", nil, &admin, http.StatusOK)
	hiddenSeen := false
	for _, p := range admin {
		hiddenSeen = hiddenSeen || (p["slug"] == "h-1" && p["hidden"] == true)
	}
	if !hiddenSeen {
		t.Fatalf("admin posts: %+v", admin)
	}

	e.call(http.MethodPost, "/api/v1/admin/posts/batch", map[string]any{"ids": []int64{post.ID}, "action": "show"}, nil, http.StatusOK)
	e.call(http.MethodGet, "/api/v1/posts/h-1", nil, nil, http.StatusOK)

	var del map[string]int
	e.call(http.MethodPost, "/api/v1/admin/notes/batch", map[string]any{"ids": []int64{note.ID}, "action": "delete"}, &del, http.StatusOK)
	if del["affected"] != 1 {
		t.Fatalf("batch delete: %v", del)
	}
	e.call(http.MethodPost, "/api/v1/admin/notes/batch", map[string]any{"ids": []int64{note.ID}, "action": "nuke"}, nil, http.StatusBadRequest)
}
