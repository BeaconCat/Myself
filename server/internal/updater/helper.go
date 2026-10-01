package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Plan struct {
	Executable string   `json:"executable"`
	Staged     string   `json:"staged"`
	Previous   string   `json:"previous"`
	Helper     string   `json:"helper"`
	Root       string   `json:"root"`
	Backup     string   `json:"backup"`
	Args       []string `json:"args"`
	CWD        string   `json:"cwd"`
	ParentPID  int      `json:"parentPid"`
	Port       string   `json:"port"`
	Version    string   `json:"version"`
	Commit     string   `json:"commit"`
	Nonce      string   `json:"nonce"`
	SHA256     string   `json:"sha256"`
	Current    Build    `json:"current"`
}

func ReadPlan(filename string) (Plan, error) {
	var p Plan
	data, err := os.ReadFile(filename)
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(p.Nonce) || !filepath.IsAbs(p.Executable) || !filepath.IsAbs(p.Root) || !filepath.IsAbs(p.CWD) {
		return p, errors.New("invalid update plan paths")
	}
	ext := ""
	if filepath.Ext(p.Executable) == ".exe" {
		ext = ".exe"
	}
	if p.Staged != p.Executable+".next-"+p.Nonce+ext || p.Previous != p.Executable+".previous-"+p.Nonce || p.Helper != p.Executable+".updater-"+p.Nonce+ext {
		return p, errors.New("invalid update plan executable names")
	}
	if filepath.Clean(filename) != filepath.Join(p.Root, ".updates", "plan.json") {
		return p, errors.New("unexpected update plan location")
	}
	rel, err := filepath.Rel(filepath.Join(p.Root, "backups"), p.Backup)
	if err != nil || filepath.Base(rel) != rel || !strings.HasSuffix(rel, ".zip") {
		return p, errors.New("invalid safety backup path")
	}
	port, err := strconv.Atoi(p.Port)
	if err != nil || port < 1 || port > 65535 || !hashRE.MatchString(p.SHA256) {
		return p, errors.New("invalid update plan metadata")
	}
	return p, nil
}

func copyExecutable(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if err != nil {
		out.Close()
		return err
	}
	if err := out.Sync(); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func Launch(filename string) error {
	plan, err := ReadPlan(filename)
	if err != nil {
		return err
	}
	if handled, err := launchInPlace(plan, filename); handled {
		return err
	}
	log, err := os.OpenFile(filepath.Join(plan.Root, ".updates", "install.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer log.Close()
	cmd := exec.Command(plan.Helper, "--apply-update", filename)
	cmd.Dir, cmd.Stdout, cmd.Stderr = plan.CWD, log, log
	hideWindow(cmd)
	return cmd.Start()
}

func savePhase(p Plan, phase, message string) error {
	filename := filepath.Join(p.Root, ".updates", "status.json")
	var state Status
	if data, err := os.ReadFile(filename); err == nil {
		_ = json.Unmarshal(data, &state)
	}
	state.Phase, state.Error = phase, message
	return writeJSONFile(filename, state)
}

// Apply runs in a copy of the old executable, so Windows can release the server
// image before it is renamed. The helper retains the previous binary for recovery.
func Apply(filename string) error {
	p, err := ReadPlan(filename)
	if err != nil {
		return err
	}
	fail := func(err error) error { _ = savePhase(p, "failed", err.Error()); return err }
	deadline := time.Now().Add(30 * time.Second)
	for p.ParentPID != os.Getpid() && processRunning(p.ParentPID) {
		if time.Now().After(deadline) {
			return fail(errors.New("old_process_did_not_exit"))
		}
		time.Sleep(100 * time.Millisecond)
	}
	defer os.Remove(p.Executable + ".update-lock")
	failBeforeInstall := func(err error) error {
		_ = savePhase(p, "failed", err.Error())
		if _, startErr := startServer(p, false); startErr != nil {
			return fmt.Errorf("%v; restart old version: %w", err, startErr)
		}
		return err
	}
	in, err := os.Open(p.Staged)
	if err != nil {
		return failBeforeInstall(err)
	}
	hash := sha256.New()
	_, err = io.Copy(hash, in)
	in.Close()
	if err != nil || hex.EncodeToString(hash.Sum(nil)) != p.SHA256 {
		return failBeforeInstall(errors.New("staged_binary_changed"))
	}
	if err := renameRetry(p.Executable, p.Previous); err != nil {
		return failBeforeInstall(err)
	}
	if err := renameRetry(p.Staged, p.Executable); err != nil {
		_ = renameRetry(p.Previous, p.Executable)
		_, _ = startServer(p, false)
		return fail(err)
	}
	cmd, err := startServer(p, true)
	if err == nil {
		err = awaitHealthy(p, cmd, p.Version, p.Commit)
	}
	if err == nil {
		if err := finishServer(p, cmd, "installed", ""); err != nil {
			return rollback(filename, p, cmd, err)
		}
		fmt.Printf("Installed %s; safety backup %s\n", p.Version, filepath.Base(p.Backup))
		return nil
	}
	return rollback(filename, p, cmd, err)
}

func renameRetry(src, dst string) error {
	var err error
	for i := 0; i < 100; i++ {
		if err = os.Rename(src, dst); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}

func startServer(p Plan, pending bool) (*exec.Cmd, error) {
	log, err := os.OpenFile(filepath.Join(p.Root, ".updates", "server.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer log.Close()
	cmd := exec.Command(p.Executable, p.Args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = p.CWD, log, log
	cmd.Env = cleanUpdateEnv()
	if pending {
		cmd.Env = append(cmd.Env, "MYSELF_UPDATE_NONCE="+p.Nonce)
	}
	hideWindow(cmd)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func cleanUpdateEnv() []string {
	var env []string
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "MYSELF_UPDATE_NONCE=") && !strings.HasPrefix(value, "MYSELF_UPDATE_EXEC=") {
			env = append(env, value)
		}
	}
	return env
}

func awaitHealthy(p Plan, cmd *exec.Cmd, version, commit string) error {
	client := &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	finished := make(chan error, 1)
	go func() { finished <- cmd.Wait() }()
	deadline := time.NewTimer(90 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-finished:
			return fmt.Errorf("new_process_exited: %v", err)
		case <-deadline.C:
			_ = cmd.Process.Kill()
			<-finished
			return errors.New("new_version_health_timeout")
		case <-ticker.C:
			req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:"+p.Port+"/api/v1/system/health", nil)
			req.Header.Set("X-Myself-Update-Probe", p.Nonce)
			response, err := client.Do(req)
			if err != nil {
				continue
			}
			var health struct {
				Version string `json:"version"`
				Commit  string `json:"commit"`
				Nonce   string `json:"nonce"`
			}
			err = json.NewDecoder(io.LimitReader(response.Body, 4096)).Decode(&health)
			response.Body.Close()
			if err == nil && response.StatusCode == http.StatusOK && health.Version == version && health.Commit == commit && health.Nonce == p.Nonce {
				return nil
			}
		}
	}
}

func rollback(filename string, p Plan, cmd *exec.Cmd, cause error) error {
	_ = savePhase(p, "rolling_back", cause.Error())
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	failed := p.Executable + ".failed-" + p.Nonce
	if err := renameRetry(p.Executable, failed); err != nil {
		return fmt.Errorf("rollback rename: %w", err)
	}
	if err := renameRetry(p.Previous, p.Executable); err != nil {
		return fmt.Errorf("rollback restore binary: %w", err)
	}
	args := append(append([]string{}, p.Args...), "--recover-update", filename)
	recover := exec.Command(p.Executable, args...)
	recover.Dir, recover.Env, recover.Stdout, recover.Stderr = p.CWD, cleanUpdateEnv(), os.Stdout, os.Stderr
	hideWindow(recover)
	if err := recover.Run(); err != nil {
		_ = savePhase(p, "failed", "rollback_backup_failed: "+err.Error())
		return err
	}
	old, err := startServer(p, true)
	if err == nil {
		err = awaitHealthy(p, old, p.Current.Version, p.Current.Commit)
	}
	if err != nil {
		_ = savePhase(p, "failed", "rollback_start_failed: "+err.Error())
		return err
	}
	_ = os.Remove(failed)
	if err := finishServer(p, old, "rolled_back", cause.Error()); err != nil {
		return err
	}
	fmt.Printf("Rolled back to %s after: %v\n", p.Current.Version, cause)
	return nil
}

// Pending keeps writes closed until the helper has accepted the new process.
func (m *Manager) Pending() bool {
	m.mu.Lock()
	pending := m.pending
	m.mu.Unlock()
	if !pending {
		return false
	}
	s := m.Status()
	if s.Phase == "installed" || s.Phase == "rolled_back" {
		m.mu.Lock()
		m.pending = false
		m.mu.Unlock()
		return false
	}
	return true
}

// Abort is called when the parent could not launch the helper after shutdown.
func Abort(filename string, cause error) {
	if p, err := ReadPlan(filename); err == nil {
		_ = savePhase(p, "failed", cause.Error())
		_ = os.Remove(p.Staged)
		_ = os.Remove(p.Helper)
		_ = os.Remove(p.Executable + ".update-lock")
	}
}
