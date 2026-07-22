from __future__ import annotations

import hashlib
import json
import os
from pathlib import Path
import subprocess
from typing import Sequence

from .local_alpha_build import native_architectures
from .local_alpha_common import ROOT


def git_commit() -> str:
    result = subprocess.run(
        ["git", "rev-parse", "--short=12", "HEAD"],
        cwd=ROOT,
        check=True,
        capture_output=True,
        text=True,
    )
    return result.stdout.strip()


def git_worktree_changes() -> list[str]:
    result = subprocess.run(
        ["git", "status", "--porcelain=v1", "--untracked-files=all"],
        cwd=ROOT,
        check=True,
        capture_output=True,
        text=True,
    )
    return [line for line in result.stdout.splitlines() if line.strip()]


def default_version(*, dirty: bool = False) -> str:
    suffix = "-dirty" if dirty else ""
    return f"local-alpha-{git_commit()}{suffix}"


def package_files(output_dir: Path) -> list[Path]:
    return sorted(
        path
        for path in output_dir.rglob("*")
        if path.is_file() and path.name != "release-manifest.json"
    )


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def write_manifest(
    platform_name: str,
    output_dir: Path,
    version: str,
    files: Sequence[Path],
    worktree_changes: Sequence[str],
) -> None:
    payload = {
        "build_shape": "local-packaged-single-player-alpha",
        "platform": platform_name,
        "architecture": native_architectures(platform_name)[0],
        "version": version,
        "commit": git_commit(),
        "dirty_worktree": bool(worktree_changes),
        "worktree_changes": list(worktree_changes),
        "files": [
            {
                "path": str(path.relative_to(output_dir)).replace(os.sep, "/"),
                "sha256": sha256(path),
                "bytes": path.stat().st_size,
            }
            for path in files
        ],
    }
    (output_dir / "release-manifest.json").write_text(
        json.dumps(payload, indent=2) + "\n",
        encoding="utf-8",
    )
