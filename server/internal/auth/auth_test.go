package auth

import (
	"testing"

	"myself/server/internal/store"
)

// Node 端 crypto.scryptSync('p@ss-word', salt, 64) 生成的哈希，验证跨实现兼容。
const nodeHash = "2eba0822da8f9275c35e581d01756cf5:11f736d0a1d6d618bcf066ede3aa2eabcc2ae775ea922ac8e8a3a141b10309b60f321a38be323f29fafa08fcb9f80d8594259485237793e444bba131658097a8"

func TestVerifyNodeHash(t *testing.T) {
	if !VerifyPassword("p@ss-word", nodeHash) {
		t.Fatal("node-generated scrypt hash must verify")
	}
	if VerifyPassword("wrong", nodeHash) {
		t.Fatal("wrong password must not verify")
	}
}

func TestHashRoundTrip(t *testing.T) {
	h, err := HashPassword("hello-world-1")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("hello-world-1", h) {
		t.Fatal("round trip failed")
	}
	if VerifyPassword("", "garbage") {
		t.Fatal("malformed hash must fail")
	}
}

func TestLegacyAdminMigratesAndMustChange(t *testing.T) {
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h, _ := HashPassword(legacyDefaultPassword)
	db.SetSetting("admin_username", "admin")
	db.SetSetting("admin_password", h)
	a := New(db)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if a.SetupCode() != "" {
		t.Fatal("migrated site must not need setup")
	}
	u, err := db.UserByLogin("admin")
	if err != nil || u == nil || u.Role != store.RoleAdmin || !u.MustChange {
		t.Fatalf("legacy admin not migrated with must-change: %+v %v", u, err)
	}
	if v, _ := db.GetSetting("admin_password"); v != "" {
		t.Fatal("legacy settings should be removed after migration")
	}
	if _, err := a.ChangePassword(u.ID, legacyDefaultPassword, "brand-new-pass"); err != nil {
		t.Fatal(err)
	}
	if u, _ = db.UserByID(u.ID); u.MustChange {
		t.Fatal("must-change flag should clear after change")
	}
	if got, _ := a.Authenticate("ADMIN", "brand-new-pass"); got == nil {
		t.Fatal("login should be case-insensitive and accept the new password")
	}
}
