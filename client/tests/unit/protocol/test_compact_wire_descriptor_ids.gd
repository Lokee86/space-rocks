extends GutTest

const DescriptorIndex := preload("res://scripts/protocol/realtime/compact_wire_descriptor_index.gd")
const DescriptorIDs := preload("res://scripts/protocol/realtime/compact_wire_descriptor_ids.gd")

func test_descriptor_index_resolves_packets_records_bindings_and_events() -> void:
	assert_eq(DescriptorIndex.packet_by_readable_id("world_full").get("compact"), "wf")
	assert_eq(DescriptorIndex.packet_by_compact_id("wf").get("id"), "world_full")
	assert_eq(DescriptorIndex.record_by_id("player_session").get("encoding"), "fixed_tuple")
	assert_eq(DescriptorIndex.binding_record_ids("session_delta", "players"), ["player_session"])
	assert_eq(DescriptorIndex.event_by_readable_type("bullet_blast").get("compact"), "bb")
	assert_eq(DescriptorIndex.event_by_compact_type("bb").get("readable"), "bullet_blast")

func test_descriptor_index_resolves_generated_key_and_value_aliases() -> void:
	assert_eq(DescriptorIndex.readable_key("pl"), "players")
	assert_eq(DescriptorIndex.compact_key("players"), "pl")
	assert_eq(DescriptorIndex.readable_packet_type("wf"), "world_full")
	assert_eq(DescriptorIndex.compact_packet_type("world_full"), "wf")
	assert_eq(DescriptorIndex.readable_event_type("bb"), "bullet_blast")
	assert_eq(DescriptorIndex.compact_event_type("bullet_blast"), "bb")
	assert_eq(DescriptorIndex.readable_lane("al"), "asteroids.lifecycle")
	assert_eq(DescriptorIndex.compact_lane("asteroids.lifecycle"), "al")
	assert_eq(DescriptorIndex.readable_snapshot_kind("f"), "full")
	assert_eq(DescriptorIndex.compact_snapshot_kind("full"), "f")

func test_descriptor_ids_expand_direct_generated_codecs() -> void:
	assert_eq(DescriptorIDs.expand_codec("player_id", 7), "player-7")
	assert_eq(DescriptorIDs.expand_codec("bullet_id", 8.0), "bullet-8")
	assert_eq(DescriptorIDs.expand_codec("asteroid_id", "9"), "asteroid-9")
	assert_eq(DescriptorIDs.expand_codec("presentation_event_id", ["pe", 10]), "presentation-event-10")
	assert_eq(DescriptorIDs.expand_codec("event_batch_id", ["eb", "11"]), "event-batch-11")
	assert_eq(DescriptorIDs.expand_codec("bullet_id", "bullet-12"), "bullet-12")

func test_descriptor_ids_preserve_malformed_and_unsupported_values() -> void:
	assert_eq(DescriptorIDs.expand_codec("player_id", 1.5), 1.5)
	assert_eq(DescriptorIDs.expand_codec("bullet_id", null), null)
	assert_eq(DescriptorIDs.expand_codec("asteroid_id", "asteroid-invalid"), "asteroid-invalid")
	assert_eq(DescriptorIDs.expand_codec("missing_codec", 4), 4)
	assert_eq(DescriptorIDs.expand_codec("bullet_id", ["wrong", 3]), ["wrong", 3])

func test_descriptor_ids_expand_selector_and_tagged_fallback() -> void:
	assert_eq(DescriptorIDs.expand_selector("source_type", "player", 2), "player-2")
	assert_eq(DescriptorIDs.expand_selector("source_type", "projectile", 3), "bullet-3")
	assert_eq(DescriptorIDs.expand_selector("source_type", "unknown", ["a", 4]), "asteroid-4")
	assert_eq(DescriptorIDs.expand_tagged(["pe", 5]), "presentation-event-5")
	assert_eq(DescriptorIDs.expand_tagged(["unknown", 5]), ["unknown", 5])
