package updater

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func waitFailed(t *testing.T, m *Manager, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s := m.Status()
		if s.Phase == "failed" {
			if s.Error != want {
				t.Fatalf("error = %s, want %s", s.Error, want)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("installation did not finish")
}

func TestDevelopmentManualInstallationKeepsVerification(t *testing.T) {
	m := testManager(t)
	m.opts.Build.Version = "dev"
	server := releaseServer(t, strings.Repeat("0", 64), Protocol, []byte("bad"))
	defer server.Close()
	m.apiBase = server.URL
	s, err := m.Check(context.Background())
	if err != nil || !s.CanApply || s.Reason != "" || s.AutomaticReason != "development_build" {
		t.Fatalf("%+v %v", s, err)
	}
	if err := m.Start("v1.1.0"); err == nil || err.Error() != "development_build" {
		t.Fatalf("automatic install: %v", err)
	}
	p := m.Preferences()
	p.AutoUpdate = true
	if err := m.Configure(p, "2026-10-02"); err == nil || err.Error() != "development_build" {
		t.Fatalf("automatic preferences: %v", err)
	}
	if err := m.StartSelected(context.Background(), s.Repository, 42); err != nil {
		t.Fatal(err)
	}
	waitFailed(t, m, "checksum_mismatch")
	data, _ := os.ReadFile(m.executable)
	if string(data) != "existing program" {
		t.Fatal("invalid binary changed executable")
	}
}

func TestManualInstallationDeploymentGates(t *testing.T) {
	m := testManager(t)
	m.opts.Disabled = true
	if err := m.StartSelected(context.Background(), m.opts.Repository, 42); err == nil || err.Error() != "disabled_by_operator" {
		t.Fatal(err)
	}
	m.opts.Disabled = false
	m.executable = filepath.Join(t.TempDir(), "go-build12345", "b001", "exe", "myself")
	if m.Status().Reason != "temporary_executable" {
		t.Fatal(m.Status())
	}
	if err := m.StartHistory(strings.Repeat("a", 32)); err == nil || err.Error() != "temporary_executable" {
		t.Fatal(err)
	}
}

func makeHistory(t *testing.T, m *Manager, n int) HistoryEntry {
	t.Helper()
	p := Plan{Root: m.opts.Root, Nonce: fmt.Sprintf("%032x", n), Previous: filepath.Join(t.TempDir(), "previous"), Current: m.opts.Build}
	p.Current.UpdateProtocol = Protocol
	if err := os.WriteFile(p.Previous, []byte("existing program"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := archivePrevious(p); err != nil {
		t.Fatal(err)
	}
	entry, err := readHistory(m.stateDir, p.Nonce)
	if err != nil {
		t.Fatal(err)
	}
	if err := archivePrevious(p); err != nil {
		t.Fatal("archive is not idempotent", err)
	}
	entry.CreatedAt = time.Date(2026, 10, 2, 0, 0, n, 0, time.UTC).Format(time.RFC3339Nano)
	if err := writeJSONFile(filepath.Join(m.stateDir, "history", entry.ID, "entry.json"), entry); err != nil {
		t.Fatal(err)
	}
	return entry
}

func TestHistoryRetentionPaginationAndDefaults(t *testing.T) {
	m := testManager(t)
	if m.Preferences().HistoryLimit != 3 {
		t.Fatal(m.Preferences())
	}
	for i := 1; i <= 7; i++ {
		makeHistory(t, m, i)
	}
	page, err := m.History(2, 3)
	if err != nil || page.Total != 7 || !page.HasNext || len(page.Items) != 3 || page.Items[0].ID != fmt.Sprintf("%032x", 4) {
		t.Fatalf("%+v %v", page, err)
	}
	page, err = m.History(999999, 3)
	if err != nil || len(page.Items) != 0 || page.HasNext {
		t.Fatalf("%+v %v", page, err)
	}
	backupDir := filepath.Join(m.opts.Root, "backups")
	os.MkdirAll(backupDir, 0700)
	os.WriteFile(filepath.Join(backupDir, "keep.zip"), []byte("site data"), 0600)
	p := m.Preferences()
	p.HistoryLimit = 3
	if err := m.Configure(p, "2026-10-02"); err != nil {
		t.Fatal(err)
	}
	page, _ = m.History(1, 10)
	if page.Total != 3 || page.Items[2].ID != fmt.Sprintf("%032x", 5) {
		t.Fatal(page)
	}
	p.HistoryLimit = 1
	if err := m.Configure(p, "2026-10-02"); err != nil {
		t.Fatal(err)
	}
	page, _ = m.History(1, 10)
	if page.Total != 1 {
		t.Fatal(page)
	}
	if _, err := os.Stat(filepath.Join(backupDir, "keep.zip")); err != nil {
		t.Fatal("site backup pruned", err)
	}
	p.HistoryLimit = 0 // An older client must preserve the saved choice.
	if err := m.Configure(p, "2026-10-02"); err != nil || m.Preferences().HistoryLimit != 1 {
		t.Fatal(err, m.Preferences())
	}
	p.HistoryLimit = 21
	if err := m.Configure(p, ""); err == nil {
		t.Fatal("accepted invalid retention")
	}
}

func TestHistoryRestoreRejectsTamperingAndWrongPlatform(t *testing.T) {
	m := testManager(t)
	entry := makeHistory(t, m, 1)
	for _, id := range []string{"../escape", "", strings.Repeat("a", 31)} {
		if err := m.StartHistory(id); err == nil {
			t.Fatal("accepted invalid ID", id)
		}
	}
	entry.Build.Arch = "different"
	filename := filepath.Join(m.stateDir, "history", entry.ID, "entry.json")
	writeJSONFile(filename, entry)
	if err := m.StartHistory(entry.ID); err == nil || err.Error() != "platform_not_available" {
		t.Fatal(err)
	}
	entry.Build.Arch = m.opts.Build.Arch
	writeJSONFile(filename, entry)
	// Same size mutation must fail SHA-256 before Prepare is called.
	os.WriteFile(filepath.Join(m.stateDir, "history", entry.ID, "program.bak"), []byte("tampered program"), 0600)
	p := m.Preferences()
	p.AutoUpdate = true
	if err := m.Configure(p, ""); err != nil {
		t.Fatal(err)
	}
	if err := m.StartHistory(entry.ID); err != nil {
		t.Fatal(err)
	}
	waitFailed(t, m, "checksum_mismatch")
	if m.Preferences().AutoUpdate {
		t.Fatal("rollback did not pause automation")
	}
}

func TestHistorySymlinkCannotEscapeOrBePruned(t *testing.T) {
	m := testManager(t)
	outside := t.TempDir()
	os.WriteFile(filepath.Join(outside, "keep"), []byte("keep"), 0600)
	os.MkdirAll(filepath.Join(m.stateDir, "history"), 0700)
	link := filepath.Join(m.stateDir, "history", strings.Repeat("a", 32))
	if err := os.Symlink(outside, link); err != nil {
		t.Skip("symlinks unavailable:", err)
	}
	if _, err := readHistory(m.stateDir, strings.Repeat("a", 32)); err == nil {
		t.Fatal("followed symlink")
	}
	if err := pruneHistory(m.stateDir, 1, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(outside, "keep")); err != nil {
		t.Fatal("outside file removed")
	}
}

func TestLegacyHelperHistoryMigrationPreservesRecoveryFileOnError(t *testing.T) {
	m := testManager(t)
	nonce := strings.Repeat("b", 32)
	p := Plan{Executable: m.executable, Staged: m.executable + ".next-" + nonce, Previous: m.executable + ".previous-" + nonce,
		Helper: m.executable + ".updater-" + nonce, Root: m.opts.Root, Backup: filepath.Join(m.opts.Root, "backups", "safety.zip"), CWD: m.opts.Root,
		Port: "3100", Nonce: nonce, SHA256: strings.Repeat("a", 64), Current: m.opts.Build}
	if err := writeJSONFile(filepath.Join(m.stateDir, "plan.json"), p); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.Previous, []byte("existing program"), 0700); err != nil {
		t.Fatal(err)
	}
	m.state.Phase = "installed"
	if s := m.Status(); s.HistoryError != "" {
		t.Fatal(s)
	}
	entry, err := readHistory(m.stateDir, nonce)
	if err != nil || entry.Build != m.opts.Build {
		t.Fatal(entry, err)
	}
	if _, err := os.Stat(p.Previous); !os.IsNotExist(err) {
		t.Fatal("legacy previous file not cleaned", err)
	}
	// A damaged existing archive must never cause the last recovery copy to vanish.
	os.WriteFile(p.Previous, []byte("existing program"), 0700)
	os.WriteFile(filepath.Join(m.stateDir, "history", nonce, "program.bak"), []byte("tampered program"), 0600)
	if s := m.Status(); s.HistoryError != "history_archive_failed" {
		t.Fatal(s)
	}
	if _, err := os.Stat(p.Previous); err != nil {
		t.Fatal("recovery file lost", err)
	}
}
