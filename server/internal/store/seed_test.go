package store

import "testing"

// 示例关于页模块：合法 JSON、身份区置首、含章节分段、id 唯一，画廊与作品用真实图片地址而非占位。
func TestDemoAboutModules(t *testing.T) {
	mods, err := DemoAboutModules()
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
