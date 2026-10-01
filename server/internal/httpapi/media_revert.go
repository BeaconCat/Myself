package httpapi

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	errNotCompressed = errors.New("not_compressed")
	errBackupMissing = errors.New("backup_missing")
	errNameTaken     = errors.New("name_taken")
)

// mediaCompression 压缩记录：压缩前的文件保存在 .precompress/<From>，可随时回退。
type mediaCompression struct {
	// From 压缩前的文件名（PNG 转 WebP 时与当前文件名不同）
	From string `json:"from"`
	// Before 压缩前体积
	Before int64  `json:"before"`
	At     string `json:"at"`
}

// compressionOf 读取压缩记录；未压缩返回 nil。
func (s *Server) compressionOf(name string) *mediaCompression {
	var c mediaCompression
	var at *string
	err := s.DB.QueryRow(`SELECT COALESCE(compressed_from, ''), COALESCE(compressed_before, 0), compressed_at FROM media WHERE name = ?`, name).
		Scan(&c.From, &c.Before, &at)
	// 来自数据库（可能源于被篡改的备份）：不是安全的单段文件名就当作未压缩，绝不拼进路径
	if err != nil || c.From == "" || safeName(c.From) == "" {
		return nil
	}
	if at != nil {
		c.At = *at
	}
	return &c
}

// clearCompression 丢弃压缩记录与压缩前备份（重新裁切后，旧备份不再对应当前内容）。
func (s *Server) clearCompression(name string) {
	if c := s.compressionOf(name); c != nil {
		os.Remove(filepath.Join(s.precompressDir, c.From))
	}
	_, _ = s.DB.Exec(`UPDATE media SET compressed_from = '', compressed_before = 0, compressed_at = NULL WHERE name = ?`, name)
}

// renameMediaRecord 素材改名（PNG 转 WebP / 回退）：记录整行跟随（显示名、哈希、裁切都保留），
// 站内所有引用同步替换：文章、随想与站点配置（身份头像、形象图、关于页模块等）。
func (s *Server) renameMediaRecord(from, to string) error {
	if _, err := s.DB.Upsert("media", []string{"name"}, []string{"name"}, nil, from); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`DELETE FROM media WHERE name = ?`, to); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`UPDATE media SET name = ? WHERE name = ?`, to, from); err != nil {
		return err
	}
	oldURL, newURL := "/uploads/"+from, "/uploads/"+to
	if err := s.replaceURLRefs(oldURL, newURL); err != nil {
		return err
	}
	_, err := s.DB.Exec("UPDATE settings SET value = REPLACE(value, ?, ?) WHERE `key` = 'site_config' AND value LIKE ?",
		`"`+oldURL+`"`, `"`+newURL+`"`, `%"`+oldURL+`"%`)
	return err
}

// markCompressed 压缩成功后记录（多次压缩保留最早的那份备份与体积）。
func (s *Server) markCompressed(name, from string, before int64) error {
	_, err := s.DB.Upsert("media", []string{"name", "compressed_from", "compressed_before", "compressed_at"}, []string{"name"}, []string{"compressed_from", "compressed_before", "compressed_at"},
		name, from, before, isoTime(time.Now()))
	return err
}

// revertOne 回退一张压缩过的素材，返回回退后的文件名。
func (s *Server) revertOne(name string) (string, int, error) {
	c := s.compressionOf(name)
	if c == nil {
		return "", http.StatusConflict, errNotCompressed
	}
	backup := filepath.Join(s.precompressDir, c.From)
	if !fileExists(backup) {
		s.clearCompression(name)
		return "", http.StatusGone, errBackupMissing
	}
	if c.From != name && fileExists(filepath.Join(s.UploadDir, c.From)) {
		return "", http.StatusConflict, errNameTaken
	}
	if err := os.Rename(backup, filepath.Join(s.UploadDir, c.From)); err != nil {
		return "", http.StatusInternalServerError, err
	}
	s.removeThumb(name)
	if c.From != name {
		os.Remove(filepath.Join(s.UploadDir, name))
		if err := s.renameMediaRecord(name, c.From); err != nil {
			return "", http.StatusInternalServerError, err
		}
	}
	_, _ = s.DB.Exec(`UPDATE media SET compressed_from = '', compressed_before = 0, compressed_at = NULL WHERE name = ?`, c.From)
	return c.From, 0, nil
}

// POST /admin/media/revert 回退压缩：{"names": [...]} → {"items": [...], "failed": n}
func (s *Server) revertMedia(w http.ResponseWriter, r *http.Request) {
	names, ok := batchNames(w, r)
	if !ok {
		return
	}
	items := []mediaItem{}
	failed := 0
	for _, name := range names {
		to, _, err := s.revertOne(name)
		if err != nil {
			failed++
			continue
		}
		if item, err := s.fileInfo(to); err == nil {
			items = append(items, item)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "failed": failed})
}
