from __future__ import annotations

from pathlib import Path
from typing import Any, Mapping

from data_sync.config import _load_toml_file
from data_sync.model.realtime_wire import (
    RealtimeWireContract,
    RealtimeWireEvent,
    RealtimeWireField,
    RealtimeWireIDCodec,
    RealtimeWireIDSelector,
    RealtimeWireIDSelectorMapping,
    RealtimeWireKeyAlias,
    RealtimeWirePacket,
    RealtimeWirePacketField,
    RealtimeWireQuantization,
    RealtimeWireRecord,
    RealtimeWireValueAlias,
    RealtimeWireValueDomain,
)


class RealtimeWireTomlError(Exception):
    """Raised when realtime wire TOML is malformed."""


def load_realtime_wire(path: Path | str) -> RealtimeWireContract:
    resolved = Path(path)
    try:
        raw = _load_toml_file(resolved)
    except Exception as exc:
        if isinstance(exc, RealtimeWireTomlError):
            raise
        raise RealtimeWireTomlError(f"failed to load realtime wire TOML {resolved}: {exc}") from exc
    return parse_realtime_wire(raw, resolved)


def parse_realtime_wire(raw: Mapping[str, Any], path: Path | None = None) -> RealtimeWireContract:
    label = str(path) if path is not None else "realtime wire TOML"
    contract = _table(raw, "wire_contract", label)
    contract_id = _string(contract, "id", "[wire_contract]")
    version = _positive_int(contract, "version", "[wire_contract]")
    readable_keys = _bool(contract, "readable_keys", "[wire_contract]")
    explicit_metadata = _bool(contract, "explicit_metadata", "[wire_contract]")
    unknown_key_passthrough = _optional_bool(contract, "unknown_key_passthrough", False, "[wire_contract]")
    unknown_event_passthrough = _optional_bool(contract, "unknown_event_passthrough", False, "[wire_contract]")
    return RealtimeWireContract(
        id=contract_id,
        version=version,
        readable_keys=readable_keys,
        explicit_metadata=explicit_metadata,
        unknown_key_passthrough=unknown_key_passthrough,
        unknown_event_passthrough=unknown_event_passthrough,
        key_aliases=tuple(_key_alias(item, index) for index, item in enumerate(_array(raw, "key_aliases", label))),
        value_domains=tuple(
            _value_domain(item, index) for index, item in enumerate(_array(raw, "value_domains", label))
        ),
        id_codecs=tuple(_id_codec(item, index) for index, item in enumerate(_array(raw, "id_codecs", label))),
        id_selectors=tuple(
            _id_selector(item, index) for index, item in enumerate(_array(raw, "id_selectors", label))
        ),
        quantizations=tuple(
            _quantization(item, index) for index, item in enumerate(_array(raw, "quantizations", label))
        ),
        packets=tuple(_packet(item, index) for index, item in enumerate(_array(raw, "packets", label))),
        records=tuple(_record(item, index) for index, item in enumerate(_array(raw, "records", label))),
        packet_fields=tuple(
            _packet_field(item, index) for index, item in enumerate(_array(raw, "packet_fields", label))
        ),
        events=tuple(_event(item, index) for index, item in enumerate(_array(raw, "events", label))),
    )


def _key_alias(raw: Any, index: int) -> RealtimeWireKeyAlias:
    table = _mapping(raw, f"key_aliases[{index}]")
    return RealtimeWireKeyAlias(
        domain=_string(table, "domain", f"key_aliases[{index}]"),
        readable=_string(table, "readable", f"key_aliases[{index}]"),
        compact=_string(table, "compact", f"key_aliases[{index}]"),
    )


def _value_domain(raw: Any, index: int) -> RealtimeWireValueDomain:
    table = _mapping(raw, f"value_domains[{index}]")
    entries = _array(table, "entries", f"value_domains[{index}]")
    return RealtimeWireValueDomain(
        id=_string(table, "id", f"value_domains[{index}]"),
        entries=tuple(
            RealtimeWireValueAlias(
                readable=_string(_mapping(item, f"value_domains[{index}].entries[{entry_index}]"), "readable", f"value_domains[{index}].entries[{entry_index}]"),
                compact=_string(_mapping(item, f"value_domains[{index}].entries[{entry_index}]"), "compact", f"value_domains[{index}].entries[{entry_index}]"),
            )
            for entry_index, item in enumerate(entries)
        ),
    )


def _id_codec(raw: Any, index: int) -> RealtimeWireIDCodec:
    table = _mapping(raw, f"id_codecs[{index}]")
    return RealtimeWireIDCodec(
        id=_string(table, "id", f"id_codecs[{index}]"),
        prefix=_string(table, "prefix", f"id_codecs[{index}]"),
        tag=_optional_string(table, "tag", f"id_codecs[{index}]"),
        numeric_suffix=_optional_bool(table, "numeric_suffix", True, f"id_codecs[{index}]"),
        preserve_malformed=_optional_bool(table, "preserve_malformed", True, f"id_codecs[{index}]"),
    )


def _id_selector(raw: Any, index: int) -> RealtimeWireIDSelector:
    label = f"id_selectors[{index}]"
    table = _mapping(raw, label)
    mappings = _array(table, "mappings", label)
    return RealtimeWireIDSelector(
        id=_string(table, "id", label),
        field=_string(table, "field", label),
        fallback_tagged=_optional_bool(table, "fallback_tagged", False, label),
        mappings=tuple(
            RealtimeWireIDSelectorMapping(
                value=_string(_mapping(item, f"{label}.mappings[{mapping_index}]"), "value", f"{label}.mappings[{mapping_index}]"),
                codec_id=_string(_mapping(item, f"{label}.mappings[{mapping_index}]"), "codec_id", f"{label}.mappings[{mapping_index}]"),
            )
            for mapping_index, item in enumerate(mappings)
        ),
    )


def _quantization(raw: Any, index: int) -> RealtimeWireQuantization:
    label = f"quantizations[{index}]"
    table = _mapping(raw, label)
    return RealtimeWireQuantization(
        path=_string(table, "path", label),
        policy=_string(table, "policy", label),
    )


def _packet(raw: Any, index: int) -> RealtimeWirePacket:
    table = _mapping(raw, f"packets[{index}]")
    return RealtimeWirePacket(
        id=_string(table, "id", f"packets[{index}]"),
        compact=_string(table, "compact", f"packets[{index}]"),
        lane=_string(table, "lane", f"packets[{index}]"),
        snapshot_kind=_optional_string(table, "snapshot_kind", f"packets[{index}]"),
        runtime=_optional_bool(table, "runtime", False, f"packets[{index}]"),
        infer_lane=_optional_bool(table, "infer_lane", False, f"packets[{index}]"),
        infer_snapshot_kind=_optional_bool(table, "infer_snapshot_kind", False, f"packets[{index}]"),
        infer_snapshot_id=_optional_bool(table, "infer_snapshot_id", False, f"packets[{index}]"),
        infer_baseline_id=_optional_bool(table, "infer_baseline_id", False, f"packets[{index}]"),
        use_baseline_sequence=_optional_bool(table, "use_baseline_sequence", False, f"packets[{index}]"),
        omit_single_chunk_metadata=_optional_bool(table, "omit_single_chunk_metadata", False, f"packets[{index}]"),
    )


def _record(raw: Any, index: int) -> RealtimeWireRecord:
    table = _mapping(raw, f"records[{index}]")
    fields = _array(table, "fields", f"records[{index}]")
    return RealtimeWireRecord(
        id=_string(table, "id", f"records[{index}]"),
        source_struct=_optional_string(table, "source_struct", f"records[{index}]"),
        encoding=_string(table, "encoding", f"records[{index}]"),
        identity_field=_optional_string(table, "identity_field", f"records[{index}]"),
        sparse_placeholder=_optional_string(table, "sparse_placeholder", f"records[{index}]"),
        sparse_trailing=_optional_bool(table, "sparse_trailing", False, f"records[{index}]"),
        preserve_unknown_fields=_optional_bool(table, "preserve_unknown_fields", False, f"records[{index}]"),
        fields=tuple(_field(item, index, field_index) for field_index, item in enumerate(fields)),
    )


def _field(raw: Any, record_index: int, field_index: int) -> RealtimeWireField:
    label = f"records[{record_index}].fields[{field_index}]"
    table = _mapping(raw, label)
    return RealtimeWireField(
        name=_string(table, "name", label),
        json=_string(table, "json", label),
        quantization=_optional_string(table, "quantization", label),
        id_codec=_optional_string(table, "id_codec", label),
        id_codec_by=_optional_string(table, "id_codec_by", label),
        value_domain=_optional_string(table, "value_domain", label),
    )


def _packet_field(raw: Any, index: int) -> RealtimeWirePacketField:
    label = f"packet_fields[{index}]"
    table = _mapping(raw, label)
    return RealtimeWirePacketField(
        packet_id=_string(table, "packet_id", label),
        readable_field=_string(table, "readable_field", label),
        record_id=_string(table, "record_id", label),
        decode_record_ids=tuple(_optional_string_array(table, "decode_record_ids", label)),
    )


def _event(raw: Any, index: int) -> RealtimeWireEvent:
    label = f"events[{index}]"
    table = _mapping(raw, label)
    return RealtimeWireEvent(
        readable=_string(table, "readable", label),
        compact=_string(table, "compact", label),
        record_id=_string(table, "record_id", label),
    )


def _array(table: Mapping[str, Any], key: str, label: str) -> list[Any]:
    value = table.get(key, [])
    if not isinstance(value, list):
        raise RealtimeWireTomlError(f"{label}.{key} must be an array")
    return value


def _table(table: Mapping[str, Any], key: str, label: str) -> Mapping[str, Any]:
    value = table.get(key)
    if not isinstance(value, Mapping):
        raise RealtimeWireTomlError(f"{label} requires [{key}]")
    return value


def _mapping(value: Any, label: str) -> Mapping[str, Any]:
    if not isinstance(value, Mapping):
        raise RealtimeWireTomlError(f"{label} must be a table")
    return value


def _string(table: Mapping[str, Any], key: str, label: str) -> str:
    value = table.get(key)
    if not isinstance(value, str) or not value:
        raise RealtimeWireTomlError(f"{label}.{key} must be a non-empty string")
    return value


def _optional_string(table: Mapping[str, Any], key: str, label: str) -> str | None:
    value = table.get(key)
    if value is None:
        return None
    if not isinstance(value, str) or not value:
        raise RealtimeWireTomlError(f"{label}.{key} must be a non-empty string")
    return value


def _optional_string_array(table: Mapping[str, Any], key: str, label: str) -> list[str]:
    value = table.get(key, [])
    if not isinstance(value, list) or any(not isinstance(item, str) or not item for item in value):
        raise RealtimeWireTomlError(f"{label}.{key} must be an array of non-empty strings")
    return value


def _positive_int(table: Mapping[str, Any], key: str, label: str) -> int:
    value = table.get(key)
    if not isinstance(value, int) or isinstance(value, bool) or value < 1:
        raise RealtimeWireTomlError(f"{label}.{key} must be a positive integer")
    return value


def _bool(table: Mapping[str, Any], key: str, label: str) -> bool:
    value = table.get(key)
    if not isinstance(value, bool):
        raise RealtimeWireTomlError(f"{label}.{key} must be a boolean")
    return value


def _optional_bool(table: Mapping[str, Any], key: str, default: bool, label: str) -> bool:
    value = table.get(key, default)
    if not isinstance(value, bool):
        raise RealtimeWireTomlError(f"{label}.{key} must be a boolean")
    return value
