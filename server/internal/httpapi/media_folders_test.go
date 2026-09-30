package httpapi

import (
	"bytes"
	"encoding/json"
	"image/color"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"
)

func TestMediaFolders(t *testing.T) {
	e := newEnv(t)
	// 上传直接进文件夹（folder 字段排在文件前）
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("folder", " 封面 / 2026 ")
	part, _ := mw.CreateFormFile("files", "a.png")
	part.Write(pngOf(color.RGBA{R: 40, A: 255}))
	mw.Close()
	res := e.do(http.MethodPost, "/api/v1/admin/media", &buf, map[string]string{"Content-Type": mw.FormDataContentType(), "Authorization": "Bearer " + e.token})
	data, _ := io.ReadAll(res.Body)
	res.Body.Close()
	var up []mediaItem
	json.Unmarshal(data, &up)
	if len(up) != 1 || up[0].Folder != "封面/2026" {
		t.Fatalf("upload into folder: %s", data)
	}
	b := e.uploadFiles(pngOf(color.RGBA{G: 40, A: 255}))[0]

	e.call(http.MethodPost, "/api/v1/admin/media/folders", map[string]string{"path": "空文件夹"}, nil, http.StatusCreated)
	for _, bad := range []string{"..", "a/../b", `a\b`, "1/2/3/4/5/6", ""} {
		e.call(http.MethodPost, "/api/v1/admin/media/folders", map[string]string{"path": bad}, nil, http.StatusBadRequest)
	}
	e.call(http.MethodPost, "/api/v1/admin/media/move", map[string]any{"names": []string{b.Name}, "folder": "封面"}, nil, http.StatusOK)

	folders := func() map[string]int {
		var out []folderInfo
		e.call(http.MethodGet, "/api/v1/admin/media/folders", nil, &out, http.StatusOK)
		m := map[string]int{}
		for _, f := range out {
			m[f.Path] = f.Count
		}
		return m
	}
	if f := folders(); f["封面"] != 1 || f["封面/2026"] != 1 || !hasKey(f, "空文件夹") {
		t.Fatalf("folders: %v", f)
	}
	// 重命名：子文件夹与素材跟着走（中文按字符截取）
	e.call(http.MethodPut, "/api/v1/admin/media/folders", map[string]string{"from": "封面", "to": "图片/头图"}, nil, http.StatusOK)
	if f := folders(); f["图片/头图"] != 1 || f["图片/头图/2026"] != 1 || hasKey(f, "封面") {
		t.Fatalf("after rename: %v", f)
	}
	e.call(http.MethodPut, "/api/v1/admin/media/folders", map[string]string{"from": "图片", "to": "图片/头图/x"}, nil, http.StatusBadRequest)
	// 删除：素材与子文件夹上移一层，文件不删
	e.call(http.MethodDelete, "/api/v1/admin/media/folders?path="+url.QueryEscape("图片/头图"), nil, nil, http.StatusOK)
	if f := folders(); f["图片"] != 1 || f["图片/2026"] != 1 || hasKey(f, "图片/头图") {
		t.Fatalf("after delete: %v", f)
	}
	info, _ := e.server.fileInfo(up[0].Name)
	if info.Folder != "图片/2026" {
		t.Fatalf("item folder after delete: %q", info.Folder)
	}
}

func hasKey(m map[string]int, k string) bool { _, ok := m[k]; return ok }
