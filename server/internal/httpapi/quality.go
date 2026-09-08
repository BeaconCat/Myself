package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"myself/server/internal/imaging"
)

var compressibleExt = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true}

// replaceURLRefs 站内 URL 引用替换：压缩改名（png→webp）后同步修正文章/随想里的链接。
func (s *Server) replaceURLRefs(oldURL, newURL string) error {
	like := "%" + oldURL + "%"
	for _, t := range []struct {
		table string
		cols  []string
	}{
		{"posts", []string{"covers", "content_md"}},
		{"notes", []string{"images", "content_md"}},
	} {
		for _, col := range t.cols {
			_, err := s.DB.Exec(`UPDATE `+t.table+` SET `+col+` = REPLACE(`+col+`, ?, ?) WHERE `+col+` LIKE ?`,
				oldURL, newURL, like)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type qualityItem struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	Size         int64  `json:"size"`
	Format       string `json:"format"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	HasAlpha     bool   `json:"hasAlpha"`
	Compressible bool   `json:"compressible"`
}

// GET /admin/quality/scan 扫描可压缩图片
func (s *Server) qualityScan(w http.ResponseWriter, _ *http.Request) {
	names, err := s.listUploads(compressibleExt)
	if err != nil {
		fail(w, err)
		return
	}
	items := []qualityItem{}
	for _, name := range names {
		full := filepath.Join(s.UploadDir, name)
		stat, err := os.Stat(full)
		if err != nil {
			continue
		}
		meta, err := imaging.Meta(full)
		if err != nil {
			continue
		}
		items = append(items, qualityItem{
			Name:         name,
			URL:          "/uploads/" + name,
			Size:         stat.Size(),
			Format:       strings.TrimPrefix(imaging.Ext(name), "."),
			Width:        meta.Width,
			Height:       meta.Height,
			HasAlpha:     meta.HasAlpha,
			Compressible: true,
		})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Size > items[j].Size })
	writeJSON(w, http.StatusOK, items)
}

type compressResult struct {
	Name    string `json:"name"`
	NewName string `json:"newName,omitempty"`
	Before  int64  `json:"before,omitempty"`
	After   int64  `json:"after,omitempty"`
	Error   string `json:"error,omitempty"`
}

// POST /admin/quality/compress { names: string[], quality: 1-100 }
// PNG → WebP（保留 alpha，文件名改 .webp，站内引用自动替换）；JPG/WebP → 原格式重压。
// 原图备份到 .originals 保留后路。
func (s *Server) qualityCompress(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}
	quality := roundInt(b.num("quality"))
	if quality == 0 {
		quality = 80
	}
	quality = clamp(quality, 1, 100)

	results := []compressResult{}
	for _, name := range b.strings("names", 0) {
		if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
			continue
		}
		full := filepath.Join(s.UploadDir, name)
		if !fileExists(full) {
			continue
		}
		ext := imaging.Ext(name)
		if !compressibleExt[ext] {
			continue
		}
		stat, err := os.Stat(full)
		if err != nil {
			continue
		}
		before := stat.Size()
		if err := copyIfMissing(full, filepath.Join(s.originalsDir, name)); err != nil {
			results = append(results, compressResult{Name: name, Error: err.Error()})
			continue
		}
		res, err := s.compressOne(name, full, ext, before, quality)
		if err != nil {
			results = append(results, compressResult{Name: name, Error: err.Error()})
			continue
		}
		results = append(results, res)
	}
	writeJSON(w, http.StatusOK, results)
}

func (s *Server) compressOne(name, full, ext string, before int64, quality int) (compressResult, error) {
	img, err := imaging.Decode(full)
	if err != nil {
		return compressResult{}, err
	}
	if ext == ".png" {
		newName := strings.TrimSuffix(name, filepath.Ext(name)) + ".webp"
		data, err := imaging.EncodeBytes(img, ".webp", quality)
		if err != nil {
			return compressResult{}, err
		}
		if err := os.WriteFile(filepath.Join(s.UploadDir, newName), data, 0o644); err != nil {
			return compressResult{}, err
		}
		os.Remove(full)
		s.removeThumb(name)
		if _, err := s.DB.Exec(`INSERT INTO media (name) VALUES (?) ON CONFLICT(name) DO NOTHING`, newName); err != nil {
			return compressResult{}, err
		}
		if _, err := s.DB.Exec(`DELETE FROM media WHERE name = ?`, name); err != nil {
			return compressResult{}, err
		}
		if err := s.replaceURLRefs("/uploads/"+name, "/uploads/"+newName); err != nil {
			return compressResult{}, err
		}
		return compressResult{Name: name, NewName: newName, Before: before, After: int64(len(data))}, nil
	}
	data, err := imaging.EncodeBytes(img, ext, quality)
	if err != nil {
		return compressResult{}, err
	}
	after := int64(len(data))
	// 只有确实变小才替换
	if after < before {
		if err := os.WriteFile(full, data, 0o644); err != nil {
			return compressResult{}, err
		}
	} else {
		after = before
	}
	return compressResult{Name: name, NewName: name, Before: before, After: after}, nil
}
