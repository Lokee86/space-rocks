extends GutTest

const RealtimeRouter := preload("res://scripts/protocol/realtime/realtime_router.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const PacketCodec := preload("res://scripts/networking/packets/packet_codec.gd")


func test_lane_packet_families_route_directly() -> void:
	var router := RealtimeRouter.new()

	router.route_lane_packet({"type": LaneMetadata.PACKET_FAMILY_WORLD[0], "baseline_id": "b1", "sequence": 1, "snapshot_id": "s1", "is_final_chunk": true})
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))

	router.route_lane_packet({"type": LaneMetadata.PACKET_FAMILY_WORLD[1], "baseline_id": "b1", "sequence": 2})
	router.route_lane_packet({"type": LaneMetadata.PACKET_FAMILY_OVERLAY[0], "baseline_id": "b2", "sequence": 1, "snapshot_id": "o1", "is_final_chunk": true})
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_OVERLAY))

	router.route_lane_packet({"type": LaneMetadata.PACKET_FAMILY_OVERLAY[1], "baseline_id": "b2", "sequence": 2})
	router.route_lane_packet({"type": LaneMetadata.PACKET_FAMILY_SESSION[0], "baseline_id": "b3", "sequence": 1, "snapshot_id": "u1", "is_final_chunk": true})
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_SESSION))

	router.route_lane_packet({"type": LaneMetadata.PACKET_FAMILY_SESSION[1], "baseline_id": "b3", "sequence": 2})
	router.route_lane_packet({"type": LaneMetadata.PACKET_FAMILY_EVENT[0], "batch_id": "batch-1", "events": [{"event_id": "event-1", "type": "spark", "payload": {}}]})
	assert_true(router.event_batch_applier.has_applied_batch("batch-1"))
	assert_true(router.event_batch_applier.has_applied_event("event-1"))


func test_lowercase_lane_fixtures_route_directly() -> void:
	var router := RealtimeRouter.new()

	var world_packet = _decode_fixture("{\"type\":\"world_full\",\"lane\":\"world\",\"sequence\":7,\"baseline_id\":\"baseline-1\",\"snapshot_id\":\"snapshot-1\",\"chunk_index\":0,\"chunk_count\":1,\"is_final_chunk\":true,\"ships\":[{\"id\":\"ship-1\",\"ship_type\":\"v_wing\",\"x\":1,\"y\":2,\"rotation\":0,\"health\":100,\"shields\":0,\"thrusting\":false,\"target_kind\":\"player\",\"target_id\":\"player-1\"}],\"bullets\":[],\"asteroids\":[],\"pickups\":[]}")
	assert_eq(world_packet["type"], "world_full")
	router.route_lane_packet(world_packet)
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))

	var overlay_packet = _decode_fixture("{\"type\":\"overlay_full\",\"lane\":\"overlay\",\"sequence\":2,\"baseline_id\":\"overlay-baseline-1\",\"snapshot_id\":\"overlay-snapshot-1\",\"chunk_index\":0,\"chunk_count\":1,\"is_final_chunk\":true,\"self_id\":\"player-1\",\"lives\":3,\"score\":120,\"respawn\":{\"delay\":2},\"primary_weapon_id\":\"laser\",\"primary_ammo_policy\":\"finite\",\"primary_cooldown_remaining\":1.5,\"primary_ammo_remaining\":9,\"secondary_weapon_id\":\"burst\",\"secondary_ammo_policy\":\"infinite\",\"secondary_cooldown_remaining\":0.5,\"secondary_ammo_remaining\":99}")
	assert_eq(overlay_packet["type"], "overlay_full")
	router.route_lane_packet(overlay_packet)
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_OVERLAY))

	var session_packet = _decode_fixture("{\"type\":\"session_full\",\"lane\":\"session\",\"sequence\":3,\"baseline_id\":\"session-baseline-1\",\"snapshot_id\":\"session-snapshot-1\",\"chunk_index\":0,\"chunk_count\":1,\"is_final_chunk\":true,\"players\":[{\"id\":\"player-1\",\"ship_type\":\"v_wing\",\"score\":8,\"lives\":3,\"respawn_cooldown\":0.25,\"primary_weapon_id\":\"pulse\",\"primary_ammo_policy\":\"limited\",\"secondary_weapon_id\":\"mine\",\"secondary_ammo_policy\":\"infinite\",\"spawn_x\":10,\"spawn_y\":20}],\"player_lifecycle\":[{\"id\":\"player-1\",\"status\":\"active\"}],\"total_asteroids\":42}")
	assert_eq(session_packet["type"], "session_full")
	router.route_lane_packet(session_packet)
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_SESSION))


func _decode_fixture(text: String) -> Dictionary:
	var decoded = PacketCodec.decode(text)
	assert_true(decoded.ok)
	return decoded.packet
