package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"myself/server/internal/imaging"
)

const (
	maxUploadFiles = 20
	// maxUploadBytes 图片单文件上限（非图片见 maxFileBytes）
	maxUploadBytes = 20 << 20
)

var safeNameRe = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

// winReserved Windows 保留设备名（不区分大小写，带扩展名同样生效，如 NUL.png）
var winReserved = regexp.MustCompile(`(?i)^(con|prn|aux|nul|com[1-9]|lpt[1-9])(\.|$)`)

// safeName 防路径穿越；非法返回空串。
func safeName(name string) string {
	if safeNameRe.MatchString(name) && !strings.Contains(name, "..") && !winReserved.MatchString(name) {
		return name
	}
	return ""
}

type mediaItem struct {
	Name string `json:"name"`
	// Title 显示名（上传时的原文件名或重命名后的名字）；为空时前端显示 Name
	Title string `json:"title"`
	// Folder 所在文件夹（空 = 未归类）
	Folder string `json:"folder"`
	// Kind 类别：image / video / audio / archive / file
	Kind string `json:"kind"`
	// Ext 小写扩展名（不带点）
	Ext string `json:"ext"`
	// Mime 直出时的 Content-Type
	Mime        string        `json:"mime"`
	URL         string        `json:"url"`
	Thumb       string        `json:"thumb"`
	Size        int64         `json:"size"`
	HasOriginal bool          `json:"hasOriginal"`
	Crop        *imaging.Rect `json:"crop"`
	CreatedAt   string        `json:"createdAt"`
	// Compressed 压缩记录（可回退）；未压缩为 nil
	Compressed *mediaCompression `json:"compressed"`
	// Duplicate 与已有素材内容相同：没有新存一份，返回的是已有的那张
	Duplicate bool `json:"duplicate,omitempty"`
}

func (s *Server) fileInfo(name string) (mediaItem, error) {
	stat, err := os.Stat(filepath.Join(s.UploadDir, name))
	if err != nil {
		return mediaItem{}, err
	}
	kind, ext, mime, _ := mediaKind(name)
	item := mediaItem{
		Name:      name,
		Kind:      kind,
		Ext:       ext,
		Mime:      mime,
		URL:       "/uploads/" + name,
		Size:      stat.Size(),
		CreatedAt: isoTime(stat.ModTime()),
	}
	// 缩略图只有图片才有
	if kind == kindImage {
		item.Thumb = thumbURL(name)
	}
	if _, err := os.Stat(filepath.Join(s.originalsDir, name)); err == nil {
		item.HasOriginal = true
	}
	item.Compressed = s.compressionOf(name)
	var cropJSON sql.NullString
	var createdAt, title string
	var folder string
	err = s.DB.QueryRow(`SELECT crop_json, created_at, COALESCE(title, ''), COALESCE(folder, '') FROM media WHERE name = ?`, name).Scan(&cropJSON, &createdAt, &title, &folder)
	if err == nil {
		item.CreatedAt = mediaTime(createdAt, stat.ModTime())
		item.Title = title
		item.Folder = folder
		if cropJSON.Valid && cropJSON.String != "" {
			var rect imaging.Rect
			if json.Unmarshal([]byte(cropJSON.String), &rect) == nil {
				item.Crop = &rect
			}
		}
	} else if err != sql.ErrNoRows {
		return mediaItem{}, err
	}
	return item, nil
}

// listUploads 列出公开素材文件名（跳过点文件、子目录与不接受的类型）；filter 为 nil 时列出所有类别。
func (s *Server) listUploads(filter map[string]bool) ([]string, error) {
	entries, err := os.ReadDir(s.UploadDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		// 用户头像（avatar-*）由个人资料管理，不进素材库
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), avatarPrefix) {
			continue
		}
		if _, _, _, ok := mediaKind(e.Name()); !ok || (filter != nil && !filter[imaging.Ext(e.Name())]) {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}

// GET /admin/media 素材列表
func (s *Server) listMedia(w http.ResponseWriter, _ *http.Request) {
	names, err := s.listUploads(nil)
	if err != nil {
		fail(w, err)
		return
	}
	items := make([]mediaItem, 0, len(names))
	for _, name := range names {
		item, err := s.fileInfo(name)
		if err != nil {
			fail(w, err)
			return
		}
		items = append(items, item)
	}
	// 同一时刻（如初始化时批量导入）按文件名兜底，保证顺序稳定
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt != items[j].CreatedAt {
			return items[i].CreatedAt > items[j].CreatedAt
		}
		return items[i].Name > items[j].Name
	})
	writeJSON(w, http.StatusOK, items)
}

// mediaTime 素材的上传时间（ISO，UTC）：取记录时间与文件修改时间中较早的那个。
// media 行可能晚于文件才建（首次裁切 / 移动 / 回退时才插入），这时文件时间才是真正的上传时间；
// 两者格式不同（SQLite datetime 与 ISO），统一解析后再比较，避免按字符串排序错位。
func mediaTime(recorded string, mod time.Time) string {
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", recorded, time.UTC); err == nil && t.Before(mod) {
		return isoTime(t)
	}
	return isoTime(mod)
}

func randomFilename(ext string) string {
	raw := make([]byte, 4)
	_, _ = rand.Read(raw)
	return strconv.FormatInt(time.Now().UnixMilli(), 36) + "-" + hex.EncodeToString(raw) + ext
}

// POST /admin/media 上传（单个/批量，字段 files）
// 图片校验文件头、上限 20MB；视频 / 音频 / 压缩包 / 其它文件只看扩展名、上限 512MB。
// 不接受的类型（可执行 / 标记类、无扩展名）静默跳过；一个都没存下且有被拒的，回 400 unsupported_type。
func (s *Server) uploadMedia(w http.ResponseWriter, r *http.Request) {
	// 全局 ReadTimeout 较短；大批量上传单独放宽
	_ = http.NewResponseController(w).SetReadDeadline(time.Now().Add(15 * time.Minute))
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxUploadFiles)*maxFileBytes+1<<20)
	reader, err := r.MultipartReader()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_multipart")
		return
	}
	var saved []string
	dups := map[string]bool{}
	rejected := 0
	// 目标文件夹：表单字段 folder（须排在文件之前）
	folder := ""
	for len(saved) < maxUploadFiles {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_multipart")
			return
		}
		if part.FormName() == "folder" && part.FileName() == "" {
			raw, _ := io.ReadAll(io.LimitReader(part, 512))
			f, ok := cleanFolder(string(raw))
			if !ok {
				writeError(w, http.StatusBadRequest, "invalid_folder")
				return
			}
			if err := s.ensureFolder(f); err != nil {
				fail(w, err)
				return
			}
			folder = f
			continue
		}
		if part.FormName() != "files" || part.FileName() == "" {
			continue
		}
		kind, ext, _, ok := mediaKind(part.FileName())
		if !ok {
			rejected++
			_, _ = io.Copy(io.Discard, part)
			continue
		}
		limit := int64(maxFileBytes)
		if kind == kindImage {
			limit = maxUploadBytes
		}
		name := randomFilename("." + ext)
		dst, err := os.Create(filepath.Join(s.UploadDir, name))
		if err != nil {
			fail(w, err)
			return
		}
		// 边写边算哈希，用于查重
		h := sha256.New()
		n, err := io.Copy(io.MultiWriter(dst, h), io.LimitReader(part, limit+1))
		dst.Close()
		if err != nil || n > limit {
			os.Remove(filepath.Join(s.UploadDir, name))
			writeError(w, http.StatusRequestEntityTooLarge, "file_too_large")
			return
		}
		// 图片：文件头必须与扩展名一致、像素数在上限内；否则不入库。
		// 其它类别不解析内容：直出时固定 Content-Type 并附带 nosniff，非媒体一律作附件下载
		if kind == kindImage {
			if _, err := imaging.Meta(filepath.Join(s.UploadDir, name)); err != nil {
				os.Remove(filepath.Join(s.UploadDir, name))
				if errors.Is(err, imaging.ErrTooLarge) {
					writeError(w, http.StatusRequestEntityTooLarge, "image_too_large")
				} else {
					writeError(w, http.StatusBadRequest, "invalid_image")
				}
				return
			}
		}
		sum := hex.EncodeToString(h.Sum(nil))
		// 与已有素材内容相同：丢弃新文件，返回已有的那张
		if err := s.ensureMediaHashes(); err != nil {
			fail(w, err)
			return
		}
		title := cleanMediaTitle(part.FileName())
		if existing := s.mediaByHash(sum, name); existing != "" {
			os.Remove(filepath.Join(s.UploadDir, name))
			_, _ = s.DB.Exec(`DELETE FROM media WHERE name = ?`, name)
			// 已有的那张还没有显示名时，补上这次上传的原文件名
			_, _ = s.DB.Exec(`UPDATE media SET title = ? WHERE name = ? AND COALESCE(title, '') = ''`, title, existing)
			saved = append(saved, existing)
			dups[existing] = true
			continue
		}
		if _, err := s.DB.Exec(`INSERT INTO media (name, sha256, title, folder) VALUES (?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET sha256 = excluded.sha256, title = excluded.title, folder = excluded.folder`, name, sum, title, folder); err != nil {
			fail(w, err)
			return
		}
		saved = append(saved, name)
	}
	if len(saved) == 0 && rejected > 0 {
		writeError(w, http.StatusBadRequest, "unsupported_type")
		return
	}
	items := make([]mediaItem, 0, len(saved))
	for _, name := range saved {
		item, err := s.fileInfo(name)
		if err != nil {
			fail(w, err)
			return
		}
		item.Duplicate = dups[name]
		items = append(items, item)
	}
	writeJSON(w, http.StatusCreated, items)
}

// copyIfMissing 首次处理时把当前文件备份为原图。
func copyIfMissing(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// POST /admin/media/{name}/crop {left,top,width,height}
// 首次裁切时把当前文件备份为原图；此后一律从原图重裁并替换公开文件，裁切框入库。
func (s *Server) cropMedia(w http.ResponseWriter, r *http.Request) {
	name := safeName(r.PathValue("name"))
	publicPath := filepath.Join(s.UploadDir, name)
	if name == "" || !fileExists(publicPath) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if !isImage(name) {
		writeError(w, http.StatusBadRequest, "not_image")
		return
	}
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_rect")
		return
	}
	rect := imaging.Rect{
		Left:   max(0, roundInt(b.num("left"))),
		Top:    max(0, roundInt(b.num("top"))),
		Width:  roundInt(b.num("width")),
		Height: roundInt(b.num("height")),
	}
	if rect.Width <= 0 || rect.Height <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_rect")
		return
	}
	originalPath := filepath.Join(s.originalsDir, name)
	if err := copyIfMissing(publicPath, originalPath); err != nil {
		fail(w, err)
		return
	}
	decodeSlots <- struct{}{}
	src, err := imaging.Decode(originalPath)
	<-decodeSlots
	if err != nil {
		writeError(w, http.StatusBadRequest, "decode_failed")
		return
	}
	bounds := src.Bounds()
	rect.Width = min(rect.Width, bounds.Dx()-rect.Left)
	rect.Height = min(rect.Height, bounds.Dy()-rect.Top)
	if rect.Width <= 0 || rect.Height <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_rect")
		return
	}
	data, err := imaging.EncodeBytes(imaging.Crop(src, rect), imaging.Ext(name), 90)
	if err != nil {
		fail(w, err)
		return
	}
	if err := os.WriteFile(publicPath, data, 0o644); err != nil {
		fail(w, err)
		return
	}
	cropJSON, _ := json.Marshal(rect)
	s.clearCompression(name)
	if _, err := s.DB.Exec(`INSERT INTO media (name, crop_json) VALUES (?, ?)
		ON CONFLICT(name) DO UPDATE SET crop_json = excluded.crop_json`, name, string(cropJSON)); err != nil {
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

// GET /admin/media/{name}/original 原图（重裁 UI 用）
func (s *Server) mediaOriginal(w http.ResponseWriter, r *http.Request) {
	name := safeName(r.PathValue("name"))
	if name == "" || !isImage(name) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	target := filepath.Join(s.originalsDir, name)
	if !fileExists(target) {
		target = filepath.Join(s.UploadDir, name)
	}
	if !fileExists(target) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	http.ServeFile(w, r, target)
}

// DELETE /admin/media/{name} 删除（连同原图与裁切记录）
func (s *Server) deleteMedia(w http.ResponseWriter, r *http.Request) {
	name := safeName(r.PathValue("name"))
	if name == "" {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	done, err := s.removeMedia(name)
	if err != nil {
		fail(w, err)
		return
	}
	if !done {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func roundInt(f float64) int {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}
	return int(math.Round(f))
}
