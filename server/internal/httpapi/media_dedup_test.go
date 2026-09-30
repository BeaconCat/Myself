package httpapi

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pngOf(c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 24, 24))
	for x := 0; x < 24; x++ {
		for y := 0; y < 24; y++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// uploadFiles 以管理员身份上传若干 PNG，返回素材条目。
func (e *env) uploadFiles(files ...[]byte) []mediaItem {
	e.t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for i, f := range files {
		part, _ := mw.CreateFormFile("files", "f"+itoa(int64(i))+".png")
		part.Write(f)
	}
	mw.Close()
	res := e.do(http.MethodPost, "/api/v1/admin/media", &buf, map[string]string{"Content-Type": mw.FormDataContentType(), "Authorization": "Bearer " + e.token})
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusCreated {
		e.t.Fatalf("upload: %d %s", res.StatusCode, data)
	}
	var out []mediaItem
	json.Unmarshal(data, &out)
	return out
}

// 查重：同一张图再传不会多存一份（返回已有的那张并标记重复）；上传前按哈希查询可直接命中；不同的图正常入库。
func TestMediaDedup(t *testing.T) {
	e := newEnv(t)
	red, blue := pngOf(color.RGBA{R: 220, A: 255}), pngOf(color.RGBA{B: 220, A: 255})

	first := e.uploadFiles(red)
	if len(first) != 1 || first[0].Duplicate {
		t.Fatalf("first upload: %+v", first)
	}
	again := e.uploadFiles(red, blue)
	if len(again) != 2 || !again[0].Duplicate || again[0].Name != first[0].Name || again[1].Duplicate {
		t.Fatalf("second upload: %+v", again)
	}
	count := 0
	entries, _ := os.ReadDir(filepath.Join(e.root, "uploads"))
	for _, en := range entries {
		if !en.IsDir() && strings.HasSuffix(en.Name(), ".png") {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("files on disk = %d, want 2", count)
	}

	sum := sha256.Sum256(red)
	hexSum := hex.EncodeToString(sum[:])
	var hits map[string]mediaItem
	e.call(http.MethodPost, "/api/v1/admin/media/lookup", map[string]any{"hashes": []string{hexSum, strings.Repeat("0", 64), "bad"}}, &hits, http.StatusOK)
	if len(hits) != 1 || hits[hexSum].Name != first[0].Name || !hits[hexSum].Duplicate {
		t.Fatalf("lookup: %+v", hits)
	}
}

func TestZipEntryName(t *testing.T) {
	used := map[string]bool{}
	cases := []struct{ title, name, want string }{
		{"猫.png", "a1-x.png", "猫.png"},
		{"猫.PNG", "a2-x.png", "猫 (2).PNG"},
		{"照片", "a3-x.jpg", "照片.jpg"},
		{"img.jpeg", "a4-x.jpg", "img.jpeg"},
		{"", "a5-x.webp", "a5-x.webp"},
		{"wrong.gif", "a6-x.png", "wrong.gif.png"},
	}
	for _, c := range cases {
		if got := zipEntryName(c.title, c.name, used); got != c.want {
			t.Errorf("zipEntryName(%q, %q) = %q, want %q", c.title, c.name, got, c.want)
		}
	}
	if got := cleanMediaTitle(`C:\fake\path\a<b>.png`); got != "ab.png" {
		t.Errorf("cleanMediaTitle = %q", got)
	}
}

// 上传保留原文件名为显示名；重命名只改显示名；批量打包按显示名命名；批量删除。
func TestMediaTitleZipBatchDelete(t *testing.T) {
	e := newEnv(t)
	items := e.uploadFiles(pngOf(color.RGBA{G: 200, A: 255}), pngOf(color.RGBA{R: 90, G: 90, A: 255}))
	if items[0].Title != "f0.png" || items[1].Title != "f1.png" {
		t.Fatalf("titles not kept: %+v", items)
	}
	var renamed mediaItem
	e.call(http.MethodPut, "/api/v1/admin/media/"+items[0].Name, map[string]any{"title": "封面/草图.png"}, &renamed, http.StatusOK)
	if renamed.Title != "草图.png" || renamed.Name != items[0].Name {
		t.Fatalf("rename: %+v", renamed)
	}

	body, _ := json.Marshal(map[string]any{"names": []string{items[0].Name, items[1].Name, "../evil"}})
	res := e.do(http.MethodPost, "/api/v1/admin/media/zip", bytes.NewReader(body), map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + e.token})
	data, _ := io.ReadAll(res.Body)
	res.Body.Close()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if res.StatusCode != http.StatusOK || err != nil {
		t.Fatalf("zip: %d %v", res.StatusCode, err)
	}
	var got []string
	for _, f := range zr.File {
		got = append(got, f.Name)
	}
	if strings.Join(got, ",") != "草图.png,f1.png" {
		t.Fatalf("zip entries: %v", got)
	}

	var del map[string]int
	e.call(http.MethodPost, "/api/v1/admin/media/delete", map[string]any{"names": []string{items[0].Name, items[1].Name}}, &del, http.StatusOK)
	if del["deleted"] != 2 {
		t.Fatalf("batch delete: %v", del)
	}
	if _, err := os.Stat(filepath.Join(e.server.UploadDir, items[0].Name)); !os.IsNotExist(err) {
		t.Fatalf("file still exists")
	}
}
