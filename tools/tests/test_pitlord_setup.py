from __future__ import annotations

import json
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
PITLORD_DIR = REPO_ROOT / "tools" / "pitlord"


def load_json(name: str) -> dict[str, object]:
    return json.loads((PITLORD_DIR / name).read_text(encoding="utf-8"))


def test_pitlord_policy_composes_repository_and_semantic_rules() -> None:
    policy = load_json("policy.json")
    assert policy["includes"] == ["repository.json", "semantic.json"]

    repository_types = {rule["type"] for rule in load_json("repository.json")["rules"]}
    semantic = load_json("semantic.json")
    semantic_types = {rule["type"] for rule in semantic["rules"]}

    assert {"forbid_content", "require_path", "forbid_path"} <= repository_types
    assert {"forbid_dependency", "require_ownership", "forbid_area_cycles"} <= semantic_types
    game_core_rule = next(rule for rule in semantic["rules"] if rule["id"] == "game-core-must-not-depend-on-orchestration")
    assert "service-player-data" in game_core_rule["to_areas"]
    area_ids = {area["id"] for area in semantic["areas"]}
    assert {"client-teams", "game-server-matchresults", "game-server-playerinventory"} <= area_ids


def test_go_semantic_areas_cover_source_and_normalized_package_paths() -> None:
    semantic = load_json("semantic.json")
    areas = {area["id"]: area for area in semantic["areas"]}

    for area_id in (
        "game-server-game",
        "game-server-matchresults",
        "game-server-networking",
        "game-server-playerinventory",
        "game-server-protocol",
        "shared-go",
    ):
        paths = areas[area_id]["paths"]
        assert any(not path.startswith("@internal/") for path in paths)
        assert any(path.startswith("@internal/") for path in paths)


def test_pitlord_runner_refreshes_and_evaluates_the_graph() -> None:
    runner = (PITLORD_DIR / "run.sh").read_text(encoding="utf-8")
    assert '"$LEXICON" scan --repo .' in runner
    assert '"$ARCANA" sync --lexicon .lexicon --state .arcana' in runner
    assert '--arcana "$ARCANA"' in runner


def test_repository_ci_builds_pinned_semantic_toolchain() -> None:
    workflow = (REPO_ROOT / ".github" / "workflows" / "ci.yml").read_text(encoding="utf-8")
    assert "GRIMOIRE_COMMIT: 42c461b2590010c5717d8f1191a9a85e9c5ff202" in workflow
    assert "go build -o \"$RUNNER_TEMP/pitlord-bin/lexicon\"" in workflow
    assert "cargo build --manifest-path" in workflow
    assert "LEXICON_ADAPTERS=" in workflow
