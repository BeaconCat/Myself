// Package httpapi 提供 /api/v1 RESTful 路由与静态 /uploads 服务。
package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/config"
	"myself/server/internal/store"

	"golang.org/x/sync/singleflight"
)

// Deps 是 Server 依赖。
type Deps struct {
	DB        *store.DB
	Auth      *auth.Service
	Config    *config.Service
	UploadDir string
	BackupDir string
	DataDir   string
	// Frontend 可选：内嵌 SPA 处理器，接管非 /api、/uploads 的请求。
	Frontend http.Handler
}

// Server 聚合全部路由处理器。
type Server struct {
	Deps
	originalsDir string

	ghMu      sync.Mutex
	ghCache   githubCache
	ghSyncLog []syncEntry
	ghFlight  singleflight.Group

	thumbsDir string
	jobs      *jobRegistry
}

// New 构造 Server 并准备目录。
func New(d Deps) *Server {
	s := &Server{
		Deps:         d,
		originalsDir: filepath.Join(d.UploadDir, ".originals"),
		thumbsDir:    filepath.Join(d.UploadDir, "thumbs"),
		jobs:         newJobRegistry(),
	}
	for _, dir := range []string{s.originalsDir, s.thumbsDir, d.BackupDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("[myself-server] mkdir %s: %v", dir, err)
		}
	}
	return s
}

// Handler 组装路由。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	const p = "/api/v1"
	admin := s.requireAuth
	ext := s.requireAPIKey

	// 公开
	mux.HandleFunc("GET "+p+"/posts", s.listPosts)
	mux.HandleFunc("GET "+p+"/posts/{slug}", s.getPost)
	mux.HandleFunc("GET "+p+"/hero", s.hero)
	mux.HandleFunc("GET "+p+"/tags", s.tags)
	mux.HandleFunc("GET "+p+"/notes", s.listNotes)
	mux.HandleFunc("GET "+p+"/img/{from}/{to}/{label}", s.placeholderImage)
	mux.HandleFunc("GET "+p+"/site-config", s.siteConfig)
	mux.HandleFunc("GET /feed", s.rssFeed)
	mux.HandleFunc("GET /feed.xml", s.rssFeed)
	mux.HandleFunc("GET "+p+"/github-status", s.githubStatus)
	mux.HandleFunc("POST "+p+"/auth/login", s.login)

	// 管理员
	mux.HandleFunc("PUT "+p+"/auth/password", admin(s.changePassword))
	mux.HandleFunc("GET "+p+"/admin/posts", admin(s.adminListPosts))
	mux.HandleFunc("GET "+p+"/admin/posts/{id}", admin(s.adminGetPost))
	mux.HandleFunc("POST "+p+"/admin/posts", admin(s.adminCreatePost))
	mux.HandleFunc("PUT "+p+"/admin/posts/{id}", admin(s.adminUpdatePost))
	mux.HandleFunc("DELETE "+p+"/admin/posts/{id}", admin(s.adminDeletePost))
	mux.HandleFunc("POST "+p+"/admin/notes", admin(s.adminCreateNote))
	mux.HandleFunc("PUT "+p+"/admin/notes/{id}", admin(s.adminUpdateNote))
	mux.HandleFunc("DELETE "+p+"/admin/notes/{id}", admin(s.adminDeleteNote))
	mux.HandleFunc("GET "+p+"/admin/settings", admin(s.adminGetSettings))
	mux.HandleFunc("PUT "+p+"/admin/settings", admin(s.adminSaveSettings))
	mux.HandleFunc("GET "+p+"/admin/apikeys", admin(s.listAPIKeys))
	mux.HandleFunc("POST "+p+"/admin/apikeys", admin(s.createAPIKey))
	mux.HandleFunc("DELETE "+p+"/admin/apikeys/{id}", admin(s.deleteAPIKey))
	mux.HandleFunc("GET "+p+"/admin/media", admin(s.listMedia))
	mux.HandleFunc("POST "+p+"/admin/media", admin(s.uploadMedia))
	mux.HandleFunc("POST "+p+"/admin/media/{name}/crop", admin(s.cropMedia))
	mux.HandleFunc("GET "+p+"/admin/media/{name}/original", admin(s.mediaOriginal))
	mux.HandleFunc("DELETE "+p+"/admin/media/{name}", admin(s.deleteMedia))
	mux.HandleFunc("GET "+p+"/admin/backups", admin(s.listBackups))
	mux.HandleFunc("POST "+p+"/admin/backups", admin(s.createBackupNow))
	mux.HandleFunc("GET "+p+"/admin/backups/{name}", admin(s.downloadBackup))
	mux.HandleFunc("DELETE "+p+"/admin/backups/{name}", admin(s.deleteBackup))
	mux.HandleFunc("GET "+p+"/admin/quality/scan", admin(s.qualityScan))
	mux.HandleFunc("POST "+p+"/admin/quality/compress", admin(s.qualityCompress))
	mux.HandleFunc("GET "+p+"/admin/quality/jobs/{id}", admin(s.compressJobStatus))
	mux.HandleFunc("POST "+p+"/admin/github/sync", admin(s.githubSync))
	mux.HandleFunc("GET "+p+"/admin/github/log", admin(s.githubLog))

	// 外部通道（X-Api-Key）
	mux.HandleFunc("GET "+p+"/ext/posts", ext(s.extListPosts))
	mux.HandleFunc("GET "+p+"/ext/posts/{id}", ext(s.extGetPost))
	mux.HandleFunc("POST "+p+"/ext/posts", ext(s.extCreatePost))
	mux.HandleFunc("PUT "+p+"/ext/posts/{id}", ext(s.extUpdatePost))
	mux.HandleFunc("DELETE "+p+"/ext/posts/{id}", ext(s.extDeletePost))
	mux.HandleFunc("GET "+p+"/ext/notes", ext(s.extListNotes))
	mux.HandleFunc("POST "+p+"/ext/notes", ext(s.extCreateNote))
	mux.HandleFunc("PUT "+p+"/ext/notes/{id}", ext(s.extUpdateNote))
	mux.HandleFunc("DELETE "+p+"/ext/notes/{id}", ext(s.extDeleteNote))

	// 静态素材：忽略点文件/点目录，缓存一天；thumbs/ 按需生成缩略图
	mux.HandleFunc("GET /uploads/thumbs/{name}", s.serveThumb)
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", s.uploadsHandler()))

	if s.Frontend != nil {
		mux.Handle("/", s.Frontend)
	}
	return recoverMiddleware(logMiddleware(mux))
}

// logMiddleware 访问日志：方法 路径 状态 耗时。
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if strings.HasPrefix(r.URL.Path, "/api/") || sw.status >= 400 {
			log.Printf("%s %s %d %s", r.Method, r.URL.RequestURI(), sw.status, time.Since(start).Round(time.Millisecond))
		}
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (s *Server) uploadsHandler() http.Handler {
	fs := http.FileServer(http.Dir(s.UploadDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, seg := range strings.Split(r.URL.Path, "/") {
			if strings.HasPrefix(seg, ".") {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Cache-Control", "public, max-age=86400")
		fs.ServeHTTP(w, r)
	})
}

// recoverMiddleware 兜底 500，等价 Express 错误处理器。
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[panic] %v\n%s", rec, debug.Stack())
				writeError(w, http.StatusInternalServerError, "internal_error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// requireAuth Bearer Token 校验。
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := s.Auth.Verify(r.Header.Get("Authorization")); err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

/* ===== 响应与请求辅助 ===== */

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("[json] encode: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}

// fail 记录内部错误并回 500。
func fail(w http.ResponseWriter, err error) {
	log.Printf("[error] %v", err)
	writeError(w, http.StatusInternalServerError, "internal_error")
}

const maxBodyBytes = 2 << 20

// readJSON 解析 JSON 请求体（上限 2MB）；空体视为空对象。
func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, os.ErrClosed) || err.Error() == "EOF" {
			return nil
		}
		return err
	}
	return nil
}

// body 是宽松的 JSON 对象体。
type body map[string]any

func (b body) str(key string) (string, bool) {
	v, ok := b[key].(string)
	return v, ok
}

func (b body) strOr(key string) string {
	switch v := b[key].(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		raw, _ := json.Marshal(v)
		return string(raw)
	}
}

func (b body) truthy(key string) bool {
	switch v := b[key].(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	case nil:
		return false
	default:
		return true
	}
}

func (b body) strings(key string, limit int) []string {
	arr, ok := b[key].([]any)
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			out = append(out, s)
		} else {
			raw, _ := json.Marshal(v)
			out = append(out, string(raw))
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (b body) num(key string) float64 {
	switch v := b[key].(type) {
	case float64:
		return v
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	case bool:
		if v {
			return 1
		}
	}
	return 0
}

// queryInt 解析 query 数字并夹在 [lo, hi]，缺省 def。
func queryInt(r *http.Request, key string, def, lo, hi int) int {
	n, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || n == 0 {
		n = def
	}
	return clamp(n, lo, hi)
}

func clamp(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}

// pathID 解析路径 {id}；非法返回 -1（查不到任何行 → 404）。
func pathID(r *http.Request) int64 {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return -1
	}
	return id
}

func nowISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

func isoTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
