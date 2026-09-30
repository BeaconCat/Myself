package httpapi

import (
	"archive/zip"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// maxBackupUpload 上传备份包的大小上限
const maxBackupUpload = 4 << 30

// 恢复时的解压上限：单个条目（大于素材上传上限）与总量，防解压炸弹塞满磁盘
const (
	restoreEntryMax = 1 << 30
	restoreTotalMax = 64 << 30
)

var errBadBackup = errors.New("invalid_backup")

// restoreBackup 从备份包恢复全站：数据库按表回灌 + 素材目录镜像还原。调用前须持有 backupMu。
//
// 不替换正在使用的数据库文件（连接池一直开着），而是把包里的 myself.db 解到临时目录后 ATTACH，
// 在一个事务里逐表「清空 → 按两边共有的列拷回」：旧版本备份缺的新列取默认值，新增的表清空，
// 恢复后的库结构始终是当前版本。素材：包里有的写回，包里没有的删除；缩略图目录清空按需重建。
func (s *Server) restoreBackup(zipPath string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return errBadBackup
	}
	defer zr.Close()

	var dbFile *zip.File
	uploads := map[string]*zip.File{}
	var total uint64
	for _, f := range zr.File {
		if f.UncompressedSize64 > restoreEntryMax {
			return errBadBackup
		}
		total += f.UncompressedSize64
		if total > restoreTotalMax {
			return errBadBackup
		}
		name := path.Clean(strings.ReplaceAll(f.Name, `\`, "/"))
		if f.FileInfo().IsDir() || strings.HasPrefix(name, "../") || path.IsAbs(name) {
			continue
		}
		switch {
		case name == "data/myself.db":
			dbFile = f
		case strings.HasPrefix(name, "uploads/"):
			rel := strings.TrimPrefix(name, "uploads/")
			if rel != "" && !strings.Contains(rel, "..") {
				uploads[rel] = f
			}
		}
	}
	if dbFile == nil {
		return errBadBackup
	}

	tmp, err := os.MkdirTemp("", "myself-restore-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	tmpDB := filepath.Join(tmp, "restore.db")
	if err := extractTo(dbFile, tmpDB); err != nil {
		return err
	}
	if err := s.restoreDB(tmpDB); err != nil {
		return err
	}
	return s.restoreUploads(uploads)
}

// restoreDB 逐表回灌（见 restoreBackup）。
func (s *Server) restoreDB(src string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	conn, err := s.DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `ATTACH DATABASE ? AS src`, filepath.ToSlash(src)); err != nil {
		return errBadBackup
	}
	defer conn.ExecContext(ctx, `DETACH DATABASE src`) //nolint:errcheck
	// 包里的库至少得有文章表，才认定是本站备份
	var n int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM src.sqlite_master WHERE type = 'table' AND name = 'posts'`).Scan(&n); err != nil || n == 0 {
		return errBadBackup
	}
	tables, err := tableNames(ctx, conn, "main")
	if err != nil {
		return err
	}
	srcTables := map[string]bool{}
	st, err := tableNames(ctx, conn, "src")
	if err != nil {
		return err
	}
	for _, t := range st {
		srcTables[t] = true
	}

	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		return err
	}
	defer conn.ExecContext(ctx, `PRAGMA foreign_keys = ON`) //nolint:errcheck
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	for _, t := range tables {
		if _, err := tx.ExecContext(ctx, `DELETE FROM main."`+t+`"`); err != nil {
			return fmt.Errorf("clear %s: %w", t, err)
		}
		if !srcTables[t] {
			continue
		}
		mainCols, err := columnNames(ctx, tx, "main", t)
		if err != nil {
			return err
		}
		srcCols, err := columnNames(ctx, tx, "src", t)
		if err != nil {
			return err
		}
		has := map[string]bool{}
		for _, c := range srcCols {
			has[c] = true
		}
		var common []string
		for _, c := range mainCols {
			if has[c] {
				common = append(common, `"`+c+`"`)
			}
		}
		if len(common) == 0 {
			continue
		}
		cols := strings.Join(common, ", ")
		if _, err := tx.ExecContext(ctx, `INSERT INTO main."`+t+`" (`+cols+`) SELECT `+cols+` FROM src."`+t+`"`); err != nil {
			return fmt.Errorf("restore %s: %w", t, err)
		}
	}
	return tx.Commit()
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func tableNames(ctx context.Context, q queryer, schema string) ([]string, error) {
	rows, err := q.QueryContext(ctx, `SELECT name FROM `+schema+`.sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func columnNames(ctx context.Context, q queryer, schema, table string) ([]string, error) {
	rows, err := q.QueryContext(ctx, `SELECT name FROM pragma_table_info(?, ?)`, table, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// restoreUploads 素材目录镜像为备份时的状态（原图备份 / 压缩前备份一并还原，缩略图清空按需重建）。
func (s *Server) restoreUploads(files map[string]*zip.File) error {
	keep := map[string]bool{}
	for rel, f := range files {
		if strings.HasPrefix(rel, "thumbs/") {
			continue
		}
		dst := filepath.Join(s.UploadDir, filepath.FromSlash(rel))
		if !strings.HasPrefix(dst, filepath.Clean(s.UploadDir)+string(os.PathSeparator)) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := extractTo(f, dst); err != nil {
			return err
		}
		keep[filepath.Clean(dst)] = true
	}
	os.RemoveAll(s.thumbsDir)
	_ = os.MkdirAll(s.thumbsDir, 0o755)
	return filepath.WalkDir(s.UploadDir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !keep[filepath.Clean(p)] {
			os.Remove(p)
		}
		return nil
	})
}

func extractTo(f *zip.File, dst string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	tmp := dst + ".restoring"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	// 按实际读出的字节限流（声明的大小可以作假）
	if n, err := io.Copy(out, io.LimitReader(rc, restoreEntryMax+1)); err != nil || n > restoreEntryMax {
		out.Close()
		os.Remove(tmp)
		if err == nil {
			err = errBadBackup
		}
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

// looksLikeBackup 是 zip 且含 data/myself.db（进一步的库结构校验在 restoreDB 里）
func looksLikeBackup(p string) bool {
	zr, err := zip.OpenReader(p)
	if err != nil {
		return false
	}
	defer zr.Close()
	for _, f := range zr.File {
		if path.Clean(strings.ReplaceAll(f.Name, `\`, "/")) == "data/myself.db" {
			return true
		}
	}
	return false
}

// runRestore 恢复前先做一份安全备份（恢复出错可回到此刻），返回安全备份名。
func (s *Server) runRestore(w http.ResponseWriter, zipPath string) {
	// 压缩任务会同时读写素材文件：进行中则拒绝，恢复期间也不接新任务
	if !s.jobs.pause() {
		writeError(w, http.StatusConflict, "job_running")
		return
	}
	defer s.jobs.resume()
	safety, err := s.createBackupWith(false)
	if err != nil {
		fail(w, err)
		return
	}
	s.backupMu.Lock()
	defer s.backupMu.Unlock()
	if err := s.restoreBackup(zipPath); err != nil {
		if errors.Is(err, errBadBackup) {
			writeError(w, http.StatusBadRequest, "invalid_backup")
			return
		}
		log.Printf("[restore] %v", err)
		fail(w, err)
		return
	}
	// 换一把签名密钥：备份里的旧密钥、以及备份之后已吊销的会话都不能再用（所有人需重新登录）
	if err := s.Auth.RotateSecret(); err != nil {
		log.Printf("[restore] rotate secret: %v", err)
	}
	log.Printf("[restore] restored from %s (safety backup %s)", filepath.Base(zipPath), safety)
	writeJSON(w, http.StatusOK, map[string]string{"safety": safety})
}

// POST /admin/backups/{name}/restore 从已有备份恢复
func (s *Server) restoreFromBackup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !backupNameRe.MatchString(name) {
		writeError(w, http.StatusBadRequest, "invalid_name")
		return
	}
	p := filepath.Join(s.BackupDir, name)
	if !fileExists(p) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if !looksLikeBackup(p) {
		writeError(w, http.StatusBadRequest, "invalid_backup")
		return
	}
	s.runRestore(w, p)
}

// POST /admin/backups/restore 上传备份包并恢复（multipart 字段 file）；上传的包同时留存在备份列表里
func (s *Server) restoreFromUpload(w http.ResponseWriter, r *http.Request) {
	_ = http.NewResponseController(w).SetReadDeadline(time.Now().Add(30 * time.Minute))
	r.Body = http.MaxBytesReader(w, r.Body, maxBackupUpload)
	mr, err := r.MultipartReader()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_multipart")
		return
	}
	for {
		part, err := mr.NextPart()
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_multipart")
			return
		}
		if part.FormName() != "file" {
			continue
		}
		name, out, err := s.newBackupFile()
		if err != nil {
			fail(w, err)
			return
		}
		dst := filepath.Join(s.BackupDir, name)
		_, err = io.Copy(out, part)
		out.Close()
		if err != nil {
			os.Remove(dst)
			writeError(w, http.StatusRequestEntityTooLarge, "file_too_large")
			return
		}
		if !looksLikeBackup(dst) {
			os.Remove(dst)
			writeError(w, http.StatusBadRequest, "invalid_backup")
			return
		}
		s.runRestore(w, dst)
		return
	}
}
