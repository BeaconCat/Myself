package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/store"
)

// 互动：随想 / 文章的回应（喜欢、灵感、会心、共鸣，线性图标，不用 emoji）与评论数。
// 回应不要求登录：登录用户按用户 id 去重，访客按服务端下发的匿名访客 Cookie（myself_vid）去重，另按 IP 限流。
// 开关：users.reactions（与用户系统总开关无关，默认开启）。

const visitorCookie = "myself_vid"

// ReactionKinds 允许的回应种类（前端按同名图标渲染）。
var ReactionKinds = []string{"like", "spark", "smile", "resonate"}

func validKind(k string) bool {
	for _, v := range ReactionKinds {
		if v == k {
			return true
		}
	}
	return false
}

// reactLimiter 每个 IP 一分钟最多 60 次回应切换。
type reactLimiter struct {
	mu sync.Mutex
	m  map[string][]time.Time
}

func (l *reactLimiter) allow(ip string) bool {
	ip = limitKey(ip)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.m == nil {
		l.m = map[string][]time.Time{}
	}
	now := time.Now()
	pruneWindows(l.m, now, time.Minute)
	kept := l.m[ip][:0]
	for _, t := range l.m[ip] {
		if now.Sub(t) < time.Minute {
			kept = append(kept, t)
		}
	}
	if len(kept) >= 60 {
		l.m[ip] = kept
		return false
	}
	l.m[ip] = append(kept, now)
	return true
}

// voterOf 回应者标识：登录用户 u:<id>；访客 v:<匿名 id>（没有则签发，create=false 时不签发）。
func (s *Server) voterOf(w http.ResponseWriter, r *http.Request, create bool) string {
	if u, _ := s.currentUser(r); u != nil {
		return "u:" + strconv.FormatInt(u.ID, 10)
	}
	if c, err := r.Cookie(visitorCookie); err == nil && len(c.Value) == 32 {
		return "v:" + c.Value
	}
	if !create {
		return ""
	}
	id, err := auth.RandomHex(16)
	if err != nil {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name: visitorCookie, Value: id, Path: "/api/", MaxAge: 365 * 24 * 3600,
		HttpOnly: true, Secure: secureRequest(r), SameSite: http.SameSiteLaxMode,
	})
	return "v:" + id
}

type engageSummary struct {
	Reactions map[string]int `json:"reactions"`
	Mine      []string       `json:"mine"`
	Comments  int            `json:"comments"`
}

// summaries 批量取回应计数、当前访客已点的回应、已公开评论数。
func (s *Server) summaries(target string, ids []int64, voter string) (map[int64]*engageSummary, error) {
	out := map[int64]*engageSummary{}
	if len(ids) == 0 {
		return out, nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := []any{target}
	for _, id := range ids {
		args = append(args, id)
		out[id] = &engageSummary{Reactions: map[string]int{}, Mine: []string{}}
	}
	rows, err := s.DB.Query(`SELECT target_id, kind, COUNT(*), SUM(voter = ?) FROM reactions
		WHERE target = ? AND target_id IN (`+ph+`) GROUP BY target_id, kind`, append([]any{voter}, args...)...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		var kind string
		var n, mine int
		if err := rows.Scan(&id, &kind, &n, &mine); err != nil {
			rows.Close()
			return nil, err
		}
		if e := out[id]; e != nil {
			e.Reactions[kind] = n
			if mine > 0 {
				e.Mine = append(e.Mine, kind)
			}
		}
	}
	rows.Close()
	crows, err := s.DB.Query(`SELECT target_id, COUNT(*) FROM comments WHERE status = 'approved' AND target = ? AND target_id IN (`+ph+`) GROUP BY target_id`, args...)
	if err != nil {
		return nil, err
	}
	defer crows.Close()
	for crows.Next() {
		var id int64
		var n int
		if err := crows.Scan(&id, &n); err != nil {
			return nil, err
		}
		if e := out[id]; e != nil {
			e.Comments = n
		}
	}
	return out, crows.Err()
}

func parseIDs(raw string, limit int) []int64 {
	var ids []int64
	for _, part := range strings.Split(raw, ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil && id > 0 && len(ids) < limit {
			ids = append(ids, id)
		}
	}
	return ids
}

// GET /engage?target=note|post&ids=1,2,3 批量互动摘要（列表用）
func (s *Server) engage(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target != "note" && target != "post" {
		writeError(w, http.StatusBadRequest, "invalid_target")
		return
	}
	out, err := s.summaries(target, parseIDs(r.URL.Query().Get("ids"), 100), s.voterOf(w, r, false))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// reactionTargets 可回应的对象 → 存在性校验（只能回应已公开的内容）。
var reactionTargets = map[string]string{
	"note":    `SELECT COUNT(*) FROM notes WHERE id = ? AND hidden = 0`,
	"post":    `SELECT COUNT(*) FROM posts WHERE id = ? AND ` + store.PublicPost,
	"comment": `SELECT COUNT(*) FROM comments WHERE id = ? AND status = 'approved'`,
}

// POST /reactions {target: note|post|comment, id, kind} 切换一个回应（已点则取消），返回该条最新摘要
func (s *Server) toggleReaction(w http.ResponseWriter, r *http.Request) {
	cfg := s.Config.Typed()
	if !cfg.Users.ReactionsOn() {
		writeError(w, http.StatusForbidden, "reactions_closed")
		return
	}
	// 跨站表单无法设置自定义头：回应接口要求 X-Requested-With，挡掉 CSRF 刷量
	if r.Header.Get(csrfHeader) != csrfValue {
		writeError(w, http.StatusForbidden, "csrf_rejected")
		return
	}
	if !s.reacts.allow(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "too_many_reactions")
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	target, kind := b.strOr("target"), b.strOr("kind")
	id := int64(b.num("id"))
	exists, ok := reactionTargets[target]
	if !ok || !validKind(kind) || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_reaction")
		return
	}
	var n int
	if err := s.DB.QueryRow(exists, id).Scan(&n); err != nil || n == 0 {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	voter := s.voterOf(w, r, true)
	if voter == "" {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	res, err := s.DB.Exec(`DELETE FROM reactions WHERE target = ? AND target_id = ? AND kind = ? AND voter = ?`, target, id, kind, voter)
	if err != nil {
		fail(w, err)
		return
	}
	if gone, _ := res.RowsAffected(); gone == 0 {
		if _, err := s.DB.Exec(`INSERT OR IGNORE INTO reactions (target, target_id, kind, voter) VALUES (?, ?, ?, ?)`, target, id, kind, voter); err != nil {
			fail(w, err)
			return
		}
	}
	out, err := s.summaries(target, []int64{id}, voter)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out[id])
}

// GET /notes/{id} 单条随想（详情页），附相邻两条的 id 便于上一条 / 下一条
func (s *Server) getNote(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	// 隐藏的随想前台 404；站长仍可打开（便于预览）
	vis := " AND hidden = 0"
	if s.isAdminReq(r) {
		vis = ""
	}
	items, err := s.DB.QueryNotes(`WHERE id = ?`+vis, false, id)
	if err != nil {
		fail(w, err)
		return
	}
	if len(items) == 0 {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	var prev, next int64
	_ = s.DB.QueryRow(`SELECT id FROM notes WHERE hidden = 0 AND (created_at < ? OR (created_at = ? AND id < ?)) ORDER BY created_at DESC, id DESC LIMIT 1`,
		items[0].CreatedAt, items[0].CreatedAt, id).Scan(&prev)
	_ = s.DB.QueryRow(`SELECT id FROM notes WHERE hidden = 0 AND (created_at > ? OR (created_at = ? AND id > ?)) ORDER BY created_at ASC, id ASC LIMIT 1`,
		items[0].CreatedAt, items[0].CreatedAt, id).Scan(&next)
	writeJSON(w, http.StatusOK, map[string]any{"note": items[0], "older": prev, "newer": next})
}
