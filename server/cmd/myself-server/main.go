// Command myself-server 是 Myself 博客的 Go 后端入口。
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	goruntime "runtime"
	"syscall"
	"time"

	"myself/server/internal/auth"
	"myself/server/internal/config"
	"myself/server/internal/httpapi"
	"myself/server/internal/store"
	"myself/server/internal/updater"
	"myself/server/web"
)

var version = "dev"
var commit = "unknown"
var buildDate = "unknown"

func main() {
	configPath := flag.String("config", "", "runtime JSON config (default: ./config.json)")
	showVersion := flag.Bool("version", false, "print version and exit")
	showVersionJSON := flag.Bool("version-json", false, "print machine-readable version and exit")
	applyPlan := flag.String("apply-update", "", "internal updater helper")
	recoverPlan := flag.String("recover-update", "", "internal update recovery")
	flag.Parse()
	if *showVersionJSON {
		_ = json.NewEncoder(os.Stdout).Encode(buildInfo())
		return
	}
	if *applyPlan != "" {
		if err := updater.Apply(*applyPlan); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *showVersion {
		fmt.Printf("Myself %s (commit %s, built %s)\n", version, commit, buildDate)
		return
	}
	filename := *configPath
	if filename == "" {
		filename = os.Getenv("MYSELF_CONFIG")
	}
	explicit := filename != ""
	if !explicit {
		filename = "config.json"
	}
	runtime, err := loadRuntime(filename, explicit)
	if err != nil {
		log.Fatalf("[myself-server] %v", err)
	}
	filename, err = filepath.Abs(filename)
	if err != nil {
		log.Fatal(err)
	}
	if *recoverPlan != "" {
		if err := recoverUpdate(runtime, *recoverPlan); err != nil {
			log.Fatal(err)
		}
		return
	}
	setupCode := ""
	for {
		next, err := serve(runtime, filename, setupCode)
		if err != nil {
			log.Fatalf("[myself-server] %v", err)
		}
		if next == nil {
			return
		}
		if next.UpdatePlan != "" {
			if err := updater.Launch(next.UpdatePlan); err != nil {
				updater.Abort(next.UpdatePlan, err)
				log.Printf("[update] helper launch failed: %v", err)
				continue
			}
			return
		}
		runtime, setupCode = next.Config, next.SetupCode
	}
}

type restartRequest struct {
	Config     runtimeOptions
	SetupCode  string
	UpdatePlan string
}

func buildInfo() updater.Build {
	return updater.Build{Version: version, Commit: commit, Date: buildDate, OS: goruntime.GOOS, Arch: goruntime.GOARCH, UpdateProtocol: updater.Protocol}
}

func recoverUpdate(runtime runtimeOptions, filename string) error {
	plan, err := updater.ReadPlan(filename)
	if err != nil {
		return err
	}
	if plan.Root != runtime.Root {
		return errors.New("recovery data root mismatch")
	}
	db, err := store.OpenConfigured(filepath.Join(runtime.Root, "data"), runtime.Database)
	if err != nil {
		return err
	}
	defer db.Close()
	srv := httpapi.New(httpapi.Deps{DB: db, Config: config.New(db), UploadDir: filepath.Join(runtime.Root, "uploads"), BackupDir: filepath.Join(runtime.Root, "backups"), DataDir: filepath.Join(runtime.Root, "data")})
	return srv.RecoverUpdate(plan.Backup)
}

func serve(runtime runtimeOptions, configFile, setupCode string) (*restartRequest, error) {
	root := runtime.Root
	db, err := store.OpenConfigured(filepath.Join(root, "data"), runtime.Database)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	authSvc := auth.New(db)
	if err := authSvc.InitWithSetupCode(setupCode); err != nil {
		return nil, fmt.Errorf("init auth: %w", err)
	}
	restart := make(chan restartRequest, 1)

	srv := httpapi.New(httpapi.Deps{
		DB:        db,
		Auth:      authSvc,
		Config:    config.New(db),
		UploadDir: filepath.Join(root, "uploads"),
		BackupDir: filepath.Join(root, "backups"),
		DataDir:   filepath.Join(root, "data"),
		Frontend:  web.Handler(),
		// 初始化选 Demo 时把示例用到的默认封面导入素材库
		DemoCovers: web.Covers(),
		Build:      buildInfo(),
		ConfigureDatabase: func(database store.DatabaseConfig, code string) error {
			if database.Driver == "current" {
				database = runtime.Database
				if database.Driver == "" {
					database.Driver = "sqlite"
				}
			}
			candidate, err := store.OpenConfigured(filepath.Join(root, "data"), database)
			if err != nil {
				return err
			}
			count, err := candidate.CountUsers(store.RoleAdmin)
			candidate.Close()
			if err != nil {
				return err
			}
			if count != 0 {
				return errors.New("target database is already initialized")
			}
			next := runtime
			next.Database = database
			if err := saveRuntime(configFile, next); err != nil {
				return err
			}
			restart <- restartRequest{Config: next, SetupCode: code}
			return nil
		},
	})
	updates, err := updater.New(updater.Options{
		Build: buildInfo(), Repository: runtime.Updates.Repository, Root: root, Port: runtime.Port,
		Args:     append([]string{}, os.Args[1:]...),
		Disabled: runtime.Updates.Disabled || os.Getenv("MYSELF_DISABLE_SELF_UPDATE") == "1",
		Prepare:  srv.PrepareUpdate,
		Resume:   srv.ResumeAfterUpdate,
		Ready:    func(plan string) error { restart <- restartRequest{UpdatePlan: plan}; return nil },
	})
	if err != nil {
		log.Printf("[update] unavailable: %v", err)
	} else {
		srv.Updates = updates
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	background, stopBackground := context.WithCancel(context.Background())
	defer stopBackground()
	srv.StartAutoBackup(background)

	port := runtime.Port
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second, // 上传接口单独放宽
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("[myself-server] listening on http://localhost:%s", port)
		serverErrors <- httpServer.ListenAndServe()
	}()
	var next *restartRequest
	select {
	case <-ctx.Done():
	case request := <-restart:
		next = &request
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return nil, err
		}
	}
	log.Print("[myself-server] shutting down")
	stopBackground()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return nil, fmt.Errorf("shutdown: %w", err)
	}
	return next, nil
}
