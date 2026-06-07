"""Constants push/diff/check support."""

from __future__ import annotations

import difflib
from dataclasses import dataclass
from pathlib import Path
from typing import Callable, Protocol

from data_sync.block_io import BlockIOError, replace_block
from data_sync.config import DataSyncConfig
from data_sync.discovery import discover_constants_files
from data_sync.generators import gds_constants, go_constants, ts_constants
from data_sync.model.constants import ConstantSection


class ConstantsSyncError(Exception):
    """Raised when constants sync cannot complete."""


class ConstantsSource(Protocol):
    def constants(self, section_name: str) -> ConstantSection:
        ...


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
    store: ConstantsSource,
    languages: tuple[str, ...],
) -> tuple[FileUpdate, ...]:
    updates: list[FileUpdate] = []
    for discovered in discover_constants_files(config, languages):
        generator = GENERATORS.get(discovered.language)
        if generator is None:
            raise ConstantsSyncError(f"unsupported constants language: {discovered.language}")

        try:
            text = discovered.path.read_text(encoding="utf-8")
        except FileNotFoundError as exc:
            raise ConstantsSyncError(f"discovered constants file does not exist: {discovered.path}") from exc
        except OSError as exc:
            raise ConstantsSyncError(f"failed to read constants file {discovered.path}: {exc}") from exc

        updated = text
        for section_name in discovered.sections:
            section = store.constants(section_name)
            generated = generator(section.name, section.values)
            try:
                updated = replace_block(updated, section_name, generated)
            except BlockIOError as exc:
                raise ConstantsSyncError(f"{discovered.path}: {exc}") from exc

        updates.append(FileUpdate(path=discovered.path, before=text, after=updated))
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
