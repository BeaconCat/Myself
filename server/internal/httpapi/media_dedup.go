package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

/*
 * 素材查重：media.sha256 记录每个文件「原始上传内容」的 SHA-256（裁切 / 压缩改的是公开文件，原图在 .originals）。
 * 上传前浏览器先算哈希调 /admin/media/lookup，已存在的直接复用、不再上传；
 * 上传时服务端也边写边算，与已有文件重复则丢弃新文件、返回已有的那张（重复重定向）。
 */

var sha256Re = regexp.MustCompile(`^[0-9a-f]{64}$`)

// hashFile 计算文件的 SHA-256。
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ensureMediaHashes 给还没有哈希的素材补上（首次查重时一次性回填；原图优先，保证与原始上传内容一致）。
func (s *Server) ensureMediaHashes() error {
	s.mediaHashMu.Lock()
	defer s.mediaHashMu.Unlock()
	names, err := s.listUploads(nil)
	if err != nil {
		return err
	}
	known := map[string]bool{}
	rows, err := s.DB.Query(`SELECT name FROM media WHERE sha256 IS NOT NULL AND sha256 != ''`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			known[n] = true
		}
	}
	rows.Close()
	for _, name := range names {
		if known[name] {
			continue
		}
		src := filepath.Join(s.originalsDir, name)
		if !fileExists(src) {
			src = filepath.Join(s.UploadDir, name)
		}
		sum, err := hashFile(src)
		if err != nil {
			continue
		}
		if _, err := s.DB.Upsert("media", []string{"name", "sha256"}, []string{"name"}, []string{"sha256"}, name, sum); err != nil {
			return err
		}
	}
	return nil
}

// mediaByHash 已存在且文件仍在的同内容素材名（排除 exclude，即正在上传的这份）；没有返回空串。
func (s *Server) mediaByHash(sum, exclude string) string {
	rows, err := s.DB.Query(`SELECT name FROM media WHERE sha256 = ? AND name != ? ORDER BY created_at, name`, sum, exclude)
	if err != nil {
		return ""
	}
	defer rows.Close()
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil && !strings.HasPrefix(n, avatarPrefix) && fileExists(filepath.Join(s.UploadDir, n)) {
			return n
		}
	}
	return ""
}

// POST /admin/media/lookup {hashes: [sha256…]} → {sha256: 已有素材}：上传前查重，命中的不必再传。
func (s *Server) lookupMedia(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if err := s.ensureMediaHashes(); err != nil {
		fail(w, err)
		return
	}
	raw, _ := b["hashes"].([]any)
	out := map[string]mediaItem{}
	for _, v := range raw {
		sum, _ := v.(string)
		sum = strings.ToLower(sum)
		if !sha256Re.MatchString(sum) || len(out) >= maxUploadFiles {
			continue
		}
		if name := s.mediaByHash(sum, ""); name != "" {
			if item, err := s.fileInfo(name); err == nil {
				item.Duplicate = true
				out[sum] = item
			}
		}
	}
	writeJSON(w, http.StatusOK, out)
}
