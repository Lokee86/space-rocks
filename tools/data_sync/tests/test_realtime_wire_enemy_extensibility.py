from __future__ import annotations

import json
from pathlib import Path

from data_sync.generators.realtime_wire_docs import generate_realtime_wire_docs
from data_sync.generators.realtime_wire_gds import generate_realtime_wire_gds
from data_sync.generators.realtime_wire_go import generate_realtime_wire_go
from data_sync.generators.realtime_wire_json import generate_realtime_wire_json
from data_sync.packet_toml import load_packet_schema_files
from data_sync.realtime_wire_toml import load_realtime_wire
from data_sync.realtime_wire_validate import validate_realtime_wire


REPO = Path(__file__).resolve().parents[3]
FIXTURES = Path(__file__).resolve().parent / "fixtures"


def test_enemy_contract_is_an_isolated_extensibility_fixture() -> None:
    packet_schema = load_packet_schema_files([FIXTURES / "realtime_wire_enemy_packets.toml"])
    contract = validate_realtime_wire(FIXTURES / "realtime_wire_enemy.toml", packet_schema)

    assert {packet.id for packet in packet_schema.packet_types} == {
        "enemies_full",
        "enemy_delta",
        "enemies_lifecycle",
        "event_batch",
        "enemy_destroyed",
    }
    packets = {packet.id: packet for packet in contract.packets}
    assert {packet.id for packet in contract.packets} == {
        "enemies_full",
        "enemy_delta",
        "enemies_lifecycle",
        "event_batch",
    }
    assert packets["enemies_full"].lane == "enemies"
    assert packets["enemy_delta"].lane == "enemies"
    assert packets["enemies_lifecycle"].lane == "enemies.lifecycle"
    assert packets["event_batch"].lane == "event"
    assert packets["enemies_full"].runtime is True
    assert packets["enemy_delta"].runtime is True
    assert packets["enemies_lifecycle"].runtime is False
    assert packets["event_batch"].runtime is False

    records = {record.id: record for record in contract.records}
    assert records["enemy_full"].encoding == "fixed_tuple"
    assert records["enemy_update"].encoding == "sparse_positional_tuple"
    assert records["enemy_lifecycle_create"].encoding == "map"
    assert records["enemy_ids"].encoding == "scalar_id_list"
    assert records["enemy_lifecycle_delete_ids"].encoding == "scalar_list"
    assert records["event_union"].encoding == "discriminated_event_tuple"
    assert records["event_enemy_destroyed"].encoding == "fixed_tuple"

    bindings = {(item.packet_id, item.readable_field): item for item in contract.packet_fields}
    assert bindings[("enemies_lifecycle", "enemy_creates")].record_id == "enemy_lifecycle_create"
    assert bindings[("enemies_lifecycle", "enemy_creates")].decode_record_ids == ("enemy_full",)
    assert bindings[("enemies_lifecycle", "enemy_deletes")].record_id == "enemy_lifecycle_delete_ids"
    assert bindings[("enemies_lifecycle", "enemy_deletes")].decode_record_ids == ("enemy_ids",)

    selector = next(item for item in contract.id_selectors if item.id == "target_kind")
    assert selector.fallback_tagged is True
    assert {mapping.value: mapping.codec_id for mapping in selector.mappings} == {
        "enemy": "enemy_id",
        "player": "player_id",
    }
    assert {item.path: item.policy for item in contract.quantizations} == {
        "enemies.full.x": "position",
        "enemies.full.y": "position",
        "enemies.full.rotation": "float_generic",
        "enemies.delta.x": "position",
        "enemies.delta.y": "position",
        "enemies.delta.rotation": "float_generic",
        "event.enemy_destroyed.x": "position",
        "event.enemy_destroyed.y": "position",
    }
    assert [(item.readable, item.compact, item.record_id) for item in contract.events] == [
        ("enemy_destroyed", "exd", "event_enemy_destroyed")
    ]

    go_output = generate_realtime_wire_go(contract)
    gds_output = generate_realtime_wire_gds(contract)
    json_output = generate_realtime_wire_json(contract)
    docs_output = generate_realtime_wire_docs(contract)

    assert '"ef":"enemies_full"' in gds_output
    assert '"ed":"enemy_delta"' in gds_output
    assert '"el":"enemies_lifecycle"' in gds_output
    assert '"eb":"event_batch"' in gds_output
    assert '"prefix":"enemy-"' in gds_output
    assert '"tag":"e"' in gds_output
    assert '"enemies_lifecycle.enemy_creates":["enemy_lifecycle_create","enemy_full"]' in gds_output
    assert '"enemies_lifecycle.enemy_deletes":["enemy_lifecycle_delete_ids","enemy_ids"]' in gds_output
    assert '"enemies.full.x":"position"' in gds_output
    assert '"event.enemy_destroyed.y":"position"' in gds_output
    assert '"record_id":"event_enemy_destroyed"' in gds_output
    assert '"event_enemy_destroyed"' in go_output

    assert json.loads(json_output) == contract.as_dict()
    assert "`enemy_full`" in docs_output
    assert "`enemy_update`" in docs_output
    assert "`enemies_lifecycle` | `enemy_creates` | `enemy_lifecycle_create` | `enemy_full`" in docs_output
    assert "`enemies_lifecycle` | `enemy_deletes` | `enemy_lifecycle_delete_ids` | `enemy_ids`" in docs_output
    assert "`enemy_destroyed` | `exd` | `event_enemy_destroyed`" in docs_output


def test_generic_runtime_files_have_no_enemy_identifier() -> None:
    paths = (
        REPO / "services/game-server/internal/protocol/realtime/compact_wire_descriptor.go",
        REPO / "client/scripts/protocol/realtime/compact_wire_descriptor_records.gd",
        REPO / "client/scripts/protocol/realtime/compact_wire_descriptor_decoder.gd",
    )
    for path in paths:
        assert "enemy" not in path.read_text(encoding="utf-8").lower()
