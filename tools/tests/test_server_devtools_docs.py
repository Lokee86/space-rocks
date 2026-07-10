from __future__ import annotations

import re
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
DOCS_ROOT = REPO_ROOT / "docs"
TASK_SKILL = REPO_ROOT / "skills" / "task" / "SKILL.md"

FORBIDDEN_TOKENS = (
    "export_devtools",
    "game-export-devtools-seams",
    "Game.Devtools",
    "DevtoolsStatusFor",
    "DevtoolsTargetPlayerIDs",
    "DevtoolsSpawnPlayerShip",
    "DevtoolsForceRespawnPlayer",
    "DevtoolsClearBullets",
    "DevtoolsClearAsteroids",
    "DevtoolsCollisionBodies",
    "DevtoolsCollisionBody",
)

FORBIDDEN_CALL_SHAPES = (
    "devtools.HandleCommand(room.GameInstance(",
    "devtools.StatusFor(gameInstance",
    "devtools.StatusFor(game,",
    "devtools.StatusesForAllPlayers(gameInstance",
    "devtools.StatusesForAllPlayers(game,",
)

FORBIDDEN_PHRASES = (
    "game export seam",
    "game-owned export seam",
    "game export method",
    "game-owned export method",
    "devtools export seam",
)


def _iter_markdown_files() -> list[Path]:
    return sorted(path for path in DOCS_ROOT.rglob("*.md") if path.is_file())


def _scan_file(path: Path) -> list[str]:
    text = path.read_text(encoding="utf-8")
    matches: list[str] = []

    for token in FORBIDDEN_TOKENS:
        if token in text:
            matches.append(token)

    for shape in FORBIDDEN_CALL_SHAPES:
        if shape in text:
            matches.append(shape)

    lowered = text.lower()
    for phrase in FORBIDDEN_PHRASES:
        if phrase in lowered:
            matches.append(phrase)

    return matches


def test_server_devtools_docs_do_not_restore_removed_architecture_names() -> None:
    violations: list[str] = []

    for path in _iter_markdown_files() + [TASK_SKILL]:
        matches = _scan_file(path)
        for match in matches:
            violations.append(f"{path.relative_to(REPO_ROOT)}: {match}")

    assert violations == [], "\n".join(violations)
