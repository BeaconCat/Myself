package httpapi

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"myself/server/internal/imaging"
)

// 缩略图：最长边 480px 的 webp，存于 uploads/thumbs/<原名>.webp，按需生成、随源文件删除。
const thumbMaxSide = 480

func (s *Server) thumbPath(name string) string {
	return filepath.Join(s.thumbsDir, name+".webp")
}

func thumbURL(name string) string {
	return "/uploads/thumbs/" + name + ".webp"
}

// decodeSlots 同时解码的图片数上限（缩略图、裁切、压缩共用），防止并发首访把内存打满。
var decodeSlots = make(chan struct{}, 2)

// ensureThumb 缺失或过期（源文件更新）时重新生成。gif 直接回源（保留动画）。
// 同一张图的并发请求经 singleflight 合并为一次解码。
// thumbFailed 解码失败的原图（名字 → 当时的修改时间）：原图不变就不再重试，防止坏图被反复请求时一次次重新分配内存
var thumbFailed sync.Map

func (s *Server) ensureThumb(name string) (string, error) {
	var mod int64
	if st, err := os.Stat(filepath.Join(s.UploadDir, name)); err == nil {
		mod = st.ModTime().UnixNano()
	}
	if v, ok := thumbFailed.Load(name); ok && v.(int64) == mod {
		return "", errThumbFailed
	}
	v, err, _ := s.thumbs.Do(name, func() (any, error) { return s.buildThumb(name) })
	if err != nil {
		thumbFailed.Store(name, mod)
		return "", err
	}
	return v.(string), nil
}

var errThumbFailed = errors.New("thumbnail failed before; source unchanged")

func (s *Server) buildThumb(name string) (string, error) {
	src := filepath.Join(s.UploadDir, name)
	dst := s.thumbPath(name)
	srcStat, err := os.Stat(src)
	if err != nil {
		return "", err
	}
	if imaging.Ext(name) == ".gif" {
		return src, nil
	}
	if dstStat, err := os.Stat(dst); err == nil && !dstStat.ModTime().Before(srcStat.ModTime()) {
		return dst, nil
	}
	decodeSlots <- struct{}{}
	defer func() { <-decodeSlots }()
	img, err := imaging.Decode(src)
	if err != nil {
		return "", err
	}
	data, err := imaging.EncodeBytes(imaging.Fit(img, thumbMaxSide), ".webp", 78)
	if err != nil {
		return "", err
	}
	return dst, os.WriteFile(dst, data, 0o644)
}

func (s *Server) removeThumb(name string) {
	os.Remove(s.thumbPath(name))
}

// GET /uploads/thumbs/{name}.webp 按需生成缩略图；源文件不存在则 404。
func (s *Server) serveThumb(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("name")
	name := safeName(strings.TrimSuffix(file, ".webp"))
	if name == "" || !strings.HasSuffix(file, ".webp") || !isImage(name) {
		http.NotFound(w, r)
		return
	}
	target, err := s.ensureThumb(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, target)
}
