from __future__ import annotations

import os
import shutil
import signal
import socket
import subprocess
import time
import urllib.error
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import IO, Sequence


@dataclass
class ManagedProcess:
    name: str
    process: subprocess.Popen[str]
    log_file: IO[str]

    def close_log(self) -> None:
        if not self.log_file.closed:
            self.log_file.close()


def find_godot(explicit: str | None) -> Path:
    candidates = [
        explicit,
        os.environ.get("SPACE_ROCKS_GODOT_EXECUTABLE"),
        r"C:\Godot.exe",
        shutil.which("godot"),
        shutil.which("godot4"),
    ]
    for candidate in candidates:
        if not candidate:
            continue
        path = Path(candidate).expanduser()
        if not path.is_file():
            continue
        if not path.name.lower().startswith("godot"):
            raise ValueError(
                f"runtime scenarios require the Godot editor executable, not an exported game: {path}"
            )
        return path.resolve()
    raise FileNotFoundError(
        "Godot editor was not found; pass --godot or set SPACE_ROCKS_GODOT_EXECUTABLE"
    )


def start_process(
    name: str,
    command: Sequence[str],
    cwd: Path,
    log_path: Path,
    *,
    env: dict[str, str] | None = None,
) -> ManagedProcess:
    log_path.parent.mkdir(parents=True, exist_ok=True)
    log_file = log_path.open("w", encoding="utf-8", errors="replace")
    creationflags = subprocess.CREATE_NEW_PROCESS_GROUP if os.name == "nt" else 0
    process = subprocess.Popen(
        list(command),
        cwd=str(cwd),
        env=env,
        stdout=log_file,
        stderr=subprocess.STDOUT,
        stdin=subprocess.DEVNULL,
        text=True,
        creationflags=creationflags,
    )
    return ManagedProcess(name=name, process=process, log_file=log_file)


def reserve_free_loopback_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as listener:
        listener.bind(("127.0.0.1", 0))
        return int(listener.getsockname()[1])


def wait_for_health(url: str, timeout_seconds: float) -> None:
    deadline = time.monotonic() + timeout_seconds
    last_error = "no response"
    while time.monotonic() < deadline:
        try:
            with urllib.request.urlopen(url, timeout=1.0) as response:
                if response.status == 200:
                    return
                last_error = f"HTTP {response.status}"
        except (OSError, urllib.error.URLError) as exc:
            last_error = str(exc)
        time.sleep(0.1)
    raise TimeoutError(f"server health check failed: {last_error}")


def health_is_available(url: str) -> bool:
    try:
        with urllib.request.urlopen(url, timeout=0.5) as response:
            return response.status == 200
    except (OSError, urllib.error.URLError):
        return False


def stop_processes(processes: list[ManagedProcess], grace_seconds: float = 4.0) -> None:
    for managed in reversed(processes):
        if managed.process.poll() is not None:
            continue
        if os.name == "nt":
            managed.process.send_signal(signal.CTRL_BREAK_EVENT)
        else:
            managed.process.terminate()
    deadline = time.monotonic() + grace_seconds
    for managed in reversed(processes):
        if managed.process.poll() is not None:
            continue
        remaining = max(deadline - time.monotonic(), 0.0)
        try:
            managed.process.wait(timeout=remaining)
        except subprocess.TimeoutExpired:
            _force_stop_process_tree(managed.process)
    for managed in processes:
        if managed.process.poll() is None:
            managed.process.wait(timeout=2.0)
        managed.close_log()


def _force_stop_process_tree(process: subprocess.Popen[str]) -> None:
    if os.name != "nt":
        process.kill()
        return
    subprocess.run(
        ["taskkill", "/PID", str(process.pid), "/T", "/F"],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        check=False,
    )
