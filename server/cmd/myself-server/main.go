// Command myself-server 是 Myself 博客的 Go 后端入口。
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/config"
	"myself/server/internal/httpapi"
	"myself/server/internal/store"
)

func main() {
	root := rootDir()
	db, err := store.Open(filepath.Join(root, "data"))
	if err != nil {
		log.Fatalf("[myself-server] open database: %v", err)
	}
	defer db.Close()

	if err := db.SeedIfEmpty(); err != nil {
		log.Fatalf("[myself-server] seed: %v", err)
	}
	authSvc := auth.New(db)
	if err := authSvc.Init(); err != nil {
		log.Fatalf("[myself-server] init auth: %v", err)
	}

	srv := httpapi.New(httpapi.Deps{
		DB:        db,
		Auth:      authSvc,
		Config:    config.New(db),
		UploadDir: filepath.Join(root, "uploads"),
		BackupDir: filepath.Join(root, "backups"),
		DataDir:   filepath.Join(root, "data"),
	})
	srv.StartAutoBackup()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3100"
	}
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("[myself-server] listening on http://localhost:%s", port)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// rootDir 返回服务根目录（data/uploads/backups 所在处）。
// 优先环境变量 MYSELF_ROOT，否则取当前工作目录。
func rootDir() string {
	if v := os.Getenv("MYSELF_ROOT"); v != "" {
		return v
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
}
