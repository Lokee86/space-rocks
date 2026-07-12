extends GutTest

const RealtimeRouter := preload("res://scripts/protocol/realtime/realtime_router.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")

func test_world_full_is_atomic_across_two_chunks() -> void:
	var router := RealtimeRouter.new()
	var first := _world_chunk(0, false, [_ship("ship-1")], [])
	var second := _world_chunk(1, true, [], [_asteroid("asteroid-1")])

	router.route_lane_packet(first)
	assert_true(router.world_lane_state.ships.is_empty())
	assert_true(router.world_lane_state.asteroids.is_empty())

	router.route_lane_packet(second)
	assert_true(router.world_lane_state.ships.has("ship-1"))
	assert_true(router.world_lane_state.asteroids.has("asteroid-1"))

func test_lifecycle_is_atomic_across_two_chunks() -> void:
	var router := RealtimeRouter.new()
	_world_full(router, "world-baseline-1")

	router.route_lane_packet(_lifecycle_chunk("asteroids_lifecycle", LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, 0, false, "world-baseline-1", [{"id": "asteroid-1"}], ["asteroid-old"]))
	assert_false(router.world_lane_state.asteroids.has("asteroid-1"))

	router.route_lane_packet(_lifecycle_chunk("asteroids_lifecycle", LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, 1, true, "world-baseline-1", [{"id": "asteroid-2"}], []))
	assert_true(router.world_lane_state.asteroids.has("asteroid-1"))
	assert_true(router.world_lane_state.asteroids.has("asteroid-2"))
	assert_false(router.world_lane_state.asteroids.has("asteroid-old"))

	router.route_lane_packet(_lifecycle_chunk("bullets_lifecycle", LaneMetadata.LANE_BULLETS_LIFECYCLE, 0, false, "world-baseline-1", [{"id": "bullet-1"}], ["bullet-old"]))
	assert_false(router.world_lane_state.bullets.has("bullet-1"))

	router.route_lane_packet(_lifecycle_chunk("bullets_lifecycle", LaneMetadata.LANE_BULLETS_LIFECYCLE, 1, true, "world-baseline-1", [{"id": "bullet-2"}], []))
	assert_true(router.world_lane_state.bullets.has("bullet-1"))
	assert_true(router.world_lane_state.bullets.has("bullet-2"))
	assert_false(router.world_lane_state.bullets.has("bullet-old"))

func test_world_series_errors_emit_one_resync_and_do_not_apply() -> void:
	var cases := [
		{"name": "duplicate", "bad": func(router): router.route_lane_packet(_world_chunk(0, false, [], []))},
		{"name": "mismatch", "bad": func(router): router.route_lane_packet(_world_chunk(1, true, [], [], "world-baseline-other"))},
		{"name": "interrupted", "bad": func(router): router.route_lane_packet(_world_chunk(0, false, [], []))},
	]
	for item in cases:
		var router := RealtimeRouter.new()
		var requests := []
		router.resync_request_required.connect(func(_lane, _baseline, _sequence, _reason): requests.append(true))
		router.route_lane_packet(_world_chunk(0, false, [_ship("old")], []))
		item.bad.call(router)
		assert_eq(requests.size(), 1, item.name)
		assert_false(router.world_lane_state.ships.has("old"), item.name)

func test_lifecycle_series_errors_emit_one_resync_and_do_not_apply() -> void:
	for lane in [LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, LaneMetadata.LANE_BULLETS_LIFECYCLE]:
		var router := RealtimeRouter.new()
		_world_full(router, "world-baseline-1")
		var requests := []
		router.resync_request_required.connect(func(_lane, _baseline, _sequence, _reason): requests.append(true))
		router.route_lane_packet(_lifecycle_chunk(_packet_type(lane), lane, 0, false, "world-baseline-1", [_record(lane, "old")], []))
		router.route_lane_packet(_lifecycle_chunk(_packet_type(lane), lane, 0, false, "world-baseline-1", [_record(lane, "duplicate")], []))
		assert_eq(requests.size(), 1)
		assert_false(_state(router, lane).has(_record_id(lane, "old")))

func test_router_replacement_clears_partial_world_and_lifecycle_series() -> void:
	var router := RealtimeRouter.new()
	var requests := []
	router.resync_request_required.connect(func(_lane, _baseline, _sequence, _reason): requests.append(true))
	router.route_lane_packet(_world_chunk(0, false, [_ship("old")], []))
	router.route_lane_packet(_lifecycle_chunk("asteroids_lifecycle", LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, 0, false, "world-baseline-1", [{"id": "old-asteroid"}], []))

	router = RealtimeRouter.new()
	router.route_lane_packet(_world_chunk(1, true, [], [_asteroid("late")]))
	router.route_lane_packet(_lifecycle_chunk("asteroids_lifecycle", LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, 1, true, "world-baseline-1", [{"id": "late-asteroid"}], []))
	assert_false(router.world_lane_state.asteroids.has("late"))
	assert_false(router.world_lane_state.asteroids.has("late-asteroid"))
	assert_eq(requests.size(), 0)

func _world_full(router: RealtimeRouter, baseline: String) -> void:
	router.route_lane_packet({"type": "world_full", "lane": LaneMetadata.LANE_WORLD, "sequence": 1, "baseline_id": baseline, "snapshot_id": "world-snapshot-1", "snapshot_kind": "full", "chunk_index": 0, "chunk_count": 1, "is_final_chunk": true, "ships": [], "bullets": [], "asteroids": [], "pickups": []})

func _world_chunk(index: int, final: bool, ships: Array, asteroids: Array, baseline := "world-baseline-1") -> Dictionary:
	return {
		"type": "world_full", "lane": LaneMetadata.LANE_WORLD, "sequence": 1,
		"baseline_id": baseline, "snapshot_id": baseline.replace("baseline", "snapshot"),
		"snapshot_kind": "full", "chunk_index": index, "chunk_count": 2,
		"is_final_chunk": final, "ships": ships, "bullets": [], "asteroids": asteroids, "pickups": [],
	}

func _lifecycle_chunk(type: String, lane: String, index: int, final: bool, baseline: String, creates: Array, deletes: Array) -> Dictionary:
	var packet := {"type": type, "lane": lane, "sequence": 1, "baseline_id": baseline,
		"snapshot_id": "lifecycle-snapshot-1", "snapshot_kind": "delta", "chunk_index": index,
		"chunk_count": 2, "is_final_chunk": final}
	packet["asteroid_creates" if lane == LaneMetadata.LANE_ASTEROIDS_LIFECYCLE else "bullet_creates"] = creates
	packet["asteroid_deletes" if lane == LaneMetadata.LANE_ASTEROIDS_LIFECYCLE else "bullet_deletes"] = deletes
	return packet

func _ship(id: String) -> Dictionary:
	return {"id": id, "x": 10, "y": 20, "rotation": 0, "velocity_x": 0, "velocity_y": 0, "thrusting": false, "health": 100, "shields": 0}

func _asteroid(id: String) -> Dictionary:
	return {"id": id, "x": 10, "y": 20, "velocity_x": 0, "velocity_y": 0, "rotation": 0, "size": 1, "health": 100, "scale": 1000, "variant": 1}

func _record(lane: String, id: String) -> Dictionary:
	return _asteroid(id) if lane == LaneMetadata.LANE_ASTEROIDS_LIFECYCLE else {"id": id, "owner_id": "player-1", "x": 10, "y": 20, "velocity_x": 0, "velocity_y": 0, "rotation": 0, "lifespan_seconds": 1, "weapon_id": "bullet", "projectile_type": "bullet"}

func _record_id(lane: String, id: String) -> String:
	return id

func _packet_type(lane: String) -> String:
	return "asteroids_lifecycle" if lane == LaneMetadata.LANE_ASTEROIDS_LIFECYCLE else "bullets_lifecycle"

func _state(router: RealtimeRouter, lane: String):
	return router.world_lane_state.asteroids if lane == LaneMetadata.LANE_ASTEROIDS_LIFECYCLE else router.world_lane_state.bullets
