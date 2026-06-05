"""Drop table sync support."""

from __future__ import annotations

from pathlib import Path

from data_sync.config import DataSyncConfig, DomainLanguageConfig
from data_sync.constants_sync import FileUpdate
from data_sync.generators.go_drop_tables import generate_drop_tables
from data_sync.drop_tables_toml import DropTablesTomlError, load_drop_tables
from data_sync.model.drop_tables import DropTablesModel


class DropTablesSyncError(Exception):
    """Raised when drop table sync cannot complete."""


def plan_drop_tables_updates(config: DataSyncConfig) -> tuple[FileUpdate, ...]:
    target = config.target("drop_tables", "go")
    return _plan_drop_tables_target(config, target)


def _plan_drop_tables_target(config: DataSyncConfig, target: DomainLanguageConfig) -> tuple[FileUpdate, ...]:
    if not target.enabled:
        raise DropTablesSyncError("[drop_tables.go] is disabled in config")
    if not target.files:
        raise DropTablesSyncError("[drop_tables.go].files must not be empty")

    model = _load_model(config.sot_paths("drop_tables"))
    rendered = generate_drop_tables(model)

    updates: list[FileUpdate] = []
    for path in target.files:
        before = _read_output(path)
        updates.append(FileUpdate(path=path, before=before, after=rendered))
    return tuple(updates)


def _load_model(paths: tuple[Path, ...]) -> DropTablesModel:
    try:
        return load_drop_tables(list(paths))
    except DropTablesTomlError as exc:
        raise DropTablesSyncError(str(exc)) from exc


def _read_output(path: Path) -> str:
    try:
        return path.read_text(encoding="utf-8")
    except FileNotFoundError as exc:
        raise DropTablesSyncError(f"configured drop tables file does not exist: {path}") from exc
    except OSError as exc:
        raise DropTablesSyncError(f"failed to read drop tables output {path}: {exc}") from exc
