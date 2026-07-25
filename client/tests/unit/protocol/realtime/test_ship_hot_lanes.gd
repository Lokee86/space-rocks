extends GutTest

const RealtimeRouter := preload("res://scripts/protocol/realtime/realtime_router.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const PacketCodec := preload("res://scripts/networking/packets/packet_codec.gd")


func test_ship_delta_updates_existing_ship_and_rejects_stale_sequence() -> void:
	var router := RealtimeRouter.new()
	_sync_empty_world(router)
	_create_ship(router, 1, "player-1", 100, 200)

	router.route_lane_packet({
		"type": "ship_delta",
		"sequence": 2,
		"ship_updates": [{"id": "player-1", "x": 420, "y": 840, "rotation": 250, "thrusting": true}],
	})
	assert_eq(router.world_lane_state.ships["player-1"]["x"], 42.0)
	assert_eq(router.world_lane_state.ships["player-1"]["y"], 84.0)
	assert_true(router.world_lane_state.ships["player-1"]["thrusting"])

	router.route_lane_packet({
		"type": "ship_delta",
		"sequence": 1,
		"ship_updates": [{"id": "player-1", "x": 10, "y": 20}],
	})
	assert_eq(router.world_lane_state.ships["player-1"]["x"], 42.0)
	assert_eq(router.world_lane_state.ships["player-1"]["y"], 84.0)


func test_ship_hot_update_before_reliable_create_is_buffered_and_applied() -> void:
	var router := RealtimeRouter.new()
	_sync_empty_world(router)

	router.route_lane_packet({
		"type": "ship_delta",
		"sequence": 1,
		"ship_updates": [{"id": "player-2", "x": 350, "y": 450, "rotation": 100, "thrusting": true}],
	})
	assert_false(router.world_lane_state.ships.has("player-2"))
	assert_true(router.world_lane_state.pending_ship_updates.has("player-2"))

	_create_ship(router, 1, "player-2", 100, 200)
	assert_true(router.world_lane_state.ships.has("player-2"))
	assert_eq(router.world_lane_state.ships["player-2"]["x"], 35.0)
	assert_eq(router.world_lane_state.ships["player-2"]["y"], 45.0)
	assert_true(router.world_lane_state.ships["player-2"]["thrusting"])
	assert_false(router.world_lane_state.pending_ship_updates.has("player-2"))


func test_ship_delete_blocks_late_hot_update_without_recreating_ship() -> void:
	var router := RealtimeRouter.new()
	_sync_empty_world(router)
	_create_ship(router, 1, "player-3", 100, 200)

	router.route_lane_packet({
		"type": "ships_lifecycle",
		"lane": LaneMetadata.LANE_SHIPS_LIFECYCLE,
		"sequence": 2,
		"baseline_id": "world-baseline-1",
		"ship_creates": [],
		"ship_deletes": ["player-3"],
	})
	assert_false(router.world_lane_state.ships.has("player-3"))
	assert_true(router.world_lane_state.deleted_ship_ids.has("player-3"))

	router.route_lane_packet({
		"type": "ship_delta",
		"sequence": 3,
		"ship_updates": [{"id": "player-3", "x": 999, "y": 999}],
	})
	assert_false(router.world_lane_state.ships.has("player-3"))
	assert_false(router.world_lane_state.pending_ship_updates.has("player-3"))


func test_ship_lifecycle_queues_until_matching_world_baseline() -> void:
	var router := RealtimeRouter.new()
	_create_ship(router, 1, "player-4", 100, 200)
	assert_false(router.world_lane_state.ships.has("player-4"))

	_sync_empty_world(router)
	assert_true(router.world_lane_state.ships.has("player-4"))


func test_compact_ship_hot_and_lifecycle_packets_decode_and_route() -> void:
	var router := RealtimeRouter.new()
	var world := _decode('{"t":"wf","q":1,"ships":[],"bullets":[],"asteroids":[],"pickups":[]}')
	router.route_lane_packet(world)

	var lifecycle := _decode('{"t":"spl","l":"sl","q":1,"b":"world-baseline-1","k":"d","sc":[{"id":"player-5","ship_type":"v_wing","x":100,"y":200,"rotation":0,"health":3,"shields":2,"thrusting":false,"target_kind":"","target_id":""}],"sx":[]}')
	router.route_lane_packet(lifecycle)
	assert_true(router.world_lane_state.ships.has("player-5"))

	var hot := _decode('{"t":"spd","q":1,"bq":1,"su":[[5,500,600,250,true]]}')
	router.route_lane_packet(hot)
	assert_eq(router.world_lane_state.ships["player-5"]["x"], 50.0)
	assert_eq(router.world_lane_state.ships["player-5"]["y"], 60.0)
	assert_true(router.world_lane_state.ships["player-5"]["thrusting"])


func _sync_empty_world(router) -> void:
	router.route_lane_packet({
		"type": "world_full",
		"lane": LaneMetadata.LANE_WORLD,
		"sequence": 1,
		"baseline_id": "world-baseline-1",
		"snapshot_id": "world-snapshot-1",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})


func _create_ship(router, sequence: int, id: String, x: int, y: int) -> void:
	router.route_lane_packet({
		"type": "ships_lifecycle",
		"lane": LaneMetadata.LANE_SHIPS_LIFECYCLE,
		"sequence": sequence,
		"baseline_id": "world-baseline-1",
		"ship_creates": [{
			"id": id,
			"ship_type": "v_wing",
			"x": x,
			"y": y,
			"rotation": 0,
			"health": 3,
			"shields": 2,
			"thrusting": false,
			"target_kind": "",
			"target_id": "",
		}],
		"ship_deletes": [],
	})


func _decode(text: String) -> Dictionary:
	var result = PacketCodec.decode(text)
	assert_true(result.ok, result.error)
	return result.packet
