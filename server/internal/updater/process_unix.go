//go:build !windows

package updater

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// exec preserves the service manager's main PID through helper and new server.
func launchInPlace(p Plan, filename string) (bool, error) {
	return true, syscall.Exec(p.Helper, []string{p.Helper, "--apply-update", filename}, append(cleanUpdateEnv(), "MYSELF_UPDATE_EXEC=1"))
}

func finishServer(p Plan, cmd *exec.Cmd, phase, message string) error {
	if os.Getenv("MYSELF_UPDATE_EXEC") != "1" {
		return savePhase(p, phase, message)
	}
	// The candidate stays in maintenance while its startup is probed. Close it
	// gracefully, then replace this helper with the verified program at the same PID.
	_ = cmd.Process.Signal(syscall.SIGTERM)
	deadline := time.Now().Add(15 * time.Second)
	for processRunning(cmd.Process.Pid) {
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			return errors.New("candidate_shutdown_failed")
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err := savePhase(p, phase, message); err != nil {
		return err
	}
	_ = os.Remove(p.Executable + ".update-lock") // exec does not run deferred cleanup.
	return syscall.Exec(p.Executable, append([]string{p.Executable}, p.Args...), cleanUpdateEnv())
}

func hideWindow(cmd *exec.Cmd)    { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
func processRunning(pid int) bool { return pid > 0 && syscall.Kill(pid, 0) == nil }
