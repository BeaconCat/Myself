package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"myself/server/internal/store"
)

// Runtime options are separate from the site settings stored in SQLite.
type runtimeOptions struct {
	Port     string               `json:"port"`
	Root     string               `json:"root"`
	Database store.DatabaseConfig `json:"database"`
	Updates  struct {
		Repository string `json:"repository,omitempty"`
		Disabled   bool   `json:"disabled,omitempty"`
	} `json:"updates,omitempty"`
}

// loadRuntime preserves legacy cwd/environment defaults when no file is present.
// File-relative roots remain stable when a service changes working directory.
func loadRuntime(filename string, explicit bool) (runtimeOptions, error) {
	cfg := runtimeOptions{Port: "3100"}
	f, err := os.Open(filename)
	if err != nil && (!os.IsNotExist(err) || explicit) {
		return cfg, fmt.Errorf("read runtime config: %w", err)
	}
	if err == nil {
		defer f.Close()
		dec := json.NewDecoder(f)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&cfg); err != nil {
			return cfg, fmt.Errorf("decode runtime config: %w", err)
		}
		if err := dec.Decode(new(any)); err != io.EOF {
			return cfg, fmt.Errorf("runtime config must contain one JSON object")
		}
		if cfg.Root != "" && !filepath.IsAbs(cfg.Root) {
			cfg.Root = filepath.Join(filepath.Dir(filename), cfg.Root)
		}
	}
	if port := os.Getenv("PORT"); port != "" {
		cfg.Port = port
	}
	if root := os.Getenv("MYSELF_ROOT"); root != "" {
		cfg.Root = root
	}
	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return cfg, fmt.Errorf("port must be an integer between 1 and 65535")
	}
	if cfg.Root == "" {
		cfg.Root = "."
	}
	cfg.Root, err = filepath.Abs(cfg.Root)
	return cfg, err
}

func saveRuntime(filename string, cfg runtimeOptions) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(filename), ".myself-config-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filename)
}
