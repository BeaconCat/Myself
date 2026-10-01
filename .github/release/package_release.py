"""Build portable releases from an already-built frontend. Python stdlib only."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tarfile
import tempfile
import zipfile

REPO = Path(__file__).resolve().parents[2]
TARGETS = ("linux-amd64", "linux-386", "linux-arm", "linux-arm64", "linux-riscv64",
           "darwin-amd64", "darwin-arm64", "windows-amd64", "windows-386", "windows-arm64")


def run(*args, **kwargs):
    return subprocess.check_output(args, cwd=REPO, text=True, encoding="utf-8", **kwargs)


def license_texts(directory, fallback=None):
    files = sorted(p for p in Path(directory).iterdir() if p.is_file() and
                   re.match(r"^(licen[cs]e|copying|notice|ofl)(\.|$|-)", p.name, re.I))
    if not files:
        if fallback and fallback.is_file():
            files = [fallback]
        else:
            raise RuntimeError(f"No license text found: {directory}")
    return "\n\n".join(p.name + "\n" + p.read_text(encoding="utf-8", errors="replace") for p in files)


def third_party_notices():
    run("go", "-C", "server", "mod", "download", "all")
    pnpm = shutil.which("pnpm.cmd" if os.name == "nt" else "pnpm")
    packages = json.loads(run(pnpm, "-C", "client", "licenses", "list", "--prod", "--json"))
    sections = ["Third-party license texts for the packaged application.\n"
                "Some dependency tools are included conservatively. Brand trademarks are not licensed by this file."]
    for group in packages.values():
        for pkg in group:
            for location in pkg["paths"]:
                metadata = json.loads((Path(location) / "package.json").read_text(encoding="utf-8"))
                fallback = REPO / "deploy/licenses" / (pkg["name"].replace("/", "_") + "-" + metadata["version"] + ".txt")
                sections.append(f"npm: {pkg['name']} {metadata['version']}\nPackage: https://www.npmjs.com/package/{pkg['name']}/v/{metadata['version']}\n" + license_texts(location, fallback))
    modules = run("go", "-C", "server", "list", "-m", "-json", "all")
    decoder = json.JSONDecoder()
    while modules.strip():
        module, end = decoder.raw_decode(modules.lstrip())
        modules = modules.lstrip()[end:]
        if not module.get("Main"):
            escaped = re.sub(r"[A-Z]", lambda m: "!" + m[0].lower(), module["Path"])
            source = f"https://proxy.golang.org/{escaped}/@v/{module['Version']}.zip"
            sections.append(f"Go: {module['Path']} {module['Version']}\nSource: {source}\n" + license_texts(module["Dir"]))
    return "\n\n" + ("\n\n" + "=" * 72 + "\n\n").join(sections)


def package(version, targets, output):
    if not re.fullmatch(r"v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z][0-9A-Za-z.-]*)?", version):
        raise ValueError("version must be vMAJOR.MINOR.PATCH or a prerelease such as v1.0.0-rc.1")
    if not (REPO / "server/web/dist/index.html").is_file():
        raise RuntimeError("Build frontend first: pnpm -C client build")
    output.mkdir(parents=True, exist_ok=True)
    if any(output.iterdir()):
        raise RuntimeError("Output directory must be empty; use a new directory for each build")
    revision = run("git", "rev-parse", "HEAD").strip()
    built = datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    notices = third_party_notices()
    artifacts = []
    for target in targets:
        goos, goarch = target.split("-")
        suffix = ".exe" if goos == "windows" else ""
        name = f"myself-{version}-{target}"
        binary = output / (name + suffix)
        env = dict(os.environ, GOOS=goos, GOARCH=goarch, GOARM="6", GOAMD64="v1", GO386="sse2", GORISCV64="rva20u64", CGO_ENABLED="0")
        run("go", "-C", "server", "build", "-tags", "nodynamic", "-trimpath", "-ldflags",
            f"-s -w -X main.version={version} -X main.commit={revision} -X main.buildDate={built}",
            "-o", str(binary), "./cmd/myself-server", env=env)
        binary.chmod(0o755)
        with tempfile.TemporaryDirectory(prefix="myself-release-") as temp:
            root = Path(temp) / name
            root.mkdir()
            shutil.copy2(binary, root / ("myself" + suffix))
            shutil.copy2(REPO / "deploy/config.example.json", root)
            shutil.copy2(REPO / "deploy/config.mysql.example.json", root)
            shutil.copy2(REPO / "deploy/README.md", root / "README.md")
            for filename in ("LICENSE", "NOTICE"):
                shutil.copy2(REPO / filename, root)
            (root / "THIRD_PARTY_LICENSES.txt").write_text(notices, encoding="utf-8")
            for folder in ("data", "uploads", "backups"):
                (root / "runtime" / folder).mkdir(parents=True)
            if goos == "windows":
                archive = output / (name + ".zip")
                with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as z:
                    for path in sorted(root.rglob("*")):
                        z.write(path, path.relative_to(root.parent).as_posix())
            else:
                archive = output / (name + ".tar.gz")
                with tarfile.open(archive, "w:gz") as tar:
                    def portable_modes(info):
                        info.uid = info.gid = 0
                        info.uname = info.gname = ""
                        info.mode = 0o755 if info.isdir() or info.name == name + "/myself" else 0o644
                        return info
                    tar.add(root, arcname=name, filter=portable_modes)
        for path, kind in ((binary, "binary"), (archive, "package")):
            with path.open("rb") as file:
                digest = hashlib.file_digest(file, "sha256").hexdigest()
            artifacts.append(dict(name=path.name, os=goos, arch=goarch, kind=kind,
                                  size=path.stat().st_size, sha256=digest))
        print(f"Built {target}: {binary.stat().st_size} bytes", flush=True)
    manifest = dict(schemaVersion=1, updateProtocol=1, version=version, commit=revision, builtAt=built, artifacts=artifacts)
    manifest_path = output / "release-manifest.json"
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    sums = [f"{item['sha256']}  {item['name']}" for item in artifacts]
    with manifest_path.open("rb") as file:
        sums.append(f"{hashlib.file_digest(file, 'sha256').hexdigest()}  {manifest_path.name}")
    (output / "SHA256SUMS").write_text("\n".join(sums) + "\n", encoding="utf-8")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", required=True)
    parser.add_argument("--target", choices=TARGETS, action="append")
    parser.add_argument("--output", default="dist-release")
    args = parser.parse_args()
    package(args.version, args.target or TARGETS, Path(args.output).resolve())
