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


func test_realtime_router_exposes_gameplay_readiness() -> void:
	var router := RealtimeRouter.new()

	assert_not_null(router.get_gameplay_readiness())
	assert_eq(router.get_gameplay_readiness(), router.gameplay_readiness)


func test_realtime_router_is_presentable_after_required_baselines() -> void:
	var router := RealtimeRouter.new()

	router.route_lane_packet({"type": "world_full", "baseline_id": "world-baseline", "sequence": 1, "snapshot_id": "world-snapshot", "is_final_chunk": true, "ships": [], "bullets": [], "asteroids": [], "pickups": []})
	router.route_lane_packet({"type": "overlay_full", "baseline_id": "overlay-baseline", "sequence": 1, "snapshot_id": "overlay-snapshot", "is_final_chunk": true})
	router.route_lane_packet({"type": "session_full", "baseline_id": "session-baseline", "sequence": 1, "snapshot_id": "session-snapshot", "is_final_chunk": true, "players": [], "player_lifecycle": [], "total_asteroids": 0})

	assert_true(router.is_presentable())


func test_lowercase_lane_fixtures_route_directly() -> void:
	var router := RealtimeRouter.new()

	var world_packet = _decode_fixture("{\"type\":\"world_full\",\"lane\":\"world\",\"sequence\":7,\"baseline_id\":\"baseline-1\",\"snapshot_id\":\"snapshot-1\",\"chunk_index\":0,\"chunk_count\":1,\"is_final_chunk\":true,\"ships\":[{\"id\":\"ship-1\",\"ship_type\":\"v_wing\",\"x\":1,\"y\":2,\"rotation\":0,\"health\":100,\"shields\":0,\"thrusting\":false,\"target_kind\":\"player\",\"target_id\":\"player-1\"}],\"bullets\":[],\"asteroids\":[],\"pickups\":[]}")
	assert_eq(world_packet["type"], "world_full")
	router.route_lane_packet(world_packet)
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))

	var overlay_packet = _decode_fixture("{\"type\":\"overlay_full\",\"lane\":\"overlay\",\"sequence\":2,\"baseline_id\":\"overlay-baseline-1\",\"snapshot_id\":\"overlay-snapshot-1\",\"chunk_index\":0,\"chunk_count\":1,\"is_final_chunk\":true,\"self_id\":\"player-1\",\"lives\":3,\"score\":120,\"respawn_cooldown\":2,\"primary_weapon_id\":\"laser\",\"primary_ammo_policy\":\"finite\",\"primary_cooldown_remaining\":1.5,\"primary_ammo_remaining\":9,\"secondary_weapon_id\":\"burst\",\"secondary_ammo_policy\":\"infinite\",\"secondary_cooldown_remaining\":0.5,\"secondary_ammo_remaining\":99}")
	assert_eq(overlay_packet["type"], "overlay_full")
	router.route_lane_packet(overlay_packet)
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_OVERLAY))

	var session_packet = _decode_fixture("{\"type\":\"session_full\",\"lane\":\"session\",\"sequence\":3,\"baseline_id\":\"session-baseline-1\",\"snapshot_id\":\"session-snapshot-1\",\"chunk_index\":0,\"chunk_count\":1,\"is_final_chunk\":true,\"players\":[{\"id\":\"player-1\",\"ship_type\":\"v_wing\",\"score\":8,\"lives\":3,\"respawn_cooldown\":0.25,\"primary_weapon_id\":\"pulse\",\"primary_ammo_policy\":\"limited\",\"secondary_weapon_id\":\"mine\",\"secondary_ammo_policy\":\"infinite\",\"spawn_x\":10,\"spawn_y\":20}],\"player_lifecycle\":[{\"id\":\"player-1\",\"status\":\"active\"}],\"total_asteroids\":42}")
	assert_eq(session_packet["type"], "session_full")
	router.route_lane_packet(session_packet)
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_SESSION))


func test_expanded_world_delta_routes_without_reexpansion() -> void:
	var router := RealtimeRouter.new()

	router.route_lane_packet({
		"type": "world_full",
		"baseline_id": "b1",
		"sequence": 1,
		"snapshot_id": "s1",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})
	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))

	router.route_lane_packet({
		"type": "world_delta",
		"baseline_id": "b1",
		"sequence": 2,
	})

	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))


func test_compact_world_full_routes_as_fallback() -> void:
	var router := RealtimeRouter.new()

	router.route_lane_packet({
		"t": "wf",
		"q": 1,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})

	assert_true(router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))


func test_asteroid_delta_routes_into_world_lane_state() -> void:
	var router := RealtimeRouter.new()
	router.world_lane_state.upsert_asteroid({"id": "asteroid-1", "x": 1.0, "y": 2.0, "rotation": 0.0})

	router.route_lane_packet({
		"type": "asteroid_delta",
		"sequence": 1,
		"asteroid_updates": [
			{"id": "asteroid-1", "x": 42, "y": 84},
			{"id": "asteroid-unknown", "x": 123, "y": 456},
		],
	})

	assert_eq(router.world_lane_state.asteroids["asteroid-1"]["x"], 4.2)
	assert_eq(router.world_lane_state.asteroids["asteroid-1"]["y"], 8.4)
	assert_false(router.world_lane_state.asteroids.has("asteroid-unknown"))


func test_bullet_delta_routes_into_world_lane_state() -> void:
	var router := RealtimeRouter.new()
	router.world_lane_state.upsert_bullet({"id": "bullet-1", "x": 1.0, "y": 2.0, "rotation": 0.0})

	router.route_lane_packet({
		"type": "bullet_delta",
		"sequence": 1,
		"bullet_updates": [
			{"id": "bullet-1", "x": 55, "y": 66},
		],
	})

	assert_eq(router.world_lane_state.bullets["bullet-1"]["x"], 5.5)
	assert_eq(router.world_lane_state.bullets["bullet-1"]["y"], 6.6)


func test_asteroids_lifecycle_routes_into_world_lane_state() -> void:
	var router := RealtimeRouter.new()

	router.route_lane_packet({
		"type": "asteroids_lifecycle",
		"lane": LaneMetadata.LANE_ASTEROIDS_LIFECYCLE,
		"asteroid_creates": [{"id": "asteroid-1", "x": 10, "y": 20, "velocity_x": 0.0, "velocity_y": 0.0, "rotation": 0.0, "size": 2, "health": 90, "scale": 1500, "variant": 3}],
		"asteroid_deletes": [],
	})

	assert_true(router.world_lane_state.asteroids.has("asteroid-1"))
	assert_eq(router.world_lane_state.asteroids["asteroid-1"]["variant"], 3)


func test_bullets_lifecycle_routes_into_world_lane_state() -> void:
	var router := RealtimeRouter.new()

	router.route_lane_packet({
		"type": "bullets_lifecycle",
		"lane": LaneMetadata.LANE_BULLETS_LIFECYCLE,
		"bullet_creates": [{"id": "bullet-1", "owner_id": "player-1", "x": 10, "y": 20, "velocity_x": 0.0, "velocity_y": 0.0, "rotation": 30, "lifespan_seconds": 1.0, "weapon_id": "torpedo", "projectile_type": "torpedo"}],
		"bullet_deletes": [],
	})

	assert_true(router.world_lane_state.bullets.has("bullet-1"))
	assert_eq(router.world_lane_state.bullets["bullet-1"]["projectile_type"], "torpedo")


func test_compact_asteroids_lifecycle_routes_into_world_lane_state() -> void:
	var router := RealtimeRouter.new()

	router.route_lane_packet({"t": "al", "l": "al", "q": 1, "k": "d", "ac": [[1, 10, 20, 2, 90, 1500, 3]]})

	assert_true(router.world_lane_state.asteroids.has("asteroid-1"))
	assert_eq(router.world_lane_state.asteroids["asteroid-1"]["variant"], 3)


func test_compact_bullets_lifecycle_routes_into_world_lane_state() -> void:
	var router := RealtimeRouter.new()

	router.route_lane_packet({"t": "bl", "l": "bl", "q": 1, "k": "d", "bc": [[1, "player-1", 10, 20, 30, "torpedo", "torpedo"]]})

	assert_true(router.world_lane_state.bullets.has("bullet-1"))
	assert_eq(router.world_lane_state.bullets["bullet-1"]["projectile_type"], "torpedo")


func test_compact_bullets_lifecycle_delete_removes_bullet() -> void:
	var router := RealtimeRouter.new()

	router.route_lane_packet({"t": "bl", "l": "bl", "q": 1, "k": "d", "bc": [[1, "player-1", 10, 20, 30, "torpedo", "torpedo"]]})
	router.route_lane_packet({"t": "bl", "l": "bl", "q": 2, "k": "d", "bx": [1]})

	assert_false(router.world_lane_state.bullets.has("bullet-1"))


func _decode_fixture(text: String) -> Dictionary:
	var decoded = PacketCodec.decode(text)
	assert_true(decoded.ok)
	return decoded.packet

