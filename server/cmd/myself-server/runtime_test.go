package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuntimeConfig(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("MYSELF_ROOT", "")
	dir := t.TempDir()
	file := filepath.Join(dir, "config.json")
	defaults, err := loadRuntime(file, false)
	if err != nil || defaults.Port != "3100" {
		t.Fatalf("legacy defaults: %+v, %v", defaults, err)
	}
	if _, err := loadRuntime(file, true); err == nil {
		t.Fatal("explicit missing file accepted")
	}
	if err := os.WriteFile(file, []byte(`{"port":"3101","root":"runtime"}`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadRuntime(file, true)
	if err != nil || cfg.Port != "3101" || cfg.Root != filepath.Join(dir, "runtime") {
		t.Fatalf("file-relative root: %+v, %v", cfg, err)
	}
	t.Setenv("PORT", "3102")
	t.Setenv("MYSELF_ROOT", dir)
	cfg, err = loadRuntime(file, true)
	if err != nil || cfg.Port != "3102" || cfg.Root != dir {
		t.Fatalf("environment precedence: %+v, %v", cfg, err)
	}
}

func TestRuntimeRejectsInvalidConfig(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("MYSELF_ROOT", "")
	for _, content := range []string{`{`, `{ "prot": "3100" }`, `{"port":"0"}`, `{"port":"70000"}`, `{"port":"abc"}`, `{} {}`} {
		t.Run(content, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(file, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadRuntime(file, true); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
}
