from __future__ import annotations

import copy
import json
from dataclasses import replace
from pathlib import Path

import pytest

from data_sync.config import load_config
from data_sync.generators.realtime_wire_docs import generate_realtime_wire_docs
from data_sync.generators.realtime_wire_gds import generate_realtime_wire_gds
from data_sync.generators.realtime_wire_go import generate_realtime_wire_go
from data_sync.generators.realtime_wire_json import generate_realtime_wire_json
from data_sync.model.realtime_wire import RealtimeWireContract
from data_sync.packet_toml import load_packet_schema_files
from data_sync.realtime_wire_toml import RealtimeWireTomlError, load_realtime_wire, parse_realtime_wire
from data_sync.realtime_wire_sync import RealtimeWireSyncError, plan_realtime_wire_updates
from data_sync.realtime_wire_validate import validate_realtime_wire, validate_realtime_wire_contract


REPO = Path(__file__).resolve().parents[3]
DATA_SYNC = Path(__file__).resolve().parents[1]
CONFIG = DATA_SYNC / "config.toml"


def _repository_contract() -> RealtimeWireContract:
    return load_realtime_wire(REPO / "shared" / "packets" / "realtime_wire.toml")


def _repository_schema():
    config = load_config(CONFIG)
    return load_packet_schema_files(config.sot_paths("packets"))


def _raw_contract() -> dict:
    return copy.deepcopy(_repository_contract().as_dict())


def _assert_invalid(raw: dict, needle: str, schema=None) -> None:
    contract = parse_realtime_wire(raw)
    errors = validate_realtime_wire_contract(contract, schema)
    assert any(needle in error for error in errors), errors


def test_repository_schema_validates_against_merged_packet_schema() -> None:
    config = load_config(CONFIG)
    packet_schema = load_packet_schema_files(config.sot_paths("packets"))
    contract = validate_realtime_wire(config.sot_path("realtime_wire"), packet_schema)
    assert contract.id == "realtime"
    assert {event.readable for event in contract.events} == {
        "bullet_blast",
        "ship_death",
        "damage_applied",
        "damage_over_time_started",
        "damage_over_time_tick",
        "radial_effect_started",
        "pickup_collected",
        "pickup_effect_applied",
        "pickup_expired",
        "pickup_dropped",
    }


def test_parser_requires_wire_contract_table(tmp_path: Path) -> None:
    path = tmp_path / "wire.toml"
    path.write_text("id = 'wrong'\n", encoding="utf-8")
    with pytest.raises(RealtimeWireTomlError, match="requires \\[wire_contract\\]"):
        load_realtime_wire(path)


def test_parsing_and_as_dict_are_deterministic() -> None:
    first = _repository_contract()
    second = parse_realtime_wire(first.as_dict())
    assert first.as_dict() == second.as_dict()
    assert first.packet("world_delta").use_baseline_sequence
    assert first.record("event_damage_applied").fields[1].json == "event_id"
    assert first.unknown_key_passthrough
    assert first.unknown_event_passthrough
    assert first.record("player_session_update").preserve_unknown_fields
    assert all(selector.fallback_tagged for selector in first.id_selectors)


def test_lifecycle_bindings_use_physical_map_and_string_list_records() -> None:
    contract = _repository_contract()
    bindings = {
        (binding.packet_id, binding.readable_field): binding.record_id
        for binding in contract.packet_fields
    }
    assert bindings[("asteroids_lifecycle", "asteroid_creates")] == "asteroid_lifecycle_create"
    assert bindings[("asteroids_lifecycle", "asteroid_deletes")] == "asteroid_lifecycle_delete_ids"
    assert bindings[("bullets_lifecycle", "bullet_creates")] == "bullet_lifecycle_create"
    assert bindings[("bullets_lifecycle", "bullet_deletes")] == "bullet_lifecycle_delete_ids"
    alternatives = {
        (binding.packet_id, binding.readable_field): binding.decode_record_ids
        for binding in contract.packet_fields
    }
    assert alternatives[("asteroids_lifecycle", "asteroid_creates")] == ("asteroid_full",)
    assert alternatives[("asteroids_lifecycle", "asteroid_deletes")] == ("asteroid_ids",)
    assert alternatives[("bullets_lifecycle", "bullet_creates")] == ("bullet_full",)
    assert alternatives[("bullets_lifecycle", "bullet_deletes")] == ("bullet_ids",)

    for record_id in (
        "asteroid_lifecycle_create",
        "bullet_lifecycle_create",
    ):
        record = contract.record(record_id)
        assert record.encoding == "map"
        assert all(field.id_codec is None and field.id_codec_by is None for field in record.fields)

    for record_id in ("asteroid_lifecycle_delete_ids", "bullet_lifecycle_delete_ids"):
        record = contract.record(record_id)
        assert record.encoding == "scalar_list"
        assert record.fields[0].id_codec is None



def test_generator_surfaces_are_complete_and_deterministic() -> None:
    contract = _repository_contract()
    json_output = generate_realtime_wire_json(contract)
    assert json_output == generate_realtime_wire_json(contract)
    assert json.loads(json_output) == contract.as_dict()
    assert '"packet_fields"' in json_output
    assert '"quantizations"' in json_output

    go_output = generate_realtime_wire_go(contract)
    assert 'package realtimewire' in go_output
    assert "RealtimeWirePacketFieldBinding" in go_output
    assert "RealtimeWireUnknownEventPassthrough" in go_output
    assert "RealtimeWireQuantizations" in go_output
    assert "RealtimeWireEventsByCompact" in go_output
    assert 'PacketID: "asteroids_lifecycle", ReadableField: "asteroid_creates", RecordIDs: []string{"asteroid_lifecycle_create"}' in go_output

    gds_output = generate_realtime_wire_gds(contract)
    assert "KEY_READABLE_BY_COMPACT" in gds_output
    assert "QUANTIZATION_POLICY_BY_PATH" in gds_output
    assert '"session.players.spawn_x":"position"' in gds_output
    assert '"overlay.primary_cooldown_remaining":"seconds"' in gds_output
    assert '"world.ships.rotation":"float_generic"' in gds_output
    assert "PACKET_FIELD_RECORD_IDS" in gds_output
    assert '"asteroids_lifecycle.asteroid_creates":["asteroid_lifecycle_create","asteroid_full"]' in gds_output
    assert "EVENTS_BY_READABLE" in gds_output

    docs_output = generate_realtime_wire_docs(contract)
    assert docs_output.startswith(
        "<!-- Code generated by data-sync; DO NOT EDIT. -->\n"
        "# Realtime Wire Contract\n\n"
        "Parent index: [Generated](./!INDEX.md)\n\n"
    )
    assert docs_output.endswith("\n")
    assert "## Quantization" in docs_output
    assert "## Packet bindings" in docs_output
    assert "Primary encode record" in docs_output
    assert "Decode alternatives" in docs_output
    assert "## Events" in docs_output


def test_plan_rejects_unsupported_output_kind() -> None:
    config = load_config(CONFIG)
    with pytest.raises(RealtimeWireSyncError, match="unsupported realtime-wire output kind"):
        plan_realtime_wire_updates(config, ("ts",))


def test_unknown_quantization_policy_is_rejected() -> None:
    raw = _raw_contract()
    raw["quantizations"][0]["policy"] = "not_a_policy"
    _assert_invalid(raw, "unknown policy")


def test_duplicate_quantization_path_is_rejected() -> None:
    raw = _raw_contract()
    raw["quantizations"].append(copy.deepcopy(raw["quantizations"][0]))
    _assert_invalid(raw, "duplicate quantization path")


def test_duplicate_binding_is_rejected() -> None:
    raw = _raw_contract()
    raw["packet_fields"].append(copy.deepcopy(raw["packet_fields"][0]))
    _assert_invalid(raw, "duplicate packet field binding")


def test_decode_alternatives_are_validated() -> None:
    raw = _raw_contract()
    binding = next(item for item in raw["packet_fields"] if item["readable_field"] == "asteroid_creates")
    binding["decode_record_ids"] = [binding["record_id"]]
    _assert_invalid(raw, "repeat primary record")

    raw = _raw_contract()
    binding = next(item for item in raw["packet_fields"] if item["readable_field"] == "asteroid_creates")
    binding["decode_record_ids"] = ["missing_record", "missing_record"]
    _assert_invalid(raw, "duplicate decode record IDs")
    _assert_invalid(raw, "unknown decode record")


def test_scalar_record_shape_is_validated() -> None:
    raw = _raw_contract()
    record = next(record for record in raw["records"] if record["id"] == "event_batch_id")
    record["fields"].append({"name": "extra", "json": "extra"})
    _assert_invalid(raw, "exactly one field")


def test_id_transform_requires_id_named_field() -> None:
    raw = _raw_contract()
    field = next(field for field in raw["records"][0]["fields"] if field["json"] == "x")
    field["id_codec"] = "player_id"
    _assert_invalid(raw, "requires an id field name")


def test_duplicate_aliases_are_rejected() -> None:
    raw = _raw_contract()
    raw["key_aliases"].append(copy.deepcopy(raw["key_aliases"][0]))
    _assert_invalid(raw, "duplicate readable key")


def test_unknown_packet_is_rejected() -> None:
    raw = _raw_contract()
    raw["packets"].append({"id": "not_a_packet", "compact": "np", "lane": "world"})
    _assert_invalid(raw, "unknown packet type", _repository_schema())


def test_unknown_struct_is_rejected() -> None:
    raw = _raw_contract()
    raw["records"][0]["source_struct"] = "NoSuchState"
    _assert_invalid(raw, "unknown source struct", _repository_schema())


def test_unknown_field_is_rejected() -> None:
    raw = _raw_contract()
    raw["records"][0]["fields"].append({"name": "bad", "json": "not_a_field"})
    _assert_invalid(raw, "unknown source field", _repository_schema())


def test_sparse_record_requires_identity() -> None:
    raw = _raw_contract()
    next(record for record in raw["records"] if record["id"] == "ship_update")["identity_field"] = None
    _assert_invalid(raw, "requires identity_field")


def test_invalid_selector_is_rejected() -> None:
    raw = _raw_contract()
    field = next(field for field in raw["records"][0]["fields"] if field["json"] == "target_id")
    field["id_codec_by"] = "missing_selector"
    _assert_invalid(raw, "unknown id selector")


def test_incompatible_transforms_are_rejected() -> None:
    raw = _raw_contract()
    ship_type = next(field for field in raw["records"][0]["fields"] if field["json"] == "ship_type")
    ship_type["quantization"] = "position"
    _assert_invalid(raw, "quantization requires a numeric source field", _repository_schema())

    raw = _raw_contract()
    numeric_field = next(field for field in raw["records"][0]["fields"] if field["json"] == "x")
    numeric_field["id_codec"] = "player_id"
    _assert_invalid(raw, "ID transform requires a string source field", _repository_schema())

    raw = _raw_contract()
    target_id = next(field for field in raw["records"][0]["fields"] if field["json"] == "target_id")
    target_id["id_codec"] = "player_id"
    _assert_invalid(raw, "cannot set both id_codec and id_codec_by")


def test_float_packet_schema_type_is_numeric_for_quantization() -> None:
    schema = _repository_schema()
    record = _repository_contract().records[0]
    source = schema.struct(record.source_struct)
    fields = tuple(
        replace(field, type="float") if field.json == "x" else field
        for field in source.fields
    )
    schema = replace(
        schema,
        structs=tuple(replace(item, fields=fields) if item.id == source.id else item for item in schema.structs),
    )
    assert not validate_realtime_wire_contract(_repository_contract(), schema)


def test_unknown_binding_and_event_are_rejected() -> None:
    raw = _raw_contract()
    raw["packet_fields"][0]["packet_id"] = "missing_packet"
    raw["events"][0]["record_id"] = "missing_record"
    errors = validate_realtime_wire_contract(parse_realtime_wire(raw))
    assert any("unknown packet" in error for error in errors)
    assert any("unknown record" in error for error in errors)
