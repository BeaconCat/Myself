package httpapi

import (
	"archive/zip"
	"fmt"
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
)

var backupNameRe = regexp.MustCompile(`^backup-[\d-]+\.zip$`)

func backupTimestamp() string {
	// 与旧实现一致：ISO 时间去掉 [:T] → 2026-07-10-18-28-22
	return time.Now().UTC().Format("2006-01-02-15-04-05")
}

// createBackup 全站备份：数据库（文章/随想/用户/配置） + 上传素材（含原图备份）。
func (s *Server) createBackup() (string, error) {
	s.backupMu.Lock()
	defer s.backupMu.Unlock()
	defer s.pruneBackups(keepBackups)
	// 先把 WAL 合并进主库文件，保证 zip 内的 .db 自洽可单独恢复。
	if _, err := s.DB.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		log.Printf("[backup] checkpoint: %v", err)
	}
	name, out, err := s.newBackupFile()
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

// newBackupFile 新建一个备份文件；同一秒内多次（如恢复前的安全备份、上传的备份包）追加序号避免重名
func (s *Server) newBackupFile() (string, *os.File, error) {
	stamp := backupTimestamp()
	name := "backup-" + stamp + ".zip"
	out, err := os.OpenFile(filepath.Join(s.BackupDir, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	for i := 1; err != nil && os.IsExist(err) && i < 100; i++ {
		name = fmt.Sprintf("backup-%s-%d.zip", stamp, i)
		out, err = os.OpenFile(filepath.Join(s.BackupDir, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	}
	return name, out, err
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
				hours := s.Config.Typed().Backup.AutoHours
				if hours <= 0 {
					continue
				}
				hours = max(hours, 1)
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

// keepBackups 保留最近的备份份数，更早的自动清理。
const keepBackups = 20

// pruneBackups 按文件名（时间戳）排序，只保留最近 keep 份。
func (s *Server) pruneBackups(keep int) {
	entries, err := os.ReadDir(s.BackupDir)
	if err != nil {
		return
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && backupNameRe.MatchString(e.Name()) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for i := 0; i < len(names)-keep; i++ {
		if err := os.Remove(filepath.Join(s.BackupDir, names[i])); err != nil {
			log.Printf("[backup] prune %s: %v", names[i], err)
		}
	}
}
