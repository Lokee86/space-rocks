"""Constants push/diff/check support."""

from __future__ import annotations

import difflib
from dataclasses import dataclass
from pathlib import Path
from typing import Callable, Protocol

from data_sync.block_io import BlockIOError, replace_block
from data_sync.config import ConfigError, DataSyncConfig
from data_sync.discovery import discover_constants_files, discover_constants_files_from_paths
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
    for language in languages:
        configured_targets = _configured_constants_targets(config, language)
        if configured_targets:
            discovered_files = discover_constants_files_from_paths(
                tuple(dict.fromkeys(path for target in configured_targets for path in target.files)),
                (language,),
            )
        else:
            discovered_files = discover_constants_files(config, (language,))

        for discovered in discovered_files:
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
            allowed_sections = (
                {section for target in configured_targets for section in target.sections}
                if configured_targets
                else None
            )
            for section_name in discovered.sections:
                if allowed_sections is not None and section_name not in allowed_sections:
                    continue
                section = store.constants(section_name)
                generated = generator(section.name, section.values)
                try:
                    updated = replace_block(updated, section_name, generated)
                except BlockIOError as exc:
                    raise ConstantsSyncError(f"{discovered.path}: {exc}") from exc

            updates.append(FileUpdate(path=discovered.path, before=text, after=updated))
    return tuple(updates)


def _configured_constants_targets(config: DataSyncConfig, language: str):
    try:
        return config.targets_for("constants", language)
    except ConfigError:
        return ()


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
