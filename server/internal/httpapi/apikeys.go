package httpapi

import (
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

// requireAPIKey X-Api-Key 认证中间件（外部 AI 发文通道）。
func (s *Server) requireAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Api-Key")
		if key == "" {
			writeError(w, http.StatusUnauthorized, "missing_api_key")
			return
		}
		var id int64
		err := s.DB.QueryRow(`SELECT id FROM api_keys WHERE key_hash = ?`, hashKey(key)).Scan(&id)
		if err == sql.ErrNoRows {
			writeError(w, http.StatusUnauthorized, "invalid_api_key")
			return
		}
		if err != nil {
			fail(w, err)
			return
		}
		if _, err := s.DB.Exec(`UPDATE api_keys SET last_used_at = datetime('now') WHERE id = ?`, id); err != nil {
			fail(w, err)
			return
		}
		next(w, r)
	}
}

type apiKeyInfo struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Prefix     string  `json:"prefix"`
	LastUsedAt *string `json:"lastUsedAt"`
	CreatedAt  string  `json:"createdAt"`
	Key        string  `json:"key,omitempty"`
}

// GET /admin/apikeys
func (s *Server) listAPIKeys(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.DB.Query(`SELECT id, name, prefix, last_used_at, created_at FROM api_keys ORDER BY id DESC`)
	if err != nil {
		fail(w, err)
		return
	}
	defer rows.Close()
	out := []apiKeyInfo{}
	for rows.Next() {
		var k apiKeyInfo
		var last sql.NullString
		if err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &last, &k.CreatedAt); err != nil {
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

// POST /admin/apikeys 创建（明文只返回这一次）
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
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		fail(w, err)
		return
	}
	key := "myk_" + hex.EncodeToString(raw)
	prefix := key[:12]
	res, err := s.DB.Exec(`INSERT INTO api_keys (name, key_hash, prefix) VALUES (?, ?, ?)`, name, hashKey(key), prefix)
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "name": name, "prefix": prefix, "key": key})
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

/* ===== 外部通道（X-Api-Key）：供外部 AI/脚本全托管内容 ===== */

// GET /ext/posts?status=all|published|draft 列表（含草稿）
func (s *Server) extListPosts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "all"
	}
	var (
		rows []store.PostRow
		err  error
	)
	if status == "all" {
		rows, err = s.DB.QueryPosts(`ORDER BY created_at DESC`)
	} else {
		rows, err = s.DB.QueryPosts(`WHERE status = ? ORDER BY created_at DESC`, status)
	}
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPosts(rows, store.PostOpts{WithStatus: true}))
}

// GET /ext/posts/{id} 详情（含 Markdown 正文）
func (s *Server) extGetPost(w http.ResponseWriter, r *http.Request) {
	s.adminGetPost(w, r)
}

// POST /ext/posts 投稿文章（默认草稿，可显式 published）
func (s *Server) extCreatePost(w http.ResponseWriter, r *http.Request) {
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
	res, err := s.DB.Exec(`INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		in.Slug, in.Title, in.Excerpt, in.ContentMd, store.JSONStrings(in.Covers), store.JSONStrings(in.Tags), in.Status)
	if store.IsUniqueErr(err) {
		writeError(w, http.StatusConflict, "slug_exists")
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "slug": in.Slug})
}

// PUT /ext/posts/{id} 更新文章（slug 不可改）
func (s *Server) extUpdatePost(w http.ResponseWriter, r *http.Request) {
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
	res, err := s.DB.Exec(`UPDATE posts SET
		title = ?, excerpt = ?, content_md = ?, covers = ?, tags = ?, status = ?, updated_at = datetime('now')
		WHERE id = ?`,
		in.Title, in.Excerpt, in.ContentMd, store.JSONStrings(in.Covers), store.JSONStrings(in.Tags), in.Status, pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	okOrNotFound(w, res)
}

// DELETE /ext/posts/{id}
func (s *Server) extDeletePost(w http.ResponseWriter, r *http.Request) {
	s.adminDeletePost(w, r)
}

// GET /ext/notes 随想列表
func (s *Server) extListNotes(w http.ResponseWriter, r *http.Request) {
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
	res, err := s.DB.Exec(`INSERT INTO notes (content_md, mood, images) VALUES (?, ?, ?)`,
		in.ContentMd, in.Mood, store.JSONStrings(in.Images))
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// PUT /ext/notes/{id} 更新随想
func (s *Server) extUpdateNote(w http.ResponseWriter, r *http.Request) {
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
