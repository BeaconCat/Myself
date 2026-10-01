package httpapi

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"myself/server/internal/imaging"
)

// maxMediaTitle 素材显示名的最大长度（字符数）
const maxMediaTitle = 120

// cleanMediaTitle 素材显示名：取文件名部分（去掉路径），剔除控制字符与路径分隔符，截断到上限。
// 显示名只用于展示与打包下载时的文件名；磁盘上的文件名（引用 URL）不变。
func cleanMediaTitle(raw string) string {
	if i := strings.LastIndexAny(raw, `/\`); i >= 0 {
		raw = raw[i+1:]
	}
	raw = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || strings.ContainsRune(`/\:*?"<>|`, r) {
			return -1
		}
		return r
	}, raw)
	raw = strings.TrimSpace(raw)
	if utf8.RuneCountInString(raw) > maxMediaTitle {
		raw = string([]rune(raw)[:maxMediaTitle])
	}
	return raw
}

// mediaTitle 素材显示名（没有记录时为空串）。
func (s *Server) mediaTitle(name string) string {
	var title string
	_ = s.DB.QueryRow(`SELECT COALESCE(title, '') FROM media WHERE name = ?`, name).Scan(&title)
	return title
}

// PUT /admin/media/{name} 重命名（只改显示名）：{"title": "..."}，空串恢复为文件名
func (s *Server) updateMedia(w http.ResponseWriter, r *http.Request) {
	name := safeName(r.PathValue("name"))
	if name == "" || !fileExists(filepath.Join(s.UploadDir, name)) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	var b struct {
		Title string `json:"title"`
	}
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	title := cleanMediaTitle(b.Title)
	if _, err := s.DB.Upsert("media", []string{"name", "title"}, []string{"name"}, []string{"title"}, name, title); err != nil {
		fail(w, err)
		return
	}
	item, err := s.fileInfo(name)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// removeMedia 删除一张素材（连同原图、缩略图与记录）；不存在返回 false。
func (s *Server) removeMedia(name string) (bool, error) {
	if !fileExists(filepath.Join(s.UploadDir, name)) {
		return false, nil
	}
	os.Remove(filepath.Join(s.UploadDir, name))
	os.Remove(filepath.Join(s.originalsDir, name))
	s.clearCompression(name)
	s.removeThumb(name)
	_, err := s.DB.Exec(`DELETE FROM media WHERE name = ?`, name)
	return true, err
}

// batchNames 读取 {"names": [...]}：去重、校验文件名，最多 500 个。
func batchNames(w http.ResponseWriter, r *http.Request) ([]string, bool) {
	var b struct {
		Names []string `json:"names"`
	}
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return nil, false
	}
	seen := map[string]bool{}
	var out []string
	for _, n := range b.Names {
		if n = safeName(n); n != "" && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	if len(out) == 0 || len(out) > 500 {
		writeError(w, http.StatusBadRequest, "invalid_names")
		return nil, false
	}
	return out, true
}

// POST /admin/media/delete 批量删除：{"names": [...]} → {"deleted": n}
func (s *Server) deleteMediaBatch(w http.ResponseWriter, r *http.Request) {
	names, ok := batchNames(w, r)
	if !ok {
		return
	}
	n := 0
	for _, name := range names {
		done, err := s.removeMedia(name)
		if err != nil {
			fail(w, err)
			return
		}
		if done {
			n++
		}
	}
	writeJSON(w, http.StatusOK, map[string]int{"deleted": n})
}

// zipEntryName 打包时的文件名：优先显示名；显示名的扩展名与真实格式不符（或没有）时补上真实扩展名；
// 重名（不区分大小写）追加「 (2)」「 (3)」…
func zipEntryName(title, name string, used map[string]bool) string {
	ext := filepath.Ext(name)
	base := title
	if base == "" {
		base = name
	}
	if te := imaging.Ext(base); te == imaging.Ext(name) || (te == ".jpeg" && ext == ".jpg") || (te == ".jpg" && ext == ".jpeg") {
		ext = filepath.Ext(base)
		base = strings.TrimSuffix(base, ext)
	}
	out := base + ext
	for i := 2; used[strings.ToLower(out)]; i++ {
		out = fmt.Sprintf("%s (%d)%s", base, i, ext)
	}
	used[strings.ToLower(out)] = true
	return out
}

// POST /admin/media/zip 打包下载：{"names": [...]} → application/zip（原文件，按显示名命名）
func (s *Server) zipMedia(w http.ResponseWriter, r *http.Request) {
	names, ok := batchNames(w, r)
	if !ok {
		return
	}
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(15 * time.Minute))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="media-%s.zip"`, time.Now().Format("20060102-150405")))
	zw := zip.NewWriter(w)
	used := map[string]bool{}
	for _, name := range names {
		path := filepath.Join(s.UploadDir, name)
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		st, _ := f.Stat()
		hdr := &zip.FileHeader{Name: zipEntryName(s.mediaTitle(name), name, used), Method: zip.Store}
		if st != nil {
			hdr.Modified = st.ModTime()
		}
		// 图片本身已压缩，直接存储
		dst, err := zw.CreateHeader(hdr)
		if err == nil {
			_, _ = io.Copy(dst, f)
		}
		f.Close()
	}
	_ = zw.Close()
}
