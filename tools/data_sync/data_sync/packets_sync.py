"""Packet push/diff/check support."""

from __future__ import annotations

from typing import Callable

from data_sync.block_io import BlockIOError, replace_block
from data_sync.config import DataSyncConfig, DomainLanguageConfig
from data_sync.constants_sync import FileUpdate
from data_sync.generators import gds_packets, go_packets, ts_packets
from data_sync.model.packets import PacketDefinition
from data_sync.toml_store import TomlStore


class PacketsSyncError(Exception):
    """Raised when packet sync cannot complete."""


Generator = Callable[[str, tuple[PacketDefinition, ...]], str]


GENERATORS: dict[str, Generator] = {
    "go": go_packets.generate_packets,
    "gds": gds_packets.generate_packets,
    "ts": ts_packets.generate_packets,
}


def plan_packets_updates(
    config: DataSyncConfig,
    store: TomlStore,
    languages: tuple[str, ...],
) -> tuple[FileUpdate, ...]:
    updates: list[FileUpdate] = []
    for language in languages:
        target = config.target("packets", language)
        updates.extend(_plan_target_updates(store, target))
    return tuple(updates)


def _plan_target_updates(store: TomlStore, target: DomainLanguageConfig) -> tuple[FileUpdate, ...]:
    generator = GENERATORS.get(target.language)
    if generator is None:
        raise PacketsSyncError(f"unsupported packets language: {target.language}")
    if not target.enabled:
        raise PacketsSyncError(f"[packets.{target.language}] is disabled in config")

    packets = store.packets()
    updates: list[FileUpdate] = []
    for path in target.files:
        try:
            text = path.read_text(encoding="utf-8")
        except FileNotFoundError as exc:
            raise PacketsSyncError(f"configured packets file does not exist: {path}") from exc
        except OSError as exc:
            raise PacketsSyncError(f"failed to read packets file {path}: {exc}") from exc

        updated = text
        for section_name in target.sections:
            if section_name != "packets":
                raise PacketsSyncError(f"unsupported packets section: {section_name}")
            generated = generator(section_name, packets)
            try:
                updated = replace_block(updated, section_name, generated)
            except BlockIOError as exc:
                raise PacketsSyncError(f"{path}: {exc}") from exc

        updates.append(FileUpdate(path=path, before=text, after=updated))
    return tuple(updates)
