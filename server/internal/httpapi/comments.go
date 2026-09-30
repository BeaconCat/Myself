package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/store"
)

// 评论：挂在文章（post，按 slug）、随想（note，按 id）或留言墙（guestbook）下。正文为纯文本（前端转义后按行渲染）。
// 可用条件：用户系统 + 读者 + 评论三级开关都开启；匿名评论另需「允许匿名」，且一律先审。
// 审核策略：all = 全部先审；first = 首条先审，通过过一次后自动公开；none = 登录用户直接公开。管理员的评论始终直接公开。

const (
	maxCommentRunes = 2000
	commentWindow   = 10 * time.Minute
	commentBurst    = 6
)

// commentLimiter 每个 IP 在窗口内最多发 commentBurst 条。
type commentLimiter struct {
	mu sync.Mutex
	m  map[string][]time.Time
}

func (c *commentLimiter) allow(ip string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = map[string][]time.Time{}
	}
	now := time.Now()
	kept := c.m[ip][:0]
	for _, t := range c.m[ip] {
		if now.Sub(t) < commentWindow {
			kept = append(kept, t)
		}
	}
	if len(kept) >= commentBurst {
		c.m[ip] = kept
		return false
	}
	c.m[ip] = append(kept, now)
	return true
}

type commentAuthor struct {
	ID     int64  `json:"id,omitempty"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Role   string `json:"role"` // admin | author | reader | guest
}

type commentItem struct {
	ID        int64         `json:"id"`
	ParentID  int64         `json:"parentId,omitempty"`
	Body      string        `json:"body"`
	CreatedAt string        `json:"createdAt"`
	Author    commentAuthor `json:"author"`
	Pending   bool          `json:"pending,omitempty"`
	// Likes / Liked 喜欢数与当前访客是否已点（仅前台列表返回；访客回应关闭时不返回）
	Likes int  `json:"likes,omitempty"`
	Liked bool `json:"liked,omitempty"`
}

// resolveTarget target + key → 目标 id；目标不存在 / 未公开返回 false。
func (s *Server) resolveTarget(target, key string) (int64, bool) {
	switch target {
	case "post":
		var id int64
		err := s.DB.QueryRow(`SELECT id FROM posts WHERE slug = ? AND `+store.PublicPost, key).Scan(&id)
		return id, err == nil
	case "note":
		id, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			return 0, false
		}
		var n int
		_ = s.DB.QueryRow(`SELECT COUNT(*) FROM notes WHERE id = ? AND hidden = 0`, id).Scan(&n)
		return id, n > 0
	case "guestbook":
		return 0, true
	}
	return 0, false
}

const commentSelect = `SELECT c.id, COALESCE(c.parent_id, 0), c.body, c.created_at, c.status, COALESCE(c.user_id, 0), c.guest_name,
	COALESCE(u.name, ''), COALESCE(u.avatar, ''), COALESCE(u.role, '')
	FROM comments c LEFT JOIN users u ON u.id = c.user_id `

func scanComment(rows *sql.Rows) (commentItem, string, error) {
	var (
		it                    commentItem
		status, guest         string
		uid                   int64
		uname, uavatar, urole string
	)
	err := rows.Scan(&it.ID, &it.ParentID, &it.Body, &it.CreatedAt, &status, &uid, &guest, &uname, &uavatar, &urole)
	if uid > 0 && uname != "" {
		it.Author = commentAuthor{ID: uid, Name: uname, Avatar: uavatar, Role: urole}
	} else {
		it.Author = commentAuthor{Name: guest, Role: "guest"}
	}
	it.Pending = status == "pending"
	return it, status, err
}

// authorOfTarget 协作作者在自己的文章下发言（回复读者）直接公开。
func (s *Server) authorOfTarget(u *store.User, target string, targetID int64) bool {
	if u == nil || u.Role != store.RoleAuthor || target != "post" {
		return false
	}
	var n int
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM posts WHERE id = ? AND author_id = ?`, targetID, u.ID).Scan(&n)
	return n > 0
}

// GET /comments?target=post|note|guestbook&key= 已公开的评论 + 当前用户自己待审的
func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	cfg := s.Config.Typed()
	if !cfg.Users.CommentsOn() {
		writeJSON(w, http.StatusOK, map[string]any{"enabled": false, "items": []commentItem{}})
		return
	}
	q := r.URL.Query()
	target := q.Get("target")
	id, ok := s.resolveTarget(target, q.Get("key"))
	if !ok {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	var uid int64 = -1
	if u, _ := s.currentUser(r); u != nil {
		uid = u.ID
	}
	rows, err := s.DB.Query(commentSelect+`WHERE c.target = ? AND c.target_id = ?
		AND (c.status = 'approved' OR (c.status = 'pending' AND c.user_id = ?)) ORDER BY c.id LIMIT 500`, target, id, uid)
	if err != nil {
		fail(w, err)
		return
	}
	defer rows.Close()
	items := []commentItem{}
	for rows.Next() {
		it, _, err := scanComment(rows)
		if err != nil {
			fail(w, err)
			return
		}
		items = append(items, it)
	}
	for i := range items {
		items[i].Author.Avatar = s.avatarOf(items[i].Author.Role, items[i].Author.Avatar)
	}
	if cfg.Users.ReactionsOn() && len(items) > 0 {
		if err := s.attachLikes(w, r, items); err != nil {
			fail(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "items": items, "reactions": cfg.Users.ReactionsOn()})
}

// POST /comments {target, key, parentId?, body, guestName?, website?}
func (s *Server) createComment(w http.ResponseWriter, r *http.Request) {
	cfg := s.Config.Typed()
	if !cfg.Users.CommentsOn() {
		writeError(w, http.StatusForbidden, "comments_closed")
		return
	}
	u, fromCookie := s.currentUser(r)
	if u != nil && fromCookie && !csrfOK(r) {
		writeError(w, http.StatusForbidden, "csrf_rejected")
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	ip := clientIP(r)
	if b.strOr("website") != "" { // 蜜罐
		writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "pending": true})
		return
	}
	if !s.comments.allow(ip) {
		writeError(w, http.StatusTooManyRequests, "too_many_comments")
		return
	}
	text := strings.TrimSpace(strings.ReplaceAll(b.strOr("body"), "\r\n", "\n"))
	if text == "" || len([]rune(text)) > maxCommentRunes {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}
	target := b.strOr("target")
	targetID, ok := s.resolveTarget(target, b.strOr("key"))
	if !ok {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	var parent any
	if pid := int64(b.num("parentId")); pid > 0 {
		var n int
		_ = s.DB.QueryRow(`SELECT COUNT(*) FROM comments WHERE id = ? AND target = ? AND target_id = ? AND status = 'approved'`,
			pid, target, targetID).Scan(&n)
		if n == 0 {
			writeError(w, http.StatusBadRequest, "invalid_parent")
			return
		}
		parent = pid
	}

	status := "pending"
	var uid any
	guest := ""
	switch {
	case u != nil:
		uid = u.ID
		switch {
		case u.Role == store.RoleAdmin, cfg.Users.Comments.Moderation == "none", s.authorOfTarget(u, target, targetID):
			status = "approved"
		case cfg.Users.Comments.Moderation == "first":
			var n int
			_ = s.DB.QueryRow(`SELECT COUNT(*) FROM comments WHERE user_id = ? AND status = 'approved'`, u.ID).Scan(&n)
			if n > 0 {
				status = "approved"
			}
		}
	case cfg.Users.Comments.Anonymous:
		guest = strings.TrimSpace(b.strOr("guestName"))
		if guest == "" || len([]rune(guest)) > 24 {
			writeError(w, http.StatusBadRequest, "invalid_name")
			return
		}
	default:
		writeError(w, http.StatusUnauthorized, "login_required")
		return
	}
	res, err := s.DB.Exec(`INSERT INTO comments (target, target_id, parent_id, user_id, guest_name, body, status, ip_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, target, targetID, parent, uid, guest, text, status, auth.HashToken(ip)[:16])
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "id": id, "pending": status == "pending"})
}

/* ===== 管理员：审核 ===== */

type adminComment struct {
	commentItem
	Status      string `json:"status"`
	Target      string `json:"target"`
	TargetTitle string `json:"targetTitle"`
	TargetLink  string `json:"targetLink"`
	// TargetKey 发表评论用的 key（文章 slug / 随想 id / 留言墙为空），后台回复时原样带回
	TargetKey string `json:"targetKey"`
	HasLink   bool   `json:"hasLink"`
	IPHash    string `json:"ipHash"`
}

// commentScope 后台评论的可见范围：协作作者只能管理自己文章下的评论。
func commentScope(r *http.Request) (string, []any) {
	if u := userOf(r); u != nil && u.Role == store.RoleAuthor {
		return ` AND c.target = 'post' AND c.target_id IN (SELECT id FROM posts WHERE author_id = ?)`, []any{u.ID}
	}
	return "", nil
}

// ownsComments 协作作者只能处理自己文章下的评论（站长不限）。
func (s *Server) ownsComments(r *http.Request, ids []int64) (bool, error) {
	scope, args := commentScope(r)
	if scope == "" || len(ids) == 0 {
		return true, nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	q := `SELECT COUNT(*) FROM comments c WHERE c.id IN (` + ph + `)` + scope
	all := make([]any, 0, len(ids)+len(args))
	for _, id := range ids {
		all = append(all, id)
	}
	var n int
	if err := s.DB.QueryRow(q, append(all, args...)...).Scan(&n); err != nil {
		return false, err
	}
	return n == len(ids), nil
}

// GET /admin/comments?status=pending|approved|spam（协作作者只看自己文章下的）
func (s *Server) adminListComments(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status != "approved" && status != "spam" {
		status = "pending"
	}
	scope, scopeArgs := commentScope(r)
	rows, err := s.DB.Query(`SELECT c.id, COALESCE(c.parent_id, 0), c.body, c.created_at, c.status, COALESCE(c.user_id, 0), c.guest_name,
		COALESCE(u.name, ''), COALESCE(u.avatar, ''), COALESCE(u.role, ''),
		c.target, c.target_id, COALESCE(p.title, ''), COALESCE(p.slug, ''), COALESCE(substr(n.content_md, 1, 40), ''), c.ip_hash
		FROM comments c
		LEFT JOIN users u ON u.id = c.user_id
		LEFT JOIN posts p ON c.target = 'post' AND p.id = c.target_id
		LEFT JOIN notes n ON c.target = 'note' AND n.id = c.target_id
		WHERE c.status = ?`+scope+` ORDER BY c.id DESC LIMIT 200`, append([]any{status}, scopeArgs...)...)
	if err != nil {
		fail(w, err)
		return
	}
	defer rows.Close()
	out := []adminComment{}
	for rows.Next() {
		var (
			it                    adminComment
			uid, targetID         int64
			guest                 string
			uname, uavatar, urole string
			ptitle, pslug, nbody  string
		)
		if err := rows.Scan(&it.ID, &it.ParentID, &it.Body, &it.CreatedAt, &it.Status, &uid, &guest, &uname, &uavatar, &urole,
			&it.Target, &targetID, &ptitle, &pslug, &nbody, &it.IPHash); err != nil {
			fail(w, err)
			return
		}
		if uid > 0 && uname != "" {
			it.Author = commentAuthor{ID: uid, Name: uname, Avatar: uavatar, Role: urole}
		} else {
			it.Author = commentAuthor{Name: guest, Role: "guest"}
		}
		switch it.Target {
		case "post":
			it.TargetTitle, it.TargetLink, it.TargetKey = ptitle, "/articles/"+pslug, pslug
		case "note":
			key := strconv.FormatInt(targetID, 10)
			it.TargetTitle, it.TargetLink, it.TargetKey = nbody, "/thoughts/"+key, key
		default:
			it.TargetTitle, it.TargetLink = "留言墙", "/about"
		}
		it.HasLink = strings.Contains(it.Body, "http://") || strings.Contains(it.Body, "https://")
		it.Author.Avatar = s.avatarOf(it.Author.Role, it.Author.Avatar)
		out = append(out, it)
	}
	writeJSON(w, http.StatusOK, out)
}

var errBadStatus = errors.New("invalid status")

func (s *Server) setCommentStatus(ids []int64, status string) error {
	if status != "approved" && status != "pending" && status != "spam" {
		return errBadStatus
	}
	for _, id := range ids {
		if _, err := s.DB.Exec(`UPDATE comments SET status = ? WHERE id = ?`, status, id); err != nil {
			return err
		}
	}
	return nil
}

// attachLikes 给评论补上喜欢数与当前访客是否已点（不为无 Cookie 的访客新建身份）。
func (s *Server) attachLikes(w http.ResponseWriter, r *http.Request, items []commentItem) error {
	ids := make([]int64, len(items))
	for i, it := range items {
		ids[i] = it.ID
	}
	sum, err := s.summaries("comment", ids, s.voterOf(w, r, false))
	if err != nil {
		return err
	}
	for i := range items {
		if e := sum[items[i].ID]; e != nil {
			items[i].Likes = e.Reactions["like"]
			items[i].Liked = slices.Contains(e.Mine, "like")
		}
	}
	return nil
}

// deleteComments 删除评论及其回复，连同它们收到的回应。
func (s *Server) deleteComments(ids []int64) error {
	for _, id := range ids {
		if _, err := s.DB.Exec(`DELETE FROM reactions WHERE target = 'comment' AND target_id IN
			(SELECT id FROM comments WHERE id = ? OR parent_id = ?)`, id, id); err != nil {
			return err
		}
		if _, err := s.DB.Exec(`DELETE FROM comments WHERE id = ? OR parent_id = ?`, id, id); err != nil {
			return err
		}
	}
	return nil
}

// PUT /admin/comments/{id} {status}
func (s *Server) adminUpdateComment(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if ok, err := s.ownsComments(r, []int64{pathID(r)}); err != nil || !ok {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if err := s.setCommentStatus([]int64{pathID(r)}, b.strOr("status")); err != nil {
		if errors.Is(err, errBadStatus) {
			writeError(w, http.StatusBadRequest, "invalid_status")
			return
		}
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DELETE /admin/comments/{id}（连同回复）
func (s *Server) adminDeleteComment(w http.ResponseWriter, r *http.Request) {
	if ok, err := s.ownsComments(r, []int64{pathID(r)}); err != nil || !ok {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if err := s.deleteComments([]int64{pathID(r)}); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /admin/comments/batch {ids, status | delete}
func (s *Server) adminBatchComments(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	raw, _ := b["ids"].([]any)
	ids := make([]int64, 0, len(raw))
	for _, v := range raw {
		if f, ok := v.(float64); ok && f > 0 && len(ids) < 500 {
			ids = append(ids, int64(f))
		}
	}
	if ok, err := s.ownsComments(r, ids); err != nil || !ok {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	var err error
	if b.truthy("delete") {
		err = s.deleteComments(ids)
	} else {
		err = s.setCommentStatus(ids, b.strOr("status"))
	}
	if errors.Is(err, errBadStatus) {
		writeError(w, http.StatusBadRequest, "invalid_status")
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "n": len(ids)})
}
