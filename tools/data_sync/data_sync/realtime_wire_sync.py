from __future__ import annotations

from collections.abc import Callable

from data_sync.config import ConfigError, DataSyncConfig
from data_sync.constants_sync import FileUpdate
from data_sync.generators.realtime_wire_docs import generate_realtime_wire_docs
from data_sync.generators.realtime_wire_gds import generate_realtime_wire_gds
from data_sync.generators.realtime_wire_go import generate_realtime_wire_go
from data_sync.generators.realtime_wire_json import generate_realtime_wire_json
from data_sync.model.realtime_wire import RealtimeWireContract
from data_sync.packet_toml import PacketTomlError, load_packet_schema_files
from data_sync.realtime_wire_toml import RealtimeWireTomlError
from data_sync.realtime_wire_validate import RealtimeWireValidationError, validate_realtime_wire


class RealtimeWireSyncError(Exception):
    pass


_RENDERERS: dict[str, Callable[[RealtimeWireContract], str]] = {
    "go": generate_realtime_wire_go,
    "gds": generate_realtime_wire_gds,
    "json": generate_realtime_wire_json,
    "docs": generate_realtime_wire_docs,
}
_OUTPUT_ORDER = {name: index for index, name in enumerate(("go", "gds", "json", "docs"))}


def plan_realtime_wire_updates(config: DataSyncConfig, languages: tuple[str, ...]) -> tuple[FileUpdate, ...]:
    try:
        packet_schema = load_packet_schema_files(config.sot_paths("packets"))
        contract = validate_realtime_wire(config.sot_path("realtime_wire"), packet_schema)
    except (ConfigError, PacketTomlError, RealtimeWireTomlError, RealtimeWireValidationError) as exc:
        raise RealtimeWireSyncError(str(exc)) from exc
    updates: list[FileUpdate] = []
    for language in sorted(languages, key=lambda item: _OUTPUT_ORDER.get(item, len(_OUTPUT_ORDER))):
        renderer = _RENDERERS.get(language)
        if renderer is None:
            raise RealtimeWireSyncError(f"unsupported realtime-wire output kind: {language}")
        try:
            targets = config.targets_for("realtime_wire", language)
        except ConfigError as exc:
            raise RealtimeWireSyncError(str(exc)) from exc
        for target in sorted(targets, key=lambda item: item.label):
            if not target.enabled:
                raise RealtimeWireSyncError(f"[{target.label}] is disabled in config")
            generated = renderer(contract)
            for path in sorted(target.files):
                before = path.read_text(encoding="utf-8") if path.exists() else ""
                updates.append(FileUpdate(path=path, before=before, after=generated))
    return tuple(updates)
