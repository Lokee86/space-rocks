from __future__ import annotations

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
GAME_INTERNAL = REPO_ROOT / "services" / "game-server" / "internal" / "game"


def test_game_package_has_no_debug_prefixed_go_files() -> None:
    debug_files = sorted(path.name for path in GAME_INTERNAL.glob("debug_*.go"))
    assert debug_files == []


def test_game_package_files_do_not_import_internal_devtools() -> None:
    violations: list[str] = []

    for path in sorted(GAME_INTERNAL.glob("*.go")):
        text = path.read_text(encoding="utf-8")
        if "/internal/devtools" in text:
            violations.append(str(path.relative_to(REPO_ROOT)))

    assert violations == []


def test_export_devtools_anchor_file_exists() -> None:
    assert (GAME_INTERNAL / "export_devtools.go").exists()


def test_export_devtools_split_files_exist() -> None:
    split_files = sorted(path.name for path in GAME_INTERNAL.glob("export_devtools_*.go"))
    assert len(split_files) >= 1
