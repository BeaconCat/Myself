package httpapi

import (
	"net/http"
	"regexp"
	"strings"

	"myself/server/internal/auth"
	"myself/server/internal/config"
	"myself/server/internal/store"
)

var slugRe = regexp.MustCompile(`^[a-z0-9-]{1,80}$`)

// POST /auth/login 失败计数按 IP 锁定；成功写入 HttpOnly 会话 Cookie（响应体不含令牌），回 mustChange 提示前端先改密
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if d := s.limiter.blocked(ip); d > 0 {
		tooMany(w, d)
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	var (
		token string
		err   error
	)
	withScryptSlot(func() { token, err = s.Auth.Login(b.strOr("username"), b.strOr("password")) })
	if err != nil {
		fail(w, err)
		return
	}
	if token == "" {
		s.limiter.fail(ip)
		writeError(w, http.StatusUnauthorized, "bad_credentials")
		return
	}
	s.limiter.success(ip)
	setSession(w, r, token)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "mustChange": s.Auth.MustChange()})
}

// PUT /auth/password 改密：其它会话全部失效，当前会话换发新 Cookie
func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if d := s.limiter.blocked(ip); d > 0 {
		tooMany(w, d)
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	newPassword, ok := b.str("newPassword")
	if !ok || len([]rune(newPassword)) < auth.MinPasswordLen || len(newPassword) > 256 {
		writeError(w, http.StatusBadRequest, "weak_password")
		return
	}
	var (
		token string
		err   error
	)
	withScryptSlot(func() { token, err = s.Auth.ChangePassword(b.strOr("oldPassword"), newPassword) })
	if err != nil {
		fail(w, err)
		return
	}
	if token == "" {
		s.limiter.fail(ip)
		writeError(w, http.StatusUnauthorized, "bad_credentials")
		return
	}
	setSession(w, r, token)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// postInput 是文章写入的规范化字段。
type postInput struct {
	Slug, Title, Excerpt, ContentMd, Status string
	Covers, Tags                            []string
	Pinned                                  bool
}

// parsePost 校验并规范化文章体。requireSlug=false 时不校验 slug（外部 PUT）。
// defaultPublished 决定 status 缺省值；coverLimit>0 时截断封面数。
func parsePost(b body, requireSlug, defaultPublished bool, coverLimit int) (postInput, bool) {
	slug, _ := b.str("slug")
	title, hasTitle := b.str("title")
	content, hasContent := b.str("contentMd")
	if requireSlug && !slugRe.MatchString(slug) {
		return postInput{}, false
	}
	if !hasTitle || strings.TrimSpace(title) == "" || !hasContent {
		return postInput{}, false
	}
	status := "draft"
	if defaultPublished {
		status = "published"
		if b.strOr("status") == "draft" {
			status = "draft"
		}
	} else if b.strOr("status") == "published" {
		status = "published"
	}
	return postInput{
		Slug:      slug,
		Title:     strings.TrimSpace(title),
		Excerpt:   b.strOr("excerpt"),
		ContentMd: content,
		Status:    status,
		Covers:    b.strings("covers", coverLimit),
		Tags:      b.strings("tags", 0),
		Pinned:    b.truthy("pinned"),
	}, true
}

// GET /admin/posts 全量（含草稿）
func (s *Server) adminListPosts(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.DB.QueryPosts(`ORDER BY created_at DESC`)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPosts(rows, store.PostOpts{WithStatus: true}))
}

// GET /admin/posts/{id} 编辑用详情
func (s *Server) adminGetPost(w http.ResponseWriter, r *http.Request) {
	row, err := s.DB.GetPost(`WHERE id = ?`, pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	if row == nil {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	writeJSON(w, http.StatusOK, row.ToPost(store.PostOpts{WithContent: true, WithStatus: true}))
}

// POST /admin/posts 新建
func (s *Server) adminCreatePost(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_post")
		return
	}
	in, ok := parsePost(b, true, true, 0)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_post")
		return
	}
	res, err := s.DB.Exec(`INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, status, pinned)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		in.Slug, in.Title, in.Excerpt, in.ContentMd, store.JSONStrings(in.Covers), store.JSONStrings(in.Tags),
		in.Status, boolInt(in.Pinned))
	if store.IsUniqueErr(err) {
		writeError(w, http.StatusConflict, "slug_exists")
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// PUT /admin/posts/{id} 更新
func (s *Server) adminUpdatePost(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_post")
		return
	}
	in, ok := parsePost(b, true, true, 0)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_post")
		return
	}
	res, err := s.DB.Exec(`UPDATE posts SET
		slug = ?, title = ?, excerpt = ?, content_md = ?, covers = ?, tags = ?, status = ?, pinned = ?,
		updated_at = datetime('now') WHERE id = ?`,
		in.Slug, in.Title, in.Excerpt, in.ContentMd, store.JSONStrings(in.Covers), store.JSONStrings(in.Tags),
		in.Status, boolInt(in.Pinned), pathID(r))
	if store.IsUniqueErr(err) {
		writeError(w, http.StatusConflict, "slug_exists")
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	okOrNotFound(w, res)
}

// DELETE /admin/posts/{id}
func (s *Server) adminDeletePost(w http.ResponseWriter, r *http.Request) {
	res, err := s.DB.Exec(`DELETE FROM posts WHERE id = ?`, pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	okOrNotFound(w, res)
}

// noteInput 是随想写入的规范化字段。
type noteInput struct {
	ContentMd, Mood string
	Images          []string
	Pinned          bool
}

func parseNote(b body) (noteInput, bool) {
	content, ok := b.str("contentMd")
	if !ok || strings.TrimSpace(content) == "" {
		return noteInput{}, false
	}
	return noteInput{
		ContentMd: strings.TrimSpace(content),
		Mood:      b.strOr("mood"),
		Images:    b.strings("images", 9),
		Pinned:    b.truthy("pinned"),
	}, true
}

// POST /admin/notes 发随想
func (s *Server) adminCreateNote(w http.ResponseWriter, r *http.Request) {
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
	res, err := s.DB.Exec(`INSERT INTO notes (content_md, mood, images, pinned) VALUES (?, ?, ?, ?)`,
		in.ContentMd, in.Mood, store.JSONStrings(in.Images), boolInt(in.Pinned))
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// PUT /admin/notes/{id} 编辑随想
func (s *Server) adminUpdateNote(w http.ResponseWriter, r *http.Request) {
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
	res, err := s.DB.Exec(`UPDATE notes SET content_md = ?, mood = ?, images = ?, pinned = ? WHERE id = ?`,
		in.ContentMd, in.Mood, store.JSONStrings(in.Images), boolInt(in.Pinned), pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	okOrNotFound(w, res)
}

// DELETE /admin/notes/{id}
func (s *Server) adminDeleteNote(w http.ResponseWriter, r *http.Request) {
	res, err := s.DB.Exec(`DELETE FROM notes WHERE id = ?`, pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	okOrNotFound(w, res)
}

// GET /site-config 公开站点配置（前台启动读取）。github 只下发白名单字段（令牌、代理等一律不出站）。
func (s *Server) siteConfig(w http.ResponseWriter, _ *http.Request) {
	cfg := s.Config.Get()
	src := config.Sub(cfg, "github")
	github := config.Map{}
	for _, k := range []string{"username", "mode", "stats", "refreshMinutes"} {
		if v, ok := src[k]; ok {
			github[k] = v
		}
	}
	need, _ := s.Auth.NeedsSetup()
	writeJSON(w, http.StatusOK, config.Map{
		"site":       cfg["site"],
		"loading":    cfg["loading"],
		"theme":      cfg["theme"],
		"hero":       cfg["hero"],
		"thoughts":   cfg["thoughts"],
		"covers":     cfg["covers"],
		"timezone":   cfg["timezone"],
		"github":     github,
		"about":      cfg["about"],
		"needsSetup": need,
	})
}

// GET /admin/settings 完整配置
func (s *Server) adminGetSettings(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.Config.Get())
}

// PUT /admin/settings 部分更新（深合并）
func (s *Server) adminSaveSettings(w http.ResponseWriter, r *http.Request) {
	var patch config.Map
	if err := readJSON(w, r, &patch); err != nil || patch == nil {
		writeError(w, http.StatusBadRequest, "invalid_config")
		return
	}
	merged, err := s.Config.Save(patch)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, merged)
}
