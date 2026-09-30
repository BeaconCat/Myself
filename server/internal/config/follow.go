package config

import (
	"net/mail"
	"strings"
)

// defaultSiteTitle 出厂站点名；仍是它的派生字段视为「没被人为改过」
const defaultSiteTitle = "Myself"

// FollowSiteTitle 站点名称变更时，把仍在沿用旧名称的派生字段一并改成新名称（写进 patch）：
//   - 启动文案 loading.bootText
//   - 邮件发件人名称（mail.from 的显示名部分，地址不变）
//
// 「沿用旧名称」= 当前值为空、等于旧站名或仍是出厂默认名；已被人为改成别的内容的不覆盖。
// patch 里显式给了这些字段时以 patch 为准。
func FollowSiteTitle(cur, patch Map) {
	next := strings.TrimSpace(Str(Sub(patch, "site"), "title"))
	old := Str(Sub(cur, "site"), "title")
	if next == "" || next == old {
		return
	}
	follows := func(v string) bool {
		v = strings.TrimSpace(v)
		return v == "" || v == old || v == defaultSiteTitle
	}

	if lp := Sub(patch, "loading"); !has(lp, "bootText") && follows(Str(Sub(cur, "loading"), "bootText")) {
		lp["bootText"] = next
		patch["loading"] = lp
	}

	mp := Sub(patch, "mail")
	if from := strings.TrimSpace(Str(Sub(cur, "mail"), "from")); !has(mp, "from") && from != "" {
		if addr, err := mail.ParseAddress(from); err == nil && follows(addr.Name) {
			mp["from"] = formatFrom(next, addr.Address)
			patch["mail"] = mp
		}
	}
}

func has(m Map, key string) bool {
	_, ok := m[key]
	return ok
}

// formatFrom 「名称 <地址>」；名称含地址语法里的特殊字符时加引号（不做 RFC 2047 编码，保持可读，发信时再编码）。
func formatFrom(name, addr string) string {
	if strings.ContainsAny(name, `,;:<>"@()[]\`) {
		name = `"` + strings.NewReplacer(`\`, `\`, `"`, `\"`).Replace(name) + `"`
	}
	return name + " <" + addr + ">"
}
