extends GutTest

const DescriptorIndex := preload("res://scripts/protocol/realtime/compact_wire_descriptor_index.gd")
const DescriptorRecords := preload("res://scripts/protocol/realtime/compact_wire_descriptor_records.gd")

func test_expand_full_fixed_tuple() -> void:
	var record := DescriptorIndex.record_by_id("player_session")
	var value := [7, "scout", 12, 3, 1, null, null, null, null, 100, 200]
	var expanded: Dictionary = DescriptorRecords.expand_record(value, record)
	assert_eq(expanded["id"], "player-7")
	assert_eq(expanded["ship_type"], "scout")
	assert_eq(expanded["spawn_y"], 200)
	assert_true(expanded.has("primary_weapon_id"))

func test_expand_sparse_positional_preserves_middle_null() -> void:
	var record := DescriptorIndex.record_by_id("ship_update")
	var expanded: Dictionary = DescriptorRecords.expand_record([2, null, 12], record)
	assert_eq(expanded["id"], "player-2")
	assert_eq(expanded["y"], 12)
	assert_false(expanded.has("x"))

func test_expand_sparse_key_value_update_preserves_unknown_key() -> void:
	var record := DescriptorIndex.record_by_id("player_session_update")
	var expanded: Dictionary = DescriptorRecords.expand_record([2, "sco", 9, "future_key", "kept"], record)
	assert_eq(expanded["id"], "player-2")
	assert_eq(expanded["score"], 9)
	assert_eq(expanded["future_key"], "kept")

func test_expand_map_lifecycle_record() -> void:
	var record := DescriptorIndex.record_by_id("asteroid_lifecycle_create")
	var expanded: Dictionary = DescriptorRecords.expand_record({"i": "asteroid-4", "x": 10, "y": 11, "h": 20}, record)
	assert_eq(expanded["id"], "asteroid-4")
	assert_eq(expanded["x"], 10)
	assert_eq(expanded["health"], 20)

func test_lifecycle_binding_accepts_map_and_decode_tuple_alternatives() -> void:
	var record_ids: Array = ["asteroid_lifecycle_create", "asteroid_full"]
	var map_value: Array = [{"i": "asteroid-4", "x": 10, "y": 11, "h": 20}]
	var tuple_value: Array = [[4, 10, 11, 2, 1, 1.0, 3]]
	var expanded_map: Array = DescriptorRecords.expand_bound(map_value, record_ids)
	var expanded_tuple: Array = DescriptorRecords.expand_bound(tuple_value, record_ids)
	assert_eq(expanded_map[0]["id"], "asteroid-4")
	assert_eq(expanded_tuple[0]["id"], "asteroid-4")
	assert_eq(expanded_tuple[0]["health"], 1)

func test_expand_scalar_direct_and_list_ids() -> void:
	var direct: Variant = DescriptorRecords.expand_bound(3, ["event_batch_id"])
	var list: Array = DescriptorRecords.expand_bound([4, "player-5"], ["player_ids"])
	assert_eq(direct, "event-batch-3")
	assert_eq(list, ["player-4", "player-5"])

func test_lifecycle_delete_alternative_decodes_numeric_and_readable_ids() -> void:
	var record_ids: Array = ["asteroid_lifecycle_delete_ids", "asteroid_ids"]
	var numeric: Array = DescriptorRecords.expand_bound([4, 5], record_ids)
	var readable: Array = DescriptorRecords.expand_bound(["asteroid-4", "asteroid-5"], record_ids)
	assert_eq(numeric, ["asteroid-4", "asteroid-5"])
	assert_eq(readable, ["asteroid-4", "asteroid-5"])

func test_expand_all_generated_event_tuple_types() -> void:
	var event_types := ["bb", "shd", "dmg", "dots", "dott", "rfx", "pcol", "pea", "pexp", "pdr"]
	for compact_type in event_types:
		var event := DescriptorIndex.event_by_compact_type(compact_type)
		var record := DescriptorIndex.record_by_id(str(event.get("record_id", "")))
		var tuple: Array = []
		for field in record.get("fields", []):
			var name := str(field.get("json", ""))
			match name:
				"type":
					tuple.append(compact_type)
				"event_id":
					tuple.append(["pe", 1])
				"source_type", "target_kind":
					tuple.append("player")
				"source_id", "target_id", "player_id":
					tuple.append(2)
				_:
					tuple.append(null)
		var expanded: Array = DescriptorRecords.expand_bound([tuple], ["event_union"])
		assert_eq(expanded[0]["type"], str(event.get("readable", "")))

func test_expand_unknown_event_passes_through_generically() -> void:
	var expanded: Array = DescriptorRecords.expand_bound([["future_event", 3]], ["event_union"])
	assert_eq(expanded, [["future_event", 3]])

func test_expand_selector_tagged_fallback() -> void:
	var record := DescriptorIndex.record_by_id("ship_full")
	var expanded: Dictionary = DescriptorRecords.expand_record([1, "scout", 1, 2, 3, 4, 5, true, "unknown", ["a", 4]], record)
	assert_eq(expanded["id"], "player-1")
	assert_eq(expanded["target_id"], "asteroid-4")

func test_expand_generic_accepts_readable_and_compact_maps() -> void:
	var expanded: Dictionary = DescriptorRecords.expand_generic({"t": "wf", "pl": [{"i": 1}], "future": ["kept"]})
	assert_eq(expanded["type"], "world_full")
	assert_true(expanded.has("players"))
	assert_eq(expanded["players"][0]["id"], 1)
	assert_eq(expanded["future"], ["kept"])
