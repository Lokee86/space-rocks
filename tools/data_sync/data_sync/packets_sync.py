"""Packet diff/check support for rich TOML packet schemas."""

from __future__ import annotations

from pathlib import Path
from typing import Callable

from data_sync.config import DataSyncConfig, DomainLanguageConfig
from data_sync.constants_sync import FileUpdate
from data_sync.generators.rich_gds_packets import (
    RichGdsPacketGenerationError,
    render_gdscript_output,
)
from data_sync.generators.rich_go_packets import RichGoPacketGenerationError, render_go_output
from data_sync.model.packets import PacketOutput, PacketSchema


class PacketsSyncError(Exception):
    """Raised when packet sync cannot complete."""


Renderer = Callable[[PacketSchema, PacketOutput], str]


RENDERERS: dict[str, Renderer] = {
    "go": render_go_output,
    "gds": render_gdscript_output,
}


def plan_packets_updates(
    config: DataSyncConfig,
    schema: PacketSchema,
    languages: tuple[str, ...],
) -> tuple[FileUpdate, ...]:
    updates: list[FileUpdate] = []
    for language in languages:
        target = config.target("packets", language)
        updates.extend(_plan_target_updates(config, schema, target))
    return tuple(updates)


def _plan_target_updates(
    config: DataSyncConfig,
    schema: PacketSchema,
    target: DomainLanguageConfig,
) -> tuple[FileUpdate, ...]:
    if not target.enabled:
        raise PacketsSyncError(f"[packets.{target.language}] is disabled in config")
    renderer = RENDERERS.get(target.language)
    if renderer is None:
        raise PacketsSyncError(f"unsupported packets language: {target.language}")

    updates: list[FileUpdate] = []
    for path in target.files:
        try:
            text = path.read_text(encoding="utf-8")
        except FileNotFoundError as exc:
            raise PacketsSyncError(f"configured packets file does not exist: {path}") from exc
        except OSError as exc:
            raise PacketsSyncError(f"failed to read packets file {path}: {exc}") from exc

        output_path = _relative_output_path(config, path)
        try:
            output = schema.output_for_path(output_path)
            generated = renderer(schema, output)
        except KeyError as exc:
            raise PacketsSyncError(f"packet TOML has no output for configured file: {output_path}") from exc
        except (RichGoPacketGenerationError, RichGdsPacketGenerationError) as exc:
            raise PacketsSyncError(str(exc)) from exc

        updates.append(FileUpdate(path=path, before=text, after=generated))
    return tuple(updates)


def _relative_output_path(config: DataSyncConfig, path: Path) -> str:
    try:
        return path.resolve().relative_to(config.root).as_posix()
    except ValueError:
        return path.as_posix()
