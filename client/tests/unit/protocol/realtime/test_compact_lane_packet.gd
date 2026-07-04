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


func test_expand_packet_converts_compact_event_batch_keys_and_values() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "eb",
		"q": 11,
		"ms": 123,
		"bid": "event-batch-11",
		"ev": [
			{"ei": "event-1", "t": "bb", "x": 10, "y": 20},
			{"ei": "event-2", "t": "dmg", "srct": "projectile", "src": "bullet-1", "tid": "player-1", "tt": "player", "dt": "explosive", "dc": "impact", "ba": 20, "ma": 17, "ah": 12, "abs": 5, "rh": 88, "rs": 0, "fx": "blast", "amt": 17, "x": 123, "y": 456},
			{"ei": "event-3", "t": "shd", "pid": "player-1", "lv": 2, "rd": 3500},
			{"ei": "event-4", "t": "dots", "srct": "asteroid", "src": "hazard-1", "fx": "radioactive", "amt": 2},
			{"ei": "event-5", "t": "dott", "srct": "asteroid", "src": "hazard-1", "fx": "radioactive", "amt": 3},
			{"ei": "event-6", "t": "rfx", "srct": "pickup", "src": "pickup-1", "fx": "pulse"},
			{"ei": "event-7", "t": "pcol", "pid": "player-1", "pkid": "pickup-1", "pkt": "shield", "x": 125, "y": 345},
			{"ei": "event-8", "t": "pea", "pid": "player-1", "pkid": "pickup-1", "pkt": "shield", "fx": "repair", "amt": 4, "lva": 3},
			{"ei": "event-9", "t": "pexp", "pkid": "pickup-1", "pkt": "shield", "x": 222, "y": 333},
			{"ei": "event-10", "t": "pdr", "pkid": "pickup-1", "pkt": "shield", "srct": "ship", "src": "ship-1", "tbl": "table-1", "x": 444, "y": 555},
		],
	})

	assert_eq(expanded["type"], "event_batch")
	assert_eq(expanded["sequence"], 11)
	assert_eq(expanded["server_sent_msec"], 123)
	assert_eq(expanded["batch_id"], "event-batch-11")
	assert_eq(expanded["events"][0]["event_id"], "event-1")
	assert_eq(expanded["events"][0]["type"], "bullet_blast")
	assert_eq(expanded["events"][1]["event_id"], "event-2")
	assert_eq(expanded["events"][1]["type"], "damage_applied")
	assert_eq(expanded["events"][1]["source_type"], "projectile")
	assert_eq(expanded["events"][1]["source_id"], "bullet-1")
	assert_eq(expanded["events"][1]["target_id"], "player-1")
	assert_eq(expanded["events"][1]["target_type"], "player")
	assert_eq(expanded["events"][1]["damage_type"], "explosive")
	assert_eq(expanded["events"][1]["damage_cause"], "impact")
	assert_eq(expanded["events"][1]["base_amount"], 20)
	assert_eq(expanded["events"][1]["modified_amount"], 17)
	assert_eq(expanded["events"][1]["applied_to_health"], 12)
	assert_eq(expanded["events"][1]["absorbed_by_shield"], 5)
	assert_eq(expanded["events"][1]["remaining_health"], 88)
	assert_eq(expanded["events"][1]["remaining_shield"], 0)
	assert_eq(expanded["events"][1]["effect_type"], "blast")
	assert_eq(expanded["events"][1]["amount"], 17)
	assert_eq(expanded["events"][2]["type"], "ship_death")
	assert_eq(expanded["events"][2]["player_id"], "player-1")
	assert_eq(expanded["events"][2]["lives"], 2)
	assert_eq(expanded["events"][2]["respawn_delay"], 3500)
	assert_eq(expanded["events"][3]["type"], "damage_over_time_started")
	assert_eq(expanded["events"][4]["type"], "damage_over_time_tick")
	assert_eq(expanded["events"][5]["type"], "radial_effect_started")
	assert_eq(expanded["events"][6]["type"], "pickup_collected")
	assert_eq(expanded["events"][7]["type"], "pickup_effect_applied")
	assert_eq(expanded["events"][8]["type"], "pickup_expired")
	assert_eq(expanded["events"][9]["type"], "pickup_dropped")


func test_expand_packet_keeps_event_batch_events_readable_for_application() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "eb",
		"q": 42,
		"ms": 9001,
		"bid": "batch-42",
		"ev": [
			{"ei": "event-1", "t": "bb", "x": 10, "y": 20},
			{"ei": "event-2", "t": "shd", "pid": "player-1", "lv": 2, "rd": 3500, "x": 30, "y": 40},
		],
	})

	assert_eq(expanded["type"], "event_batch")
	assert_eq(expanded["batch_id"], "batch-42")
	assert_eq(expanded["events"][0]["event_id"], "event-1")
	assert_eq(expanded["events"][0]["type"], "bullet_blast")
	assert_eq(expanded["events"][1]["event_id"], "event-2")
	assert_eq(expanded["events"][1]["type"], "ship_death")
	assert_eq(expanded["events"][1]["player_id"], "player-1")
	assert_eq(expanded["events"][1]["lives"], 2)
	assert_eq(expanded["events"][1]["respawn_delay"], 3500)


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

