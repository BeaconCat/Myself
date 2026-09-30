package httpapi

import (
	"net/http"
	"strings"

	"myself/server/internal/store"
)

// isAdminReq 当前请求是否来自站长会话（公开接口里用来放行隐藏内容）。
func (s *Server) isAdminReq(r *http.Request) bool {
	u, _ := s.currentUser(r)
	return u != nil && u.Role == store.RoleAdmin
}

// batchReq 批量操作：{"ids": [...], "action": "hide" | "show" | "delete"}，最多 500 条。
type batchReq struct {
	IDs    []int64 `json:"ids"`
	Action string  `json:"action"`
}

func readBatch(w http.ResponseWriter, r *http.Request) (batchReq, bool) {
	var b batchReq
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return b, false
	}
	if len(b.IDs) == 0 || len(b.IDs) > 500 {
		writeError(w, http.StatusBadRequest, "invalid_ids")
		return b, false
	}
	switch b.Action {
	case "hide", "show", "delete":
		return b, true
	}
	writeError(w, http.StatusBadRequest, "invalid_action")
	return b, false
}

// runBatch 对 table 中 ids 执行 hide / show / delete；scope 追加额外条件（作者只能动自己的文章）。
func (s *Server) runBatch(w http.ResponseWriter, table string, b batchReq, scope string, scopeArgs []any) {
	marks := strings.TrimSuffix(strings.Repeat("?,", len(b.IDs)), ",")
	args := make([]any, 0, len(b.IDs)+len(scopeArgs)+1)
	var sql string
	switch b.Action {
	case "delete":
		sql = `DELETE FROM ` + table + ` WHERE id IN (` + marks + `)` + scope
	default:
		sql = `UPDATE ` + table + ` SET hidden = ? WHERE id IN (` + marks + `)` + scope
		args = append(args, boolInt(b.Action == "hide"))
	}
	for _, id := range b.IDs {
		args = append(args, id)
	}
	res, err := s.DB.Exec(sql, append(args, scopeArgs...)...)
	if err != nil {
		fail(w, err)
		return
	}
	n, _ := res.RowsAffected()
	writeJSON(w, http.StatusOK, map[string]int64{"affected": n})
}

// POST /admin/posts/batch 批量隐藏 / 取消隐藏 / 删除文章（作者只作用于自己的）
func (s *Server) adminBatchPosts(w http.ResponseWriter, r *http.Request) {
	b, ok := readBatch(w, r)
	if !ok {
		return
	}
	scope, args := authorScope(r)
	s.runBatch(w, "posts", b, scope, args)
}

// POST /admin/notes/batch 批量隐藏 / 取消隐藏 / 删除随想
func (s *Server) adminBatchNotes(w http.ResponseWriter, r *http.Request) {
	b, ok := readBatch(w, r)
	if !ok {
		return
	}
	s.runBatch(w, "notes", b, "", nil)
}
