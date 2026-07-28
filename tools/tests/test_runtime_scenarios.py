from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.model import Scenario, ScenarioError
from runtime_scenarios.processes import find_godot


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
    ("filename", "clients", "bots", "seed"),
    [
        ("network_interest_lifecycle_v1.json", 2, 6, 27072701),
        ("receiver_scale_1c_7b_v1.json", 1, 7, 27072801),
        ("receiver_scale_2c_6b_v1.json", 2, 6, 27072801),
        ("receiver_scale_4c_4b_v1.json", 4, 4, 27072801),
        ("receiver_scale_8c_0b_v1.json", 8, 0, 27072801),
        ("simulation_scale_1c_7b_v1.json", 1, 7, 27072802),
    ],
)
def test_repository_scenarios_are_valid(
    filename: str, clients: int, bots: int, seed: int
) -> None:
    repo_root = Path(__file__).resolve().parents[2]
    scenario = Scenario.load(
        repo_root / "tools" / "runtime_scenarios" / "scenarios" / filename
    )
    assert scenario.clients.total == clients
    assert scenario.bots == bots
    assert scenario.seed == seed
    assert scenario.clients.total + scenario.bots == 8


def test_find_godot_accepts_editor_executable(tmp_path: Path) -> None:
    editor = tmp_path / "Godot.exe"
    editor.write_bytes(b"")
    assert find_godot(str(editor)) == editor.resolve()


def test_find_godot_rejects_exported_game_executable(tmp_path: Path) -> None:
    exported_game = tmp_path / "SpaceRocks.exe"
    exported_game.write_bytes(b"")
    with pytest.raises(ValueError, match="not an exported game"):
        find_godot(str(exported_game))
