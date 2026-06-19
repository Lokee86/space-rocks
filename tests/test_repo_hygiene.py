from __future__ import annotations

import os
from pathlib import Path


def test_no_python_cache_artifacts_in_doc_ledgers_worktree() -> None:
    repo_root = Path(__file__).resolve().parents[1]
    offenders: list[str] = []

    for path in _walk_repo_files(repo_root):
        if "__pycache__" in path.parts:
            offenders.append(path.relative_to(repo_root).as_posix())
            continue
        if path.suffix in {".pyc", ".pyo"}:
            offenders.append(path.relative_to(repo_root).as_posix())

    assert offenders == [], f"Python cache artifacts found: {offenders}"


def _walk_repo_files(root: Path):
    for current_root, dirnames, filenames in os.walk(root):
        dirnames[:] = [name for name in dirnames if name != ".git"]
        current_root_path = Path(current_root)
        for filename in filenames:
            yield current_root_path / filename
