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

func test_chunked_ship_lifecycle_preserves_updates_from_every_chunk() -> void:
	var router := RealtimeRouter.new()
	_world_full(router, "world-baseline-1", [_ship("ship-remote")])

	router.route_lane_packet(_lifecycle_chunk(
		"ships_lifecycle",
		LaneMetadata.LANE_SHIPS_LIFECYCLE,
		0,
		false,
		"world-baseline-1",
		[],
		[],
		[{"id": "ship-remote", "health": 75}]
	))
	assert_eq(router.world_lane_state.ships["ship-remote"]["health"], 100)

	router.route_lane_packet(_lifecycle_chunk(
		"ships_lifecycle",
		LaneMetadata.LANE_SHIPS_LIFECYCLE,
		1,
		true,
		"world-baseline-1",
		[],
		[],
		[{"id": "ship-remote", "shields": 5}]
	))
	assert_eq(router.world_lane_state.ships["ship-remote"]["health"], 75)
	assert_eq(router.world_lane_state.ships["ship-remote"]["shields"], 5)


func test_lifecycle_is_atomic_across_two_chunks() -> void:
	var router := RealtimeRouter.new()
	_world_full(router, "world-baseline-1")

	router.route_lane_packet(_lifecycle_chunk("ships_lifecycle", LaneMetadata.LANE_SHIPS_LIFECYCLE, 0, false, "world-baseline-1", [_ship("ship-1")], ["ship-old"]))
	assert_false(router.world_lane_state.ships.has("ship-1"))

	router.route_lane_packet(_lifecycle_chunk("ships_lifecycle", LaneMetadata.LANE_SHIPS_LIFECYCLE, 1, true, "world-baseline-1", [_ship("ship-2")], []))
	assert_true(router.world_lane_state.ships.has("ship-1"))
	assert_true(router.world_lane_state.ships.has("ship-2"))
	assert_false(router.world_lane_state.ships.has("ship-old"))

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
	for lane in [LaneMetadata.LANE_SHIPS_LIFECYCLE, LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, LaneMetadata.LANE_BULLETS_LIFECYCLE]:
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

func _world_full(router: RealtimeRouter, baseline: String, ships: Array = []) -> void:
	router.route_lane_packet({"type": "world_full", "lane": LaneMetadata.LANE_WORLD, "sequence": 1, "baseline_id": baseline, "snapshot_id": "world-snapshot-1", "snapshot_kind": "full", "chunk_index": 0, "chunk_count": 1, "is_final_chunk": true, "ships": ships, "bullets": [], "asteroids": [], "pickups": []})

func _world_chunk(index: int, final: bool, ships: Array, asteroids: Array, baseline := "world-baseline-1") -> Dictionary:
	return {
		"type": "world_full", "lane": LaneMetadata.LANE_WORLD, "sequence": 1,
		"baseline_id": baseline, "snapshot_id": baseline.replace("baseline", "snapshot"),
		"snapshot_kind": "full", "chunk_index": index, "chunk_count": 2,
		"is_final_chunk": final, "ships": ships, "bullets": [], "asteroids": asteroids, "pickups": [],
	}

func _lifecycle_chunk(type: String, lane: String, index: int, final: bool, baseline: String, creates: Array, deletes: Array, updates: Array = []) -> Dictionary:
	var packet := {"type": type, "lane": lane, "sequence": 1, "baseline_id": baseline,
		"snapshot_id": "lifecycle-snapshot-1", "snapshot_kind": "delta", "chunk_index": index,
		"chunk_count": 2, "is_final_chunk": final}
	match lane:
		LaneMetadata.LANE_SHIPS_LIFECYCLE:
			packet["ship_creates"] = creates
			packet["ship_updates"] = updates
			packet["ship_deletes"] = deletes
		LaneMetadata.LANE_ASTEROIDS_LIFECYCLE:
			packet["asteroid_creates"] = creates
			packet["asteroid_deletes"] = deletes
		_:
			packet["bullet_creates"] = creates
			packet["bullet_deletes"] = deletes
	return packet

func _ship(id: String) -> Dictionary:
	return {"id": id, "x": 10, "y": 20, "rotation": 0, "velocity_x": 0, "velocity_y": 0, "thrusting": false, "health": 100, "shields": 0}

func _asteroid(id: String) -> Dictionary:
	return {"id": id, "x": 10, "y": 20, "velocity_x": 0, "velocity_y": 0, "rotation": 0, "size": 1, "health": 100, "scale": 1000, "variant": 1}

func _record(lane: String, id: String) -> Dictionary:
	if lane == LaneMetadata.LANE_SHIPS_LIFECYCLE:
		return _ship(id)
	if lane == LaneMetadata.LANE_ASTEROIDS_LIFECYCLE:
		return _asteroid(id)
	return {"id": id, "owner_id": "player-1", "x": 10, "y": 20, "velocity_x": 0, "velocity_y": 0, "rotation": 0, "lifespan_seconds": 1, "weapon_id": "bullet", "projectile_type": "bullet"}

func _record_id(lane: String, id: String) -> String:
	return id

func _packet_type(lane: String) -> String:
	if lane == LaneMetadata.LANE_SHIPS_LIFECYCLE:
		return "ships_lifecycle"
	return "asteroids_lifecycle" if lane == LaneMetadata.LANE_ASTEROIDS_LIFECYCLE else "bullets_lifecycle"

func _state(router: RealtimeRouter, lane: String):
	if lane == LaneMetadata.LANE_SHIPS_LIFECYCLE:
		return router.world_lane_state.ships
	return router.world_lane_state.asteroids if lane == LaneMetadata.LANE_ASTEROIDS_LIFECYCLE else router.world_lane_state.bullets

const WorldFullChunkAssembler := preload("res://scripts/protocol/realtime/world_full_chunk_assembler.gd")
const RealtimeReceiveLimits := preload("res://scripts/protocol/realtime/realtime_receive_limits.gd")

func _assembler_packet(index = 0, count = 1, final = true, sequence = 1, baseline = "world-baseline-1") -> Dictionary:
	return {"match_id": "match-1", "lane": LaneMetadata.LANE_WORLD, "sequence": sequence, "baseline_id": baseline, "snapshot_id": "snapshot-1", "snapshot_kind": "full", "chunk_index": index, "chunk_count": count, "is_final_chunk": final, "ships": [], "bullets": [], "asteroids": [], "pickups": []}

func test_direct_assembler_rejects_raw_metadata() -> void:
	var assembler := WorldFullChunkAssembler.new()
	for key_value in [["sequence", "1"], ["chunk_index", 1.5], ["chunk_count", true], ["baseline_id", 10], ["is_final_chunk", "true"]]:
		var packet := _assembler_packet()
		packet[key_value[0]] = key_value[1]
		assert_eq(assembler.accept(packet).reason, "invalid_chunk_metadata")

func test_direct_assembler_rejects_chunk_cap() -> void:
	var assembler := WorldFullChunkAssembler.new()
	assert_eq(assembler.accept(_assembler_packet(0, RealtimeReceiveLimits.MAX_WORLD_FULL_CHUNKS_PER_ASSEMBLY + 1, false)).reason, "chunk_limit")

func test_direct_assembler_rejects_record_cap() -> void:
	var assembler := WorldFullChunkAssembler.new()
	var packet := _assembler_packet()
	packet["ships"] = range(RealtimeReceiveLimits.MAX_WORLD_FULL_RECORDS_PER_ASSEMBLY + 1)
	assert_eq(assembler.accept(packet).reason, "record_limit")

func test_direct_assembler_rejects_byte_cap() -> void:
	var assembler := WorldFullChunkAssembler.new()
	var packet := _assembler_packet()
	packet["padding"] = "x".repeat(RealtimeReceiveLimits.MAX_WORLD_FULL_ESTIMATED_BYTES_PER_ASSEMBLY)
	assert_eq(assembler.accept(packet).reason, "byte_limit")

func test_direct_assembler_expires_and_resets() -> void:
	var now := [0]
	var assembler := WorldFullChunkAssembler.new(func(): return now[0])
	assert_eq(assembler.accept(_assembler_packet(0, 2, false)).status, WorldFullChunkAssembler.INCOMPLETE)
	now[0] = RealtimeReceiveLimits.WORLD_FULL_ASSEMBLY_LIFETIME_MSEC
	assert_eq(assembler.accept(_assembler_packet(1, 2, true)).reason, "expired")
	assert_eq(assembler.accept(_assembler_packet()).status, WorldFullChunkAssembler.COMPLETE)

func test_direct_assembler_accepts_bounded_two_chunk_series() -> void:
	var assembler := WorldFullChunkAssembler.new()
	var first := _assembler_packet(0, 2, false)
	first["ships"] = [_ship("ship-1")]
	var second := _assembler_packet(1, 2, true)
	second["asteroids"] = [_asteroid("asteroid-1")]
	assert_eq(assembler.accept(first).status, WorldFullChunkAssembler.INCOMPLETE)
	var result := assembler.accept(second)
	assert_eq(result.status, WorldFullChunkAssembler.COMPLETE)
	assert_eq(result.packet.ships.size(), 1)
	assert_eq(result.packet.asteroids.size(), 1)

func test_router_assembly_limit_requests_world_resync_without_partial_state() -> void:
	var router := RealtimeRouter.new()
	var requests := []
	router.resync_request_required.connect(func(lane, _baseline, _sequence, reason): requests.append([lane, reason]))
	var packet := _world_chunk(0, false, [_ship("partial")], [])
	packet["chunk_count"] = RealtimeReceiveLimits.MAX_WORLD_FULL_CHUNKS_PER_ASSEMBLY + 1
	router.route_lane_packet(packet)
	assert_eq(requests.size(), 1)
	assert_eq(requests[0][0], LaneMetadata.LANE_WORLD)
	assert_eq(requests[0][1], "chunk_limit")
	assert_false(router.world_lane_state.ships.has("partial"))

const LifecycleChunkAssembler := preload("res://scripts/protocol/realtime/lifecycle_chunk_assembler.gd")
const LifecycleReceiveLimits := preload("res://scripts/protocol/realtime/realtime_receive_limits.gd")

func _lifecycle_assembler_packet(index = 0, count = 1, final = true, sequence = 1, baseline = "world-baseline-1", creates: Array = [], deletes: Array = []) -> Dictionary:
	return {"match_id": "match-1", "lane": LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, "sequence": sequence, "baseline_id": baseline, "snapshot_id": "snapshot-1", "snapshot_kind": "delta", "chunk_index": index, "chunk_count": count, "is_final_chunk": final, "asteroid_creates": creates, "asteroid_deletes": deletes}

func test_direct_lifecycle_assembler_rejects_raw_metadata() -> void:
	var assembler := LifecycleChunkAssembler.new()
	for key_value in [["sequence", "1"], ["chunk_index", 1.5], ["chunk_count", true], ["baseline_id", 10], ["is_final_chunk", "true"]]:
		var packet := _lifecycle_assembler_packet()
		packet[key_value[0]] = key_value[1]
		assert_eq(assembler.accept(packet, "asteroid_creates", "asteroid_deletes").reason, "invalid_lifecycle_chunk_metadata")

func test_direct_lifecycle_assembler_rejects_chunk_cap() -> void:
	var assembler := LifecycleChunkAssembler.new()
	assert_eq(assembler.accept(_lifecycle_assembler_packet(0, LifecycleReceiveLimits.MAX_LIFECYCLE_CHUNKS_PER_ASSEMBLY + 1, false), "asteroid_creates", "asteroid_deletes").reason, "chunk_limit")

func test_direct_lifecycle_assembler_rejects_record_cap() -> void:
	var assembler := LifecycleChunkAssembler.new()
	var packet := _lifecycle_assembler_packet()
	packet["asteroid_creates"] = range(LifecycleReceiveLimits.MAX_LIFECYCLE_RECORDS_PER_ASSEMBLY + 1)
	assert_eq(assembler.accept(packet, "asteroid_creates", "asteroid_deletes").reason, "record_limit")

func test_direct_lifecycle_assembler_rejects_byte_cap() -> void:
	var assembler := LifecycleChunkAssembler.new()
	var packet := _lifecycle_assembler_packet()
	packet["padding"] = "x".repeat(LifecycleReceiveLimits.MAX_LIFECYCLE_ESTIMATED_BYTES_PER_ASSEMBLY)
	assert_eq(assembler.accept(packet, "asteroid_creates", "asteroid_deletes").reason, "byte_limit")

func test_direct_lifecycle_assembler_expires_resets_and_accepts_again() -> void:
	var now := [0]
	var assembler := LifecycleChunkAssembler.new(func(): return now[0])
	assert_eq(assembler.accept(_lifecycle_assembler_packet(0, 2, false), "asteroid_creates", "asteroid_deletes").status, LifecycleChunkAssembler.INCOMPLETE)
	now[0] = LifecycleReceiveLimits.LIFECYCLE_ASSEMBLY_LIFETIME_MSEC
	assert_eq(assembler.accept(_lifecycle_assembler_packet(1, 2, true), "asteroid_creates", "asteroid_deletes").reason, "expired")
	assert_eq(assembler.accept(_lifecycle_assembler_packet(), "asteroid_creates", "asteroid_deletes").status, LifecycleChunkAssembler.COMPLETE)

func test_direct_lifecycle_assembler_accepts_bounded_two_chunk_series() -> void:
	var assembler := LifecycleChunkAssembler.new()
	assert_eq(assembler.accept(_lifecycle_assembler_packet(0, 2, false, 1, "world-baseline-1", [{"id": "a"}], ["old"]), "asteroid_creates", "asteroid_deletes").status, LifecycleChunkAssembler.INCOMPLETE)
	var result := assembler.accept(_lifecycle_assembler_packet(1, 2, true, 1, "world-baseline-1", [{"id": "b"}], []), "asteroid_creates", "asteroid_deletes")
	assert_eq(result.status, LifecycleChunkAssembler.COMPLETE)
	assert_eq(result.packet.asteroid_creates.size(), 2)
	assert_eq(result.packet.asteroid_deletes.size(), 1)

func test_router_lifecycle_assembly_limits_resync_without_partial_state() -> void:
	for lane in [LaneMetadata.LANE_SHIPS_LIFECYCLE, LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, LaneMetadata.LANE_BULLETS_LIFECYCLE]:
		var router := RealtimeRouter.new()
		var requests := []
		router.resync_request_required.connect(func(requested_lane, _baseline, _sequence, reason): requests.append([requested_lane, reason]))
		var packet := _lifecycle_chunk(_packet_type(lane), lane, 0, false, "world-baseline-1", [_record(lane, "partial")], [])
		packet["chunk_count"] = LifecycleReceiveLimits.MAX_LIFECYCLE_CHUNKS_PER_ASSEMBLY + 1
		router.route_lane_packet(packet)
		assert_eq(requests.size(), 1)
		assert_eq(requests[0][0], LaneMetadata.LANE_WORLD)
		assert_eq(requests[0][1], "chunk_limit")
		assert_false(_state(router, lane).has("partial"))
