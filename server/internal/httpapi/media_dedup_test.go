package httpapi

import (
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
