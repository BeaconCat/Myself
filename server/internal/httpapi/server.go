// Package httpapi 提供 /api/v1 RESTful 路由与静态 /uploads 服务。
package httpapi

import (
	"encoding/json"
	"errors"
	"io/fs"
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
	// DemoCovers 可选：内嵌的默认封面（NN.webp）；初始化选择 Demo 时导入素材库。
	DemoCovers fs.FS
}

// Server 聚合全部路由处理器。
type Server struct {
	Deps
	originalsDir string
	// precompressDir 压缩前的文件（回退用）
	precompressDir string

	ghMu      sync.Mutex
	ghCache   githubCache
	ghSyncLog []syncEntry
	ghFlight  singleflight.Group
	ghFailAt  time.Time
	ghFailKey string

	thumbsDir string
	jobs      *jobRegistry
	limiter   *attemptLimiter
	// signupLimiter 注册 / 找回密码按 IP 计次（与登录锁定分开，避免互相影响）
	signupLimiter *attemptLimiter
	// acctLimiter 按登录名计失败次数（与按 IP 叠加）
	acctLimiter *attemptLimiter
	// mailLimiter 按用户限制「确认邮箱」等外发邮件（15 分钟 3 封）
	mailLimiter *attemptLimiter
	oauth       oauthStates
	comments    commentLimiter
	reacts      reactLimiter
	thumbs      singleflight.Group
	backupMu    sync.Mutex
	// mediaHashMu 串行化素材哈希回填
	mediaHashMu sync.Mutex
	// mailer 发信实现；nil 用 SMTP（sendWith），测试里替换成捕获函数
	mailer func(c config.Mail, to, subject, text, htmlBody string) error
}

// New 构造 Server 并准备目录。
func New(d Deps) *Server {
	s := &Server{
		Deps:           d,
		originalsDir:   filepath.Join(d.UploadDir, ".originals"),
		precompressDir: filepath.Join(d.UploadDir, ".precompress"),
		thumbsDir:      filepath.Join(d.UploadDir, "thumbs"),
		jobs:           newJobRegistry(),
		limiter:        newAttemptLimiter(),
		signupLimiter:  newAttemptLimiter(),
		acctLimiter:    &attemptLimiter{m: map[string]*failEntry{}, limit: accountFailLimit},
		mailLimiter:    &attemptLimiter{m: map[string]*failEntry{}, limit: 3},
	}
	for _, dir := range []string{s.originalsDir, s.precompressDir, s.thumbsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("[myself-server] mkdir %s: %v", dir, err)
		}
	}
	// 备份含数据库（JWT 密钥、口令哈希）：仅属主可读
	if err := os.MkdirAll(d.BackupDir, 0o700); err != nil {
		log.Fatalf("[myself-server] mkdir %s: %v", d.BackupDir, err)
	}
	// 登录有效期跟随后台配置
	if d.Auth != nil {
		d.Auth.TTL = func() time.Duration { return d.Config.Typed().SessionTTL() }
	}
	return s
}

// Handler 组装路由。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	const p = "/api/v1"
	admin := s.requireRole("admin")
	staff := s.requireRole("admin", "author")
	member := s.requireRole("admin", "author", "reader")
	ext := s.requireAPIKey

	// 公开
	mux.HandleFunc("GET "+p+"/posts", s.listPosts)
	mux.HandleFunc("GET "+p+"/posts/{slug}", s.getPost)
	mux.HandleFunc("GET "+p+"/hero", s.hero)
	mux.HandleFunc("GET "+p+"/tags", s.tags)
	mux.HandleFunc("GET "+p+"/notes", s.listNotes)
	mux.HandleFunc("GET "+p+"/notes/{id}", s.getNote)
	mux.HandleFunc("GET "+p+"/engage", s.engage)
	mux.HandleFunc("POST "+p+"/reactions", s.toggleReaction)
	mux.HandleFunc("GET "+p+"/img/{from}/{to}/{label}", s.placeholderImage)
	mux.HandleFunc("GET "+p+"/site-config", s.siteConfig)
	mux.HandleFunc("GET /feed", s.rssFeed)
	mux.HandleFunc("GET /feed.xml", s.rssFeed)
	mux.HandleFunc("GET "+p+"/github-status", s.githubStatus)
	mux.HandleFunc("POST "+p+"/auth/login", withCSRF(s.login))
	mux.HandleFunc("POST "+p+"/auth/logout", withCSRF(s.logout))
	mux.HandleFunc("GET "+p+"/auth/session", s.session)
	mux.HandleFunc("POST "+p+"/auth/register", withCSRF(s.register))
	mux.HandleFunc("POST "+p+"/auth/verify", withCSRF(s.verifyEmail))
	mux.HandleFunc("POST "+p+"/auth/forgot", withCSRF(s.forgotPassword))
	mux.HandleFunc("GET "+p+"/auth/reset", s.resetInfo)
	mux.HandleFunc("POST "+p+"/auth/reset", withCSRF(s.resetPassword))
	mux.HandleFunc("GET "+p+"/auth/invite", s.inviteInfo)
	mux.HandleFunc("GET "+p+"/auth/github/start", s.githubStart)
	mux.HandleFunc("GET "+p+"/auth/github/callback", s.githubCallback)
	mux.HandleFunc("GET "+p+"/comments", s.listComments)
	mux.HandleFunc("POST "+p+"/comments", s.createComment)
	mux.HandleFunc("GET "+p+"/archive/{name}", s.archiveListing)
	mux.HandleFunc("GET "+p+"/setup", s.setupStatus)
	mux.HandleFunc("POST "+p+"/setup", withCSRF(s.setup))
	mux.HandleFunc("POST "+p+"/setup/verify", withCSRF(s.setupVerify))

	// 管理员
	mux.HandleFunc("PUT "+p+"/auth/password", member(s.changePassword))
	mux.HandleFunc("PUT "+p+"/me", member(s.updateMe))
	mux.HandleFunc("POST "+p+"/me/avatar", member(s.uploadAvatar))
	mux.HandleFunc("DELETE "+p+"/me/avatar", member(s.deleteAvatar))
	mux.HandleFunc("GET "+p+"/avatar-pending/{name}", member(s.servePendingAvatar))
	mux.HandleFunc("PUT "+p+"/me/login", member(s.changeLogin))
	mux.HandleFunc("PUT "+p+"/me/email", member(s.changeEmail))
	mux.HandleFunc("DELETE "+p+"/me/email/pending", member(s.cancelEmailChange))
	// 文章与素材上传：管理员 + 协作作者（作者只能看到 / 修改自己的文章）
	mux.HandleFunc("GET "+p+"/admin/posts", staff(s.adminListPosts))
	mux.HandleFunc("GET "+p+"/admin/posts/{id}", staff(s.adminGetPost))
	mux.HandleFunc("POST "+p+"/admin/posts", staff(s.adminCreatePost))
	mux.HandleFunc("PUT "+p+"/admin/posts/{id}", staff(s.adminUpdatePost))
	mux.HandleFunc("DELETE "+p+"/admin/posts/{id}", staff(s.adminDeletePost))
	mux.HandleFunc("POST "+p+"/admin/posts/batch", staff(s.adminBatchPosts))
	mux.HandleFunc("POST "+p+"/admin/notes/batch", admin(s.adminBatchNotes))
	mux.HandleFunc("POST "+p+"/admin/notes", admin(s.adminCreateNote))
	mux.HandleFunc("PUT "+p+"/admin/notes/{id}", admin(s.adminUpdateNote))
	mux.HandleFunc("DELETE "+p+"/admin/notes/{id}", admin(s.adminDeleteNote))
	mux.HandleFunc("GET "+p+"/admin/settings", admin(s.adminGetSettings))
	mux.HandleFunc("PUT "+p+"/admin/settings", admin(s.adminSaveSettings))
	mux.HandleFunc("GET "+p+"/admin/apikeys", admin(s.listAPIKeys))
	mux.HandleFunc("POST "+p+"/admin/apikeys", admin(s.createAPIKey))
	mux.HandleFunc("DELETE "+p+"/admin/apikeys/{id}", admin(s.deleteAPIKey))
	mux.HandleFunc("GET "+p+"/admin/api-logs", admin(s.adminListAPILogs))
	mux.HandleFunc("GET "+p+"/admin/media", admin(s.listMedia))
	mux.HandleFunc("POST "+p+"/admin/media", staff(s.uploadMedia))
	mux.HandleFunc("POST "+p+"/admin/media/lookup", staff(s.lookupMedia))
	mux.HandleFunc("POST "+p+"/admin/media/{name}/crop", admin(s.cropMedia))
	mux.HandleFunc("GET "+p+"/admin/media/{name}/original", admin(s.mediaOriginal))
	mux.HandleFunc("DELETE "+p+"/admin/media/{name}", admin(s.deleteMedia))
	mux.HandleFunc("PUT "+p+"/admin/media/{name}", admin(s.updateMedia))
	mux.HandleFunc("POST "+p+"/admin/media/delete", admin(s.deleteMediaBatch))
	mux.HandleFunc("POST "+p+"/admin/media/zip", admin(s.zipMedia))
	mux.HandleFunc("POST "+p+"/admin/media/revert", admin(s.revertMedia))
	mux.HandleFunc("GET "+p+"/admin/backups", admin(s.listBackups))
	mux.HandleFunc("POST "+p+"/admin/backups", admin(s.createBackupNow))
	mux.HandleFunc("GET "+p+"/admin/backups/{name}", admin(s.downloadBackup))
	mux.HandleFunc("DELETE "+p+"/admin/backups/{name}", admin(s.deleteBackup))
	mux.HandleFunc("POST "+p+"/admin/backups/{name}/restore", admin(s.restoreFromBackup))
	mux.HandleFunc("POST "+p+"/admin/backups/restore", admin(s.restoreFromUpload))
	mux.HandleFunc("GET "+p+"/admin/export/markdown", admin(s.exportMarkdown))
	mux.HandleFunc("POST "+p+"/admin/import", admin(s.importPosts))
	mux.HandleFunc("GET "+p+"/admin/quality/scan", admin(s.qualityScan))
	mux.HandleFunc("POST "+p+"/admin/quality/compress", admin(s.qualityCompress))
	mux.HandleFunc("GET "+p+"/admin/quality/jobs/{id}", admin(s.compressJobStatus))
	mux.HandleFunc("POST "+p+"/admin/github/sync", admin(s.githubSync))
	mux.HandleFunc("GET "+p+"/admin/github/log", admin(s.githubLog))
	mux.HandleFunc("GET "+p+"/admin/users", admin(s.adminListUsers))
	mux.HandleFunc("GET "+p+"/admin/users/export", admin(s.adminExportUsers))
	mux.HandleFunc("PUT "+p+"/admin/users/{id}", admin(s.adminUpdateUser))
	mux.HandleFunc("DELETE "+p+"/admin/users/{id}", admin(s.adminDeleteUser))
	mux.HandleFunc("POST "+p+"/admin/users/{id}/reset", admin(s.adminResetLink))
	mux.HandleFunc("PUT "+p+"/admin/users/{id}/avatar", admin(s.adminReviewAvatar))
	mux.HandleFunc("GET "+p+"/admin/invites", admin(s.adminListInvites))
	mux.HandleFunc("POST "+p+"/admin/invites", admin(s.adminCreateInvite))
	mux.HandleFunc("DELETE "+p+"/admin/invites/{id}", admin(s.adminDeleteInvite))
	mux.HandleFunc("GET "+p+"/admin/comments", staff(s.adminListComments))
	mux.HandleFunc("PUT "+p+"/admin/comments/{id}", staff(s.adminUpdateComment))
	mux.HandleFunc("DELETE "+p+"/admin/comments/{id}", staff(s.adminDeleteComment))
	mux.HandleFunc("POST "+p+"/admin/comments/batch", staff(s.adminBatchComments))
	mux.HandleFunc("POST "+p+"/admin/mail/test", admin(s.mailTest))

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
	return recoverMiddleware(securityHeaders(logMiddleware(mux)))
}

// logMiddleware 访问日志：方法 路径 状态 耗时。
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if strings.HasPrefix(r.URL.Path, "/api/") || sw.status >= 400 {
			// 只记路径：重置令牌、邀请码、OAuth code 都在查询串里，不能进日志
			log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond))
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
