extends GutTest

const CompactLanePacket := preload("res://scripts/protocol/realtime/compact_lane_packet.gd")
const RealtimeRouter := preload("res://scripts/protocol/realtime/realtime_router.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")


func test_expand_packet_converts_compact_world_delta_keys_and_values() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"l": "w",
		"q": 7,
		"b": "player-1",
		"sid": "player-1",
		"ms": 123,
		"k": "d",
		"su": [
			{"i": "ship-1", "x": 10, "y": 20, "r": 3142, "th": false},
		],
	})

	assert_eq(expanded["type"], "world_delta")
	assert_eq(expanded["lane"], "world")
	assert_eq(expanded["sequence"], 7)
	assert_eq(expanded["baseline_id"], "player-1")
	assert_eq(expanded["snapshot_id"], "player-1")
	assert_eq(expanded["server_sent_msec"], 123)
	assert_eq(expanded["snapshot_kind"], "delta")
	assert_eq(expanded["ship_updates"][0]["id"], "ship-1")
	assert_eq(expanded["ship_updates"][0]["rotation"], 3142)
	assert_false(expanded["ship_updates"][0]["thrusting"])


func test_expand_packet_derives_minimal_compact_world_delta_runtime_metadata() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"q": 7,
		"bq": 5,
		"su": [
			{"i": "ship-1"},
		],
	})

	assert_eq(expanded["type"], "world_delta")
	assert_eq(expanded["lane"], "world")
	assert_eq(expanded["snapshot_kind"], "delta")
	assert_eq(expanded["snapshot_id"], "world-snapshot-7")
	assert_eq(expanded["baseline_sequence"], 5)
	assert_eq(expanded["baseline_id"], "world-baseline-5")
	assert_eq(expanded["chunk_index"], 0)
	assert_eq(expanded["chunk_count"], 1)
	assert_true(expanded["is_final_chunk"])


func test_expand_packet_derives_minimal_compact_world_full_runtime_metadata() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wf",
		"q": 3,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})

	assert_eq(expanded["type"], "world_full")
	assert_eq(expanded["lane"], "world")
	assert_eq(expanded["snapshot_kind"], "full")
	assert_eq(expanded["snapshot_id"], "world-baseline-3")
	assert_eq(expanded["baseline_id"], "world-baseline-3")
	assert_eq(expanded["chunk_index"], 0)
	assert_eq(expanded["chunk_count"], 1)
	assert_true(expanded["is_final_chunk"])


func test_expand_packet_derives_chunk_finality_when_missing() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wf",
		"q": 4,
		"ci": 1,
		"cc": 3,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})

	assert_eq(expanded["chunk_index"], 1)
	assert_eq(expanded["chunk_count"], 3)
	assert_false(expanded["is_final_chunk"])

	var final_chunk := CompactLanePacket.expand_packet({
		"t": "wf",
		"q": 4,
		"ci": 2,
		"cc": 3,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})

	assert_true(final_chunk["is_final_chunk"])


func test_legacy_long_key_packets_still_route_to_existing_appliers() -> void:
	var router := RealtimeRouter.new()
	var packet := {
		"type": "world_full",
		"lane": "world",
		"sequence": 7,
		"baseline_id": "baseline-1",
		"snapshot_id": "snapshot-1",
		"server_sent_msec": 123,
		"snapshot_kind": "full",
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
		"is_final_chunk": true,
	}

	router.route_lane_packet(packet)
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))
