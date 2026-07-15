from __future__ import annotations

from pathlib import Path


# The inventory is intentionally exact: the player-data bridge has been fully
# removed, so any future legacy call is an architecture violation.
PLAYER_DATA_BRIDGE_INVENTORY: tuple[str, ...] = ()
LEGACY_CALL_MARKERS = (
    "logging.HTTP.",
    "logging.Runtime.",
    "logging.Store.",
    "logging.Server.",
    "playerlogging.HTTP.",
    "playerlogging.Runtime.",
    "playerlogging.Store.",
    "playerlogging.Server.",
)


def test_player_data_bridge_inventory_is_exact() -> None:
    repo_root = Path(__file__).resolve().parents[1]
    observed: list[str] = []
    for path in sorted((repo_root / "services" / "player-data").rglob("*.go")):
        if path.name.endswith("_test.go") or path.name == "logger.go":
            continue
        text = path.read_text(encoding="utf-8")
        if any(marker in text for marker in LEGACY_CALL_MARKERS):
            observed.append(path.relative_to(repo_root).as_posix())
    assert tuple(observed) == PLAYER_DATA_BRIDGE_INVENTORY


def test_player_data_production_does_not_emit_legacy_log_message() -> None:
    repo_root = Path(__file__).resolve().parents[1]
    offenders: list[str] = []
    for path in sorted((repo_root / "services" / "player-data").rglob("*.go")):
        if path.name.endswith("_test.go") or path.name == "logger.go":
            continue
        if "log_message" in path.read_text(encoding="utf-8"):
            offenders.append(path.relative_to(repo_root).as_posix())
    assert offenders == []
