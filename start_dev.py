from __future__ import annotations
import atexit
import os
import shutil
import signal
import subprocess
import sys
import threading
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parent
IS_WIN = sys.platform.startswith("win")

SERVICES = [
    ("server",   ["go", "run", "."],      ROOT / "server" / "main"),
    ("tools",    ["go", "run", "."],      ROOT / "tools" / "main"),
    (
        "frontend",
        ["npm", "run", "watch", "--", "--base-href", "/design/"],
        ROOT / "tools" / "main" / "spa",
    ),
]

procs: list[tuple[str, subprocess.Popen]] = []


def resolve_cmd(cmd: list[str]) -> list[str]:
    """On Windows, resolve bare names like 'npm' to 'npm.cmd' so we can
    avoid shell=True (which causes 'Terminate batch job?' prompts)."""
    exe = shutil.which(cmd[0])
    if exe:
        return [exe, *cmd[1:]]
    return cmd


def kill_all():
    for name, p in procs:
        if p.poll() is None:
            if IS_WIN:
                subprocess.run(["taskkill", "/PID", str(p.pid), "/T", "/F"],
                               stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            else:
                try:
                    os.killpg(p.pid, signal.SIGTERM)
                except ProcessLookupError:
                    pass

atexit.register(kill_all)


def stream(name: str, proc: subprocess.Popen):
    for line in proc.stdout:
        print(f"[{name}] {line}", end="")


def ensure_frontend_dependencies() -> None:
    frontend_dir = ROOT / "tools" / "main" / "spa"
    angular_cli = frontend_dir / "node_modules" / ".bin" / "ng"
    if angular_cli.exists():
        return

    print("Installing frontend dependencies: npm ci")
    subprocess.run([*resolve_cmd(["npm", "ci"])], cwd=frontend_dir, check=True)


def ensure_docker() -> None:
    if not shutil.which("docker"):
        raise RuntimeError("Docker is not installed or is missing from PATH.")

    def ready() -> bool:
        try:
            return subprocess.run(
                ["docker", "info"], stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL, timeout=5,
            ).returncode == 0
        except subprocess.TimeoutExpired:
            return False

    if ready():
        return
    if sys.platform != "darwin" or os.environ.get("DOCKER_HOST") or os.environ.get("DOCKER_CONTEXT"):
        raise RuntimeError("Docker is unavailable. Start your Docker engine and check your Docker context.")
    context = subprocess.run(
        ["docker", "context", "show"], capture_output=True, text=True, check=True,
    ).stdout.strip()
    if context not in ("default", "desktop-linux"):
        raise RuntimeError(f"Docker context '{context}' is unavailable. Start its Docker engine.")

    print("Starting Docker Desktop; waiting up to 120 seconds for the engine...")
    subprocess.run(["open", "-a", "Docker"], check=True)
    deadline = time.monotonic() + 120
    while time.monotonic() < deadline:
        if ready():
            return
        time.sleep(2)
    raise RuntimeError("Docker did not become ready. Open Docker Desktop and check its startup status.")


def ensure_mongodb() -> None:
    ensure_docker()
    print("Starting MongoDB: docker compose up --detach --wait mongodb")
    subprocess.run(
        [*resolve_cmd(["docker", "compose", "up", "--detach", "--wait", "mongodb"])],
        cwd=ROOT,
        check=True,
    )


def ensure_server_data() -> None:
    data_file = ROOT / "server" / "main" / "data" / "areas.json"
    if data_file.exists():
        return

    print("Generating game data: go run . deploy bloop")
    subprocess.run(
        [*resolve_cmd(["go", "run", ".", "deploy", "bloop"])],
        cwd=ROOT / "tools" / "main",
        check=True,
    )


def main() -> int:
    env = os.environ.copy()
    env.setdefault(
        "WORLD_DESIGN_DIR", str(ROOT / "tools" / "main" / "spa" / "dist" / "spa" / "browser")
    )
    ensure_frontend_dependencies()
    ensure_mongodb()
    ensure_server_data()

    for name, cmd, cwd in SERVICES:
        resolved = resolve_cmd(cmd)
        print(f"Starting {name}: {' '.join(cmd)}  (cwd={cwd})")
        p = subprocess.Popen(
            resolved, cwd=cwd,
            env=env,
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
            text=True, encoding="utf-8", errors="replace",
            # Prevent Ctrl+C from reaching children directly; we kill them
            # ourselves via taskkill, avoiding "Terminate batch job?" prompts.
            creationflags=subprocess.CREATE_NEW_PROCESS_GROUP if IS_WIN else 0,
            start_new_session=not IS_WIN,
        )
        procs.append((name, p))
        threading.Thread(target=stream, args=(name, p), daemon=True).start()

    try:
        while True:
            for name, p in procs:
                if p.poll() is not None:
                    print(f"\n'{name}' exited (code {p.returncode}). Stopping all.")
                    kill_all()
                    return p.returncode or 1
            time.sleep(0.5)
    except KeyboardInterrupt:
        print("\nInterrupted.")
        kill_all()
        return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (RuntimeError, OSError, subprocess.SubprocessError) as exc:
        print(f"Startup failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
