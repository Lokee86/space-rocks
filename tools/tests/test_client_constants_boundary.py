from __future__ import annotations

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
CLIENT_SCRIPTS = REPO_ROOT / "client" / "scripts"
GENERATED_CONSTANTS = CLIENT_SCRIPTS / "constants" / "constants.gd"

FORBIDDEN_REFERENCES = (
    "Constants.PLAYER_STARTING_LIVES",
    "Constants.PLAYER_RESPAWN_DELAY",
    "Constants.ASTEROID_SIZE_SCALE",
)


def test_active_client_scripts_do_not_reference_server_owned_constants() -> None:
    violations: list[str] = []

    for path in sorted(CLIENT_SCRIPTS.rglob("*.gd")):
        if path == GENERATED_CONSTANTS:
            continue

        text = path.read_text(encoding="utf-8")
        for line_number, line in enumerate(text.splitlines(), start=1):
            for reference in FORBIDDEN_REFERENCES:
                if reference in line:
                    relative_path = path.relative_to(REPO_ROOT)
                    violations.append(f"{relative_path}:{line_number}: {reference}")

    assert violations == []
