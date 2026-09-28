// Command myself-server 是 Myself 博客的 Go 后端入口。
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/config"
	"myself/server/internal/httpapi"
	"myself/server/internal/store"
	"myself/server/web"
)

func main() {
	root := rootDir()
	db, err := store.Open(filepath.Join(root, "data"))
	if err != nil {
		log.Fatalf("[myself-server] open database: %v", err)
	}
	defer db.Close()

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
		Frontend:  web.Handler(),
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
		ReadTimeout:       60 * time.Second, // 上传接口单独放宽
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Printf("[myself-server] listening on http://localhost:%s", port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	log.Print("[myself-server] shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("[myself-server] shutdown: %v", err)
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
