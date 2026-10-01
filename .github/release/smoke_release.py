"""Verify checksums and boot an extracted native package in an isolated directory."""
import hashlib
import argparse
import http.cookiejar
import json
from pathlib import Path
import re
import socket
import struct
import subprocess
import sys
import tarfile
import tempfile
import time
import urllib.request
import zipfile
import zlib


def smoke(directory, target, runner=None):
    manifest = json.loads((directory / "release-manifest.json").read_text(encoding="utf-8"))
    for line in (directory / "SHA256SUMS").read_text(encoding="utf-8").splitlines():
        digest, name = line.split("  ", 1)
        assert Path(name).name == name
        with (directory / name).open("rb") as file:
            assert hashlib.file_digest(file, "sha256").hexdigest() == digest, name
    goos, goarch = target.split("-")
    artifact = next(a for a in manifest["artifacts"] if a["os"] == goos and a["arch"] == goarch and a["kind"] == "package")
    with tempfile.TemporaryDirectory(prefix="myself-smoke-") as temp:
        temp = Path(temp)
        archive = directory / artifact["name"]
        if goos == "windows":
            with zipfile.ZipFile(archive) as z:
                z.extractall(temp)
        else:
            with tarfile.open(archive) as tar:
                tar.extractall(temp, filter="data")
        root = next(temp.iterdir())
        exe = root / ("myself.exe" if goos == "windows" else "myself")
        command = ([runner] if runner else []) + [str(exe)]
        result = subprocess.check_output(command + ["--version"], text=True)
        assert manifest["version"] in result and manifest["commit"] in result
        for folder in ("data", "uploads", "backups"):
            assert not any((root / "runtime" / folder).iterdir()), "Package contains site data"
        assert (root / "THIRD_PARTY_LICENSES.txt").stat().st_size > 1000
        with socket.socket() as sock:
            sock.bind(("127.0.0.1", 0))
            port = sock.getsockname()[1]
        cfg = json.loads((root / "config.example.json").read_text(encoding="utf-8"))
        cfg["port"] = str(port)
        config = root / "config.json"
        config.write_text(json.dumps(cfg), encoding="utf-8")
        import os
        env = {k: v for k, v in os.environ.items() if k not in ("PORT", "MYSELF_ROOT", "MYSELF_CONFIG")}
        with (temp / "server.log").open("w", encoding="utf-8") as log:
            proc = subprocess.Popen(command + ["--config", str(config)], cwd=temp, env=env, stdout=log, stderr=log)
            try:
                base = f"http://127.0.0.1:{port}"
                for _ in range(300):
                    if proc.poll() is not None:
                        raise RuntimeError("Packaged server exited early: " + (temp / "server.log").read_text(encoding="utf-8"))
                    try:
                        with urllib.request.urlopen(base + "/api/v1/setup", timeout=1) as response:
                            setup = json.load(response)
                        break
                    except OSError:
                        time.sleep(0.1)
                else:
                    raise RuntimeError("Packaged server did not start")
                assert setup["needsSetup"] is True, setup
                with urllib.request.urlopen(base + "/setup") as response:
                    html = response.read().decode()
                asset = re.search(r'src="(/assets/[^\"]+\.js)"', html)
                assert asset, "Embedded frontend missing"
                with urllib.request.urlopen(base + asset[1]) as response:
                    assert response.status == 200 and len(response.read()) > 0
                assert (root / "runtime/data/myself.db").is_file(), "Config-relative data root not used"
                assert not (temp / "data").exists(), "Data leaked into cwd"
                # Exercise the embedded SQLite engine and pure-Go image encoder.
                opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
                code = re.search(r"初始化码 ([A-F0-9-]+)", (temp / "server.log").read_text(encoding="utf-8"))[1]
                headers = {"Origin": base, "X-Requested-With": "myself", "Content-Type": "application/json"}
                body = json.dumps({"code": code, "username": "smoke", "password": "release-smoke-test-only", "demo": False}).encode()
                with opener.open(urllib.request.Request(base + "/api/v1/setup", data=body, headers=headers), timeout=30) as response:
                    assert json.load(response)["ok"]
                def chunk(kind, data):
                    return struct.pack(">I", len(data)) + kind + data + struct.pack(">I", zlib.crc32(kind + data))
                png = b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", 2, 2, 8, 2, 0, 0, 0))
                png += chunk(b"IDAT", zlib.compress(b"\x00\xff\x00\x00\x00\xff\x00" * 2)) + chunk(b"IEND", b"")
                multipart = b'--myself-smoke\r\nContent-Disposition: form-data; name="files"; filename="smoke.png"\r\nContent-Type: image/png\r\n\r\n' + png + b"\r\n--myself-smoke--\r\n"
                headers["Content-Type"] = "multipart/form-data; boundary=myself-smoke"
                with opener.open(urllib.request.Request(base + "/api/v1/admin/media", data=multipart, headers=headers), timeout=30) as response:
                    media = json.load(response)[0]
                with opener.open(base + media["thumb"], timeout=30) as response:
                    webp = response.read()
                    assert webp[:4] == b"RIFF" and webp[8:12] == b"WEBP", "WebP encoder failed"
            finally:
                proc.terminate()
                proc.wait(timeout=15)
    print(f"PASS: {target} checksums, clean archive, embedded frontend/SQLite, initialization, upload, WebP thumbnail")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directory", type=Path)
    parser.add_argument("target")
    parser.add_argument("--runner")
    args = parser.parse_args()
    smoke(args.directory.resolve(), args.target, args.runner)
