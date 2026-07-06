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


func test_expand_compact_presentation_event_id_rebuilds_numeric_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_presentation_event_id(1), "presentation-event-1")
	assert_eq(CompactLanePacket.expand_compact_presentation_event_id(1.0), "presentation-event-1")
	assert_eq(CompactLanePacket.expand_compact_presentation_event_id("1"), "presentation-event-1")
	assert_eq(CompactLanePacket.expand_compact_presentation_event_id("presentation-event-1"), "presentation-event-1")
	assert_eq(CompactLanePacket.expand_compact_presentation_event_id(1.5), 1.5)
	assert_null(CompactLanePacket.expand_compact_presentation_event_id(null))


func test_expand_compact_event_batch_id_rebuilds_numeric_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_event_batch_id(11), "event-batch-11")
	assert_eq(CompactLanePacket.expand_compact_event_batch_id(11.0), "event-batch-11")
	assert_eq(CompactLanePacket.expand_compact_event_batch_id("11"), "event-batch-11")
	assert_eq(CompactLanePacket.expand_compact_event_batch_id("event-batch-11"), "event-batch-11")
	assert_eq(CompactLanePacket.expand_compact_event_batch_id(11.5), 11.5)
	assert_null(CompactLanePacket.expand_compact_event_batch_id(null))


func test_expand_packet_converts_compact_event_batch_keys_and_values() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "eb",
		"q": 11,
		"ms": 123,
		"bid": 11,
		"ev": [
			["bb", 1, 10, 20],
			["shd", 2, 1, 3, 250, 30, 40],
			["dmg", 3, "projectile", 1, "blast", 17, 50, 60],
			["dots", 4, "asteroid", "hazard-1", "radioactive", 2],
			["dott", 5, "asteroid", "hazard-1", "radioactive", 3, 70, 80],
			["rfx", 6, "pickup", 1, "pulse", 90, 100],
			["pcol", 7, 1, 1, "shield", 110, 120],
			["pea", 8, 1, 1, "shield", "repair", 4, 3],
			["pexp", 9, 1, "shield", 130, 140],
			["pdr", 10, 1, "shield", "ship", 1, 1, 150, 160],
		],
	})

	assert_eq(expanded["type"], "event_batch")
	assert_eq(expanded["batch_id"], "event-batch-11")
	assert_eq(expanded["sequence"], 11)
	assert_eq(expanded["server_sent_msec"], 123)
	assert_eq(expanded["events"][0], {"event_id": "presentation-event-1", "type": "bullet_blast", "x": 10, "y": 20})
	assert_eq(expanded["events"][1], {"event_id": "presentation-event-2", "type": "ship_death", "player_id": "player-1", "lives": 3, "respawn_delay": 250, "x": 30, "y": 40})
	assert_eq(expanded["events"][2], {"event_id": "presentation-event-3", "type": "damage_applied", "source_type": "projectile", "source_id": "bullet-1", "effect_type": "blast", "amount": 17, "x": 50, "y": 60})
	assert_eq(expanded["events"][3], {"event_id": "presentation-event-4", "type": "damage_over_time_started", "source_type": "asteroid", "source_id": "hazard-1", "effect_type": "radioactive", "amount": 2})
	assert_eq(expanded["events"][4], {"event_id": "presentation-event-5", "type": "damage_over_time_tick", "source_type": "asteroid", "source_id": "hazard-1", "effect_type": "radioactive", "amount": 3, "x": 70, "y": 80})
	assert_eq(expanded["events"][5], {"event_id": "presentation-event-6", "type": "radial_effect_started", "source_type": "pickup", "source_id": "pickup-1", "effect_type": "pulse", "x": 90, "y": 100})
	assert_eq(expanded["events"][6], {"event_id": "presentation-event-7", "type": "pickup_collected", "player_id": "player-1", "pickup_id": "pickup-1", "pickup_type": "shield", "x": 110, "y": 120})
	assert_eq(expanded["events"][7], {"event_id": "presentation-event-8", "type": "pickup_effect_applied", "player_id": "player-1", "pickup_id": "pickup-1", "pickup_type": "shield", "effect_type": "repair", "amount": 4, "lives_after": 3})
	assert_eq(expanded["events"][8], {"event_id": "presentation-event-9", "type": "pickup_expired", "pickup_id": "pickup-1", "pickup_type": "shield", "x": 130, "y": 140})
	assert_eq(expanded["events"][9], {"event_id": "presentation-event-10", "type": "pickup_dropped", "pickup_id": "pickup-1", "pickup_type": "shield", "source_type": "ship", "source_id": "ship-1", "table_id": "table-1", "x": 150, "y": 160})


func test_expand_packet_keeps_compact_map_shaped_event_records_compatible() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "eb",
		"bid": 11,
		"ev": [
			{"ei": "presentation-event-1", "t": "bb", "x": 10, "y": 20},
		],
	})

	assert_eq(expanded["type"], "event_batch")
	assert_eq(expanded["batch_id"], "event-batch-11")
	assert_eq(expanded["events"][0], {"event_id": "presentation-event-1", "type": "bullet_blast", "x": 10, "y": 20})


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



func test_expand_compact_bullet_id_rebuilds_numeric_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_bullet_id(123), "bullet-123")


func test_expand_compact_bullet_id_rebuilds_float_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_bullet_id(1.0), "bullet-1")


func test_expand_compact_bullet_id_keeps_non_integer_float() -> void:
	assert_eq(CompactLanePacket.expand_compact_bullet_id(1.5), 1.5)


func test_expand_compact_bullet_id_rebuilds_string_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_bullet_id("123"), "bullet-123")


func test_expand_compact_bullet_id_keeps_full_id() -> void:
	assert_eq(CompactLanePacket.expand_compact_bullet_id("bullet-123"), "bullet-123")


func test_expand_compact_player_id_rebuilds_numeric_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_player_id(123), "player-123")


func test_expand_compact_player_id_rebuilds_float_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_player_id(1.0), "player-1")


func test_expand_compact_player_id_keeps_non_integer_float() -> void:
	assert_eq(CompactLanePacket.expand_compact_player_id(1.5), 1.5)


func test_expand_compact_player_id_rebuilds_string_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_player_id("123"), "player-123")


func test_expand_compact_player_id_keeps_full_id() -> void:
	assert_eq(CompactLanePacket.expand_compact_player_id("player-123"), "player-123")


func test_expand_compact_player_id_keeps_null_and_unsupported_values() -> void:
	assert_null(CompactLanePacket.expand_compact_player_id(null))
	assert_eq(CompactLanePacket.expand_compact_player_id({"id": 123}), {"id": 123})


func test_expand_packet_converts_compact_world_full_ship_tuples() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wf",
		"ships": [[1, "v_wing", 10, 20, 30, 100, 50, true, "player", "player-2"]],
	})

	assert_eq(expanded["ships"][0]["id"], "player-1")
	assert_eq(expanded["ships"][0]["ship_type"], "v_wing")
	assert_eq(expanded["ships"][0]["x"], 10)
	assert_eq(expanded["ships"][0]["y"], 20)
	assert_eq(expanded["ships"][0]["rotation"], 30)
	assert_eq(expanded["ships"][0]["health"], 100)
	assert_eq(expanded["ships"][0]["shields"], 50)
	assert_eq(expanded["ships"][0]["thrusting"], true)
	assert_eq(expanded["ships"][0]["target_kind"], "player")
	assert_eq(expanded["ships"][0]["target_id"], "player-2")


func test_expand_packet_converts_compact_world_delta_ship_create_tuples() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"sc": [[1, "v_wing", 10, 20, 30, 100, 50, true, "player", "player-2"]],
	})

	assert_eq(expanded["ship_creates"][0]["id"], "player-1")
	assert_eq(expanded["ship_creates"][0]["ship_type"], "v_wing")
	assert_eq(expanded["ship_creates"][0]["x"], 10)
	assert_eq(expanded["ship_creates"][0]["y"], 20)
	assert_eq(expanded["ship_creates"][0]["rotation"], 30)
	assert_eq(expanded["ship_creates"][0]["health"], 100)
	assert_eq(expanded["ship_creates"][0]["shields"], 50)
	assert_eq(expanded["ship_creates"][0]["thrusting"], true)
	assert_eq(expanded["ship_creates"][0]["target_kind"], "player")
	assert_eq(expanded["ship_creates"][0]["target_id"], "player-2")


func test_expand_packet_converts_compact_world_delta_ship_update_tuple_xy_rotation_thrusting() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"su": [[1, 10, 20, 30, true]],
	})

	assert_eq(expanded["ship_updates"][0]["id"], "player-1")
	assert_eq(expanded["ship_updates"][0]["x"], 10)
	assert_eq(expanded["ship_updates"][0]["y"], 20)
	assert_eq(expanded["ship_updates"][0]["rotation"], 30)
	assert_eq(expanded["ship_updates"][0]["thrusting"], true)


func test_expand_packet_converts_compact_world_delta_ship_update_tuple_x_only() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"su": [[1, 10]],
	})

	assert_eq(expanded["ship_updates"][0]["id"], "player-1")
	assert_eq(expanded["ship_updates"][0]["x"], 10)
	assert_false(expanded["ship_updates"][0].has("y"))
	assert_false(expanded["ship_updates"][0].has("rotation"))
	assert_false(expanded["ship_updates"][0].has("thrusting"))


func test_expand_packet_converts_compact_world_delta_ship_update_tuple_y_only() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"su": [[1, null, 20]],
	})

	assert_eq(expanded["ship_updates"][0]["id"], "player-1")
	assert_false(expanded["ship_updates"][0].has("x"))
	assert_eq(expanded["ship_updates"][0]["y"], 20)
	assert_false(expanded["ship_updates"][0].has("rotation"))
	assert_false(expanded["ship_updates"][0].has("thrusting"))


func test_expand_packet_converts_compact_world_delta_ship_update_tuple_rotation_only() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"su": [[1, null, null, 30]],
	})

	assert_eq(expanded["ship_updates"][0]["id"], "player-1")
	assert_false(expanded["ship_updates"][0].has("x"))
	assert_false(expanded["ship_updates"][0].has("y"))
	assert_eq(expanded["ship_updates"][0]["rotation"], 30)
	assert_false(expanded["ship_updates"][0].has("thrusting"))


func test_expand_packet_converts_compact_world_delta_ship_update_tuple_thrusting_only_false() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"su": [[1, null, null, null, false]],
	})

	assert_eq(expanded["ship_updates"][0]["id"], "player-1")
	assert_false(expanded["ship_updates"][0].has("x"))
	assert_false(expanded["ship_updates"][0].has("y"))
	assert_false(expanded["ship_updates"][0].has("rotation"))
	assert_eq(expanded["ship_updates"][0]["thrusting"], false)


func test_expand_packet_converts_compact_world_delta_ship_update_tuple_zero_values() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"su": [[1, 0, 0, 0, false]],
	})

	assert_eq(expanded["ship_updates"][0]["id"], "player-1")
	assert_eq(expanded["ship_updates"][0]["x"], 0)
	assert_eq(expanded["ship_updates"][0]["y"], 0)
	assert_eq(expanded["ship_updates"][0]["rotation"], 0)
	assert_eq(expanded["ship_updates"][0]["thrusting"], false)


func test_expand_packet_converts_compact_world_delta_ship_delete_ids() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"sx": [1, "2", "player-3"],
	})

	assert_eq(expanded["ship_deletes"], ["player-1", "player-2", "player-3"])


func test_expand_packet_converts_compact_session_full_player_tuples() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "sf",
		"pl": [[1, "v_wing", 100, 3, 250, "pulse", "limited", "mine", "limited", 10, 20]],
	})

	assert_eq(expanded["players"][0]["id"], "player-1")
	assert_eq(expanded["players"][0]["ship_type"], "v_wing")
	assert_eq(expanded["players"][0]["score"], 100)
	assert_eq(expanded["players"][0]["lives"], 3)
	assert_eq(expanded["players"][0]["respawn_cooldown"], 250)
	assert_eq(expanded["players"][0]["primary_weapon_id"], "pulse")
	assert_eq(expanded["players"][0]["primary_ammo_policy"], "limited")
	assert_eq(expanded["players"][0]["secondary_weapon_id"], "mine")
	assert_eq(expanded["players"][0]["secondary_ammo_policy"], "limited")
	assert_eq(expanded["players"][0]["spawn_x"], 10)
	assert_eq(expanded["players"][0]["spawn_y"], 20)


func test_expand_packet_converts_compact_session_delta_player_create_tuples() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "sd",
		"pl": [[1, "v_wing", 100, 3, 250, "pulse", "limited", "mine", "limited", 10, 20]],
	})

	assert_eq(expanded["players"][0]["id"], "player-1")
	assert_eq(expanded["players"][0]["ship_type"], "v_wing")
	assert_eq(expanded["players"][0]["score"], 100)
	assert_eq(expanded["players"][0]["lives"], 3)
	assert_eq(expanded["players"][0]["respawn_cooldown"], 250)
	assert_eq(expanded["players"][0]["primary_weapon_id"], "pulse")
	assert_eq(expanded["players"][0]["primary_ammo_policy"], "limited")
	assert_eq(expanded["players"][0]["secondary_weapon_id"], "mine")
	assert_eq(expanded["players"][0]["secondary_ammo_policy"], "limited")
	assert_eq(expanded["players"][0]["spawn_x"], 10)
	assert_eq(expanded["players"][0]["spawn_y"], 20)


func test_expand_packet_converts_compact_session_delta_player_session_updates() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "sd",
		"psu": [[1, "sco", 100, "lv", 2, "rcd", 0]],
	})

	assert_eq(expanded["player_session_updates"][0], {"id": "player-1", "score": 100, "lives": 2, "respawn_cooldown": 0})


func test_expand_packet_converts_compact_session_delta_player_session_deletes() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "sd",
		"psx": [1, "2", "player-3"],
	})

	assert_eq(expanded["player_session_deletes"], ["player-1", "player-2", "player-3"])


func test_expand_packet_converts_compact_session_lifecycle_tuples() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "sd",
		"plc": [[1, "active"]],
		"plu": [[1, "respawning"]],
	})

	assert_eq(expanded["player_lifecycle"][0], {"player_id": "player-1", "status": "active"})
	assert_eq(expanded["player_lifecycle_updates"][0], {"player_id": "player-1", "status": "respawning"})


func test_expand_packet_converts_compact_session_lifecycle_deletes() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "sd",
		"plx": [1],
	})

	assert_eq(expanded["player_lifecycle_deletes"], ["player-1"])


func test_expand_compact_bullet_id_keeps_null_and_unsupported_values() -> void:
	assert_null(CompactLanePacket.expand_compact_bullet_id(null))
	assert_eq(CompactLanePacket.expand_compact_bullet_id({"id": 123}), {"id": 123})


func test_expand_packet_converts_compact_world_full_bullet_tuples() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wf",
		"bullets": [[1, "player-1", 10, 20, 30, "pulse", "laser"]],
	})

	assert_eq(expanded["bullets"][0]["id"], "bullet-1")
	assert_eq(expanded["bullets"][0]["owner_id"], "player-1")
	assert_eq(expanded["bullets"][0]["x"], 10)
	assert_eq(expanded["bullets"][0]["y"], 20)
	assert_eq(expanded["bullets"][0]["rotation"], 30)
	assert_eq(expanded["bullets"][0]["weapon_id"], "pulse")
	assert_eq(expanded["bullets"][0]["projectile_type"], "laser")


func test_expand_packet_converts_compact_world_delta_bullet_create_tuples() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"bc": [[1, "player-1", 10, 20, 30, "pulse", "laser"]],
	})

	assert_eq(expanded["bullet_creates"][0]["id"], "bullet-1")
	assert_eq(expanded["bullet_creates"][0]["owner_id"], "player-1")
	assert_eq(expanded["bullet_creates"][0]["x"], 10)
	assert_eq(expanded["bullet_creates"][0]["y"], 20)
	assert_eq(expanded["bullet_creates"][0]["rotation"], 30)
	assert_eq(expanded["bullet_creates"][0]["weapon_id"], "pulse")
	assert_eq(expanded["bullet_creates"][0]["projectile_type"], "laser")


func test_expand_packet_converts_compact_world_delta_bullet_update_tuple_xy_rotation() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"bu": [[1, 10, 20, 30]],
	})

	assert_eq(expanded["bullet_updates"][0]["id"], "bullet-1")
	assert_eq(expanded["bullet_updates"][0]["x"], 10)
	assert_eq(expanded["bullet_updates"][0]["y"], 20)
	assert_eq(expanded["bullet_updates"][0]["rotation"], 30)


func test_expand_packet_converts_compact_world_delta_bullet_update_tuple_x_only() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"bu": [[1, 10]],
	})

	assert_eq(expanded["bullet_updates"][0]["id"], "bullet-1")
	assert_eq(expanded["bullet_updates"][0]["x"], 10)
	assert_false(expanded["bullet_updates"][0].has("y"))
	assert_false(expanded["bullet_updates"][0].has("rotation"))


func test_expand_packet_converts_compact_world_delta_bullet_update_tuple_y_only() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"bu": [[1, null, 20]],
	})

	assert_eq(expanded["bullet_updates"][0]["id"], "bullet-1")
	assert_false(expanded["bullet_updates"][0].has("x"))
	assert_eq(expanded["bullet_updates"][0]["y"], 20)
	assert_false(expanded["bullet_updates"][0].has("rotation"))


func test_expand_packet_converts_compact_world_delta_bullet_update_tuple_rotation_only() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"bu": [[1, null, null, 30]],
	})

	assert_eq(expanded["bullet_updates"][0]["id"], "bullet-1")
	assert_false(expanded["bullet_updates"][0].has("x"))
	assert_false(expanded["bullet_updates"][0].has("y"))
	assert_eq(expanded["bullet_updates"][0]["rotation"], 30)


func test_expand_packet_converts_compact_world_delta_bullet_update_tuple_zero_values() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"bu": [[1, 0, 0, 0]],
	})

	assert_eq(expanded["bullet_updates"][0]["id"], "bullet-1")
	assert_eq(expanded["bullet_updates"][0]["x"], 0)
	assert_eq(expanded["bullet_updates"][0]["y"], 0)
	assert_eq(expanded["bullet_updates"][0]["rotation"], 0)


func test_expand_packet_converts_compact_world_delta_bullet_delete_ids() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"bx": [1, "2", "bullet-3"],
	})

	assert_eq(expanded["bullet_deletes"], ["bullet-1", "bullet-2", "bullet-3"])


func test_expand_compact_asteroid_id_rebuilds_numeric_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_asteroid_id(123), "asteroid-123")


func test_expand_compact_asteroid_id_rebuilds_float_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_asteroid_id(1.0), "asteroid-1")


func test_expand_compact_asteroid_id_keeps_non_integer_float() -> void:
	assert_eq(CompactLanePacket.expand_compact_asteroid_id(1.5), 1.5)

func test_expand_compact_asteroid_id_rebuilds_string_suffix() -> void:
	assert_eq(CompactLanePacket.expand_compact_asteroid_id("123"), "asteroid-123")


func test_expand_compact_asteroid_id_keeps_full_id() -> void:
	assert_eq(CompactLanePacket.expand_compact_asteroid_id("asteroid-123"), "asteroid-123")


func test_expand_compact_asteroid_id_keeps_null_and_unsupported_values() -> void:
	assert_null(CompactLanePacket.expand_compact_asteroid_id(null))
	assert_eq(CompactLanePacket.expand_compact_asteroid_id({"id": 123}), {"id": 123})


func test_expand_packet_converts_compact_world_full_asteroid_tuples() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wf",
		"asteroids": [[1, 10, 20, 2, 90, 1500, 3]],
	})

	assert_eq(expanded["asteroids"][0]["id"], "asteroid-1")
	assert_eq(expanded["asteroids"][0]["x"], 10)
	assert_eq(expanded["asteroids"][0]["variant"], 3)


func test_expand_packet_converts_compact_world_delta_asteroid_create_tuples() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"ac": [[1, 10, 20, 2, 90, 1500, 3]],
	})

	assert_eq(expanded["asteroid_creates"][0]["id"], "asteroid-1")

func test_expand_packet_converts_compact_world_delta_asteroid_update_tuple_xy() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"au": [[1, 10, 20]],
	})

	assert_eq(expanded["asteroid_updates"][0]["id"], "asteroid-1")
	assert_eq(expanded["asteroid_updates"][0]["x"], 10)
	assert_eq(expanded["asteroid_updates"][0]["y"], 20)


func test_expand_packet_converts_compact_world_delta_asteroid_update_tuple_x_only() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"au": [[1, 10]],
	})

	assert_eq(expanded["asteroid_updates"][0]["id"], "asteroid-1")
	assert_eq(expanded["asteroid_updates"][0]["x"], 10)
	assert_false(expanded["asteroid_updates"][0].has("y"))


func test_expand_packet_converts_compact_world_delta_asteroid_update_tuple_y_only() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"au": [[1, null, 20]],
	})

	assert_eq(expanded["asteroid_updates"][0]["id"], "asteroid-1")
	assert_false(expanded["asteroid_updates"][0].has("x"))
	assert_eq(expanded["asteroid_updates"][0]["y"], 20)


func test_expand_packet_converts_compact_world_delta_asteroid_update_tuple_zero_values() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"au": [[1, 0, 0]],
	})

	assert_eq(expanded["asteroid_updates"][0]["id"], "asteroid-1")
	assert_eq(expanded["asteroid_updates"][0]["x"], 0)
	assert_eq(expanded["asteroid_updates"][0]["y"], 0)

func test_expand_packet_converts_compact_world_delta_asteroid_delete_ids() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"ax": [1, "2", "asteroid-3"],
	})

	assert_eq(expanded["asteroid_deletes"], ["asteroid-1", "asteroid-2", "asteroid-3"])



func test_expand_packet_converts_compact_world_delta_ship_update_tuple_sparse_interior_nulls_preserve_later_values() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"su": [[1, null, 20]],
	})

	assert_eq(expanded["ship_updates"][0], {"id": "player-1", "y": 20})

func test_expand_packet_converts_compact_world_delta_ship_update_tuple_rotation_only_with_sparse_nulls() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"su": [[1, null, null, 30]],
	})

	assert_eq(expanded["ship_updates"][0], {"id": "player-1", "rotation": 30})

func test_expand_packet_converts_compact_world_delta_ship_update_tuple_thrusting_only_with_sparse_nulls() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"su": [[1, null, null, null, true]],
	})

	assert_eq(expanded["ship_updates"][0], {"id": "player-1", "thrusting": true})

func test_expand_packet_converts_compact_world_delta_bullet_update_tuple_sparse_interior_nulls_preserve_later_values() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"bu": [[1, null, 20]],
	})

	assert_eq(expanded["bullet_updates"][0], {"id": "bullet-1", "y": 20})

func test_expand_packet_converts_compact_world_delta_bullet_update_tuple_rotation_only_with_sparse_nulls() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "wd",
		"bu": [[1, null, null, 30]],
	})

	assert_eq(expanded["bullet_updates"][0], {"id": "bullet-1", "rotation": 30})




func test_expand_packet_rehydrates_tagged_compact_ids_in_tuples() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "eb",
		"bid": ["eb", 11],
		"ev": [
			["dmg", ["pe", 3], "mystery", ["p", 2], "blast", 17, 50, 60],
		],
	})

	assert_eq(expanded["batch_id"], "event-batch-11")
	assert_eq(expanded["events"][0], {"event_id": "presentation-event-3", "type": "damage_applied", "source_type": "mystery", "source_id": "player-2", "effect_type": "blast", "amount": 17, "x": 50, "y": 60})




func test_expand_packet_preserves_unknown_and_malformed_tuple_ids() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "eb",
		"bid": "event-batch-bad",
		"ev": [
			["pdr", "event-batch-1", "pickup-bad", "shield", "mystery", "hazard-1", "table-bad", 150, 160],
		],
	})

	assert_eq(expanded["batch_id"], "event-batch-bad")
	assert_eq(expanded["events"][0], {"event_id": "event-batch-1", "type": "pickup_dropped", "pickup_id": "pickup-bad", "pickup_type": "shield", "source_type": "mystery", "source_id": "hazard-1", "table_id": "table-bad", "x": 150, "y": 160})




func test_expand_packet_converts_compact_asteroid_delta_updates() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "ad",
		"au": [[1, 10, 20]],
	})

	assert_eq(expanded["type"], "asteroid_delta")
	assert_eq(expanded["asteroid_updates"][0]["id"], "asteroid-1")
	assert_eq(expanded["asteroid_updates"][0]["x"], 10)
	assert_eq(expanded["asteroid_updates"][0]["y"], 20)
	assert_false(expanded.has("asteroid_creates"))
	assert_false(expanded.has("asteroid_deletes"))
	assert_false(expanded.has("bullet_creates"))
	assert_false(expanded.has("bullet_deletes"))


func test_expand_packet_converts_compact_bullet_delta_updates() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "bd",
		"bu": [[1, 10, 20, 30]],
	})

	assert_eq(expanded["type"], "bullet_delta")
	assert_eq(expanded["bullet_updates"][0]["id"], "bullet-1")
	assert_eq(expanded["bullet_updates"][0]["x"], 10)
	assert_eq(expanded["bullet_updates"][0]["y"], 20)
	assert_eq(expanded["bullet_updates"][0]["rotation"], 30)
	assert_false(expanded.has("asteroid_creates"))
	assert_false(expanded.has("asteroid_deletes"))
	assert_false(expanded.has("bullet_creates"))
	assert_false(expanded.has("bullet_deletes"))
