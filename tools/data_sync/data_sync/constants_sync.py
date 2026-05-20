"""Constants push/diff/check support."""

from __future__ import annotations

import difflib
from dataclasses import dataclass
from pathlib import Path
from typing import Callable

from data_sync.block_io import BlockIOError, replace_block
from data_sync.config import DataSyncConfig, DomainLanguageConfig
from data_sync.generators import gds_constants, go_constants, ts_constants
from data_sync.toml_store import TomlStore


class ConstantsSyncError(Exception):
    """Raised when constants sync cannot complete."""


Generator = Callable[[str, tuple[tuple[str, object], ...]], str]


GENERATORS: dict[str, Generator] = {
    "go": go_constants.generate_constants,
    "gds": gds_constants.generate_constants,
    "ts": ts_constants.generate_constants,
}


@dataclass(frozen=True)
class FileUpdate:
    path: Path
    before: str
    after: str

    @property
    def changed(self) -> bool:
        return self.before != self.after


def plan_constants_updates(
    config: DataSyncConfig,
    store: TomlStore,
    languages: tuple[str, ...],
) -> tuple[FileUpdate, ...]:
    updates: list[FileUpdate] = []
    for language in languages:
        target = config.target("constants", language)
        updates.extend(_plan_target_updates(store, target))
    return tuple(updates)


def apply_updates(updates: tuple[FileUpdate, ...]) -> None:
    for update in updates:
        if update.changed:
            update.path.write_text(update.after, encoding="utf-8")


def unified_diff(updates: tuple[FileUpdate, ...]) -> str:
    chunks: list[str] = []
    for update in updates:
        if not update.changed:
            continue
        chunks.extend(
            difflib.unified_diff(
                update.before.splitlines(keepends=True),
                update.after.splitlines(keepends=True),
                fromfile=str(update.path),
                tofile=str(update.path),
            )
        )
    return "".join(chunks)


def all_synced(updates: tuple[FileUpdate, ...]) -> bool:
    return all(not update.changed for update in updates)


def _plan_target_updates(store: TomlStore, target: DomainLanguageConfig) -> tuple[FileUpdate, ...]:
    generator = GENERATORS.get(target.language)
    if generator is None:
        raise ConstantsSyncError(f"unsupported constants language: {target.language}")
    if not target.enabled:
        raise ConstantsSyncError(f"[constants.{target.language}] is disabled in config")

    updates: list[FileUpdate] = []
    for path in target.files:
        try:
            text = path.read_text(encoding="utf-8")
        except FileNotFoundError as exc:
            raise ConstantsSyncError(f"configured constants file does not exist: {path}") from exc
        except OSError as exc:
            raise ConstantsSyncError(f"failed to read constants file {path}: {exc}") from exc

        updated = text
        for section_name in target.sections:
            section = store.constants(section_name)
            generated = generator(section.name, section.values)
            try:
                updated = replace_block(updated, section_name, generated)
            except BlockIOError as exc:
                raise ConstantsSyncError(f"{path}: {exc}") from exc

        updates.append(FileUpdate(path=path, before=text, after=updated))
    return tuple(updates)
