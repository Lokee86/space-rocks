from __future__ import annotations

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
PLAYER_DATA_ROOT = REPO_ROOT / "services" / "player-data"
FORBIDDEN_IMPORTS = (
    "github.com/Lokee86/space-rocks/server/internal/",
    "services/game-server",
)
SKIP_PATH_PARTS = {".git", "vendor", "build", "dist", "node_modules", "__pycache__"}


def test_player_data_does_not_import_game_server_internals() -> None:
    violations: list[str] = []

    for path in sorted(PLAYER_DATA_ROOT.rglob("*.go")):
        if any(part in SKIP_PATH_PARTS for part in path.parts):
            continue

        text = path.read_text(encoding="utf-8")
        for line_number, line in enumerate(text.splitlines(), start=1):
            if not line.startswith("import "):
                continue
            if any(forbidden in line for forbidden in FORBIDDEN_IMPORTS):
                relative_path = path.relative_to(REPO_ROOT)
                violations.append(f"{relative_path}:{line_number}: {line.strip()}")

    assert violations == []
