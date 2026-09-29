package httpapi

import (
	"net/http"
	"regexp"
	"strings"

	"myself/server/internal/config"
	"myself/server/internal/store"
)

var slugRe = regexp.MustCompile(`^[a-z0-9-]{1,80}$`)

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

// authorScope 协作作者只能看到 / 修改自己的文章；管理员不受限。
func authorScope(r *http.Request) (string, []any) {
	if u := userOf(r); u != nil && u.Role == store.RoleAuthor {
		return " AND author_id = ?", []any{u.ID}
	}
	return "", nil
}

// authorStatus 作者不能置顶；站点未开「作者可直接发布」时只能存草稿（修改已发布的文章也会退回草稿，重新等待审阅）。
func (s *Server) authorStatus(r *http.Request, in *postInput) {
	if u := userOf(r); u != nil && u.Role == store.RoleAuthor {
		in.Pinned = false
		if !s.Config.Typed().Users.Authors.DirectPublish {
			in.Status = "draft"
		}
	}
}

// GET /admin/posts 全量（含草稿）；作者只看自己的
func (s *Server) adminListPosts(w http.ResponseWriter, r *http.Request) {
	scope, args := authorScope(r)
	rows, err := s.DB.QueryPosts(`WHERE 1 = 1`+scope+` ORDER BY created_at DESC`, args...)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.toPosts(rows, store.PostOpts{WithStatus: true}))
}

// GET /admin/posts/{id} 编辑用详情
func (s *Server) adminGetPost(w http.ResponseWriter, r *http.Request) {
	scope, args := authorScope(r)
	row, err := s.DB.GetPost(`WHERE id = ?`+scope, append([]any{pathID(r)}, args...)...)
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
	s.authorStatus(r, &in)
	var author any
	if u := userOf(r); u != nil && u.Role == store.RoleAuthor {
		author = u.ID
	}
	res, err := s.DB.Exec(`INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, status, pinned, author_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.Slug, in.Title, in.Excerpt, in.ContentMd, store.JSONStrings(in.Covers), store.JSONStrings(in.Tags),
		in.Status, boolInt(in.Pinned), author)
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
	s.authorStatus(r, &in)
	scope, args := authorScope(r)
	res, err := s.DB.Exec(`UPDATE posts SET
		slug = ?, title = ?, excerpt = ?, content_md = ?, covers = ?, tags = ?, status = ?, pinned = ?,
		updated_at = datetime('now') WHERE id = ?`+scope,
		append([]any{in.Slug, in.Title, in.Excerpt, in.ContentMd, store.JSONStrings(in.Covers), store.JSONStrings(in.Tags),
			in.Status, boolInt(in.Pinned), pathID(r)}, args...)...)
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
	scope, args := authorScope(r)
	res, err := s.DB.Exec(`DELETE FROM posts WHERE id = ?`+scope, append([]any{pathID(r)}, args...)...)
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
		"users":      publicUsersConfig(s.Config.Typed()),
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

// publicUsersConfig 公开下发的用户系统开关（不含任何发信 / OAuth 密钥）。
func publicUsersConfig(t config.Typed) config.Map {
	u := t.Users
	return config.Map{
		"enabled": u.Enabled,
		"readers": config.Map{"enabled": u.ReadersOn(), "signup": u.Readers.Signup},
		"authors": config.Map{"enabled": u.AuthorsOn(), "directPublish": u.AuthorsOn() && u.Authors.DirectPublish},
		"comments": config.Map{
			"enabled": u.CommentsOn(), "anonymous": u.CommentsOn() && u.Comments.Anonymous, "moderation": u.Comments.Moderation,
		},
		"login":     config.Map{"github": t.GitHubLoginReady(), "mailReset": t.Mail.Ready()},
		"reactions": u.ReactionsOn(),
	}
}
