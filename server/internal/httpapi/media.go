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

var allowedExt = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true}

const (
	maxUploadFiles = 20
	maxUploadBytes = 20 << 20
)

var safeNameRe = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

// safeName 防路径穿越；非法返回空串。
func safeName(name string) string {
	if safeNameRe.MatchString(name) && !strings.Contains(name, "..") {
		return name
	}
	return ""
}

type mediaItem struct {
	Name        string        `json:"name"`
	URL         string        `json:"url"`
	Thumb       string        `json:"thumb"`
	Size        int64         `json:"size"`
	HasOriginal bool          `json:"hasOriginal"`
	Crop        *imaging.Rect `json:"crop"`
	CreatedAt   string        `json:"createdAt"`
	// Duplicate 与已有素材内容相同：没有新存一份，返回的是已有的那张
	Duplicate bool `json:"duplicate,omitempty"`
}

func (s *Server) fileInfo(name string) (mediaItem, error) {
	stat, err := os.Stat(filepath.Join(s.UploadDir, name))
	if err != nil {
		return mediaItem{}, err
	}
	item := mediaItem{
		Name:      name,
		URL:       "/uploads/" + name,
		Thumb:     thumbURL(name),
		Size:      stat.Size(),
		CreatedAt: isoTime(stat.ModTime()),
	}
	if _, err := os.Stat(filepath.Join(s.originalsDir, name)); err == nil {
		item.HasOriginal = true
	}
	var cropJSON sql.NullString
	var createdAt string
	err = s.DB.QueryRow(`SELECT crop_json, created_at FROM media WHERE name = ?`, name).Scan(&cropJSON, &createdAt)
	if err == nil {
		item.CreatedAt = createdAt
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

// listUploads 列出公开素材文件名（跳过点文件与不支持的扩展）。
func (s *Server) listUploads(filter map[string]bool) ([]string, error) {
	entries, err := os.ReadDir(s.UploadDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		// 用户头像（avatar-*）由个人资料管理，不进素材库
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), avatarPrefix) || !filter[imaging.Ext(e.Name())] {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}

// GET /admin/media 素材列表
func (s *Server) listMedia(w http.ResponseWriter, _ *http.Request) {
	names, err := s.listUploads(allowedExt)
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
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })
	writeJSON(w, http.StatusOK, items)
}

func randomFilename(ext string) string {
	raw := make([]byte, 4)
	_, _ = rand.Read(raw)
	return strconv.FormatInt(time.Now().UnixMilli(), 36) + "-" + hex.EncodeToString(raw) + ext
}

// POST /admin/media 上传（单个/批量，字段 files）
func (s *Server) uploadMedia(w http.ResponseWriter, r *http.Request) {
	// 全局 ReadTimeout 较短；大批量上传单独放宽
	_ = http.NewResponseController(w).SetReadDeadline(time.Now().Add(15 * time.Minute))
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxUploadFiles)*maxUploadBytes+1<<20)
	reader, err := r.MultipartReader()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_multipart")
		return
	}
	var saved []string
	dups := map[string]bool{}
	for len(saved) < maxUploadFiles {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_multipart")
			return
		}
		if part.FormName() != "files" || part.FileName() == "" {
			continue
		}
		ext := imaging.Ext(part.FileName())
		if !allowedExt[ext] {
			_, _ = io.Copy(io.Discard, part)
			continue
		}
		name := randomFilename(ext)
		dst, err := os.Create(filepath.Join(s.UploadDir, name))
		if err != nil {
			fail(w, err)
			return
		}
		// 边写边算哈希，用于查重
		h := sha256.New()
		n, err := io.Copy(io.MultiWriter(dst, h), io.LimitReader(part, maxUploadBytes+1))
		dst.Close()
		if err != nil || n > maxUploadBytes {
			os.Remove(filepath.Join(s.UploadDir, name))
			writeError(w, http.StatusRequestEntityTooLarge, "file_too_large")
			return
		}
		// 文件头必须与扩展名一致、像素数在上限内；否则不入库
		if _, err := imaging.Meta(filepath.Join(s.UploadDir, name)); err != nil {
			os.Remove(filepath.Join(s.UploadDir, name))
			if errors.Is(err, imaging.ErrTooLarge) {
				writeError(w, http.StatusRequestEntityTooLarge, "image_too_large")
			} else {
				writeError(w, http.StatusBadRequest, "invalid_image")
			}
			return
		}
		sum := hex.EncodeToString(h.Sum(nil))
		// 与已有素材内容相同：丢弃新文件，返回已有的那张
		if err := s.ensureMediaHashes(); err != nil {
			fail(w, err)
			return
		}
		if existing := s.mediaByHash(sum, name); existing != "" {
			os.Remove(filepath.Join(s.UploadDir, name))
			_, _ = s.DB.Exec(`DELETE FROM media WHERE name = ?`, name)
			saved = append(saved, existing)
			dups[existing] = true
			continue
		}
		if _, err := s.DB.Exec(`INSERT INTO media (name, sha256) VALUES (?, ?)
			ON CONFLICT(name) DO UPDATE SET sha256 = excluded.sha256`, name, sum); err != nil {
			fail(w, err)
			return
		}
		saved = append(saved, name)
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
	src, err := imaging.Decode(originalPath)
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
	if name == "" {
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
	if name == "" || !fileExists(filepath.Join(s.UploadDir, name)) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	os.Remove(filepath.Join(s.UploadDir, name))
	os.Remove(filepath.Join(s.originalsDir, name))
	s.removeThumb(name)
	if _, err := s.DB.Exec(`DELETE FROM media WHERE name = ?`, name); err != nil {
		fail(w, err)
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
