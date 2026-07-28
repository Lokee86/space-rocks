from __future__ import annotations

import os
import shlex
from collections.abc import Mapping
from pathlib import Path

_RUNTIME_OUTPUT = "SPACE_ROCKS_RUNTIME_SCENARIO_OUTPUT"


def runtime_server_command(
    server_root: Path,
    runtime_env: Mapping[str, str],
    *,
    platform: str | None = None,
) -> list[str]:
    """Launch current server source without starting a Windows game-server.exe."""
    host_platform = os.name if platform is None else platform
    if host_platform != "nt":
        return ["go", "run", "./cmd/game-server"]

    wsl_env = dict(runtime_env)
    output_path = Path(wsl_env[_RUNTIME_OUTPUT])
    wsl_env[_RUNTIME_OUTPUT] = os.path.relpath(output_path, server_root).replace("\\", "/")
    assignments = " ".join(
        f"{key}={shlex.quote(value)}" for key, value in sorted(wsl_env.items())
    )
    shell_command = f"exec env {assignments} go run ./cmd/game-server"
    return [
        "wsl.exe",
        "--cd",
        str(server_root),
        "--",
        "bash",
        "-lic",
        shell_command,
    ]
