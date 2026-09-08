// Package web 内嵌前端构建产物（client 构建输出到 server/web/dist），提供 SPA 静态服务。
package web

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

//go:embed all:dist
var dist embed.FS

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
		http.ServeContent(w, r, "index.html", built, bytes.NewReader(index))
	})
}
