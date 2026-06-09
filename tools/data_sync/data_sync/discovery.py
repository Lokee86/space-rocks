"""Constants destination discovery data shapes."""

from __future__ import annotations

import fnmatch
from dataclasses import dataclass
from pathlib import Path

from data_sync.block_io import find_all_blocks
from data_sync.config import DataSyncConfig


@dataclass(frozen=True)
class DiscoveredConstantsBlock:
    path: Path
    language: str
    section: str


@dataclass(frozen=True)
class DiscoveredConstantsFile:
    path: Path
    language: str
    sections: tuple[str, ...]


def language_for_path(path: Path) -> str | None:
    suffix = path.suffix.lower()
    if suffix == ".go":
        return "go"
    if suffix == ".gd":
        return "gds"
    if suffix == ".ts":
        return "ts"
    return None


def discover_constants_files(
    config: DataSyncConfig,
    languages: tuple[str, ...] | list[str],
) -> tuple[DiscoveredConstantsFile, ...]:
    discovered: list[DiscoveredConstantsFile] = []
    for path in _included_paths(config):
        language = language_for_path(path)
        if language is None:
            continue
        if language not in languages:
            continue
        text = path.read_text(encoding="utf-8")
        sections = tuple(
            block.section
            for block in find_all_blocks(text)
            if block.section.startswith("constants.")
        )
        if not sections:
            continue
        discovered.append(
            DiscoveredConstantsFile(
                path=path,
                language=language,
                sections=sections,
            )
        )
    return tuple(discovered)


def _included_paths(config: DataSyncConfig) -> tuple[Path, ...]:
    paths: list[Path] = []
    for pattern in config.constants_scan.include:
        for path in sorted(config.root.glob(pattern)):
            if not path.is_file():
                continue
            relative_path = path.relative_to(config.root)
            if _is_excluded(relative_path, config.constants_scan.exclude):
                continue
            paths.append(path)
    return tuple(sorted(paths))


def _is_excluded(relative_path: Path, exclude_patterns: tuple[str, ...]) -> bool:
    relative_text = relative_path.as_posix()
    return any(fnmatch.fnmatchcase(relative_text, pattern) for pattern in exclude_patterns)
