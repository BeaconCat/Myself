package httpapi

import (
	"archive/zip"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"myself/server/internal/config"
)

var backupNameRe = regexp.MustCompile(`^backup-[\d-]+\.zip$`)

func backupTimestamp() string {
	// 与旧实现一致：ISO 时间去掉 [:T] → 2026-07-10-18-28-22
	return time.Now().UTC().Format("2006-01-02-15-04-05")
}

// createBackup 全站备份：数据库（文章/随想/用户/配置） + 上传素材（含原图备份）。
func (s *Server) createBackup() (string, error) {
	// 先把 WAL 合并进主库文件，保证 zip 内的 .db 自洽可单独恢复。
	if _, err := s.DB.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		log.Printf("[backup] checkpoint: %v", err)
	}
	name := "backup-" + backupTimestamp() + ".zip"
	out, err := os.Create(filepath.Join(s.BackupDir, name))
	if err != nil {
		return "", err
	}
	zw := zip.NewWriter(out)
	for _, dir := range []struct{ path, prefix string }{
		{s.DataDir, "data"},
		{s.UploadDir, "uploads"},
	} {
		if !dirExists(dir.path) {
			continue
		}
		if err := addDirToZip(zw, dir.path, dir.prefix); err != nil {
			zw.Close()
			out.Close()
			os.Remove(filepath.Join(s.BackupDir, name))
			return "", err
		}
	}
	if err := zw.Close(); err != nil {
		out.Close()
		return "", err
	}
	return name, out.Close()
}

func addDirToZip(zw *zip.Writer, root, prefix string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = prefix + "/" + filepath.ToSlash(rel)
		hdr.Method = zip.Deflate
		wr, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(wr, f)
		return err
	})
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

var autoBackupOnce sync.Once

// StartAutoBackup 自动备份调度：间隔小时数取自站点配置，0 关闭；每分钟核对一次。
func (s *Server) StartAutoBackup() {
	autoBackupOnce.Do(func() {
		go func() {
			var last time.Time
			for range time.Tick(time.Minute) {
				hours := config.Num(config.Sub(s.Config.Get(), "backup"), "autoHours")
				if hours <= 0 {
					continue
				}
				if time.Since(last) >= time.Duration(hours*float64(time.Hour)) {
					last = time.Now()
					if _, err := s.createBackup(); err != nil {
						log.Printf("[auto-backup] %v", err)
					}
				}
			}
		}()
	})
}

type backupInfo struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"createdAt"`
}

// GET /admin/backups 备份列表
func (s *Server) listBackups(w http.ResponseWriter, _ *http.Request) {
	entries, err := os.ReadDir(s.BackupDir)
	if err != nil {
		fail(w, err)
		return
	}
	items := []backupInfo{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".zip") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, backupInfo{Name: e.Name(), Size: info.Size(), CreatedAt: isoTime(info.ModTime())})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })
	writeJSON(w, http.StatusOK, items)
}

// POST /admin/backups 立即备份
func (s *Server) createBackupNow(w http.ResponseWriter, _ *http.Request) {
	name, err := s.createBackup()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"name": name})
}

// GET /admin/backups/{name} 下载
func (s *Server) downloadBackup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !backupNameRe.MatchString(name) {
		writeError(w, http.StatusBadRequest, "bad_name")
		return
	}
	full := filepath.Join(s.BackupDir, name)
	if !fileExists(full) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Content-Type", "application/zip")
	http.ServeFile(w, r, full)
}

// DELETE /admin/backups/{name}
func (s *Server) deleteBackup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !backupNameRe.MatchString(name) {
		writeError(w, http.StatusBadRequest, "bad_name")
		return
	}
	os.Remove(filepath.Join(s.BackupDir, name))
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
