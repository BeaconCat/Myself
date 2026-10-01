package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"myself/server/internal/store"
)

type apiLogPage struct {
	Items    []store.APILog `json:"items"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
	Total    int            `json:"total"`
	Counts   map[string]int `json:"counts"`
}

func TestAPICallLogs(t *testing.T) {
	e := newEnv(t)
	var key struct {
		ID  int64  `json:"id"`
		Key string `json:"key"`
	}
	e.call(http.MethodPost, "/api/v1/admin/apikeys", map[string]string{"name": "写作助手", "scope": "full"}, &key, http.StatusCreated)
	hdr := map[string]string{"Content-Type": "application/json", "X-Api-Key": key.Key, "User-Agent": "agent/1.0"}

	// 成功：201 与 200 各一条
	res := e.do(http.MethodPost, "/api/v1/ext/posts", strings.NewReader(`{"slug":"log-1","title":"L","contentMd":"secret body"}`), hdr)
	res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("ext create: %d", res.StatusCode)
	}
	res = e.do(http.MethodGet, "/api/v1/ext/posts?status=draft", nil, hdr)
	res.Body.Close()
	// 失败：无效 Key、缺失 Key、有效 Key 的 404
	res = e.do(http.MethodGet, "/api/v1/ext/posts", nil, map[string]string{"X-Api-Key": "myk_deadbeefcafe0000"})
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("invalid key: %d", res.StatusCode)
	}
	res = e.do(http.MethodGet, "/api/v1/ext/notes", nil, nil)
	res.Body.Close()
	res = e.do(http.MethodGet, "/api/v1/ext/posts/999999", nil, hdr)
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("missing post: %d", res.StatusCode)
	}

	var all apiLogPage
	e.call(http.MethodGet, "/api/v1/admin/api-logs", nil, &all, http.StatusOK)
	if all.Total != 5 || len(all.Items) != 5 || all.Counts["all"] != 5 || all.Counts["error"] != 3 || all.Counts["ok"] != 2 {
		t.Fatalf("logs: %+v", all)
	}
	// 新到旧：最后一条是 404
	last, first := all.Items[0], all.Items[4]
	if last.Status != http.StatusNotFound || last.KeyID == nil || *last.KeyID != key.ID || last.KeyName != "写作助手" {
		t.Fatalf("latest entry: %+v", last)
	}
	if first.Method != http.MethodPost || first.Path != "/api/v1/ext/posts" || first.Status != http.StatusCreated ||
		first.UA != "agent/1.0" || first.IP == "" || first.KeyPrefix != key.Key[:12] {
		t.Fatalf("first entry: %+v", first)
	}
	if all.Items[3].Path != "/api/v1/ext/posts?status=draft" {
		t.Fatalf("query string not kept: %q", all.Items[3].Path)
	}
	invalid := all.Items[2]
	if invalid.KeyID != nil || invalid.Status != http.StatusUnauthorized || invalid.KeyPrefix != "myk_deadbeef" {
		t.Fatalf("invalid key entry: %+v", invalid)
	}
	if missing := all.Items[1]; missing.KeyID != nil || missing.KeyPrefix != "" || missing.Status != http.StatusUnauthorized {
		t.Fatalf("missing key entry: %+v", missing)
	}
	// 明文 Key 与请求体从不落库
	var leaked int
	if err := e.server.DB.QueryRow(`SELECT COUNT(*) FROM api_logs WHERE path || key_name || key_prefix || user_agent LIKE ? OR path LIKE '%secret%'`,
		"%"+key.Key+"%").Scan(&leaked); err != nil || leaked != 0 {
		t.Fatalf("secret leaked into logs: %d %v", leaked, err)
	}

	// 筛选：按 Key、按状态
	var byKey apiLogPage
	e.call(http.MethodGet, "/api/v1/admin/api-logs?key="+itoa(key.ID), nil, &byKey, http.StatusOK)
	if byKey.Total != 3 || byKey.Counts["error"] != 1 {
		t.Fatalf("key filter: %+v", byKey)
	}
	var errs apiLogPage
	e.call(http.MethodGet, "/api/v1/admin/api-logs?status=error", nil, &errs, http.StatusOK)
	if errs.Total != 3 || len(errs.Items) != 3 {
		t.Fatalf("error filter: %+v", errs)
	}
	for _, l := range errs.Items {
		if l.Status < 400 {
			t.Fatalf("ok row in error filter: %+v", l)
		}
	}
	var keyOK apiLogPage
	e.call(http.MethodGet, "/api/v1/admin/api-logs?key="+itoa(key.ID)+"&status=ok", nil, &keyOK, http.StatusOK)
	if keyOK.Total != 2 || len(keyOK.Items) != 2 {
		t.Fatalf("key+ok filter: %+v", keyOK)
	}

	// 分页
	var p2 apiLogPage
	e.call(http.MethodGet, "/api/v1/admin/api-logs?page=2&pageSize=2", nil, &p2, http.StatusOK)
	if p2.Page != 2 || p2.PageSize != 2 || p2.Total != 5 || len(p2.Items) != 2 || p2.Items[0].ID != all.Items[2].ID {
		t.Fatalf("page 2: %+v", p2)
	}
	var p3 apiLogPage
	e.call(http.MethodGet, "/api/v1/admin/api-logs?page=3&pageSize=2", nil, &p3, http.StatusOK)
	if len(p3.Items) != 1 {
		t.Fatalf("page 3: %+v", p3)
	}

	// Key 列表带调用数
	var keys []apiKeyInfo
	e.call(http.MethodGet, "/api/v1/admin/apikeys", nil, &keys, http.StatusOK)
	if len(keys) != 1 || keys[0].Calls != 3 || keys[0].Errors != 1 || keys[0].LastUsedAt == nil {
		t.Fatalf("key usage: %+v", keys)
	}

	// 吊销后日志保留名字快照
	e.call(http.MethodDelete, "/api/v1/admin/apikeys/"+itoa(key.ID), nil, nil, http.StatusOK)
	e.call(http.MethodGet, "/api/v1/admin/api-logs?key="+itoa(key.ID), nil, &byKey, http.StatusOK)
	if byKey.Total != 3 || byKey.Items[0].KeyName != "写作助手" {
		t.Fatalf("logs after revoke: %+v", byKey)
	}

	// 仅管理员
	if c, _ := e.as("", http.MethodGet, "/api/v1/admin/api-logs", nil); c != http.StatusUnauthorized {
		t.Fatalf("anonymous logs: %d", c)
	}
}

func TestAPILogRetention(t *testing.T) {
	e := newEnv(t)
	db := e.server.DB
	// 灌入超额的旧记录：前 10 条是 100 天前的
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`INSERT INTO api_logs (created_at, method, path, status) VALUES (?, 'GET', '/api/v1/ext/posts', 200)`)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < store.APILogKeep+120; i++ {
		at := time.Now().UTC()
		if i < 10 {
			at = at.AddDate(0, 0, -100)
		}
		if _, err := stmt.Exec(at.Format("2006-01-02 15:04:05")); err != nil {
			t.Fatal(err)
		}
	}
	stmt.Close()
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	// 顺带清理只在 id 为 50 的倍数时触发：写到下一个倍数为止
	for {
		var maxID int64
		if err := db.QueryRow(`SELECT MAX(id) FROM api_logs`).Scan(&maxID); err != nil {
			t.Fatal(err)
		}
		if err := db.InsertAPILog(store.APILog{Method: "GET", Path: "/x", Status: 200}); err != nil {
			t.Fatal(err)
		}
		if (maxID+1)%50 == 0 {
			break
		}
	}
	var n, old int
	if err := db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(created_at < ?), 0) FROM api_logs`, time.Now().UTC().AddDate(0, 0, -90).Format("2006-01-02 15:04:05")).Scan(&n, &old); err != nil {
		t.Fatal(err)
	}
	if n > store.APILogKeep || old != 0 {
		t.Fatalf("retention: %d rows, %d older than 90 days", n, old)
	}
	var total apiLogPage
	e.call(http.MethodGet, fmt.Sprintf("/api/v1/admin/api-logs?pageSize=%d", 100), nil, &total, http.StatusOK)
	if total.Total != n || len(total.Items) != 100 {
		t.Fatalf("paged after prune: total %d items %d", total.Total, len(total.Items))
	}
}
