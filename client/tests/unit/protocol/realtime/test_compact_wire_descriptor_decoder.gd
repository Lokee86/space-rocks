extends GutTest

const DescriptorDecoder := preload("res://scripts/protocol/realtime/compact_wire_descriptor_decoder.gd")
const DescriptorIndex := preload("res://scripts/protocol/realtime/compact_wire_descriptor_index.gd")
const DescriptorRecords := preload("res://scripts/protocol/realtime/compact_wire_descriptor_records.gd")

func test_generated_scalar_binding_expands_global_batch_id() -> void:
	assert_eq(DescriptorIndex.scalar_binding_record_ids("batch_id"), ["event_batch_id"])
	var generic: Dictionary = DescriptorRecords.expand_generic({"bid": 7})
	assert_eq(generic["batch_id"], "event-batch-7")
	var expanded: Dictionary = DescriptorDecoder.expand_packet({"t": "eb", "bid": 7})
	assert_eq(expanded["batch_id"], "event-batch-7")

func test_runtime_full_infers_descriptor_metadata_and_chunk_defaults() -> void:
	var expanded: Dictionary = DescriptorDecoder.expand_packet({"t": "wf", "q": 7})
	assert_eq(expanded["type"], "world_full")
	assert_eq(expanded["lane"], "world")
	assert_eq(expanded["snapshot_kind"], "full")
	assert_eq(expanded["snapshot_id"], "world-baseline-7")
	assert_eq(expanded["baseline_id"], "world-baseline-7")
	assert_eq(expanded["chunk_index"], 0)
	assert_eq(expanded["chunk_count"], 1)
	assert_true(expanded["is_final_chunk"])

func test_runtime_delta_uses_baseline_sequence_for_baseline_id() -> void:
	var expanded: Dictionary = DescriptorDecoder.expand_packet({"t": "wd", "q": 9, "bq": 4})
	assert_eq(expanded["type"], "world_delta")
	assert_eq(expanded["lane"], "world")
	assert_eq(expanded["snapshot_kind"], "delta")
	assert_eq(expanded["snapshot_id"], "world-snapshot-9")
	assert_eq(expanded["baseline_id"], "world-baseline-4")

func test_explicit_runtime_metadata_is_preserved() -> void:
	var expanded: Dictionary = DescriptorDecoder.expand_packet({
		"t": "wf",
		"q": 7,
		"l": "custom",
		"k": "d",
		"sid": "explicit-snapshot",
		"b": "explicit-baseline",
		"ci": 2,
		"cc": 4,
		"fc": false,
	})
	assert_eq(expanded["lane"], "custom")
	assert_eq(expanded["snapshot_kind"], "delta")
	assert_eq(expanded["snapshot_id"], "explicit-snapshot")
	assert_eq(expanded["baseline_id"], "explicit-baseline")
	assert_eq(expanded["chunk_index"], 2)
	assert_eq(expanded["chunk_count"], 4)
	assert_false(expanded["is_final_chunk"])

func test_non_runtime_lifecycle_packet_does_not_infer_metadata() -> void:
	var expanded: Dictionary = DescriptorDecoder.expand_packet({"t": "al", "q": 1})
	assert_eq(expanded["type"], "asteroids_lifecycle")
	assert_eq(expanded["sequence"], 1)
	assert_false(expanded.has("lane"))
	assert_false(expanded.has("snapshot_kind"))
	assert_false(expanded.has("chunk_index"))

func test_unknown_packet_uses_generic_key_and_value_expansion() -> void:
	var expanded: Dictionary = DescriptorDecoder.expand_packet({
		"t": "future_packet",
		"pl": [{"i": 1}],
		"future": {"k": "d"},
	})
	assert_eq(expanded["type"], "future_packet")
	assert_eq(expanded["players"][0]["id"], 1)
	assert_eq(expanded["future"]["snapshot_kind"], "delta")
	assert_false(expanded.has("lane"))
