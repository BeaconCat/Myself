package httpapi

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

/* ===== 安全响应头 ===== */

// securityHeaders 全站基础安全头。页面（SPA）的完整 CSP 由 web.Handler 设置；
// 这里给其余响应（API、素材）一个最严的默认 CSP，禁止被嵌入 iframe、禁止 MIME 嗅探。
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			h.Set("Cache-Control", "no-store")
		} else if strings.HasPrefix(r.URL.Path, "/uploads/") {
			// 素材直链：即便被当作文档打开也处于沙箱中
			h.Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; sandbox")
		}
		next.ServeHTTP(w, r)
	})
}

/* ===== 登录 / 初始化码防爆破 ===== */

const (
	failWindow  = 15 * time.Minute
	failLimit   = 5
	lockoutSpan = 15 * time.Minute
)

type failEntry struct {
	count int
	first time.Time
	until time.Time
}

// attemptLimiter 按客户端 IP 统计失败次数：窗口内失败 failLimit 次即锁定 lockoutSpan。
type attemptLimiter struct {
	mu sync.Mutex
	m  map[string]*failEntry
}

func newAttemptLimiter() *attemptLimiter {
	return &attemptLimiter{m: map[string]*failEntry{}}
}

// blocked 返回剩余锁定时长；0 表示放行。
func (l *attemptLimiter) blocked(ip string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.m[ip]
	if e == nil {
		return 0
	}
	if d := time.Until(e.until); d > 0 {
		return d
	}
	return 0
}

func (l *attemptLimiter) fail(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	e := l.m[ip]
	if e == nil || now.Sub(e.first) > failWindow {
		e = &failEntry{first: now}
		l.m[ip] = e
	}
	e.count++
	if e.count >= failLimit {
		e.until = now.Add(lockoutSpan)
		e.count = 0
		e.first = now
	}
	// 顺手清理过期条目，防止表无限增长
	if len(l.m) > 4096 {
		for k, v := range l.m {
			if now.Sub(v.first) > failWindow && now.After(v.until) {
				delete(l.m, k)
			}
		}
	}
}

func (l *attemptLimiter) success(ip string) {
	l.mu.Lock()
	delete(l.m, ip)
	l.mu.Unlock()
}

// trustProxy 部署在反向代理后时设置 MYSELF_TRUST_PROXY=1，按 X-Forwarded-For 最左一段识别客户端。
var trustProxy = os.Getenv("MYSELF_TRUST_PROXY") == "1"

func clientIP(r *http.Request) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			first, _, _ := strings.Cut(xff, ",")
			if ip := strings.TrimSpace(first); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// tooMany 被锁定时回 429 与 Retry-After。
func tooMany(w http.ResponseWriter, d time.Duration) {
	w.Header().Set("Retry-After", strconv.Itoa(int(d.Seconds())+1))
	writeError(w, http.StatusTooManyRequests, "too_many_attempts")
}

// scryptSlots 限制同时进行的口令哈希计算（每次约 16MB 内存），防止并发登录拖垮进程。
var scryptSlots = make(chan struct{}, 4)

func withScryptSlot(fn func()) {
	scryptSlots <- struct{}{}
	defer func() { <-scryptSlots }()
	fn()
}
