package store

import (
	"encoding/json"
	"strings"
	"testing"
)

// 示例关于页模块：合法 JSON、身份区置首、含章节分段、id 唯一，画廊与作品用真实图片地址而非占位。
func TestDemoAboutModules(t *testing.T) {
	mods, err := DemoAboutModules(nil)
	if err != nil || len(mods) == 0 {
		t.Fatalf("demo modules: %v", err)
	}
	if first, _ := mods[0].(map[string]any); first["type"] != "profile" {
		t.Fatalf("first demo module = %v", first["type"])
	}
	seen, chapters := map[string]bool{}, 0
	for _, raw := range mods {
		m, _ := raw.(map[string]any)
		id, _ := m["id"].(string)
		if id == "" || seen[id] {
			t.Fatalf("empty or duplicate id %q", id)
		}
		seen[id] = true
		switch m["type"] {
		case "chapter":
			chapters++
		case "gallery":
			for _, im := range m["data"].(map[string]any)["images"].([]any) {
				if src, _ := im.(map[string]any)["src"].(string); src == "" {
					t.Fatal("gallery demo image without src")
				}
			}
		}
	}
	if chapters < 2 {
		t.Fatalf("chapters = %d", chapters)
	}
}

// 地址改写：关于页示例模块里的默认封面全部经映射改写，其余内容不变。
func TestDemoAboutModulesMapsCovers(t *testing.T) {
	mods, err := DemoAboutModules(func(u string) string { return "/uploads/x" + u[len("/covers/"):] })
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(mods)
	if strings.Contains(string(raw), "/covers/") || !strings.Contains(string(raw), "/uploads/x05.webp") {
		t.Fatalf("covers not mapped: %.200s", raw)
	}
}
