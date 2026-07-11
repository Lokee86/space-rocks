from __future__ import annotations

from pathlib import Path

from data_sync.model.packets import PacketSchema
from data_sync.model.realtime_wire import ENCODING_MODES, RealtimeWireContract
from data_sync.realtime_wire_toml import RealtimeWireTomlError, load_realtime_wire

NUMERIC_TYPES = frozenset({"int", "uint32", "float", "float32", "float64"})
STRING_TYPES = frozenset({"string"})
REQUIRED_ID_MODES = frozenset(
    {"sparse_positional_tuple", "sparse_key_value_tuple", "scalar_id_list", "scalar_list"}
)
SYNTHETIC_ID_FIELDS = frozenset({"id", "event_id", "player_id", "source_id", "target_id"})
QUANTIZATION_POLICIES = frozenset(
    {
        "float_generic",
        "ratio_0_1",
        "percent_0_100",
        "seconds",
        "signed_seconds",
        "angle_turn",
        "position",
        "velocity",
        "angular_velocity",
    }
)


class RealtimeWireValidationError(Exception):
    def __init__(self, errors: list[str]) -> None:
        self.errors = errors
        super().__init__("\n".join(errors))


def validate_realtime_wire(path: Path | str, packet_schema: PacketSchema | None = None) -> RealtimeWireContract:
    contract = load_realtime_wire(path)
    errors = validate_realtime_wire_contract(contract, packet_schema)
    if errors:
        raise RealtimeWireValidationError(errors)
    return contract


def validate_realtime_wire_contract(
    contract: RealtimeWireContract, packet_schema: PacketSchema | None = None
) -> list[str]:
    errors: list[str] = []
    _validate_aliases(contract, errors)
    _validate_quantizations(contract, errors)
    _validate_codecs_and_selectors(contract, errors)
    _validate_records(contract, packet_schema, errors)
    _validate_packets(contract, packet_schema, errors)
    _validate_bindings(contract, errors)
    _validate_events(contract, errors)
    return errors


def _validate_aliases(contract: RealtimeWireContract, errors: list[str]) -> None:
    for domain in {alias.domain for alias in contract.key_aliases}:
        aliases = [alias for alias in contract.key_aliases if alias.domain == domain]
        _unique((alias.readable for alias in aliases), f"readable key in domain {domain}", errors)
        _unique((alias.compact for alias in aliases), f"compact key in domain {domain}", errors)
    _unique((domain.id for domain in contract.value_domains), "value domain", errors)
    for domain in contract.value_domains:
        _unique((entry.readable for entry in domain.entries), f"readable value in domain {domain.id}", errors)
        _unique((entry.compact for entry in domain.entries), f"compact value in domain {domain.id}", errors)


def _validate_quantizations(contract: RealtimeWireContract, errors: list[str]) -> None:
    _unique((item.path for item in contract.quantizations), "quantization path", errors)
    for item in contract.quantizations:
        if item.policy not in QUANTIZATION_POLICIES:
            errors.append(f"quantization {item.path} references unknown policy: {item.policy}")


def _validate_codecs_and_selectors(contract: RealtimeWireContract, errors: list[str]) -> None:
    _unique((codec.id for codec in contract.id_codecs), "id codec", errors)
    _unique((selector.id for selector in contract.id_selectors), "id selector", errors)
    for selector in contract.id_selectors:
        _unique((mapping.value for mapping in selector.mappings), f"selector value in {selector.id}", errors)
        for mapping in selector.mappings:
            codec = next((codec for codec in contract.id_codecs if codec.id == mapping.codec_id), None)
            if codec is None:
                errors.append(f"id selector {selector.id} references unknown codec: {mapping.codec_id}")
            elif selector.fallback_tagged and not codec.tag:
                errors.append(f"id selector {selector.id} fallback codec must be tagged: {mapping.codec_id}")


def _validate_records(contract: RealtimeWireContract, packet_schema: PacketSchema | None, errors: list[str]) -> None:
    codecs = {codec.id for codec in contract.id_codecs}
    selectors = {selector.id: selector for selector in contract.id_selectors}
    domains = {domain.id for domain in contract.value_domains}
    structs = {}
    if packet_schema is not None:
        structs = {struct.id: {field.json: field for field in struct.fields} for struct in packet_schema.structs}
    _unique((record.id for record in contract.records), "record descriptor", errors)
    for record in contract.records:
        _unique((field.name for field in record.fields), f"name in record {record.id}", errors)
        _unique((field.json for field in record.fields), f"field in record {record.id}", errors)
        fields = {field.json: field for field in record.fields}
        if record.encoding not in ENCODING_MODES:
            errors.append(f"record {record.id} has unsupported encoding mode: {record.encoding}")
        if record.encoding in {"scalar_id", "scalar_list"} and len(record.fields) != 1:
            errors.append(f"record {record.id} must have exactly one field for {record.encoding}")
        if record.encoding == "scalar_id" and record.identity_field:
            errors.append(f"record {record.id} must not declare identity_field for scalar_id")
        if record.encoding in REQUIRED_ID_MODES and not record.identity_field:
            errors.append(f"record {record.id} requires identity_field for {record.encoding}")
        if record.identity_field and record.identity_field not in fields:
            if not (record.encoding == "discriminated_event_tuple" and record.identity_field == "event_id"):
                errors.append(f"record {record.id} identity_field is not a declared field: {record.identity_field}")
        if record.encoding == "fixed_tuple" and record.identity_field and record.fields:
            if record.fields[0].json != record.identity_field:
                errors.append(f"record {record.id} identity_field must be first in fixed_tuple")
        source_fields = None
        if record.source_struct:
            source_fields = structs.get(record.source_struct)
            if packet_schema is not None and source_fields is None:
                errors.append(f"record {record.id} references unknown source struct: {record.source_struct}")
        for field in record.fields:
            source_field = source_fields.get(field.json) if source_fields else None
            if source_fields is not None and source_field is None:
                if not record.preserve_unknown_fields and not (record.id.startswith("event_") and field.json == "event_id"):
                    errors.append(f"record {record.id} references unknown source field: {field.json}")
            if field.quantization:
                if field.quantization not in QUANTIZATION_POLICIES:
                    errors.append(f"record {record.id}.{field.json} references unknown quantization policy: {field.quantization}")
                if source_field and source_field.type not in NUMERIC_TYPES:
                    errors.append(f"record {record.id}.{field.json} quantization requires a numeric source field")
            if field.value_domain and field.value_domain not in domains:
                errors.append(f"record {record.id}.{field.json} references unknown value domain: {field.value_domain}")
            if field.id_codec and field.id_codec_by:
                errors.append(f"record {record.id}.{field.json} cannot set both id_codec and id_codec_by")
            if field.id_codec and field.id_codec not in codecs:
                errors.append(f"record {record.id}.{field.json} references unknown id codec: {field.id_codec}")
            if field.id_codec_by:
                selector = selectors.get(field.id_codec_by)
                if selector is None:
                    errors.append(f"record {record.id}.{field.json} references unknown id selector: {field.id_codec_by}")
                elif selector.field not in fields:
                    errors.append(f"record {record.id}.{field.json} selector field is not in the same record: {selector.field}")
            if field.id_codec or field.id_codec_by:
                if field.json != "id" and not field.json.endswith("_id"):
                    errors.append(f"record {record.id}.{field.json} ID transform requires an id field name")
                if source_field and source_field.type not in STRING_TYPES and field.json not in SYNTHETIC_ID_FIELDS:
                    errors.append(f"record {record.id}.{field.json} ID transform requires a string source field")


def _validate_packets(contract: RealtimeWireContract, packet_schema: PacketSchema | None, errors: list[str]) -> None:
    packet_ids = {packet.id for packet in packet_schema.packet_types} if packet_schema else set()
    aliases = {entry.readable: entry.compact for entry in _domain(contract, "packet_type")}
    lanes = {entry.readable for entry in _domain(contract, "lane")}
    snapshots = {entry.readable for entry in _domain(contract, "snapshot_kind")}
    _unique((packet.id for packet in contract.packets), "packet descriptor", errors)
    for packet in contract.packets:
        if packet_schema is not None and packet.id not in packet_ids:
            errors.append(f"packet descriptor references unknown packet type: {packet.id}")
        if aliases.get(packet.id) != packet.compact:
            errors.append(f"packet {packet.id} has no matching packet_type value alias: {packet.compact}")
        if packet.lane not in lanes:
            errors.append(f"packet {packet.id} references unknown lane value: {packet.lane}")
        if packet.snapshot_kind and packet.snapshot_kind not in snapshots:
            errors.append(f"packet {packet.id} references unknown snapshot_kind value: {packet.snapshot_kind}")


def _validate_bindings(contract: RealtimeWireContract, errors: list[str]) -> None:
    packets = {packet.id for packet in contract.packets}
    records = {record.id for record in contract.records}
    record_by_id = {record.id: record for record in contract.records}
    seen: dict[tuple[str, str], list[str]] = {}
    for binding in contract.packet_fields:
        seen.setdefault((binding.packet_id, binding.readable_field), []).append(binding.record_id)
    for (packet_id, readable_field), record_ids in seen.items():
        if len(record_ids) > 1 and not (
            packet_id == "event_batch"
            and readable_field == "events"
            and all(
                record_by_id.get(record_id) is not None
                and record_by_id[record_id].encoding == "discriminated_event_tuple"
                for record_id in record_ids
            )
        ):
            errors.append(f"duplicate packet field binding: {(packet_id, readable_field)}")
    for binding in contract.packet_fields:
        if binding.packet_id not in packets:
            errors.append(f"packet field binding references unknown packet: {binding.packet_id}")
        if binding.record_id not in records:
            errors.append(f"packet field binding references unknown record: {binding.record_id}")
        if len(set(binding.decode_record_ids)) != len(binding.decode_record_ids):
            errors.append(f"packet field binding has duplicate decode record IDs: {(binding.packet_id, binding.readable_field)}")
        if binding.record_id in binding.decode_record_ids:
            errors.append(f"packet field binding decode alternatives repeat primary record: {binding.record_id}")
        for decode_record_id in binding.decode_record_ids:
            if decode_record_id not in records:
                errors.append(f"packet field binding references unknown decode record: {decode_record_id}")
        if not binding.readable_field:
            errors.append("packet field binding readable_field must be non-empty")


def _validate_events(contract: RealtimeWireContract, errors: list[str]) -> None:
    records = {record.id for record in contract.records}
    event_values = {entry.readable: entry.compact for entry in _domain(contract, "event_type")}
    _unique((event.readable for event in contract.events), "readable event type", errors)
    _unique((event.compact for event in contract.events), "compact event type", errors)
    for event in contract.events:
        if event.record_id not in records:
            errors.append(f"event {event.readable} references unknown record: {event.record_id}")
        if event_values.get(event.readable) != event.compact:
            errors.append(f"event {event.readable} has no matching event_type value alias: {event.compact}")


def _domain(contract: RealtimeWireContract, domain_id: str):
    for domain in contract.value_domains:
        if domain.id == domain_id:
            return domain.entries
    return ()


def _unique(values, label: str, errors: list[str]) -> None:
    seen: set[str] = set()
    for value in values:
        if value in seen:
            errors.append(f"duplicate {label}: {value}")
        seen.add(value)
