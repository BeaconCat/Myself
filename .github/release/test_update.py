"""Exercise real binary replacement and rollback, including database/file recovery."""
import hashlib
import http.cookiejar
import json
import os
from pathlib import Path
import re
import secrets
import shutil
import signal
import socket
import subprocess
import sys
import tempfile
import time
import urllib.request

REPO = Path(__file__).resolve().parents[2]
FLAGS = {"creationflags": subprocess.CREATE_NO_WINDOW} if os.name == "nt" else {}


def metadata(exe):
    return json.loads(subprocess.check_output([str(exe), "--version-json"], text=True, **FLAGS))


def exercise(old_binary, new_binary, should_rollback):
    old_meta, new_meta = metadata(old_binary), metadata(new_binary)
    with tempfile.TemporaryDirectory(prefix="myself-update-test-") as temp:
        root = Path(temp).resolve()
        exe = root / ("myself.exe" if os.name == "nt" else "myself")
        shutil.copy2(old_binary, exe)
        (root / "config.json").write_text(json.dumps({"root": str(root), "port": "3100"}), encoding="utf-8")
        with socket.socket() as sock:
            sock.bind(("127.0.0.1", 0))
            port = sock.getsockname()[1]
        config = root / "config.json"
        config.write_text(json.dumps({"root": str(root), "port": str(port)}), encoding="utf-8")
        base = f"http://127.0.0.1:{port}"
        opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
        env = {k: v for k, v in os.environ.items() if k not in ("PORT", "MYSELF_ROOT", "MYSELF_CONFIG", "MYSELF_UPDATE_NONCE", "MYSELF_UPDATE_EXEC")}
        password = secrets.token_urlsafe(24)
        process = helper_process = None
        installed_pid = None

        def call(method, endpoint, data=None, headers=None):
            body = json.dumps(data).encode() if data is not None else None
            request = urllib.request.Request(base + endpoint, data=body, method=method, headers={"Origin": base, "X-Requested-With": "myself", "Content-Type": "application/json", **(headers or {})})
            with opener.open(request, timeout=3) as response:
                return json.load(response)

        log_path = root / "test-server.log"
        with log_path.open("w", encoding="utf-8") as log:
            try:
                process = subprocess.Popen([exe, "--config", config], cwd=root, env=env, stdout=log, stderr=log, **FLAGS)
                for _ in range(150):
                    if process.poll() is not None:
                        raise RuntimeError(log_path.read_text(encoding="utf-8"))
                    try:
                        call("GET", "/api/v1/system/health")
                        break
                    except OSError:
                        time.sleep(.1)
                code = re.search(r"初始化码 ([A-F0-9-]+)", log_path.read_text(encoding="utf-8"))[1]
                call("POST", "/api/v1/setup", {"code": code, "username": "owner", "password": password, "demo": False})
                call("POST", "/api/v1/admin/posts", {"slug": "before-update", "title": "Before", "contentMd": "Before", "status": "published"})
                proof = root / "uploads/recovery-proof.txt"
                proof.write_text("before", encoding="utf-8")
                backup = call("POST", "/api/v1/admin/backups")["name"]
                process.terminate()
                process.wait(timeout=15)
                nonce = secrets.token_hex(16)
                extension = ".exe" if os.name == "nt" else ""
                staged = Path(str(exe) + ".next-" + nonce + extension)
                helper = Path(str(exe) + ".updater-" + nonce + extension)
                shutil.copy2(new_binary, staged)
                shutil.copy2(old_binary, helper)
                with staged.open("rb") as file:
                    digest = hashlib.file_digest(file, "sha256").hexdigest()
                plan = {"executable": str(exe), "staged": str(staged), "previous": str(exe) + ".previous-" + nonce,
                        "helper": str(helper), "root": str(root), "backup": str(root / "backups" / backup),
                        "args": ["--config", str(config)], "cwd": str(root), "parentPid": process.pid,
                        "port": str(port), "version": new_meta["version"], "commit": new_meta["commit"],
                        "nonce": nonce, "sha256": digest, "current": old_meta}
                updates = root / ".updates"
                updates.mkdir(exist_ok=True)
                plan_file = updates / "plan.json"
                plan_file.write_text(json.dumps(plan), encoding="utf-8")
                (updates / "status.json").write_text(json.dumps({"phase": "restarting", "current": old_meta}), encoding="utf-8")
                helper_env = dict(env)
                if os.name != "nt":
                    helper_env["MYSELF_UPDATE_EXEC"] = "1"
                helper_process = subprocess.Popen([helper, "--apply-update", plan_file], cwd=root, env=helper_env, stdout=log, stderr=log, **FLAGS)
                expected_phase = "rolled_back" if should_rollback else "installed"
                expected_meta = old_meta if should_rollback else new_meta
                for _ in range(1200):
                    try:
                        state = json.loads((updates / "status.json").read_text(encoding="utf-8"))
                        if state.get("phase") == "failed":
                            raise RuntimeError(str(state) + "\n" + log_path.read_text(encoding="utf-8"))
                        if state.get("phase") == expected_phase:
                            health = call("GET", "/api/v1/system/health", headers={"X-Myself-Update-Probe": nonce})
                            installed_pid = health.get("pid", helper_process.pid)
                            status = call("GET", "/api/v1/admin/system")
                            if status["current"]["version"] == expected_meta["version"]:
                                break
                    except (OSError, json.JSONDecodeError):
                        pass
                    time.sleep(.1)
                else:
                    raise RuntimeError("update timed out\n" + log_path.read_text(encoding="utf-8"))
                assert call("GET", "/api/v1/posts/before-update")["contentMd"] == "Before"
                assert proof.read_text(encoding="utf-8") == "before"
                assert metadata(exe)["version"] == expected_meta["version"]
                if os.name != "nt":
                    assert installed_pid == helper_process.pid, "Unix update lost the service PID"
                print(json.dumps({"case": expected_phase, "version": expected_meta["version"], "database": "preserved", "uploads": "preserved"}), flush=True)
            finally:
                if installed_pid:
                    try:
                        os.kill(installed_pid, signal.SIGTERM)
                    except OSError:
                        pass
                for child in (process, helper_process):
                    if child and child.poll() is None:
                        child.terminate()
                        child.wait(timeout=15)
                time.sleep(.3)


def crash_fixture(old):
    # This candidate passes version validation, mutates data like a bad migration,
    # and exits. Recovery must restore both the database and uploaded files.
    meta = metadata(old)
    meta["version"] = "v99.0.0"
    source = r'''package main
import ("database/sql"; "encoding/json"; "flag"; "fmt"; "os"; "path/filepath"; _ "modernc.org/sqlite")
func main() {
 version := flag.Bool("version-json", false, "")
 config := flag.String("config", "", "")
 flag.Parse()
 if *version { fmt.Print(METADATA); return }
 var cfg struct { Root string `json:"root"` }
 b, _ := os.ReadFile(*config); _ = json.Unmarshal(b, &cfg)
 db, err := sql.Open("sqlite", filepath.Join(cfg.Root, "data", "myself.db")); if err != nil { panic(err) }
 if _, err = db.Exec("UPDATE posts SET content_md = 'bad migration' WHERE slug = 'before-update'"); err != nil { panic(err) }
 db.Close()
 os.WriteFile(filepath.Join(cfg.Root, "uploads", "recovery-proof.txt"), []byte("bad migration"), 0600)
 os.Exit(42)
}
'''.replace("METADATA", "`" + json.dumps(meta) + "`")
    with tempfile.TemporaryDirectory(prefix="myself-crash-fixture-") as temp:
        file = Path(temp) / "main.go"
        file.write_text(source, encoding="utf-8")
        binary = Path(temp) / ("crash.exe" if os.name == "nt" else "crash")
        subprocess.run(["go", "-C", str(REPO / "server"), "build", "-o", str(binary), str(file)], check=True)
        exercise(old, binary, True)


if __name__ == "__main__":
    old, new = Path(sys.argv[1]).resolve(), Path(sys.argv[2]).resolve()
    exercise(old, new, False)
    crash_fixture(old)
