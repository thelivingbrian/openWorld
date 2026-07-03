from __future__ import annotations
import os, shutil, subprocess, sys, threading, time, atexit
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
                p.kill()

atexit.register(kill_all)


def stream(name: str, proc: subprocess.Popen):
    for line in proc.stdout:
        print(f"[{name}] {line}", end="")


def main() -> int:
    env = os.environ.copy()
    env.setdefault(
        "WORLD_DESIGN_DIR", str(ROOT / "tools" / "main" / "spa" / "dist" / "spa" / "browser")
    )

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
    raise SystemExit(main())
