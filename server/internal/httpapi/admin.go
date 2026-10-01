package httpapi

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"myself/server/internal/config"
	"myself/server/internal/store"
)

var slugRe = regexp.MustCompile(`^[a-z0-9-]{1,80}$`)

// postInput 是文章写入的规范化字段。
type postInput struct {
	Slug, Title, Excerpt, ContentMd, Status, PublishAt string
	Covers, Tags                                       []string
	Pinned                                             bool
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
	if b.strOr("status") == "scheduled" {
		status = "scheduled"
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
			in.PublishAt = ""
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
	s.publicationMu.Lock()
	defer s.publicationMu.Unlock()
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
	var pubErr error
	in.Status, in.PublishAt, pubErr = s.publicationInput(b, "published", "")
	if pubErr != nil {
		writeError(w, http.StatusBadRequest, pubErr.Error())
		return
	}
	s.authorStatus(r, &in)
	var author any
	if u := userOf(r); u != nil && u.Role == store.RoleAuthor {
		author = u.ID
	}
	res, err := s.DB.Exec(`INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, status, pinned, author_id, publish_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.Slug, in.Title, in.Excerpt, in.ContentMd, store.JSONStrings(in.Covers), store.JSONStrings(in.Tags),
		in.Status, boolInt(in.Pinned), author, in.PublishAt)
	if store.IsUniqueErr(err) {
		writeError(w, http.StatusConflict, "slug_exists")
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "status": in.Status, "publishAt": store.PublishTime(in.PublishAt)})
}

// PUT /admin/posts/{id} 更新
func (s *Server) adminUpdatePost(w http.ResponseWriter, r *http.Request) {
	s.publicationMu.Lock()
	defer s.publicationMu.Unlock()
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
	scope, args := authorScope(r)
	previous, err := s.DB.GetPost(`WHERE id = ?`+scope, append([]any{pathID(r)}, args...)...)
	if err != nil {
		fail(w, err)
		return
	}
	if previous == nil {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	in.Status, in.PublishAt, err = s.publicationInput(b, previous.Status, previous.PublishAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.authorStatus(r, &in)
	res, err := s.DB.Exec(`UPDATE posts SET
		slug = ?, title = ?, excerpt = ?, content_md = ?, covers = ?, tags = ?, status = ?, pinned = ?, publish_at = ?,
		updated_at = CURRENT_TIMESTAMP WHERE id = ?`+scope,
		append([]any{in.Slug, in.Title, in.Excerpt, in.ContentMd, store.JSONStrings(in.Covers), store.JSONStrings(in.Tags),
			in.Status, boolInt(in.Pinned), in.PublishAt, pathID(r)}, args...)...)
	if store.IsUniqueErr(err) {
		writeError(w, http.StatusConflict, "slug_exists")
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	publicationSaved(w, res, in.Status, in.PublishAt)
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
	ContentMd, Mood, Status, PublishAt string
	Images                             []string
	Pinned                             bool
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
	s.publicationMu.Lock()
	defer s.publicationMu.Unlock()
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
	var pubErr error
	in.Status, in.PublishAt, pubErr = s.publicationInput(b, "published", "")
	if pubErr != nil {
		writeError(w, http.StatusBadRequest, pubErr.Error())
		return
	}
	res, err := s.DB.Exec(`INSERT INTO notes (content_md, mood, images, pinned, status, publish_at) VALUES (?, ?, ?, ?, ?, ?)`,
		in.ContentMd, in.Mood, store.JSONStrings(in.Images), boolInt(in.Pinned), in.Status, in.PublishAt)
	if err != nil {
		fail(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "status": in.Status, "publishAt": store.PublishTime(in.PublishAt)})
}

// PUT /admin/notes/{id} 编辑随想
func (s *Server) adminUpdateNote(w http.ResponseWriter, r *http.Request) {
	s.publicationMu.Lock()
	defer s.publicationMu.Unlock()
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
	previous, err := s.DB.QueryNotes(`WHERE id = ?`, true, pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	if len(previous) == 0 {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	in.Status, in.PublishAt, err = s.publicationInput(b, previous[0].Status, previous[0].PublishAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.DB.Exec(`UPDATE notes SET content_md = ?, mood = ?, images = ?, pinned = ?, status = ?, publish_at = ? WHERE id = ?`,
		in.ContentMd, in.Mood, store.JSONStrings(in.Images), boolInt(in.Pinned), in.Status, in.PublishAt, pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	publicationSaved(w, res, in.Status, in.PublishAt)
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
		"motion":     cfg["motion"],
		"thoughts":   cfg["thoughts"],
		"covers":     cfg["covers"],
		"timezone":   cfg["timezone"],
		"github":     github,
		"about":      cfg["about"],
		"users":      publicUsersConfig(s.Config.Typed()),
		"needsSetup": need,
	})
}

// secretMask 设置里密钥类字段回显用的占位：浏览器拿不到明文（后台万一被 XSS 也偷不走）；
// 保存时收到原样的占位视为「不修改」，清空才是删除
const secretMask = "••••••••"

// secretFields 只写不读的配置字段（段 → 子段… → 字段）
var secretFields = [][]string{{"mail", "password"}, {"oauth", "github", "clientSecret"}, {"github", "token"}}

// walkSecret 找到密钥字段所在的父对象（不存在时返回 nil）
func walkSecret(m config.Map, path []string) config.Map {
	for _, k := range path[:len(path)-1] {
		next, ok := m[k].(config.Map)
		if !ok {
			return nil
		}
		m = next
	}
	return m
}

// maskSecrets 就地把已设置的密钥字段替换为占位
func maskSecrets(cfg config.Map) config.Map {
	for _, p := range secretFields {
		if parent := walkSecret(cfg, p); parent != nil {
			if v, _ := parent[p[len(p)-1]].(string); v != "" {
				parent[p[len(p)-1]] = secretMask
			}
		}
	}
	return cfg
}

// GET /admin/settings 完整配置（密钥字段打码）
func (s *Server) adminGetSettings(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, maskSecrets(s.Config.Get()))
}

// PUT /admin/settings 部分更新（深合并）
func (s *Server) adminSaveSettings(w http.ResponseWriter, r *http.Request) {
	var patch config.Map
	if err := readJSON(w, r, &patch); err != nil || patch == nil {
		writeError(w, http.StatusBadRequest, "invalid_config")
		return
	}
	// 密钥字段回传的仍是打码占位：去掉，保留原值
	for _, sp := range secretFields {
		if parent := walkSecret(patch, sp); parent != nil && parent[sp[len(sp)-1]] == secretMask {
			delete(parent, sp[len(sp)-1])
		}
	}
	// 改站点名称时，仍沿用旧名的启动文案 / 发件人名称跟着改
	config.FollowSiteTitle(s.Config.Get(), patch)
	oldName := aboutName(s.Config.Get())
	merged, err := s.Config.Save(patch)
	if err != nil {
		fail(w, err)
		return
	}
	// 身份名字改了：仍沿用旧名字的站长账号昵称跟着改（评论、留言署名一致）；人为改过的不动
	if newName := aboutName(merged); oldName != "" && newName != "" && newName != oldName {
		if _, err := s.DB.Exec(`UPDATE users SET name = ? WHERE role = 'admin' AND name = ?`, limitRunes(newName, 40), oldName); err != nil {
			log.Printf("[settings] follow identity name: %v", err)
		}
	}
	writeJSON(w, http.StatusOK, maskSecrets(merged))
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

// aboutName 配置里的身份名字（about.name，去首尾空白）。
func aboutName(cfg config.Map) string {
	about, _ := cfg["about"].(config.Map)
	return strings.TrimSpace(str(about["name"]))
}

// GET /admin/notes/{id} includes drafts and scheduled notes for editing.
func (s *Server) adminGetNote(w http.ResponseWriter, r *http.Request) {
	notes, err := s.DB.QueryNotes(`WHERE id = ?`, true, pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	if len(notes) == 0 {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	writeJSON(w, http.StatusOK, notes[0])
}
