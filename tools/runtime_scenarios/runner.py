from __future__ import annotations

import json
import os
import sys
import time
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from runtime_scenarios.model import Scenario
from runtime_scenarios.phase_markers import phase_markers_for_scenario
from runtime_scenarios.server_command import runtime_server_command
from runtime_scenarios.processes import (
    ManagedProcess,
    find_godot,
    reserve_free_loopback_port,
    start_process,
    stop_processes,
    wait_for_health,
)


@dataclass(frozen=True)
class RunOptions:
    repo_root: Path
    output_root: Path
    godot: str | None = None
    headless_coordinator: bool = False

class ScenarioRunner:
    def __init__(self, scenario: Scenario, options: RunOptions) -> None:
        self.scenario = scenario
        self.options = options
        self.run_directory = self._new_run_directory()
        self.processes: list[ManagedProcess] = []
        self.status_paths: dict[str, Path] = {}
        self.started_at = datetime.now(timezone.utc)
        self.port = reserve_free_loopback_port()
        self.health_url = f"http://127.0.0.1:{self.port}/health"
        self.websocket_url = f"ws://127.0.0.1:{self.port}/ws"

    def run(self) -> int:
        self.run_directory.mkdir(parents=True, exist_ok=False)
        phase_markers = phase_markers_for_scenario(self.scenario.raw)
        summary: dict[str, Any] = {
            "scenario_id": self.scenario.scenario_id,
            "scenario_path": str(self.scenario.path),
            "seed": self.scenario.seed,
            "started_at": self.started_at.isoformat(),
            "run_directory": str(self.run_directory),
            "phase_markers": phase_markers,
            "phase_markers_path": str(self.run_directory / "phase-markers.json"),
            "success": False,
        }
        try:
            godot = find_godot(self.options.godot)
            summary["execution"] = {
                "godot_editor": str(godot),
                "source_project": str(self.options.repo_root / "client"),
                "coordinator_headless": self.options.headless_coordinator,
                "packaged_client_started": False,
                "bundled_local_server_started": False,
                "server_launch": (
                    "WSL go run ./cmd/game-server"
                    if os.name == "nt"
                    else "go run ./cmd/game-server"
                ),
                "harness_server_binary_created": False,
                "windows_game_server_executable_started": False,
            }
            self._start_server()
            wait_for_health(self.health_url, self.scenario.setup_timeout_seconds)

            coordinator = self._start_client(
                godot=godot,
                client_id="coordinator-1",
                role="coordinator",
                headless=self.options.headless_coordinator,
            )
            room_status = self._wait_for_status(
                coordinator,
                accepted={"room_ready"},
                timeout=self.scenario.setup_timeout_seconds,
            )
            room_code = str(room_status.get("room_code", "")).strip()
            if not room_code:
                raise RuntimeError("coordinator did not publish a room code")

            participants: list[ManagedProcess] = []
            for index in range(self.scenario.clients.headless):
                participants.append(
                    self._start_client(
                        godot=godot,
                        client_id=f"participant-{index + 1}",
                        role="participant",
                        headless=True,
                        room_code=room_code,
                    )
                )

            client_processes = [coordinator, *participants]
            final_statuses = self._wait_for_completion(client_processes)
            failures = {
                name: status
                for name, status in final_statuses.items()
                if status.get("state") != "completed"
            }
            if failures:
                raise RuntimeError(f"one or more clients failed: {failures}")

            summary["success"] = True
            summary["room_code"] = room_code
            summary["clients"] = final_statuses
            return 0
        except Exception as exc:  # noqa: BLE001 - CLI boundary records full failure
            summary["error"] = str(exc)
            summary["clients"] = self._read_all_statuses()
            return 1
        finally:
            stop_processes(self.processes)
            summary["ended_at"] = datetime.now(timezone.utc).isoformat()
            summary["processes"] = self._process_summary()
            self._write_json(
                self.run_directory / "phase-markers.json",
                {
                    "scenario_id": self.scenario.scenario_id,
                    "source": "configured_phase_durations",
                    "phases": phase_markers,
                },
            )
            self._write_json(self.run_directory / "summary.json", summary)
            print(json.dumps(summary, indent=2), file=sys.stdout)

    def _start_server(self) -> None:
        server_root = self.options.repo_root / "services" / "game-server"
        runtime_env = {
            "BUILD_VERSION": "runtime-scenario",
            "ENVIRONMENT": "development",
            "SPACE_ROCKS_RUNTIME_SCENARIO_SEED": str(self.scenario.seed),
            "SPACE_ROCKS_RUNTIME_SCENARIO_AUTH": "1",
            "SPACE_ROCKS_RUNTIME_SCENARIO_OUTPUT": str(self.run_directory),
            "SPACE_ROCKS_RUNTIME_SCENARIO_PORT": str(self.port),
        }
        env = os.environ.copy()
        env.update(runtime_env)
        managed = start_process(
            "game-server",
            runtime_server_command(server_root, runtime_env),
            server_root,
            self.run_directory / "game-server.log",
            env=env,
        )
        self.processes.append(managed)

    def _start_client(
        self,
        *,
        godot: Path,
        client_id: str,
        role: str,
        headless: bool,
        room_code: str = "",
    ) -> ManagedProcess:
        status_path = self.run_directory / f"{client_id}-status.json"
        self.status_paths[client_id] = status_path
        command = [str(godot), "--path", str(self.options.repo_root / "client")]
        if headless:
            command.append("--headless")
        command.extend(
            [
                "--",
                f"--runtime-scenario={self.scenario.path}",
                f"--runtime-scenario-role={role}",
                f"--runtime-scenario-client-id={client_id}",
                f"--runtime-scenario-status={status_path}",
                f"--runtime-scenario-server-url={self.websocket_url}",
            ]
        )
        if room_code:
            command.append(f"--runtime-scenario-room-code={room_code}")
        managed = start_process(
            client_id,
            command,
            self.options.repo_root / "client",
            self.run_directory / f"{client_id}.log",
        )
        self.processes.append(managed)
        return managed

    def _wait_for_completion(self, clients: list[ManagedProcess]) -> dict[str, dict[str, Any]]:
        deadline = time.monotonic() + self.scenario.timeout_seconds
        terminal = {"completed", "failed"}
        while time.monotonic() < deadline:
            statuses = self._read_all_statuses()
            for client in clients:
                return_code = client.process.poll()
                state = str(statuses.get(client.name, {}).get("state", ""))
                if return_code is not None and state not in terminal:
                    raise RuntimeError(
                        f"{client.name} exited with code {return_code} before reporting completion"
                    )
            if all(str(statuses.get(client.name, {}).get("state", "")) in terminal for client in clients):
                return {client.name: statuses[client.name] for client in clients}
            time.sleep(0.1)
        raise TimeoutError("runtime scenario exceeded its timeout")

    def _wait_for_status(
        self, client: ManagedProcess, *, accepted: set[str], timeout: float
    ) -> dict[str, Any]:
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            status = self._read_status(client.name)
            state = str(status.get("state", ""))
            if state == "failed":
                raise RuntimeError(f"{client.name} failed: {status.get('error', 'unknown error')}")
            if state in accepted:
                return status
            return_code = client.process.poll()
            if return_code is not None:
                raise RuntimeError(
                    f"{client.name} exited with code {return_code} while waiting for {sorted(accepted)}"
                )
            time.sleep(0.1)
        raise TimeoutError(f"timed out waiting for {client.name} status {sorted(accepted)}")

    def _read_status(self, client_id: str) -> dict[str, Any]:
        path = self.status_paths.get(client_id)
        if path is None or not path.exists():
            return {}
        try:
            payload = json.loads(path.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError):
            return {}
        return payload if isinstance(payload, dict) else {}

    def _read_all_statuses(self) -> dict[str, dict[str, Any]]:
        return {client_id: self._read_status(client_id) for client_id in self.status_paths}

    def _process_summary(self) -> dict[str, dict[str, Any]]:
        return {
            managed.name: {
                "return_code": managed.process.poll(),
                "log": str(self.run_directory / f"{managed.name}.log"),
            }
            for managed in self.processes
        }

    def _new_run_directory(self) -> Path:
        stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
        return self.options.output_root / f"{self.scenario.scenario_id}-{stamp}"

    @staticmethod
    def _write_json(path: Path, payload: dict[str, Any]) -> None:
        path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
