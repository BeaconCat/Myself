package httpapi

import (
	"net/http"
	"testing"
)

// 身份名字改动时，仍沿用旧名字的站长昵称跟着改；人为改过的昵称不被覆盖。
func TestIdentityNameFollowsAdminNickname(t *testing.T) {
	e := newEnv(t)
	adminName := func() string {
		var me struct {
			User struct {
				Name string `json:"name"`
			} `json:"user"`
		}
		e.call(http.MethodGet, "/api/v1/auth/session", nil, &me, http.StatusOK)
		return me.User.Name
	}
	setAbout := func(name string) {
		e.call(http.MethodPut, "/api/v1/admin/settings", map[string]any{"about": map[string]any{"name": name}}, nil, http.StatusOK)
	}

	setAbout("旧名字")
	e.call(http.MethodPut, "/api/v1/me", map[string]any{"name": "旧名字"}, nil, http.StatusOK)
	setAbout("新名字")
	if got := adminName(); got != "新名字" {
		t.Fatalf("nickname should follow identity: %q", got)
	}

	e.call(http.MethodPut, "/api/v1/me", map[string]any{"name": "自定义"}, nil, http.StatusOK)
	setAbout("再换一个")
	if got := adminName(); got != "自定义" {
		t.Fatalf("custom nickname overwritten: %q", got)
	}
}
