from __future__ import annotations

import re
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
GAME_INTERNAL = REPO_ROOT / "services" / "game-server" / "internal" / "game"
DEVTOOLS_INTERNAL = REPO_ROOT / "services" / "game-server" / "internal" / "devtools"
ROOT_GAME_IMPORT = '"github.com/Lokee86/space-rocks/services/game-server/internal/game"'
DEVTOOLS_METHOD_RE = re.compile(r"func\s*\([^)]*\*Game\s*\)\s*Devtools\w*\s*\(")


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


def test_game_package_files_do_not_own_devtools_continuous_bullet_stream_state() -> None:
    violations: list[str] = []
    forbidden_terms = (
        "activeDebugBulletStreams",
        "DevtoolsContinuousBulletStream",
        "stepDevtoolsContinuousBulletStreams",
        "continuousBulletStreams",
    )

    for path in sorted(GAME_INTERNAL.glob("*.go")):
        text = path.read_text(encoding="utf-8")
        if any(term in text for term in forbidden_terms):
            violations.append(str(path.relative_to(REPO_ROOT)))

    assert violations == []


def test_control_anchor_and_focused_control_files_exist() -> None:
    assert (GAME_INTERNAL / "control.go").exists()
    assert list(GAME_INTERNAL.glob("control_*.go")), "expected at least one focused control_*.go file"


def test_no_legacy_export_devtools_files_remain_in_game() -> None:
    legacy_files = sorted(path.name for path in GAME_INTERNAL.glob("export_devtools*.go"))
    assert legacy_files == []


def test_no_game_devtools_methods_remain() -> None:
    violations: list[str] = []
    for path in sorted(GAME_INTERNAL.glob("*.go")):
        text = path.read_text(encoding="utf-8")
        if DEVTOOLS_METHOD_RE.search(text):
            violations.append(str(path.relative_to(REPO_ROOT)))
    assert violations == []


def test_devtools_production_files_do_not_import_root_game_package() -> None:
    violations: list[str] = []
    for path in sorted(DEVTOOLS_INTERNAL.glob("*.go")):
        if path.name.endswith("_test.go"):
            continue
        text = path.read_text(encoding="utf-8")
        if ROOT_GAME_IMPORT in text:
            violations.append(str(path.relative_to(REPO_ROOT)))
    assert violations == []


def test_devtools_package_allows_game_subpackage_imports_and_has_controller_targets() -> None:
    assert (DEVTOOLS_INTERNAL / "controller.go").exists()
    assert (DEVTOOLS_INTERNAL / "target.go").exists()
