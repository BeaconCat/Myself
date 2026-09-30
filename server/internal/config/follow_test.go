package config

import "testing"

func TestFollowSiteTitle(t *testing.T) {
	cur := Map{
		"site":    Map{"title": "Old Lab"},
		"loading": Map{"bootText": "Old Lab"},
		"mail":    Map{"from": "Old Lab <bot@example.com>"},
	}
	patch := Map{"site": Map{"title": "New Lab"}}
	FollowSiteTitle(cur, patch)
	if got := Str(Sub(patch, "loading"), "bootText"); got != "New Lab" {
		t.Fatalf("bootText = %q", got)
	}
	if got := Str(Sub(patch, "mail"), "from"); got != "New Lab <bot@example.com>" {
		t.Fatalf("from = %q", got)
	}

	// 已被人为改过的不覆盖；patch 显式给的值优先
	cur = Map{
		"site":    Map{"title": "Old Lab"},
		"loading": Map{"bootText": "欢迎光临"},
		"mail":    Map{"from": "站长 <bot@example.com>"},
	}
	patch = Map{"site": Map{"title": "New Lab"}}
	FollowSiteTitle(cur, patch)
	if _, ok := patch["loading"]; ok {
		t.Fatalf("custom bootText overwritten: %v", patch)
	}
	if _, ok := patch["mail"]; ok {
		t.Fatalf("custom sender overwritten: %v", patch)
	}

	// 出厂默认名、无显示名的发件地址也跟随；名称含特殊字符加引号
	cur = Map{"site": Map{"title": "Myself"}, "loading": Map{"bootText": "Myself"}, "mail": Map{"from": "bot@example.com"}}
	patch = Map{"site": Map{"title": "A, B"}, "loading": Map{"bootText": "手动"}}
	FollowSiteTitle(cur, patch)
	if got := Str(Sub(patch, "loading"), "bootText"); got != "手动" {
		t.Fatalf("explicit bootText replaced: %q", got)
	}
	if got := Str(Sub(patch, "mail"), "from"); got != `"A, B" <bot@example.com>` {
		t.Fatalf("quoted from = %q", got)
	}
}
