from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.model import Scenario, ScenarioError
from runtime_scenarios.phase_markers import phase_markers_for_scenario
from runtime_scenarios.processes import find_godot, prepare_godot_project
from runtime_scenarios.rounds import expand_rounds


def write_scenario(path: Path, **overrides: object) -> Path:
    payload: dict[str, object] = {
        "id": "test_scenario",
        "seed": 7,
        "timeout_seconds": 10,
        "setup_timeout_seconds": 2,
        "clients": {"visible": 1, "headless": 1},
        "bots": 2,
        "phases": [{"name": "warmup", "duration_seconds": 1}],
    }
    payload.update(overrides)
    path.write_text(json.dumps(payload), encoding="utf-8")
    return path


def test_loads_valid_runtime_scenario(tmp_path: Path) -> None:
    scenario = Scenario.load(write_scenario(tmp_path / "scenario.json"))
    assert scenario.scenario_id == "test_scenario"
    assert scenario.seed == 7
    assert scenario.clients.total == 2
    assert scenario.room_count == 1
    assert scenario.bots == 2


def test_requires_exactly_one_visible_coordinator(tmp_path: Path) -> None:
    path = write_scenario(
        tmp_path / "scenario.json",
        clients={"visible": 2, "headless": 0},
    )
    with pytest.raises(ScenarioError, match="exactly one visible coordinator"):
        Scenario.load(path)


def test_rejects_missing_phases(tmp_path: Path) -> None:
    path = write_scenario(tmp_path / "scenario.json", phases=[])
    with pytest.raises(ScenarioError, match="non-empty array"):
        Scenario.load(path)


def test_loads_valid_match_churn_rounds(tmp_path: Path) -> None:
    path = tmp_path / "scenario.json"
    path.write_text(
        json.dumps(
            {
                "id": "match_churn",
                "seed": 8,
                "clients": {"visible": 1, "headless": 1},
                "bots": 2,
                "rounds": [
                    {
                        "name": "cycle-1",
                        "lives": 2,
                        "phases": [{"name": "pressure", "duration_seconds": 1}],
                    }
                ],
            }
        ),
        encoding="utf-8",
    )
    scenario = Scenario.load(path)
    assert scenario.raw["rounds"][0]["name"] == "cycle-1"


def test_expands_repeated_rounds_and_phase_markers() -> None:
    payload = {
        "rounds": [
            {
                "name": "soak",
                "repeat": 3,
                "phases": [{"name": "pressure", "duration_seconds": 2}],
            }
        ]
    }
    rounds = expand_rounds(payload["rounds"])
    markers = phase_markers_for_scenario(payload)

    assert [round_payload["name"] for round_payload in rounds] == [
        "soak-001",
        "soak-002",
        "soak-003",
    ]
    assert markers[-1]["round"] == 3
    assert markers[-1]["end_seconds"] == 6.0


def test_rejects_invalid_heap_profile_round(tmp_path: Path) -> None:
    path = write_scenario(tmp_path / "scenario.json", heap_profile_rounds=[0])
    with pytest.raises(ScenarioError, match="heap_profile_rounds"):
        Scenario.load(path)


def test_rejects_nonpositive_round_repeat(tmp_path: Path) -> None:
    path = tmp_path / "scenario.json"
    path.write_text(
        json.dumps(
            {
                "id": "invalid-repeat",
                "seed": 9,
                "clients": {"visible": 1, "headless": 0},
                "rounds": [
                    {
                        "name": "cycle",
                        "repeat": 0,
                        "phases": [{"name": "pressure", "duration_seconds": 1}],
                    }
                ],
            }
        ),
        encoding="utf-8",
    )
    with pytest.raises(ScenarioError, match="repeat"):
        Scenario.load(path)


def test_rejects_phases_and_rounds_together(tmp_path: Path) -> None:
    path = write_scenario(
        tmp_path / "scenario.json",
        rounds=[
            {
                "name": "cycle-1",
                "phases": [{"name": "pressure", "duration_seconds": 1}],
            }
        ],
    )
    with pytest.raises(ScenarioError, match="phases or rounds"):
        Scenario.load(path)


def test_rejects_negative_setup_asteroid_count(tmp_path: Path) -> None:
    path = write_scenario(
        tmp_path / "scenario.json",
        setup={"asteroid_spawns": -1},
    )
    with pytest.raises(ScenarioError, match="asteroid_spawns"):
        Scenario.load(path)


def test_rejects_negative_phase_bullet_streams(tmp_path: Path) -> None:
    path = write_scenario(
        tmp_path / "scenario.json",
        phases=[{"name": "pressure", "duration_seconds": 1, "bullet_streams": -1}],
    )
    with pytest.raises(ScenarioError, match="bullet_streams"):
        Scenario.load(path)


@pytest.mark.parametrize(
    ("filename", "clients", "bots", "seed", "room_count"),
    [
        ("network_interest_lifecycle_v1.json", 2, 6, 27072701, 1),
        ("match_churn_2c_6b_v1.json", 2, 6, 27072803, 1),
        ("match_churn_soak_2c_6b_v1.json", 2, 6, 27072804, 1),
        ("match_churn_heap_profile_2c_6b_v1.json", 2, 6, 27072806, 1),
        ("multi_room_3x1c_7b_v1.json", 1, 7, 27072901, 3),
        ("receiver_scale_1c_7b_v1.json", 1, 7, 27072801, 1),
        ("receiver_scale_2c_6b_v1.json", 2, 6, 27072801, 1),
        ("receiver_scale_4c_4b_v1.json", 4, 4, 27072801, 1),
        ("receiver_scale_8c_0b_v1.json", 8, 0, 27072801, 1),
        ("simulation_scale_1c_7b_v1.json", 1, 7, 27072802, 1),
    ],
)
def test_repository_scenarios_are_valid(
    filename: str, clients: int, bots: int, seed: int, room_count: int
) -> None:
    repo_root = Path(__file__).resolve().parents[2]
    scenario = Scenario.load(
        repo_root / "tools" / "runtime_scenarios" / "scenarios" / filename
    )
    assert scenario.clients.total == clients
    assert scenario.bots == bots
    assert scenario.seed == seed
    assert scenario.room_count == room_count
    assert scenario.clients.total + scenario.bots == 8


def test_prepare_godot_project_runs_headless_editor_scan(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    captured: dict[str, object] = {}

    class FakeProcess:
        def wait(self, timeout: float) -> int:
            captured["timeout"] = timeout
            return 0

    class FakeManaged:
        def __init__(self) -> None:
            self.process = FakeProcess()
            self.closed = False

        def close_log(self) -> None:
            self.closed = True

    managed = FakeManaged()

    def fake_start_process(
        name: str,
        command: list[str],
        cwd: Path,
        log_path: Path,
        **_kwargs: object,
    ) -> FakeManaged:
        captured.update(
            name=name,
            command=command,
            cwd=cwd,
            log_path=log_path,
        )
        return managed

    monkeypatch.setattr("runtime_scenarios.processes.start_process", fake_start_process)
    godot = tmp_path / "Godot.exe"
    project = tmp_path / "client"
    log = tmp_path / "scan.log"

    prepare_godot_project(godot, project, log)

    assert captured["name"] == "godot-project-scan"
    assert captured["command"] == [
        str(godot),
        "--headless",
        "--editor",
        "--path",
        str(project),
        "--quit",
    ]
    assert captured["cwd"] == project
    assert captured["log_path"] == log
    assert managed.closed is True


def test_find_godot_accepts_editor_executable(tmp_path: Path) -> None:
    editor = tmp_path / "Godot.exe"
    editor.write_bytes(b"")
    assert find_godot(str(editor)) == editor.resolve()


def test_find_godot_rejects_exported_game_executable(tmp_path: Path) -> None:
    exported_game = tmp_path / "SpaceRocks.exe"
    exported_game.write_bytes(b"")
    with pytest.raises(ValueError, match="not an exported game"):
        find_godot(str(exported_game))
