package httpapi

import (
	"archive/zip"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"myself/server/internal/imaging"
)

/*
 * 素材类型：图片之外还收视频、音频、压缩包与其它文件。
 * 类型只看扩展名（固定映射，不依赖系统 MIME 表，Windows 与 Linux 结果一致）；
 * 图片 / 视频 / 音频内联直出（支持 Range，视频可拖动），其余一律以附件下载，并加沙箱 CSP。
 */

// 素材类别
const (
	kindImage   = "image"
	kindVideo   = "video"
	kindAudio   = "audio"
	kindArchive = "archive"
	kindFile    = "file"
)

const (
	// maxFileBytes 非图片单文件上限
	maxFileBytes = 512 << 20
	// maxArchiveEntries 压缩包预览最多列出的条目数
	maxArchiveEntries = 5000
)

// imageExt 可上传的图片扩展名（上传时还会校验文件头）
var imageExt = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true}

// knownMedia 已知扩展名 → 类别与直出时的 Content-Type
var knownMedia = map[string]struct{ kind, mime string }{
	"png": {kindImage, "image/png"}, "jpg": {kindImage, "image/jpeg"}, "jpeg": {kindImage, "image/jpeg"},
	"webp": {kindImage, "image/webp"}, "gif": {kindImage, "image/gif"},

	"mp4": {kindVideo, "video/mp4"}, "m4v": {kindVideo, "video/mp4"}, "webm": {kindVideo, "video/webm"},
	"mov": {kindVideo, "video/quicktime"}, "ogv": {kindVideo, "video/ogg"},

	"mp3": {kindAudio, "audio/mpeg"}, "m4a": {kindAudio, "audio/mp4"}, "aac": {kindAudio, "audio/aac"},
	"ogg": {kindAudio, "audio/ogg"}, "oga": {kindAudio, "audio/ogg"}, "wav": {kindAudio, "audio/wav"},
	"flac": {kindAudio, "audio/flac"}, "opus": {kindAudio, "audio/ogg"},

	"zip": {kindArchive, "application/zip"}, "7z": {kindArchive, "application/x-7z-compressed"},
	"rar": {kindArchive, "application/vnd.rar"}, "tar": {kindArchive, "application/x-tar"},
	"gz": {kindArchive, "application/gzip"}, "tgz": {kindArchive, "application/gzip"},
	"bz2": {kindArchive, "application/x-bzip2"}, "xz": {kindArchive, "application/x-xz"},
}

// blockedExt 可在浏览器里执行或被当作标记解析的类型：一律拒收、不直出
var blockedExt = map[string]bool{
	"html": true, "htm": true, "xhtml": true, "shtml": true, "svg": true, "svgz": true,
	"xml": true, "xsl": true, "xslt": true, "js": true, "mjs": true, "cjs": true,
	"php": true, "phtml": true, "asp": true, "aspx": true, "jsp": true, "cgi": true,
	"pl": true, "py": true, "sh": true, "bat": true, "cmd": true, "ps1": true,
	"vbs": true, "hta": true, "swf": true,
}

var fileExtRe = regexp.MustCompile(`^[a-z0-9]{1,10}$`)

// mediaKind 按扩展名判断类别与 Content-Type；不接受的类型 ok 为 false。
func mediaKind(name string) (kind, ext, mime string, ok bool) {
	ext = strings.TrimPrefix(imaging.Ext(name), ".")
	if m, hit := knownMedia[ext]; hit {
		return m.kind, ext, m.mime, true
	}
	if blockedExt[ext] || !fileExtRe.MatchString(ext) {
		return "", ext, "", false
	}
	return kindFile, ext, "application/octet-stream", true
}

// isImage 是否为图片素材（裁切、缩略图、压缩只对图片生效）。
func isImage(name string) bool {
	return imageExt[imaging.Ext(name)]
}

// attrChars RFC 5987 attr-char 中除字母数字外可原样保留的字符
const attrChars = "!#$&+-.^_`|~"

// contentDisposition 附件下载头：ASCII 回退名用磁盘文件名，filename* 带 UTF-8 显示名。
func contentDisposition(display, fallback string) string {
	var b strings.Builder
	for _, c := range []byte(display) {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.IndexByte(attrChars, c) >= 0 {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return `attachment; filename="` + fallback + `"; filename*=UTF-8''` + b.String()
}

// uploadsHandler 素材直链：只接受单段、通过 safeName 且类型可接受的文件名；
// 不列目录，不暴露点目录（.originals）与子目录（含 Windows 8.3 短名绕过）。
// 图片 / 视频 / 音频内联（Range 由 ServeContent 处理），其余作为附件下载。
func (s *Server) uploadsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := safeName(r.URL.Path)
		kind, _, mime, ok := mediaKind(name)
		if name == "" || strings.HasPrefix(name, ".") || !ok {
			http.NotFound(w, r)
			return
		}
		f, err := os.Open(filepath.Join(s.UploadDir, name))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		st, err := f.Stat()
		if err != nil || st.IsDir() {
			http.NotFound(w, r)
			return
		}
		h := w.Header()
		h.Set("Cache-Control", "public, max-age=86400")
		h.Set("Content-Type", mime)
		h.Set("X-Content-Type-Options", "nosniff")
		switch kind {
		case kindImage, kindVideo, kindAudio:
			// 直接在标签页打开时浏览器会生成媒体文档，需放行同源的图片与媒体
			h.Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; media-src 'self'; style-src 'unsafe-inline'; sandbox")
		default:
			h.Set("Content-Security-Policy", "default-src 'none'; sandbox")
			h.Set("Content-Disposition", contentDisposition(zipEntryName(s.mediaTitle(name), name, map[string]bool{}), name))
		}
		http.ServeContent(w, r, "", st.ModTime(), f)
	})
}

type archiveEntry struct {
	Name       string `json:"name"`
	Size       uint64 `json:"size"`
	Compressed uint64 `json:"compressed"`
	Dir        bool   `json:"dir"`
	Modified   string `json:"modified"`
}

// archiveSlots 同时读取的压缩包目录数上限（条目极多时目录本身也占内存）
var archiveSlots = make(chan struct{}, 2)

// GET /archive/{name} 压缩包在线预览：只读 zip 中央目录（从不解压，不怕压缩炸弹），最多列 maxArchiveEntries 条。
func (s *Server) archiveListing(w http.ResponseWriter, r *http.Request) {
	name := safeName(r.PathValue("name"))
	path := filepath.Join(s.UploadDir, name)
	if name == "" || strings.HasPrefix(name, ".") || !fileExists(path) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if imaging.Ext(name) != ".zip" {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_type")
		return
	}
	archiveSlots <- struct{}{}
	defer func() { <-archiveSlots }()
	zr, err := zip.OpenReader(path)
	// 含不安全路径（../、绝对路径）的包仍可列出：这里只展示名字，不落盘
	if err != nil && !errors.Is(err, zip.ErrInsecurePath) {
		writeError(w, http.StatusUnprocessableEntity, "invalid_archive")
		return
	}
	defer zr.Close()
	var size int64
	if st, err := os.Stat(path); err == nil {
		size = st.Size()
	}
	title := s.mediaTitle(name)
	if title == "" {
		title = name
	}
	entries := make([]archiveEntry, 0, min(len(zr.File), maxArchiveEntries))
	for _, f := range zr.File[:min(len(zr.File), maxArchiveEntries)] {
		e := archiveEntry{
			Name:       strings.ToValidUTF8(f.Name, "�"),
			Size:       f.UncompressedSize64,
			Compressed: f.CompressedSize64,
			Dir:        strings.HasSuffix(f.Name, "/") || f.FileInfo().IsDir(),
		}
		if !f.Modified.IsZero() {
			e.Modified = f.Modified.Format(time.RFC3339)
		}
		entries = append(entries, e)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name":      name,
		"title":     title,
		"size":      size,
		"entries":   entries,
		"total":     len(zr.File),
		"truncated": len(zr.File) > maxArchiveEntries,
	})
}
