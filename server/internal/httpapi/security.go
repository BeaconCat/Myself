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
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
		// 经 HTTPS 访问时要求浏览器此后只走 HTTPS
		if secureRequest(r) {
			h.Set("Strict-Transport-Security", "max-age=31536000")
		}
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
	// limit 触发锁定的失败次数（0 = failLimit）
	limit int
}

func newAttemptLimiter() *attemptLimiter {
	return &attemptLimiter{m: map[string]*failEntry{}}
}

// accountFailLimit 按账号的失败上限：比按 IP 宽松，挡住多 IP 分散爆破，又不让人轻易把站长锁在门外
const accountFailLimit = 20

// blocked 返回剩余锁定时长；0 表示放行。
func (l *attemptLimiter) blocked(ip string) time.Duration {
	ip = limitKey(ip)
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
	ip = limitKey(ip)
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	e := l.m[ip]
	if e == nil || now.Sub(e.first) > failWindow {
		e = &failEntry{first: now}
		l.m[ip] = e
	}
	e.count++
	limit := l.limit
	if limit == 0 {
		limit = failLimit
	}
	if e.count >= limit {
		e.until = now.Add(lockoutSpan)
		e.count = 0
		e.first = now
	}
	// 顺手清理过期条目，防止表无限增长；仍超出硬上限时丢弃未锁定的条目
	if len(l.m) > 4096 {
		for k, v := range l.m {
			if now.Sub(v.first) > failWindow && now.After(v.until) {
				delete(l.m, k)
			}
		}
	}
	if len(l.m) > 65536 {
		for k, v := range l.m {
			if now.After(v.until) {
				delete(l.m, k)
			}
		}
	}
}

func (l *attemptLimiter) success(ip string) {
	ip = limitKey(ip)
	l.mu.Lock()
	delete(l.m, ip)
	l.mu.Unlock()
}

// trustProxy 部署在反向代理后时设置 MYSELF_TRUST_PROXY=1，按 X-Forwarded-For 识别客户端。
var trustProxy = os.Getenv("MYSELF_TRUST_PROXY") == "1"

// clientIP 客户端地址。信任代理时取 X-Forwarded-For 的最右一段：那是紧邻本服务的受信代理追加的真实来源，
// 左侧各段可由客户端随意伪造（取最左会让所有按 IP 的限流形同虚设）。
func clientIP(r *http.Request) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); net.ParseIP(ip) != nil {
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

// pruneWindows 滑动窗口限流表过大时清掉窗口已过的来源，防止伪造来源把表撑爆
func pruneWindows(m map[string][]time.Time, now time.Time, window time.Duration) {
	if len(m) <= 4096 {
		return
	}
	for k, ts := range m {
		if len(ts) == 0 || now.Sub(ts[len(ts)-1]) >= window {
			delete(m, k)
		}
	}
}

// limitKey 限流用的来源键：IPv6 按 /64 聚合（一个用户通常分到整个 /64，按单地址计数形同虚设）。
func limitKey(ip string) string {
	p := net.ParseIP(ip)
	if p == nil || p.To4() != nil {
		return ip
	}
	return p.Mask(net.CIDRMask(64, 128)).String() + "/64"
}

// tooMany 被锁定时回 429 与 Retry-After。
func tooMany(w http.ResponseWriter, d time.Duration) {
	w.Header().Set("Retry-After", strconv.Itoa(int(d.Seconds())+1))
	writeError(w, http.StatusTooManyRequests, "too_many_attempts")
}

// scryptSlots 限制同时进行的口令哈希计算（当前参数每次约 128MB 内存），防止并发登录拖垮进程。
var scryptSlots = make(chan struct{}, 2)

func withScryptSlot(fn func()) {
	scryptSlots <- struct{}{}
	defer func() { <-scryptSlots }()
	fn()
}
