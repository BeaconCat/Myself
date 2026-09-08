package httpapi

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"

	"myself/server/internal/store"
)

type pagedPosts struct {
	Items    []store.Post `json:"items"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
	Total    int          `json:"total"`
}

func toPosts(rows []store.PostRow, o store.PostOpts) []store.Post {
	out := make([]store.Post, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.ToPost(o))
	}
	return out
}

// GET /posts?page=&pageSize=&tag=&q= 文章列表（分页 + 标签过滤 + 关键词搜索）
func (s *Server) listPosts(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1, 1, 1<<30)
	pageSize := queryInt(r, "pageSize", 10, 1, 50)
	tag := r.URL.Query().Get("tag")
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	where := []string{"status = 'published'"}
	var args []any
	if tag != "" {
		// tags 为 JSON 数组文本，用引号包裹精确匹配单个标签
		where = append(where, "tags LIKE ?")
		args = append(args, `%"`+strings.ReplaceAll(tag, `"`, "")+`"%`)
	}
	if q != "" {
		where = append(where, "(title LIKE ? OR excerpt LIKE ? OR content_md LIKE ?)")
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM posts WHERE `+whereSQL, args...).Scan(&total); err != nil {
		fail(w, err)
		return
	}
	rows, err := s.DB.QueryPosts(`WHERE `+whereSQL+` ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		append(args, pageSize, (page-1)*pageSize)...)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pagedPosts{
		Items: toPosts(rows, store.PostOpts{}), Page: page, PageSize: pageSize, Total: total,
	})
}

// GET /posts/{slug} 文章详情（含 Markdown 正文）
func (s *Server) getPost(w http.ResponseWriter, r *http.Request) {
	row, err := s.DB.GetPost(`WHERE slug = ? AND status = 'published'`, r.PathValue("slug"))
	if err != nil {
		fail(w, err)
		return
	}
	if row == nil {
		writeError(w, http.StatusNotFound, "post_not_found")
		return
	}
	writeJSON(w, http.StatusOK, row.ToPost(store.PostOpts{WithContent: true}))
}

// GET /hero 首页轮播条目：按配置规则取（最新 n 条，置顶优先/无视）
func (s *Server) hero(w http.ResponseWriter, _ *http.Request) {
	hero := s.Config.Typed().Hero
	count := hero.Count
	if count == 0 {
		count = 4
	}
	count = clamp(count, 1, 10)
	order := "created_at DESC"
	if hero.PinnedRule == "pinned-first" {
		order = "pinned DESC, created_at DESC"
	}
	rows, err := s.DB.QueryPosts(`WHERE status = 'published' ORDER BY `+order+` LIMIT ?`, count)
	if err != nil {
		fail(w, err)
		return
	}
	interval := hero.IntervalMs
	if interval == 0 {
		interval = 3000
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"intervalMs": max(1000, interval),
		"items":      toPosts(rows, store.PostOpts{}),
	})
}

// GET /tags 标签及计数（SQL 层聚合，按最近使用排序）
func (s *Server) tags(w http.ResponseWriter, _ *http.Request) {
	out, err := s.DB.TagCounts()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// GET /notes?page=&pageSize=&q=&media=1&from=&to= 随想信息流
func (s *Server) listNotes(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	page := queryInt(r, "page", 1, 1, 1<<30)
	pageSize := queryInt(r, "pageSize", 20, 1, 50)
	q := strings.TrimSpace(qs.Get("q"))

	where := []string{"1=1"}
	var args []any
	if q != "" {
		where = append(where, "(content_md LIKE ? OR mood LIKE ?)")
		like := "%" + q + "%"
		args = append(args, like, like)
	}
	if qs.Get("media") == "1" {
		where = append(where, "images != '[]'")
	}
	if from := qs.Get("from"); dateRe.MatchString(from) {
		where = append(where, "created_at >= ?")
		args = append(args, from)
	}
	if to := qs.Get("to"); dateRe.MatchString(to) {
		where = append(where, "created_at < date(?, '+1 day')")
		args = append(args, to)
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM notes WHERE `+whereSQL, args...).Scan(&total); err != nil {
		fail(w, err)
		return
	}
	items, err := s.DB.QueryNotes(`WHERE `+whereSQL+` ORDER BY pinned DESC, created_at DESC LIMIT ? OFFSET ?`,
		true, append(args, pageSize, (page-1)*pageSize)...)
	if err != nil {
		fail(w, err)
		return
	}
	if items == nil {
		items = []store.Note{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items, "page": page, "pageSize": pageSize, "total": total,
	})
}

var hexRe = regexp.MustCompile(`^[0-9a-fA-F]{6}$`)

// GET /img/{from}/{to}/{label} 本地渐变占位图（演示配图，无外部资源）
func (s *Server) placeholderImage(w http.ResponseWriter, r *http.Request) {
	from, to := r.PathValue("from"), r.PathValue("to")
	if !hexRe.MatchString(from) || !hexRe.MatchString(to) {
		writeError(w, http.StatusBadRequest, "bad_color")
		return
	}
	label := []rune(r.PathValue("label"))
	if len(label) > 24 {
		label = label[:24]
	}
	safe := html.EscapeString(strings.NewReplacer("<", "", ">", "", "&", "", `"`, "", "'", "").Replace(string(label)))
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="900" height="900">
  <defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
    <stop offset="0" stop-color="#%s"/><stop offset="1" stop-color="#%s"/>
  </linearGradient></defs>
  <rect width="900" height="900" fill="url(#g)"/>
  <text x="50%%" y="52%%" text-anchor="middle" font-family="sans-serif" font-size="52"
    fill="rgba(255,255,255,.85)" font-weight="700">%s</text>
</svg>`, from, to, safe)
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write([]byte(svg))
}
