package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"

	"myself/server/internal/store"
)

func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// API Key 权限档位。
const (
	scopeFull    = "full"    // 全托管：读写全部文章与随想，可直接发布
	scopeContrib = "contrib" // 仅投稿：只能创建草稿文章，只能读 / 改 / 删自己创建的草稿
)

type keyCtxKey struct{}

// apiKeyCtx 经认证的 Key 身份，挂在请求 context 上供外部通道处理器判定权限。
type apiKeyCtx struct {
	ID    int64
	Scope string
}

func keyOf(r *http.Request) apiKeyCtx {
	k, _ := r.Context().Value(keyCtxKey{}).(apiKeyCtx)
	return k
}

// requireAPIKey X-Api-Key 认证中间件（外部 AI 发文通道）。
func (s *Server) requireAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Api-Key")
		if key == "" {
			writeError(w, http.StatusUnauthorized, "missing_api_key")
			return
		}
		var k apiKeyCtx
		err := s.DB.QueryRow(`SELECT id, scope FROM api_keys WHERE key_hash = ?`, hashKey(key)).Scan(&k.ID, &k.Scope)
		if err == sql.ErrNoRows {
			writeError(w, http.StatusUnauthorized, "invalid_api_key")
			return
		}
		if err != nil {
			fail(w, err)
			return
		}
		if _, err := s.DB.Exec(`UPDATE api_keys SET last_used_at = datetime('now') WHERE id = ?`, k.ID); err != nil {
			fail(w, err)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), keyCtxKey{}, k)))
	}
}

type apiKeyInfo struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Prefix     string  `json:"prefix"`
	Scope      string  `json:"scope"`
	LastUsedAt *string `json:"lastUsedAt"`
	CreatedAt  string  `json:"createdAt"`
	Key        string  `json:"key,omitempty"`
}

// GET /admin/apikeys
func (s *Server) listAPIKeys(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.DB.Query(`SELECT id, name, prefix, scope, last_used_at, created_at FROM api_keys ORDER BY id DESC`)
	if err != nil {
		fail(w, err)
		return
	}
	defer rows.Close()
	out := []apiKeyInfo{}
	for rows.Next() {
		var k apiKeyInfo
		var last sql.NullString
		if err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &k.Scope, &last, &k.CreatedAt); err != nil {
			fail(w, err)
			return
		}
		if last.Valid {
			k.LastUsedAt = &last.String
		}
		out = append(out, k)
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /admin/apikeys 创建（明文只返回这一次）。scope 缺省为仅投稿。
func (s *Server) createAPIKey(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "name_required")
		return
	}
	name := strings.TrimSpace(b.strOr("name"))
	if name == "" {
		writeError(w, http.StatusBadRequest, "name_required")
		return
	}
	name = limitRunes(name, 60)
	scope := b.strOr("scope")
	if scope != scopeFull {
		scope = scopeContrib
	}
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		fail(w, err)
		return
	}
	key := "myk_" + hex.EncodeToString(raw)
	prefix := key[:12]
	res, err := s.DB.Exec(`INSERT INTO api_keys (name, key_hash, prefix, scope) VALUES (?, ?, ?, ?)`, name, hashKey(key), prefix, scope)
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "name": name, "prefix": prefix, "scope": scope, "key": key})
}

// DELETE /admin/apikeys/{id} 吊销
func (s *Server) deleteAPIKey(w http.ResponseWriter, r *http.Request) {
	res, err := s.DB.Exec(`DELETE FROM api_keys WHERE id = ?`, pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	okOrNotFound(w, res)
}

/* ===== 外部通道（X-Api-Key）：全托管 Key 管理全部内容；仅投稿 Key 只碰自己创建的草稿 ===== */

// ownWhere 仅投稿 Key 追加「自己创建的草稿」条件。
func ownWhere(k apiKeyCtx) (string, []any) {
	if k.Scope == scopeFull {
		return "", nil
	}
	return " AND source_key = ? AND status = 'draft'", []any{k.ID}
}

// GET /ext/posts?status=all|published|draft 列表
func (s *Server) extListPosts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	where, args := "WHERE 1 = 1", []any{}
	if status == "published" || status == "draft" {
		where += " AND status = ?"
		args = append(args, status)
	}
	own, ownArgs := ownWhere(keyOf(r))
	rows, err := s.DB.QueryPosts(where+own+` ORDER BY created_at DESC`, append(args, ownArgs...)...)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.toPosts(rows, store.PostOpts{WithStatus: true}))
}

// GET /ext/posts/{id} 详情（含 Markdown 正文）
func (s *Server) extGetPost(w http.ResponseWriter, r *http.Request) {
	own, ownArgs := ownWhere(keyOf(r))
	row, err := s.DB.GetPost(`WHERE id = ?`+own, append([]any{pathID(r)}, ownArgs...)...)
	if err != nil {
		fail(w, err)
		return
	}
	if row == nil {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	writeJSON(w, http.StatusOK, s.onePost(*row, store.PostOpts{WithContent: true, WithStatus: true}))
}

// POST /ext/posts 投稿文章（默认草稿；全托管 Key 可显式 published）
func (s *Server) extCreatePost(w http.ResponseWriter, r *http.Request) {
	k := keyOf(r)
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_post")
		return
	}
	in, ok := parsePost(b, true, false, 3)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_post")
		return
	}
	if k.Scope != scopeFull {
		in.Status = "draft"
	}
	res, err := s.DB.Exec(`INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, status, source_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		in.Slug, in.Title, in.Excerpt, in.ContentMd, store.JSONStrings(in.Covers), store.JSONStrings(in.Tags), in.Status, k.ID)
	if store.IsUniqueErr(err) {
		writeError(w, http.StatusConflict, "slug_exists")
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "slug": in.Slug, "status": in.Status})
}

// PUT /ext/posts/{id} 更新文章（slug 不可改；仅投稿 Key 不能发布）
func (s *Server) extUpdatePost(w http.ResponseWriter, r *http.Request) {
	k := keyOf(r)
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_post")
		return
	}
	in, ok := parsePost(b, false, false, 3)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_post")
		return
	}
	if k.Scope != scopeFull {
		in.Status = "draft"
	}
	own, ownArgs := ownWhere(k)
	args := append([]any{in.Title, in.Excerpt, in.ContentMd, store.JSONStrings(in.Covers), store.JSONStrings(in.Tags), in.Status, pathID(r)}, ownArgs...)
	res, err := s.DB.Exec(`UPDATE posts SET
		title = ?, excerpt = ?, content_md = ?, covers = ?, tags = ?, status = ?, updated_at = datetime('now')
		WHERE id = ?`+own, args...)
	if err != nil {
		fail(w, err)
		return
	}
	okOrNotFound(w, res)
}

// DELETE /ext/posts/{id}
func (s *Server) extDeletePost(w http.ResponseWriter, r *http.Request) {
	own, ownArgs := ownWhere(keyOf(r))
	res, err := s.DB.Exec(`DELETE FROM posts WHERE id = ?`+own, append([]any{pathID(r)}, ownArgs...)...)
	if err != nil {
		fail(w, err)
		return
	}
	okOrNotFound(w, res)
}

// fullOnly 随想发出即公开，只开放给全托管 Key。
func fullOnly(w http.ResponseWriter, r *http.Request) bool {
	if keyOf(r).Scope == scopeFull {
		return true
	}
	writeError(w, http.StatusForbidden, "scope_forbidden")
	return false
}

// GET /ext/notes 随想列表
func (s *Server) extListNotes(w http.ResponseWriter, r *http.Request) {
	if !fullOnly(w, r) {
		return
	}
	pageSize := queryInt(r, "pageSize", 50, 1, 100)
	items, err := s.DB.QueryNotes(`ORDER BY created_at DESC LIMIT ?`, false, pageSize)
	if err != nil {
		fail(w, err)
		return
	}
	if items == nil {
		items = []store.Note{}
	}
	writeJSON(w, http.StatusOK, items)
}

// POST /ext/notes 投稿随想
func (s *Server) extCreateNote(w http.ResponseWriter, r *http.Request) {
	if !fullOnly(w, r) {
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_note")
		return
	}
	in, ok := parseNote(b)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_note")
		return
	}
	res, err := s.DB.Exec(`INSERT INTO notes (content_md, mood, images, source_key) VALUES (?, ?, ?, ?)`,
		in.ContentMd, in.Mood, store.JSONStrings(in.Images), keyOf(r).ID)
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// PUT /ext/notes/{id} 更新随想
func (s *Server) extUpdateNote(w http.ResponseWriter, r *http.Request) {
	if !fullOnly(w, r) {
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_note")
		return
	}
	in, ok := parseNote(b)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_note")
		return
	}
	res, err := s.DB.Exec(`UPDATE notes SET content_md = ?, mood = ?, images = ? WHERE id = ?`,
		in.ContentMd, in.Mood, store.JSONStrings(in.Images), pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	okOrNotFound(w, res)
}

// DELETE /ext/notes/{id}
func (s *Server) extDeleteNote(w http.ResponseWriter, r *http.Request) {
	if !fullOnly(w, r) {
		return
	}
	s.adminDeleteNote(w, r)
}

// okOrNotFound 依据受影响行数回 {ok:true} 或 404。
func okOrNotFound(w http.ResponseWriter, res sql.Result) {
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
