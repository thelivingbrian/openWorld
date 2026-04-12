from __future__ import annotations

import signal
import shutil
import subprocess
import sys
import threading
import time
import re
from pathlib import Path


ROOT = Path(__file__).resolve().parent


class ManagedProcess:
    def __init__(self, name: str, command: list[str], cwd: Path):
        self.name = name
        self.command = command
        self.cwd = cwd
        self.process: subprocess.Popen[str] | None = None
        self.reader_thread: threading.Thread | None = None

    def start(self) -> None:
        resolved_executable = resolve_executable(self.command[0])
        if resolved_executable is None:
            raise FileNotFoundError(f"Could not find executable '{self.command[0]}' in PATH")

        command = [resolved_executable, *self.command[1:]]
        creation_flags = 0
        popen_kwargs: dict[str, object] = {}
        if sys.platform.startswith("win"):
            creation_flags = subprocess.CREATE_NEW_PROCESS_GROUP
        else:
            popen_kwargs["start_new_session"] = True

        self.process = subprocess.Popen(
            command,
            cwd=self.cwd,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            encoding="utf-8",
            errors="replace",
            creationflags=creation_flags,
            **popen_kwargs,
        )
        self.reader_thread = threading.Thread(target=self._read_output, daemon=True)
        self.reader_thread.start()

    def _read_output(self) -> None:
        if self.process is None or self.process.stdout is None:
            return
        for line in self.process.stdout:
            print(f"[{self.name}] {line}", end="")

    def poll(self) -> int | None:
        if self.process is None:
            return None
        return self.process.poll()

    def terminate(self) -> None:
        if self.process is None:
            return
        if self.process.poll() is None:
            if sys.platform.startswith("win"):
                self._taskkill(force=False)
            else:
                self.process.terminate()

    def kill(self) -> None:
        if self.process is None:
            return
        if self.process.poll() is None:
            if sys.platform.startswith("win"):
                self._taskkill(force=True)
            else:
                self.process.kill()

    def wait(self, timeout: float | None = None) -> int | None:
        if self.process is None:
            return None
        return self.process.wait(timeout=timeout)

    def _taskkill(self, force: bool) -> None:
        if self.process is None or self.process.pid is None:
            return

        command = ["taskkill", "/PID", str(self.process.pid), "/T"]
        if force:
            command.append("/F")

        subprocess.run(
            command,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            check=False,
        )


def resolve_executable(executable: str) -> str | None:
    direct_match = shutil.which(executable)
    if direct_match:
        return direct_match

    if sys.platform.startswith("win") and not executable.lower().endswith((".exe", ".cmd", ".bat")):
        for extension in (".exe", ".cmd", ".bat"):
            candidate = shutil.which(f"{executable}{extension}")
            if candidate:
                return candidate

    return None


def find_listening_pids_on_port(port: int) -> list[int]:
    try:
        result = subprocess.run(
            ["netstat", "-ano", "-p", "tcp"],
            capture_output=True,
            text=True,
            check=False,
        )
    except OSError:
        return []

    pids: set[int] = set()
    pattern = re.compile(r"\s+", re.ASCII)

    for line in result.stdout.splitlines():
        normalized = pattern.sub(" ", line.strip())
        if not normalized:
            continue

        parts = normalized.split(" ")
        if len(parts) < 5:
            continue

        local_address = parts[1]
        state = parts[3].upper()
        pid_text = parts[4]

        if state != "LISTENING":
            continue
        if not local_address.endswith(f":{port}"):
            continue

        try:
            pids.add(int(pid_text))
        except ValueError:
            continue

    return sorted(pids)


def kill_pid_tree(pid: int) -> bool:
    if sys.platform.startswith("win"):
        result = subprocess.run(
            ["taskkill", "/PID", str(pid), "/T", "/F"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            check=False,
        )
        return result.returncode == 0

    try:
        subprocess.run(["kill", "-TERM", str(pid)], check=False)
    except OSError:
        return False
    return True


def clear_stale_port_listeners(ports: list[int]) -> None:
    for port in ports:
        pids = find_listening_pids_on_port(port)
        if not pids:
            continue

        print(f"Found existing listener(s) on :{port} -> PID(s) {pids}. Attempting cleanup...")
        for pid in pids:
            if kill_pid_tree(pid):
                print(f"Stopped PID {pid} on :{port}.")
            else:
                print(f"Could not stop PID {pid} on :{port}; start may fail.")


def main() -> int:
    clear_stale_port_listeners([4444])

    processes = [
        ManagedProcess("server", ["go", "run", "."], ROOT / "server" / "main"),
        ManagedProcess("tools", ["go", "run", "."], ROOT / "tools" / "main"),
        ManagedProcess("frontend", ["npm", "run", "watch"], ROOT / "tools" / "main" / "spa"),
    ]

    try:
        for proc in processes:
            print(f"Starting {proc.name}: {' '.join(proc.command)} (cwd={proc.cwd})")
            proc.start()
    except FileNotFoundError as exc:
        print(f"Failed to start '{proc.name}': {exc}")
        for started_proc in processes:
            started_proc.terminate()
        return 1

    stop_requested = False

    def request_stop(_signum: int, _frame: object) -> None:
        nonlocal stop_requested
        stop_requested = True

    signal.signal(signal.SIGINT, request_stop)
    if hasattr(signal, "SIGTERM"):
        signal.signal(signal.SIGTERM, request_stop)

    exit_code = 0
    try:
        while True:
            if stop_requested:
                break

            for proc in processes:
                code = proc.poll()
                if code is not None:
                    print(f"\nProcess '{proc.name}' exited with code {code}. Stopping all processes.")
                    if proc.name == "server" and code != 0:
                        print("Hint: the game server requires MongoDB at localhost:27017 (see readme.md).")
                    exit_code = code if code != 0 else exit_code
                    stop_requested = True
                    break

            if stop_requested:
                break

            threading.Event().wait(0.5)
    finally:
        print("\nShutting down...")
        for proc in processes:
            proc.kill()

        time.sleep(0.5)

        for proc in processes:
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                proc.kill()

        for proc in processes:
            if proc.reader_thread is not None:
                proc.reader_thread.join(timeout=1)

    return exit_code


if __name__ == "__main__":
    raise SystemExit(main())
