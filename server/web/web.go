// Package web 内嵌前端构建产物（client 构建输出到 server/web/dist），提供 SPA 静态服务。
package web

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
)

//go:embed all:dist
var dist embed.FS

var inlineScriptRe = regexp.MustCompile(`(?s)<script>(.*?)</script>`)

// ContentSecurityPolicy 依据 index.html 的内联脚本（首帧主题）生成 CSP：脚本仅同源 + 这些内联脚本的哈希，
// 因此注入的 <script>、事件属性与 javascript: 链接都不会执行。样式允许内联（组件 :style 绑定）；
// 图片允许 https（正文可引用外链图片），其余资源仅同源。
func ContentSecurityPolicy(index []byte) string {
	scripts := []string{"'self'"}
	for _, m := range inlineScriptRe.FindAllSubmatch(index, -1) {
		sum := sha256.Sum256(m[1])
		scripts = append(scripts, "'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'")
	}
	return strings.Join([]string{
		"default-src 'self'",
		"script-src " + strings.Join(scripts, " "),
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data: blob: https:",
		"font-src 'self' data:",
		"connect-src 'self'",
		"object-src 'none'",
		"base-uri 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
	}, "; ")
}

// Handler 返回 SPA 处理器：命中静态文件直接返回（assets 长缓存），否则回落 index.html。
// dist 中无 index.html（未构建）时返回 404 提示。
func Handler() http.Handler {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	index, indexErr := fs.ReadFile(sub, "index.html")
	hasIndex := indexErr == nil
	built := time.Now()
	csp := ContentSecurityPolicy(index)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasIndex {
			http.Error(w, "frontend not built: run `pnpm build` at repo root", http.StatusNotFound)
			return
		}
		p := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if p == "" {
			p = "index.html"
		}
		// index.html 交给 FileServer 会被 301 到 "/"，统一走 SPA 回落直出
		if st, err := fs.Stat(sub, p); err == nil && !st.IsDir() && p != "index.html" {
			if strings.HasPrefix(p, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}
			fileServer.ServeHTTP(w, r)
			return
		}
		// SPA 回落：直出 index.html
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Security-Policy", csp)
		http.ServeContent(w, r, "index.html", built, bytes.NewReader(index))
	})
}
