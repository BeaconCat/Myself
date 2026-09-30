package httpapi

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

type namedFile struct {
	name string
	data []byte
}

// uploadNamed 按给定文件名上传，返回状态码与响应体。
func (e *env) uploadNamed(files ...namedFile) (int, []byte) {
	e.t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for _, f := range files {
		part, _ := mw.CreateFormFile("files", f.name)
		part.Write(f.data)
	}
	mw.Close()
	res := e.do(http.MethodPost, "/api/v1/admin/media", &buf, map[string]string{"Content-Type": mw.FormDataContentType(), "Authorization": "Bearer " + e.token})
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	return res.StatusCode, data
}

func (e *env) uploadOne(name string, data []byte) mediaItem {
	e.t.Helper()
	code, body := e.uploadNamed(namedFile{name, data})
	var out []mediaItem
	if code != http.StatusCreated || json.Unmarshal(body, &out) != nil || len(out) != 1 {
		e.t.Fatalf("upload %s: %d %s", name, code, body)
	}
	return out[0]
}

func zipOf(t *testing.T, files map[string]string, dirs ...string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, d := range dirs {
		if _, err := zw.Create(d); err != nil {
			t.Fatal(err)
		}
	}
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(content))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// 非图片素材：类别 / 扩展名 / MIME 正确、没有缩略图，列表里能看到；可执行与标记类型拒收。
func TestMediaKindsUpload(t *testing.T) {
	e := newEnv(t)
	cases := []struct{ file, kind, ext, mime string }{
		{"clip.MP4", kindVideo, "mp4", "video/mp4"},
		{"song.mp3", kindAudio, "mp3", "audio/mpeg"},
		{"bundle.zip", kindArchive, "zip", "application/zip"},
		{"报告.pdf", kindFile, "pdf", "application/octet-stream"},
	}
	names := map[string]bool{}
	for i, c := range cases {
		data := []byte("content-" + c.file)
		if c.ext == "zip" {
			data = zipOf(t, map[string]string{"a.txt": "a"})
		}
		item := e.uploadOne(c.file, data)
		if item.Kind != c.kind || item.Ext != c.ext || item.Mime != c.mime || item.Thumb != "" || item.Title != c.file {
			t.Fatalf("case %d: %+v", i, item)
		}
		if !strings.HasSuffix(item.Name, "."+c.ext) {
			t.Fatalf("disk name %q", item.Name)
		}
		names[item.Name] = true
	}
	// 同内容再传：查重命中
	again := e.uploadOne("again.mp3", []byte("content-song.mp3"))
	if !again.Duplicate || !names[again.Name] {
		t.Fatalf("dedup: %+v", again)
	}

	for _, bad := range []string{"x.html", "x.svg", "noext", "x.JS"} {
		code, body := e.uploadNamed(namedFile{bad, []byte("<script>alert(1)</script>")})
		if code != http.StatusBadRequest || !strings.Contains(string(body), "unsupported_type") {
			t.Fatalf("%s: %d %s", bad, code, body)
		}
	}
	// 混传：被拒的静默跳过，其余照常入库
	code, body := e.uploadNamed(namedFile{"x.html", []byte("<b>")}, namedFile{"notes.txt", []byte("hello")})
	var mixed []mediaItem
	if code != http.StatusCreated || json.Unmarshal(body, &mixed) != nil || len(mixed) != 1 || mixed[0].Kind != kindFile {
		t.Fatalf("mixed: %d %s", code, body)
	}

	var list []mediaItem
	e.call(http.MethodGet, "/api/v1/admin/media", nil, &list, http.StatusOK)
	kinds := map[string]int{}
	for _, it := range list {
		kinds[it.Kind]++
		if it.Kind != kindImage && it.Thumb != "" {
			t.Fatalf("non-image thumb: %+v", it)
		}
	}
	for _, k := range []string{kindImage, kindVideo, kindAudio, kindArchive, kindFile} {
		if kinds[k] == 0 {
			t.Fatalf("list missing kind %s: %v", k, kinds)
		}
	}
}

// 直出：文件以附件下载（显示名 + nosniff + 沙箱），视频内联且支持 Range；缩略图与裁切只对图片。
func TestMediaKindsServe(t *testing.T) {
	e := newEnv(t)
	pdf := e.uploadOne("季度 报告.pdf", []byte("%PDF-1.4 fake"))
	res := e.do(http.MethodGet, pdf.URL, nil, nil)
	res.Body.Close()
	cd := res.Header.Get("Content-Disposition")
	if res.StatusCode != http.StatusOK || res.Header.Get("Content-Type") != "application/octet-stream" ||
		res.Header.Get("X-Content-Type-Options") != "nosniff" ||
		!strings.HasPrefix(cd, "attachment;") ||
		!strings.Contains(cd, "filename*=UTF-8''%E5%AD%A3%E5%BA%A6%20%E6%8A%A5%E5%91%8A.pdf") ||
		!strings.Contains(res.Header.Get("Content-Security-Policy"), "sandbox") {
		t.Fatalf("pdf: %d %v", res.StatusCode, res.Header)
	}

	video := e.uploadOne("clip.mp4", []byte("0123456789abcdef"))
	res = e.do(http.MethodGet, video.URL, nil, map[string]string{"Range": "bytes=4-7"})
	data, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusPartialContent || string(data) != "4567" ||
		res.Header.Get("Content-Type") != "video/mp4" || res.Header.Get("Content-Disposition") != "" {
		t.Fatalf("range: %d %q %v", res.StatusCode, data, res.Header)
	}

	res = e.do(http.MethodGet, "/uploads/thumbs/"+video.Name+".webp", nil, nil)
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("thumb of video: %d", res.StatusCode)
	}
	var errBody map[string]string
	e.call(http.MethodPost, "/api/v1/admin/media/"+video.Name+"/crop", map[string]int{"left": 0, "top": 0, "width": 1, "height": 1}, &errBody, http.StatusBadRequest)
	if errBody["error"] != "not_image" {
		t.Fatalf("crop: %v", errBody)
	}
	// 质量扫描只列图片
	var scan []qualityItem
	e.call(http.MethodGet, "/api/v1/admin/quality/scan", nil, &scan, http.StatusOK)
	for _, it := range scan {
		if it.Name == video.Name || it.Name == pdf.Name {
			t.Fatalf("scan lists non-image: %+v", it)
		}
	}
	// 打包下载保留真实扩展名
	res = e.do(http.MethodPost, "/api/v1/admin/media/zip", strings.NewReader(`{"names":["`+pdf.Name+`"]}`),
		map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + e.token})
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil || len(zr.File) != 1 || zr.File[0].Name != "季度 报告.pdf" {
		t.Fatalf("zip: %v %d", err, len(raw))
	}
}

// 压缩包预览：只读目录，列出条目；非 zip 415，不存在 404。
func TestArchiveListing(t *testing.T) {
	e := newEnv(t)
	item := e.uploadOne("pack.zip", zipOf(t, map[string]string{"docs/readme.txt": strings.Repeat("x", 1000)}, "docs/"))
	var out struct {
		Name    string         `json:"name"`
		Title   string         `json:"title"`
		Size    int64          `json:"size"`
		Entries []archiveEntry `json:"entries"`
		Total   int            `json:"total"`
		Trunc   bool           `json:"truncated"`
	}
	e.call(http.MethodGet, "/api/v1/archive/"+item.Name, nil, &out, http.StatusOK)
	if out.Name != item.Name || out.Title != "pack.zip" || out.Size != item.Size || out.Total != 2 || out.Trunc || len(out.Entries) != 2 {
		t.Fatalf("listing: %+v", out)
	}
	if !out.Entries[0].Dir || out.Entries[0].Name != "docs/" || out.Entries[1].Dir ||
		out.Entries[1].Name != "docs/readme.txt" || out.Entries[1].Size != 1000 || out.Entries[1].Modified == "" {
		t.Fatalf("entries: %+v", out.Entries)
	}

	txt := e.uploadOne("a.txt", []byte("plain"))
	e.call(http.MethodGet, "/api/v1/archive/"+txt.Name, nil, nil, http.StatusUnsupportedMediaType)
	e.call(http.MethodGet, "/api/v1/archive/missing.zip", nil, nil, http.StatusNotFound)
	e.call(http.MethodGet, "/api/v1/archive/..%2Fx.zip", nil, nil, http.StatusNotFound)
}
