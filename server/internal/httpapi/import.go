package httpapi

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"myself/server/internal/imaging"
	"myself/server/internal/store"
)

// 导入的体积上限：单文件 / 整批
const (
	importFileMax  = 64 << 20
	importTotalMax = 1 << 30
	importMaxFiles = 20000
)

// importSkipDir 构建产物、主题、依赖目录：不含文章，跳过
var importSkipDir = regexp.MustCompile(`(^|/)(node_modules|\.git|public|_site|resources|themes|\.github|\.obsidian|\.vscode)(/|$)`)

// importFS 上传的站点源码目录（相对路径 → 内容），已去掉共同的顶层目录
type importFS map[string][]byte

func (f importFS) has(p string) bool { _, ok := f[p]; return ok }

// readImportFS 读取上传内容：单个 .zip，或多个带相对路径的文件（浏览器选择文件夹时文件名即相对路径）。
func readImportFS(r *http.Request) (importFS, error) {
	mr, err := r.MultipartReader()
	if err != nil {
		return nil, errors.New("invalid_multipart")
	}
	files := importFS{}
	var total int64
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.New("invalid_multipart")
		}
		if part.FormName() != "files" || part.FileName() == "" {
			continue
		}
		// part.FileName() 只保留末段；选择文件夹上传时需要完整相对路径，直接读 Content-Disposition
		raw := part.FileName()
		if _, params, err := mime.ParseMediaType(part.Header.Get("Content-Disposition")); err == nil && params["filename"] != "" {
			raw = params["filename"]
		}
		name := cleanImportPath(raw)
		data, err := io.ReadAll(io.LimitReader(part, importFileMax+1))
		if err != nil {
			return nil, errors.New("invalid_multipart")
		}
		if int64(len(data)) > importFileMax {
			continue
		}
		total += int64(len(data))
		if total > importTotalMax {
			return nil, errors.New("too_large")
		}
		if strings.HasSuffix(strings.ToLower(name), ".zip") && len(files) == 0 {
			err := readImportZip(files, data)
			if err == nil {
				continue
			}
			if msg := err.Error(); msg == "too_large" || msg == "too_many_files" {
				return nil, err
			}
		}
		if len(files) >= importMaxFiles {
			return nil, errors.New("too_many_files")
		}
		if name != "" && !importSkipDir.MatchString(name) {
			files[name] = data
		}
	}
	return stripCommonRoot(files), nil
}

func readImportZip(files importFS, data []byte) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	if len(zr.File) > importMaxFiles {
		return errors.New("too_many_files")
	}
	var total int64
	for _, f := range zr.File {
		name := cleanImportPath(f.Name)
		if f.FileInfo().IsDir() || name == "" || importSkipDir.MatchString(name) || f.UncompressedSize64 > importFileMax {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		b, err := io.ReadAll(io.LimitReader(rc, importFileMax+1))
		rc.Close()
		// 按实际解压出的字节累计（声明的大小可作假），超出整批上限即中止
		total += int64(len(b))
		if total > importTotalMax {
			return errors.New("too_large")
		}
		if err == nil && int64(len(b)) <= importFileMax {
			files[name] = b
		}
	}
	return nil
}

func cleanImportPath(p string) string {
	p = path.Clean("/" + strings.ReplaceAll(p, `\`, "/"))
	return strings.TrimPrefix(p, "/")
}

// stripCommonRoot 选择文件夹上传时所有路径都带着文件夹名，去掉这一层
func stripCommonRoot(files importFS) importFS {
	var root string
	for p := range files {
		first, _, ok := strings.Cut(p, "/")
		if !ok {
			return files
		}
		if root == "" {
			root = first
		} else if root != first {
			return files
		}
	}
	if root == "" {
		return files
	}
	out := importFS{}
	for p, b := range files {
		out[strings.TrimPrefix(p, root+"/")] = b
	}
	return out
}

// detectPlatform 按目录结构识别来源；返回平台名与文章所在目录前缀（草稿目录另算）
func detectPlatform(files importFS) (string, []string, []string) {
	hasPrefix := func(pre string) bool {
		for p := range files {
			if strings.HasPrefix(p, pre) {
				return true
			}
		}
		return false
	}
	switch {
	case hasPrefix("source/_posts/"):
		return "hexo", []string{"source/_posts/"}, []string{"source/_drafts/"}
	case hasPrefix("_posts/"):
		return "jekyll", []string{"_posts/"}, []string{"_drafts/"}
	case hasPrefix("content/"):
		for _, d := range []string{"content/posts/", "content/post/", "content/blog/", "content/articles/"} {
			if hasPrefix(d) {
				return "hugo", []string{d}, nil
			}
		}
		return "hugo", []string{"content/"}, nil
	}
	return "markdown", []string{""}, nil
}

var (
	mdImageRe   = regexp.MustCompile(`!\[([^\]]*)\]\(\s*<?([^)\s>]+)>?(\s+"[^"]*")?\s*\)`)
	htmlImgRe   = regexp.MustCompile(`(<img[^>]*?\ssrc=["'])([^"']+)(["'])`)
	hexoAssetRe = regexp.MustCompile(`\{%\s*asset_img\s+(\S+)(?:\s+([^%]*?))?\s*%\}`)
	hugoFigRe   = regexp.MustCompile(`\{\{<\s*figure\s+[^>]*?src="([^"]+)"[^>]*?>\}\}`)
	moreRe      = regexp.MustCompile(`(?m)^\s*<!--\s*more\s*-->\s*$`)
	jekyllDate  = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)
	slugBadRe   = regexp.MustCompile(`[^a-z0-9]+`)
)

var importDateLayouts = []string{
	time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05 -0700", "2006-01-02 15:04:05 Z07:00",
	"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02", "2006/01/02 15:04:05", "2006/01/02",
}

// parseImportDate 解析常见日期写法；没有时区的按服务器本地时区理解，统一存 UTC
func parseImportDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	for _, l := range importDateLayouts {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Trim(slugBadRe.ReplaceAllString(s, "-"), "-")
	if len(s) > 80 {
		s = strings.Trim(s[:80], "-")
	}
	return s
}

type importItem struct {
	File   string `json:"file"`
	Title  string `json:"title"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
	Images int    `json:"images"`
	Reason string `json:"reason,omitempty"`
}

type importer struct {
	s        *Server
	files    importFS
	platform string
	saved    map[string]string // 源路径 → /uploads/… （同一张图只导一次）
	images   int
}

// resolveAsset 按各平台约定找到正文里引用的本地图片
func (im *importer) resolveAsset(mdPath, ref string) (string, bool) {
	if ref == "" || strings.HasPrefix(ref, "data:") || strings.Contains(ref, "://") || strings.HasPrefix(ref, "//") {
		return "", false
	}
	ref = strings.SplitN(strings.SplitN(ref, "?", 2)[0], "#", 2)[0]
	if u, err := url.PathUnescape(ref); err == nil {
		ref = u
	}
	dir := path.Dir(mdPath)
	base := strings.TrimSuffix(path.Base(mdPath), path.Ext(mdPath))
	var cands []string
	if strings.HasPrefix(ref, "/") {
		r := strings.TrimPrefix(ref, "/")
		cands = append(cands, "source/"+r, "static/"+r, "assets/"+r, r)
	} else {
		cands = append(cands,
			path.Join(dir, ref),       // 相对路径 / Hugo 页面包
			path.Join(dir, base, ref), // Hexo 资源文件夹（post_asset_folder）
			path.Join("source", ref), path.Join("static", ref), ref,
		)
	}
	for _, c := range cands {
		if im.files.has(c) {
			return c, true
		}
	}
	return "", false
}

// importAsset 把本地图片存进素材库（按内容查重），返回站内地址
func (im *importer) importAsset(p string) (string, bool) {
	if u, ok := im.saved[p]; ok {
		return u, true
	}
	data := im.files[p]
	ext := imaging.Ext(p)
	if !imageExt[ext] || len(data) == 0 || len(data) > maxUploadBytes {
		return "", false
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	_ = im.s.ensureMediaHashes()
	if existing := im.s.mediaByHash(hash, ""); existing != "" {
		u := "/uploads/" + existing
		im.saved[p] = u
		return u, true
	}
	name := randomFilename(ext)
	dst := filepath.Join(im.s.UploadDir, name)
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return "", false
	}
	if _, err := imaging.Meta(dst); err != nil {
		os.Remove(dst)
		return "", false
	}
	if _, err := im.s.DB.Exec(`INSERT INTO media (name, sha256, title) VALUES (?, ?, ?)`, name, hash, cleanMediaTitle(path.Base(p))); err != nil {
		os.Remove(dst)
		return "", false
	}
	u := "/uploads/" + name
	im.saved[p] = u
	im.images++
	return u, true
}

// rewriteBody 转换平台专有语法并把本地图片导入素材库、改写为站内地址
func (im *importer) rewriteBody(mdPath, body string) (string, int) {
	n := 0
	body = moreRe.ReplaceAllString(body, "")
	body = hexoAssetRe.ReplaceAllStringFunc(body, func(m string) string {
		g := hexoAssetRe.FindStringSubmatch(m)
		return fmt.Sprintf("![%s](%s)", strings.Trim(strings.TrimSpace(g[2]), `"'`), g[1])
	})
	body = hugoFigRe.ReplaceAllString(body, "![]($1)")
	fix := func(ref string) string {
		if p, ok := im.resolveAsset(mdPath, ref); ok {
			if u, ok := im.importAsset(p); ok {
				n++
				return u
			}
		}
		return ref
	}
	body = mdImageRe.ReplaceAllStringFunc(body, func(m string) string {
		g := mdImageRe.FindStringSubmatch(m)
		return fmt.Sprintf("![%s](%s%s)", g[1], fix(g[2]), g[3])
	})
	body = htmlImgRe.ReplaceAllStringFunc(body, func(m string) string {
		g := htmlImgRe.FindStringSubmatch(m)
		return g[1] + fix(g[2]) + g[3]
	})
	return strings.TrimSpace(body), n
}

// POST /admin/import?draft=1&dry=1 从 Hexo / Hugo / Jekyll（或任意带 front matter 的 Markdown 目录）导入文章。
// draft=1 全部导入为草稿（默认）；dry=1 只识别不写入，用于预览。
func (s *Server) importPosts(w http.ResponseWriter, r *http.Request) {
	_ = http.NewResponseController(w).SetReadDeadline(time.Now().Add(30 * time.Minute))
	r.Body = http.MaxBytesReader(w, r.Body, importTotalMax+(8<<20))
	files, err := readImportFS(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	asDraft := r.URL.Query().Get("draft") != "0"
	dry := r.URL.Query().Get("dry") == "1"

	platform, postDirs, draftDirs := detectPlatform(files)
	im := &importer{s: s, files: files, platform: platform, saved: map[string]string{}}

	var paths []string
	for p := range files {
		ext := strings.ToLower(path.Ext(p))
		if ext != ".md" && ext != ".markdown" {
			continue
		}
		base := strings.ToLower(path.Base(p))
		if base == "_index.md" || base == "readme.md" || base == "license.md" || base == "changelog.md" {
			continue
		}
		for _, d := range append(append([]string{}, postDirs...), draftDirs...) {
			if strings.HasPrefix(p, d) {
				paths = append(paths, p)
				break
			}
		}
	}
	sort.Strings(paths)

	usedSlugs := map[string]bool{}
	var items []importItem
	nCreated := 0
	for _, p := range paths {
		meta, body := parseFrontMatter(string(files[p]))
		it := importItem{File: p}
		title := fmString(meta, "title")
		if title == "" {
			if platform == "markdown" {
				it.Reason = "no_title"
				items = append(items, it)
				continue
			}
			title = strings.TrimSuffix(path.Base(p), path.Ext(p))
		}
		it.Title = title

		// 草稿：draft: true / published: false / 草稿目录 / 选项
		isDraft := asDraft
		if v, ok := fmBool(meta, "draft"); ok && v {
			isDraft = true
		}
		if v, ok := fmBool(meta, "published"); ok && !v {
			isDraft = true
		}
		for _, d := range draftDirs {
			if strings.HasPrefix(p, d) {
				isDraft = true
			}
		}
		it.Status = "published"
		if isDraft {
			it.Status = "draft"
		}

		// 日期：front matter → Jekyll 文件名 → 现在
		fileBase := strings.TrimSuffix(path.Base(p), path.Ext(p))
		if fileBase == "index" { // Hugo 页面包 / Hexo 目录形式
			fileBase = path.Base(path.Dir(p))
		}
		created := time.Now().UTC()
		if t, ok := parseImportDate(fmString(meta, "date", "publishdate", "published_at", "created")); ok {
			created = t
		} else if m := jekyllDate.FindStringSubmatch(fileBase); m != nil {
			if t, ok := parseImportDate(m[1]); ok {
				created = t
			}
		}
		updated := created
		if t, ok := parseImportDate(fmString(meta, "updated", "lastmod", "last_modified_at", "modified")); ok {
			updated = t
		}

		// slug：slug → permalink / url 末段 → 文件名（去掉 Jekyll 日期前缀）→ 日期兜底；重名追加序号
		slugSrc := fmString(meta, "slug", "abbrlink")
		if slugSrc == "" {
			if pl := strings.Trim(fmString(meta, "permalink", "url"), "/"); pl != "" {
				slugSrc = path.Base(pl)
			}
		}
		if slugSrc == "" {
			slugSrc = fileBase
			if m := jekyllDate.FindStringSubmatch(fileBase); m != nil {
				slugSrc = m[2]
			}
		}
		slug := slugify(slugSrc)
		if slug == "" {
			slug = "post-" + created.Format("20060102")
		}
		base := slug
		for i := 2; usedSlugs[slug] || s.slugTaken(slug); i++ {
			slug = fmt.Sprintf("%s-%d", base, i)
		}
		usedSlugs[slug] = true
		it.Slug = slug

		if dry {
			items = append(items, it)
			continue
		}
		content, nImg := im.rewriteBody(p, body)
		it.Images = nImg
		var covers []string
		if c := fmString(meta, "cover", "image", "thumbnail", "banner", "featured_image", "og_image", "images", "cover_image"); c != "" {
			if src, ok := im.resolveAsset(p, c); ok {
				if u, ok := im.importAsset(src); ok {
					covers = append(covers, u)
				}
			} else if strings.HasPrefix(c, "https://") {
				covers = append(covers, c)
			}
		}
		tags := fmStrings(meta, "tags", "categories", "category", "tag")
		excerpt := fmString(meta, "excerpt", "description", "summary", "subtitle")
		if _, err := s.DB.Exec(`INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			slug, title, excerpt, content, store.JSONStrings(covers), store.JSONStrings(tags), it.Status,
			created.Format("2006-01-02 15:04:05"), updated.Format("2006-01-02 15:04:05")); err != nil {
			it.Reason = "save_failed"
		} else {
			nCreated++
		}
		items = append(items, it)
	}
	if items == nil {
		items = []importItem{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"platform": platform, "dry": dry, "found": len(paths), "created": nCreated, "images": im.images, "items": items,
	})
}

func (s *Server) slugTaken(slug string) bool {
	var n int
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM posts WHERE slug = ?`, slug).Scan(&n)
	return n > 0
}
