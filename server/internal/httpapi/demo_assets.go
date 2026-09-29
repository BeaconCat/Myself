package httpapi

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"myself/server/internal/store"
)

// demoCoverURLRe 默认封面地址 /covers/NN.webp。
var demoCoverURLRe = regexp.MustCompile(`^/covers/([0-9]{2})\.webp$`)

// demoAssets 返回 Demo 内容的图片地址改写函数：首次引用某张默认封面时把它复制进素材库
// （uploads/demo-cover-NN.webp），之后复用同一文件，地址改写为 /uploads/…。
// 没有内嵌封面（前端未构建）或复制失败时保留原地址，Demo 仍可正常显示。
func (s *Server) demoAssets() store.AssetMapper {
	done := map[string]string{}
	return func(u string) string {
		if v, ok := done[u]; ok {
			return v
		}
		out := u
		if m := demoCoverURLRe.FindStringSubmatch(u); m != nil && s.DemoCovers != nil {
			name := "demo-cover-" + m[1] + ".webp"
			if err := s.copyDemoAsset(m[1]+".webp", name); err != nil {
				log.Printf("[myself-server] import demo cover %s: %v", m[1], err)
			} else {
				out = "/uploads/" + name
			}
		}
		done[u] = out
		return out
	}
}

// copyDemoAsset 从内嵌封面复制到素材目录；目标已存在时不覆盖。
func (s *Server) copyDemoAsset(src, name string) error {
	dst := filepath.Join(s.UploadDir, name)
	if _, err := os.Stat(dst); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	data, err := fs.ReadFile(s.DemoCovers, src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.UploadDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
