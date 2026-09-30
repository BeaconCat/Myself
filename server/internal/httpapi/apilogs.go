package httpapi

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"myself/server/internal/store"
)

// API 调用日志：外部通道请求的审计记录与后台查询。

// 调用日志字段长度上限（按字符）。
const (
	apiLogPathMax = 300
	apiLogUAMax   = 200
)

// keyPrefixOf 无效 Key 只留与后台列表同长的前缀（形如 myk_1a2b3c4d），便于辨认已吊销的旧 Key；不是 myk_ 格式的一律不记。
func keyPrefixOf(key string) string {
	if !strings.HasPrefix(key, "myk_") || len(key) < 12 {
		return ""
	}
	return key[:12]
}

// logAPICall 写一条调用日志；写失败只打日志，不影响已返回的响应。
func (s *Server) logAPICall(r *http.Request, e store.APILog, status int, start time.Time) {
	e.Method = r.Method
	e.Path = limitRunes(r.URL.RequestURI(), apiLogPathMax)
	e.Status = status
	e.Ms = time.Since(start).Milliseconds()
	e.IP = clientIP(r)
	e.UA = limitRunes(r.UserAgent(), apiLogUAMax)
	if err := s.DB.InsertAPILog(e); err != nil {
		log.Printf("[api-log] %v", err)
	}
}

// GET /admin/api-logs?page=&pageSize=&key=<id>&status=ok|error
// → {items, page, pageSize, total, counts:{all, ok, error}}；counts 只受 key 筛选影响，供状态分段显示数量。
func (s *Server) adminListAPILogs(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1, 1, 1<<30)
	pageSize := queryInt(r, "pageSize", 20, 1, 100)
	var conds []string
	var args []any
	if id, err := strconv.ParseInt(r.URL.Query().Get("key"), 10, 64); err == nil && id > 0 {
		conds = append(conds, "key_id = ?")
		args = append(args, id)
	}
	keyWhere := ""
	if len(conds) > 0 {
		keyWhere = "WHERE " + strings.Join(conds, " AND ")
	}
	var all, bad int
	if err := s.DB.QueryRow(`SELECT COUNT(*), COALESCE(SUM(status >= 400), 0) FROM api_logs `+keyWhere, args...).Scan(&all, &bad); err != nil {
		fail(w, err)
		return
	}
	total := all
	switch r.URL.Query().Get("status") {
	case "ok":
		conds, total = append(conds, "status < 400"), all-bad
	case "error":
		conds, total = append(conds, "status >= 400"), bad
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	items, err := s.DB.QueryAPILogs(where, args, pageSize, (page-1)*pageSize)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items, "page": page, "pageSize": pageSize, "total": total,
		"counts": map[string]int{"all": all, "ok": all - bad, "error": bad},
	})
}
