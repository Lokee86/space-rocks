"""Check command support."""

from __future__ import annotations

from data_sync.config import DataSyncConfig
from data_sync.constants_sync import all_synced, plan_constants_updates
from data_sync.toml_store import TomlStore


def check_constants(config: DataSyncConfig, store: TomlStore, languages: tuple[str, ...]) -> bool:
    return all_synced(plan_constants_updates(config, store, languages))
