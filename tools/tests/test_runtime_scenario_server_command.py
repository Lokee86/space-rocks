from __future__ import annotations

import sys
from pathlib import Path

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.server_command import runtime_server_command


def runtime_env(output: Path) -> dict[str, str]:
    return {
        "BUILD_VERSION": "runtime-scenario",
        "ENVIRONMENT": "development",
        "SPACE_ROCKS_RUNTIME_SCENARIO_AUTH": "1",
        "SPACE_ROCKS_RUNTIME_SCENARIO_OUTPUT": str(output),
        "SPACE_ROCKS_RUNTIME_SCENARIO_PORT": "12345",
        "SPACE_ROCKS_RUNTIME_SCENARIO_SEED": "7",
    }


def test_windows_runtime_server_uses_wsl_go_run(tmp_path: Path) -> None:
    server_root = tmp_path / "services" / "game-server"
    output = tmp_path / ".ci-artifacts" / "runtime-scenarios" / "run"
    command = runtime_server_command(server_root, runtime_env(output), platform="nt")

    assert command[:2] == ["wsl.exe", "--cd"]
    assert command[2] == str(server_root)
    assert command[3:6] == ["--", "bash", "-lic"]
    assert "go run ./cmd/game-server" in command[-1]
    assert "game-server.exe" not in " ".join(command).lower()
    assert "go build" not in command[-1]
    assert "-o " not in command[-1]
    assert "../../.ci-artifacts/runtime-scenarios/run" in command[-1]


def test_non_windows_runtime_server_uses_direct_go_run(tmp_path: Path) -> None:
    command = runtime_server_command(
        tmp_path / "services" / "game-server",
        runtime_env(tmp_path / "run"),
        platform="posix",
    )
    assert command == ["go", "run", "./cmd/game-server"]
