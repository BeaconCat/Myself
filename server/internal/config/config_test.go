package config

import (
	"strings"
	"testing"
)

// 出厂关于页模块：只有身份区、站点数字、收尾格言（示例内容由初始化时的 Demo 数据写入），id 唯一。
func TestDefaultAboutModules(t *testing.T) {
	about, ok := Default()["about"].(Map)
	if !ok {
		t.Fatal("about missing")
	}
	mods, ok := about["modules"].([]any)
	if !ok || len(mods) == 0 {
		t.Fatal("about.modules missing")
	}
	var types []string
	seen := map[string]bool{}
	for _, raw := range mods {
		m, _ := raw.(Map)
		id, _ := m["id"].(string)
		if id == "" || seen[id] {
			t.Fatalf("empty or duplicate module id %q", id)
		}
		seen[id] = true
		types = append(types, m["type"].(string))
	}
	if strings.Join(types, ",") != "profile,stats,motto" {
		t.Fatalf("factory modules = %v, want profile,stats,motto", types)
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

// 界面风格默认值：clean（简洁），且允许访客切换卡片。
func TestDefaultThemeStyle(t *testing.T) {
	theme := Sub(Default(), "theme")
	if got := Str(theme, "defaultStyle"); got != "clean" {
		t.Fatalf("theme.defaultStyle = %q, want clean", got)
	}
	if !Bool(theme, "allowUserStyle") {
		t.Fatal("theme.allowUserStyle should default to true")
	}
}

func TestDropPlaceholders(t *testing.T) {
	cfg := Map{
		"github": Map{"username": "your-github"},
		"about": Map{"links": []any{
			Map{"name": "GitHub", "url": "https://github.com/your-github", "primary": true},
			Map{"name": "邮件", "url": "mailto:hi@example.com"},
			Map{"name": "RSS", "url": "/feed"},
		}},
	}
	dropPlaceholders(cfg)
	if u := cfg["github"].(Map)["username"]; u != "" {
		t.Fatalf("username = %v", u)
	}
	links := cfg["about"].(Map)["links"].([]any)
	if len(links) != 1 || links[0].(Map)["url"] != "/feed" || links[0].(Map)["primary"] != true {
		t.Fatalf("links = %v", links)
	}
}
