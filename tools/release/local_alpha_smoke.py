from __future__ import annotations

import os
from pathlib import Path
import socket
import subprocess
import sys
import tempfile
import time

from .local_alpha_common import ReleaseGateError, run

DEFAULT_LOCAL_SERVER_PORT = 8080


def find_available_loopback_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as probe:
        probe.bind(("127.0.0.1", 0))
        return int(probe.getsockname()[1])


def assert_port_free(port: int) -> None:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as probe:
        probe.settimeout(0.25)
        if probe.connect_ex(("127.0.0.1", port)) == 0:
            raise ReleaseGateError(f"127.0.0.1:{port} is already in use")


def wait_for_port_closed(port: int, timeout_seconds: float = 5.0) -> None:
    deadline = time.monotonic() + timeout_seconds
    while time.monotonic() < deadline:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as probe:
            probe.settimeout(0.2)
            if probe.connect_ex(("127.0.0.1", port)) != 0:
                return
        time.sleep(0.1)
    raise ReleaseGateError(f"bundled server still owns 127.0.0.1:{port} after the client exited")


def smoke_environment(
    platform_name: str,
    data_root: Path,
    server_port: int = DEFAULT_LOCAL_SERVER_PORT,
) -> dict[str, str]:
    environment = os.environ.copy()
    environment["SPACE_ROCKS_RELEASE_GATE"] = "1"
    environment["SPACE_ROCKS_LOCAL_SERVER_PORT"] = str(server_port)
    if platform_name == "windows":
        environment["APPDATA"] = str(data_root / "AppData" / "Roaming")
        environment["LOCALAPPDATA"] = str(data_root / "AppData" / "Local")
        environment["USERPROFILE"] = str(data_root)
        environment["HOME"] = str(data_root)
    else:
        environment["HOME"] = str(data_root)
        environment["XDG_CONFIG_HOME"] = str(data_root / ".config")
    return environment


def print_smoke_diagnostics(data_root: Path) -> None:
    candidates = sorted(
        path
        for path in data_root.rglob("*")
        if path.is_file()
        and (
            path.name == "release_gate_smoke_result.json"
            or path.suffix.lower() in {".log", ".jsonl"}
        )
    )
    if not candidates:
        print("release smoke produced no diagnostic files", file=sys.stderr)
        return
    for path in candidates:
        print(f"--- release smoke diagnostic: {path.relative_to(data_root)} ---", file=sys.stderr)
        try:
            contents = path.read_text(encoding="utf-8", errors="replace")
        except OSError as error:
            print(f"could not read diagnostic: {error}", file=sys.stderr)
            continue
        print(contents[-20000:], file=sys.stderr)


def run_packaged_smoke(platform_name: str, client_executable: Path) -> None:
    server_port = find_available_loopback_port()
    assert_port_free(server_port)
    with tempfile.TemporaryDirectory(prefix="space-rocks-local-alpha-") as temporary:
        data_root = Path(temporary)
        environment = smoke_environment(platform_name, data_root, server_port)
        try:
            for phase in ("seed", "verify"):
                run(
                    [client_executable, "--headless", "--", f"--local-alpha-smoke={phase}"],
                    cwd=client_executable.parent,
                    env=environment,
                    timeout=120,
                )
                wait_for_port_closed(server_port)
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired, ReleaseGateError):
            print_smoke_diagnostics(data_root)
            raise
