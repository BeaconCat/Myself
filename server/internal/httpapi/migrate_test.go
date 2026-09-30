package httpapi

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"image/color"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

// multipartBody 组装 multipart：field 相同的多个文件（文件名即相对路径）
func multipartBody(t *testing.T, field string, files map[string][]byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for name, data := range files {
		part, err := mw.CreateFormFile(field, name)
		if err != nil {
			t.Fatal(err)
		}
		part.Write(data)
	}
	mw.Close()
	return &buf, mw.FormDataContentType()
}

// 恢复：备份后改动数据（新文章、删素材），从备份恢复后回到备份时的状态；恢复前自动留一份安全备份。
func TestBackupRestore(t *testing.T) {
	e := newEnv(t)
	img := e.uploadFiles(pngOf(color.RGBA{R: 10, G: 200, B: 90, A: 255}))[0]
	var b struct{ Name string }
	e.call(http.MethodPost, "/api/v1/admin/backups", nil, &b, http.StatusCreated)

	e.call(http.MethodPost, "/api/v1/admin/posts", map[string]any{"slug": "after-backup", "title": "After", "contentMd": "x", "status": "published"}, nil, http.StatusCreated)
	e.call(http.MethodDelete, "/api/v1/admin/media/"+img.Name, nil, nil, http.StatusOK)

	var res struct{ Safety string }
	e.call(http.MethodPost, "/api/v1/admin/backups/"+b.Name+"/restore", nil, &res, http.StatusOK)
	if res.Safety == "" || res.Safety == b.Name {
		t.Fatalf("safety backup: %+v", res)
	}
	e.call(http.MethodGet, "/api/v1/posts/after-backup", nil, nil, http.StatusNotFound)
	if !fileExists(e.server.UploadDir + "/" + img.Name) {
		t.Fatal("media not restored")
	}
	// 恢复后会话仍有效（同一站点的备份，密钥与令牌版本不变）
	e.call(http.MethodGet, "/api/v1/admin/posts", nil, nil, http.StatusOK)

	// 上传一个不是备份的 zip：拒绝
	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)
	f, _ := zw.Create("hello.txt")
	f.Write([]byte("hi"))
	zw.Close()
	body, ct := multipartBody(t, "file", map[string][]byte{"x.zip": zbuf.Bytes()})
	r := e.do(http.MethodPost, "/api/v1/admin/backups/restore", body, map[string]string{"Content-Type": ct, "Authorization": "Bearer " + e.token})
	r.Body.Close()
	if r.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad backup accepted: %d", r.StatusCode)
	}
}

// 导出：每篇文章一个带 YAML front matter 的 .md，随想在 notes/
func TestExportMarkdown(t *testing.T) {
	e := newEnv(t)
	e.call(http.MethodPost, "/api/v1/admin/posts", map[string]any{"slug": "exp-1", "title": `He said "hi"`, "contentMd": "Body", "tags": []string{"a", "b"}, "status": "draft"}, nil, http.StatusCreated)
	r := e.do(http.MethodGet, "/api/v1/admin/export/markdown", nil, map[string]string{"Authorization": "Bearer " + e.token})
	data, _ := io.ReadAll(r.Body)
	r.Body.Close()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var post string
	notes := 0
	for _, f := range zr.File {
		if f.Name == "posts/exp-1.md" {
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			rc.Close()
			post = string(b)
		}
		if strings.HasPrefix(f.Name, "notes/") {
			notes++
		}
	}
	for _, want := range []string{"title: \"He said \\\"hi\\\"\"", "tags: [\"a\", \"b\"]", "draft: true", "Body"} {
		if !strings.Contains(post, want) {
			t.Fatalf("export missing %q in:\n%s", want, post)
		}
	}
	if notes == 0 {
		t.Fatal("notes not exported")
	}
	// 导出的 front matter 能被导入解析回来
	meta, body := parseFrontMatter(post)
	if meta["title"] != `He said "hi"` || strings.TrimSpace(body) != "Body" {
		t.Fatalf("roundtrip: %v %q", meta, body)
	}
}

// 导入：Hexo 目录（资源文件夹图片 + asset_img + 草稿）与 Hugo 页面包（TOML）
func TestImportPosts(t *testing.T) {
	e := newEnv(t)
	png := pngOf(color.RGBA{R: 200, G: 20, B: 90, A: 255})
	hexo := map[string][]byte{
		"blog/_config.yml": []byte("title: x\n"),
		"blog/source/_posts/hello-world.md": []byte("---\ntitle: 你好世界\ndate: 2021-03-04 05:06:07\ntags:\n  - go\n  - 随笔\ncategories: [技术]\ncover: /images/c.png\n---\n摘要\n<!-- more -->\n正文 ![图](hello-world/a.png) 和 {% asset_img a.png 说明 %}\n"),
		"blog/source/_posts/hello-world/a.png": png,
		"blog/source/images/c.png":             png,
		"blog/source/_drafts/wip.md":           []byte("title: 未完成\n---\n草稿\n"),
		"blog/node_modules/x/readme.md":        []byte("---\ntitle: nope\n---\n"),
	}
	body, ct := multipartBody(t, "files", hexo)
	r := e.do(http.MethodPost, "/api/v1/admin/import?draft=0", body, map[string]string{"Content-Type": ct, "Authorization": "Bearer " + e.token})
	var out struct {
		Platform string
		Created  int
		Images   int
		Items    []importItem
	}
	json.NewDecoder(r.Body).Decode(&out)
	r.Body.Close()
	if out.Platform != "hexo" || out.Created != 2 || out.Images != 1 {
		t.Fatalf("hexo import: %+v", out)
	}
	var got struct {
		Title     string
		ContentMd string
		Tags      []string
		Covers    []string
		Status    string
		CreatedAt string
	}
	e.call(http.MethodGet, "/api/v1/posts/hello-world", nil, &got, http.StatusOK)
	if got.Title != "你好世界" || strings.Contains(got.ContentMd, "more") || strings.Count(got.ContentMd, "/uploads/") != 2 ||
		len(got.Tags) != 3 || len(got.Covers) != 1 || !strings.HasPrefix(got.CreatedAt, "2021-03-0") {
		t.Fatalf("hello-world: %+v", got)
	}
	for _, it := range out.Items {
		if it.Title == "未完成" && it.Status != "draft" {
			t.Fatalf("draft dir not draft: %+v", it)
		}
	}

	// Hugo：TOML front matter + 页面包；重名 slug 自动加序号；dry 只预览不写入
	hugo := map[string][]byte{
		"site/hugo.toml": []byte("baseURL = 'x'\n"),
		"site/content/posts/hello-world/index.md": []byte("+++\ntitle = \"Bundle\"\ndate = 2022-01-02T03:04:05+08:00\ntags = [\"a\", \"b\"]\ndraft = true\n+++\n{{< figure src=\"pic.png\" >}}\n"),
		"site/content/posts/hello-world/pic.png":  png,
	}
	body, ct = multipartBody(t, "files", hugo)
	r = e.do(http.MethodPost, "/api/v1/admin/import?dry=1", body, map[string]string{"Content-Type": ct, "Authorization": "Bearer " + e.token})
	json.NewDecoder(r.Body).Decode(&out)
	r.Body.Close()
	if out.Platform != "hugo" || out.Created != 0 || len(out.Items) != 1 || out.Items[0].Slug != "hello-world-2" || out.Items[0].Status != "draft" {
		t.Fatalf("hugo dry: %+v", out)
	}
}

// 全量备份：数据库（含用户、文章、设置）+ 全部素材，含压缩前原版与裁切原图；缩略图不进包
func TestBackupIsComplete(t *testing.T) {
	e := newEnv(t)
	img := e.uploadFiles(pngOf(color.RGBA{R: 1, G: 2, B: 3, A: 255}))[0]
	if res := e.server.compressNamed(img.Name, 70); res == nil || res.Error != "" {
		t.Fatalf("compress: %+v", res)
	}
	r := e.do(http.MethodGet, "/uploads/thumbs/"+strings.TrimSuffix(img.Name, ".png")+".webp", nil, nil)
	r.Body.Close()
	var b struct{ Name string }
	e.call(http.MethodPost, "/api/v1/admin/backups", nil, &b, http.StatusCreated)
	zr, err := zip.OpenReader(e.server.BackupDir + "/" + b.Name)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	has := map[string]bool{}
	for _, f := range zr.File {
		has[f.Name] = true
		if strings.HasPrefix(f.Name, "uploads/thumbs/") || strings.HasSuffix(f.Name, "-wal") {
			t.Fatalf("unexpected entry %s", f.Name)
		}
	}
	for _, want := range []string{"data/myself.db", "uploads/.precompress/" + img.Name, "uploads/.originals/" + img.Name} {
		if !has[want] {
			t.Fatalf("backup missing %s: %v", want, has)
		}
	}
}
