package httpapi

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"myself/server/internal/store"
)

// yamlStr 双引号 JSON 字符串同时是合法的 YAML 标量，省去转义规则的差异
func yamlStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func yamlList(items []string) string {
	q := make([]string, len(items))
	for i, s := range items {
		q[i] = yamlStr(s)
	}
	return "[" + strings.Join(q, ", ") + "]"
}

// frontMatter 按顺序写出 YAML front matter（字段名取 Hexo / Hugo / Jekyll 通用的写法，便于迁出）
func frontMatter(fields [][2]string) string {
	var b strings.Builder
	b.WriteString("---\n")
	for _, f := range fields {
		if f[1] == "" {
			continue
		}
		b.WriteString(f[0] + ": " + f[1] + "\n")
	}
	b.WriteString("---\n\n")
	return b.String()
}

func postMarkdown(p store.PostRow) string {
	fields := [][2]string{
		{"title", yamlStr(p.Title)},
		{"slug", p.Slug},
		{"date", yamlStr(p.CreatedAt)},
		{"updated", yamlStr(p.UpdatedAt)},
	}
	if tags := store.ParseStrings(p.Tags); len(tags) > 0 {
		fields = append(fields, [2]string{"tags", yamlList(tags)})
	}
	if covers := store.ParseStrings(p.Covers); len(covers) > 0 {
		fields = append(fields, [2]string{"cover", yamlStr(covers[0])}, [2]string{"covers", yamlList(covers)})
	}
	if p.Excerpt != "" {
		fields = append(fields, [2]string{"excerpt", yamlStr(p.Excerpt)})
	}
	if p.Status == "draft" {
		fields = append(fields, [2]string{"draft", "true"})
	}
	if p.Pinned != 0 {
		fields = append(fields, [2]string{"pinned", "true"})
	}
	if p.Hidden != 0 {
		fields = append(fields, [2]string{"hidden", "true"})
	}
	return frontMatter(fields) + strings.TrimSpace(p.ContentMd) + "\n"
}

// GET /admin/export/markdown?media=1 导出全部文章与随想为 Markdown（zip）；media=1 时连同素材一起打包，
// 正文里的 /uploads/… 地址与包内 uploads/ 目录对应。
func (s *Server) exportMarkdown(w http.ResponseWriter, r *http.Request) {
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(30 * time.Minute))
	posts, err := s.DB.QueryPosts(`ORDER BY created_at`)
	if err != nil {
		fail(w, err)
		return
	}
	notes, err := s.DB.QueryNotes(`ORDER BY created_at`, true)
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="myself-markdown-%s.zip"`, time.Now().Format("20060102-150405")))
	zw := zip.NewWriter(w)
	defer zw.Close()

	add := func(name, content string) {
		if f, err := zw.Create(name); err == nil {
			_, _ = io.WriteString(f, content)
		}
	}
	for _, p := range posts {
		add("posts/"+p.Slug+".md", postMarkdown(p))
	}
	for _, n := range notes {
		fields := [][2]string{{"date", yamlStr(n.CreatedAt)}}
		if n.Mood != "" {
			fields = append(fields, [2]string{"mood", yamlStr(n.Mood)})
		}
		if len(n.Images) > 0 {
			fields = append(fields, [2]string{"images", yamlList(n.Images)})
		}
		if n.Pinned != nil && *n.Pinned {
			fields = append(fields, [2]string{"pinned", "true"})
		}
		if n.Hidden {
			fields = append(fields, [2]string{"hidden", "true"})
		}
		day := strings.ReplaceAll(strings.SplitN(n.CreatedAt, " ", 2)[0], "/", "-")
		add(fmt.Sprintf("notes/%s-%d.md", day, n.ID), frontMatter(fields)+strings.TrimSpace(n.ContentMd)+"\n")
	}
	add("README.md", fmt.Sprintf("# Myself 导出\n\n- 文章 %d 篇（posts/），随想 %d 条（notes/）\n- 导出时间：%s\n- front matter 为 YAML；正文里的 /uploads/… 对应包内 uploads/ 目录（导出时勾选了素材才有）\n",
		len(posts), len(notes), time.Now().Format("2006-01-02 15:04:05")))

	if r.URL.Query().Get("media") != "1" {
		return
	}
	names, err := s.listUploads(nil)
	if err != nil {
		return
	}
	for _, name := range names {
		f, err := os.Open(filepath.Join(s.UploadDir, name))
		if err != nil {
			continue
		}
		hdr := &zip.FileHeader{Name: "uploads/" + name, Method: zip.Store}
		if st, err := f.Stat(); err == nil {
			hdr.Modified = st.ModTime()
		}
		if wr, err := zw.CreateHeader(hdr); err == nil {
			_, _ = io.Copy(wr, f)
		}
		f.Close()
	}
}
