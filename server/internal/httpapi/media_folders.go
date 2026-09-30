package httpapi

import (
	"net/http"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 素材文件夹：纯组织用的虚拟路径（media.folder，如「封面/2026」），磁盘文件名与引用地址不受影响。
// 空文件夹记在 media_folders 里；有素材的文件夹即使没有记录也会列出。

const (
	folderMaxDepth = 5
	folderMaxSeg   = 40
)

// cleanFolder 规范化文件夹路径：按 / 分段、去空白、去空段；非法（控制字符、反斜杠、. / ..、过长、过深）返回 ok=false。
// 空串表示根目录（未归类）。
func cleanFolder(raw string) (string, bool) {
	var segs []string
	for _, seg := range strings.Split(raw, "/") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if seg == "." || seg == ".." || utf8.RuneCountInString(seg) > folderMaxSeg || strings.ContainsAny(seg, `\`) ||
			strings.IndexFunc(seg, unicode.IsControl) >= 0 {
			return "", false
		}
		segs = append(segs, seg)
	}
	if len(segs) > folderMaxDepth {
		return "", false
	}
	return strings.Join(segs, "/"), true
}

// subfolders 匹配某文件夹所有子路径的 LIKE 模式（转义通配符，ESCAPE '\'）
func subfolders(folder string) string {
	return strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(folder) + "/%"
}

// ensureFolder 记录文件夹及其各级上级（空文件夹也能列出）
func (s *Server) ensureFolder(folder string) error {
	if folder == "" {
		return nil
	}
	segs := strings.Split(folder, "/")
	for i := range segs {
		if _, err := s.DB.Exec(`INSERT INTO media_folders (path) VALUES (?) ON CONFLICT(path) DO NOTHING`, strings.Join(segs[:i+1], "/")); err != nil {
			return err
		}
	}
	return nil
}

type folderInfo struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// GET /admin/media/folders 文件夹列表（含各自直接包含的素材数）
func (s *Server) listFolders(w http.ResponseWriter, _ *http.Request) {
	counts := map[string]int{}
	rows, err := s.DB.Query(`SELECT folder, COUNT(*) FROM media WHERE folder != '' GROUP BY folder`)
	if err != nil {
		fail(w, err)
		return
	}
	for rows.Next() {
		var f string
		var n int
		if rows.Scan(&f, &n) == nil {
			counts[f] = n
		}
	}
	rows.Close()
	paths := map[string]bool{}
	if rows, err := s.DB.Query(`SELECT path FROM media_folders`); err == nil {
		for rows.Next() {
			var p string
			if rows.Scan(&p) == nil {
				paths[p] = true
			}
		}
		rows.Close()
	}
	// 有素材的文件夹连同上级都列出
	for f := range counts {
		segs := strings.Split(f, "/")
		for i := range segs {
			paths[strings.Join(segs[:i+1], "/")] = true
		}
	}
	out := make([]folderInfo, 0, len(paths))
	for p := range paths {
		out = append(out, folderInfo{Path: p, Count: counts[p]})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	writeJSON(w, http.StatusOK, out)
}

// POST /admin/media/folders {path} 新建文件夹
func (s *Server) createFolder(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Path string `json:"path"`
	}
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	folder, ok := cleanFolder(b.Path)
	if !ok || folder == "" {
		writeError(w, http.StatusBadRequest, "invalid_folder")
		return
	}
	if err := s.ensureFolder(folder); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"path": folder})
}

// PUT /admin/media/folders {from, to} 重命名 / 移动文件夹（连同子文件夹与其中素材）
func (s *Server) renameFolder(w http.ResponseWriter, r *http.Request) {
	var b struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	from, ok1 := cleanFolder(b.From)
	to, ok2 := cleanFolder(b.To)
	if !ok1 || !ok2 || from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "invalid_folder")
		return
	}
	if to == from || strings.HasPrefix(to+"/", from+"/") {
		writeError(w, http.StatusBadRequest, "invalid_folder") // 不能移到自己或自己的子文件夹里
		return
	}
	if len(strings.Split(to, "/"))+s.folderDepthBelow(from) > folderMaxDepth {
		writeError(w, http.StatusBadRequest, "folder_too_deep")
		return
	}
	tx, err := s.DB.Begin()
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback() //nolint:errcheck
	// 自己与所有子路径：前缀替换
	for _, q := range []string{
		`UPDATE media SET folder = ? || substr(folder, ?) WHERE folder = ? OR folder LIKE ? ESCAPE '\'`,
		`UPDATE OR IGNORE media_folders SET path = ? || substr(path, ?) WHERE path = ? OR path LIKE ? ESCAPE '\'`,
	} {
		if _, err := tx.Exec(q, to, utf8.RuneCountInString(from)+1, from, subfolders(from)); err != nil {
			fail(w, err)
			return
		}
	}
	// UPDATE OR IGNORE 遇到目标已存在时留下的旧记录
	if _, err := tx.Exec(`DELETE FROM media_folders WHERE path = ? OR path LIKE ? ESCAPE '\'`, from, subfolders(from)); err != nil {
		fail(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		fail(w, err)
		return
	}
	if err := s.ensureFolder(to); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": to})
}

// folderDepthBelow 文件夹下最深的子路径比它多几层
func (s *Server) folderDepthBelow(folder string) int {
	deepest := 0
	collect := func(q string) {
		rows, err := s.DB.Query(q, subfolders(folder))
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var p string
			if rows.Scan(&p) == nil {
				deepest = max(deepest, strings.Count(strings.TrimPrefix(p, folder), "/"))
			}
		}
	}
	collect(`SELECT DISTINCT folder FROM media WHERE folder LIKE ? ESCAPE '\'`)
	collect(`SELECT path FROM media_folders WHERE path LIKE ? ESCAPE '\'`)
	return deepest
}

// DELETE /admin/media/folders?path= 删除文件夹：其中的素材与子文件夹上移到父级，不删任何文件
func (s *Server) deleteFolder(w http.ResponseWriter, r *http.Request) {
	folder, ok := cleanFolder(r.URL.Query().Get("path"))
	if !ok || folder == "" {
		writeError(w, http.StatusBadRequest, "invalid_folder")
		return
	}
	parent := ""
	if i := strings.LastIndex(folder, "/"); i >= 0 {
		parent = folder[:i]
	}
	prefix := subfolders(folder)
	tx, err := s.DB.Begin()
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback() //nolint:errcheck
	// 直接包含的素材 → 父级；子文件夹整体上移一层
	stmts := []struct {
		q    string
		args []any
	}{
		{`UPDATE media SET folder = ? WHERE folder = ?`, []any{parent, folder}},
		{`UPDATE media SET folder = CASE WHEN ? = '' THEN substr(folder, ?) ELSE ? || '/' || substr(folder, ?) END WHERE folder LIKE ? ESCAPE '\'`,
			[]any{parent, utf8.RuneCountInString(folder) + 2, parent, utf8.RuneCountInString(folder) + 2, prefix}},
		{`UPDATE OR IGNORE media_folders SET path = CASE WHEN ? = '' THEN substr(path, ?) ELSE ? || '/' || substr(path, ?) END WHERE path LIKE ? ESCAPE '\'`,
			[]any{parent, utf8.RuneCountInString(folder) + 2, parent, utf8.RuneCountInString(folder) + 2, prefix}},
		{`DELETE FROM media_folders WHERE path = ? OR path LIKE ? ESCAPE '\'`, []any{folder, prefix}},
	}
	for _, st := range stmts {
		if _, err := tx.Exec(st.q, st.args...); err != nil {
			fail(w, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /admin/media/move {names, folder} 批量移动素材到文件夹（空串 = 未归类）
func (s *Server) moveMedia(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Names  []string `json:"names"`
		Folder string   `json:"folder"`
	}
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	folder, ok := cleanFolder(b.Folder)
	if !ok || len(b.Names) == 0 || len(b.Names) > 500 {
		writeError(w, http.StatusBadRequest, "invalid_folder")
		return
	}
	if err := s.ensureFolder(folder); err != nil {
		fail(w, err)
		return
	}
	n := 0
	for _, name := range b.Names {
		if name = safeName(name); name == "" || !fileExists(s.UploadDir+"/"+name) {
			continue
		}
		if _, err := s.DB.Exec(`INSERT INTO media (name, folder) VALUES (?, ?) ON CONFLICT(name) DO UPDATE SET folder = excluded.folder`, name, folder); err != nil {
			fail(w, err)
			return
		}
		n++
	}
	writeJSON(w, http.StatusOK, map[string]int{"moved": n})
}
