"""Diff command support."""

from __future__ import annotations

from data_sync.config import DataSyncConfig
from data_sync.constants_sync import plan_constants_updates, unified_diff
from data_sync.toml_store import TomlStore


def diff_constants(config: DataSyncConfig, store: TomlStore, languages: tuple[str, ...]) -> str:
    return unified_diff(plan_constants_updates(config, store, languages))
