package config

import "testing"

// 默认关于页模块：合法 JSON、profile 置首、含 chapter 分段、id 唯一。
func TestDefaultAboutModules(t *testing.T) {
	about, ok := Default()["about"].(Map)
	if !ok {
		t.Fatal("about missing")
	}
	mods, ok := about["modules"].([]any)
	if !ok || len(mods) == 0 {
		t.Fatal("about.modules missing")
	}
	first, _ := mods[0].(Map)
	if first["type"] != "profile" {
		t.Fatalf("first module = %v, want profile", first["type"])
	}
	seen := map[string]bool{}
	chapters := 0
	for _, raw := range mods {
		m, _ := raw.(Map)
		id, _ := m["id"].(string)
		if id == "" || seen[id] {
			t.Fatalf("empty or duplicate module id %q", id)
		}
		seen[id] = true
		if m["type"] == "chapter" {
			chapters++
		}
	}
	if chapters < 2 {
		t.Fatalf("chapters = %d, want >= 2", chapters)
	}
}

// 已有模块的配置不被默认值覆盖；空模块列表回落默认。
func TestMigrateAboutModules(t *testing.T) {
	cfg := Map{"about": Map{"modules": []any{Map{"id": "x", "type": "motto", "data": Map{"text": "hi"}}}}}
	out := migrateAboutModules(cfg)
	if n := len(out["about"].(Map)["modules"].([]any)); n != 1 {
		t.Fatalf("modules len = %d, want 1", n)
	}
	empty := migrateAboutModules(Map{"about": Map{}})
	if n := len(empty["about"].(Map)["modules"].([]any)); n == 0 {
		t.Fatal("empty modules should fall back to defaults")
	}
}

// 界面风格默认值：cards，且允许访客切换。
func TestDefaultThemeStyle(t *testing.T) {
	theme := Sub(Default(), "theme")
	if got := Str(theme, "defaultStyle"); got != "cards" {
		t.Fatalf("theme.defaultStyle = %q, want cards", got)
	}
	if !Bool(theme, "allowUserStyle") {
		t.Fatal("theme.allowUserStyle should default to true")
	}
}
