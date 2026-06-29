extends GutTest

const WorldLaneApplier := preload("res://scripts/protocol/realtime/world_lane_applier.gd")
const WorldLaneState := preload("res://scripts/protocol/realtime/world_lane_state.gd")
const BaselineTracker := preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")


func test_world_full_replaces_lane_and_removes_missing_entities() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	world_lane_state.upsert_ship(_ship_packet("ship-1", 10, 20))
	world_lane_state.upsert_ship(_ship_packet("ship-2", 30, 40))

	applier.apply_world_full(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 1,
			"snapshot_id": "snapshot-1",
			"ships": [_ship_packet("ship-1", 11, 21)],
			"bullets": [_bullet_packet("bullet-1", 5, 6)],
			"asteroids": [_asteroid_packet("asteroid-1", 7, 8)],
			"pickups": [_pickup_packet("pickup-1", 9, 10)],
			"is_final_chunk": true,
		}
	)

	assert_false(world_lane_state.ships.has("ship-2"))
	assert_eq(world_lane_state.ships["ship-1"]["x"], 11)
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 5)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 7)
	assert_eq(world_lane_state.pickups["pickup-1"]["x"], 9)
	assert_false(baseline_tracker.needs_resync(LaneMetadata.LANE_WORLD))


func test_world_delta_updates_creates_and_deletes_entities() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	applier.apply_world_full(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 1,
			"ships": [_ship_packet("ship-1", 10, 20)],
			"bullets": [],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	var applied := applier.apply_world_delta(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 2,
			"ship_creates": [_ship_packet("ship-2", 30, 40)],
			"ship_updates": [_ship_packet("ship-1", 11, 21)],
			"ship_deletes": ["ship-3"],
			"bullet_creates": [_bullet_packet("bullet-1", 5, 6)],
			"bullet_updates": [],
			"bullet_deletes": [],
			"asteroid_creates": [],
			"asteroid_updates": [],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [],
			"pickup_deletes": [],
		}
	)

	assert_true(applied)
	assert_eq(world_lane_state.ships["ship-1"]["x"], 11)
	assert_eq(world_lane_state.ships["ship-2"]["x"], 30)
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 5)


func test_world_delta_missing_entities_leave_lane_unchanged() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	applier.apply_world_full(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 1,
			"ships": [_ship_packet("ship-1", 10, 20)],
			"bullets": [],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_world_delta(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 2,
			"ship_creates": [],
			"ship_updates": [],
			"ship_deletes": [],
			"bullet_creates": [],
			"bullet_updates": [],
			"bullet_deletes": [],
			"asteroid_creates": [],
			"asteroid_updates": [],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [],
			"pickup_deletes": [],
		}
	)

	assert_eq(world_lane_state.ships["ship-1"]["x"], 10)


func test_world_delta_delete_removes_entity() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	applier.apply_world_full(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 1,
			"ships": [_ship_packet("ship-1", 10, 20)],
			"bullets": [],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_world_delta(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 2,
			"ship_creates": [],
			"ship_updates": [],
			"ship_deletes": ["ship-1"],
			"bullet_creates": [],
			"bullet_updates": [],
			"bullet_deletes": [],
			"asteroid_creates": [],
			"asteroid_updates": [],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [],
			"pickup_deletes": [],
		}
	)

	assert_false(world_lane_state.ships.has("ship-1"))


func test_world_delta_rejected_when_unsynced() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	var baseline_tracker := BaselineTracker.new()

	var applied := applier.apply_world_delta(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 1,
			"ship_creates": [_ship_packet("ship-1", 10, 20)],
			"ship_updates": [],
			"ship_deletes": [],
			"bullet_creates": [],
			"bullet_updates": [],
			"bullet_deletes": [],
			"asteroid_creates": [],
			"asteroid_updates": [],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [],
			"pickup_deletes": [],
		}
	)

	assert_false(applied)
	assert_false(world_lane_state.ships.has("ship-1"))


func test_world_delta_wrong_baseline_marks_resync_needed() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	applier.apply_world_full(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 1,
			"ships": [_ship_packet("ship-1", 10, 20)],
			"bullets": [],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	var applied := applier.apply_world_delta(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-2",
			"sequence": 2,
			"ship_creates": [],
			"ship_updates": [],
			"ship_deletes": [],
			"bullet_creates": [],
			"bullet_updates": [],
			"bullet_deletes": [],
			"asteroid_creates": [],
			"asteroid_updates": [],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [],
			"pickup_deletes": [],
		}
	)

	assert_false(applied)
	assert_true(baseline_tracker.needs_resync(LaneMetadata.LANE_WORLD))


static func _ship_packet(id: String, x: int, y: int) -> Dictionary:
	return {
		"id": id,
		"x": x,
		"y": y,
		"rotation": 0.0,
		"velocity_x": 0.0,
		"velocity_y": 0.0,
		"thrusting": false,
		"health": 100,
		"shields": 0,
	}


static func _bullet_packet(id: String, x: int, y: int) -> Dictionary:
	return {
		"id": id,
		"x": x,
		"y": y,
		"velocity_x": 0.0,
		"velocity_y": 0.0,
		"rotation": 0.0,
		"owner_id": "ship-1",
		"lifespan_seconds": 1.0,
	}


static func _asteroid_packet(id: String, x: int, y: int) -> Dictionary:
	return {
		"id": id,
		"x": x,
		"y": y,
		"velocity_x": 0.0,
		"velocity_y": 0.0,
		"rotation": 0.0,
		"size": 1,
		"health": 100,
	}


static func _pickup_packet(id: String, x: int, y: int) -> Dictionary:
	return {
		"id": id,
		"x": x,
		"y": y,
		"pickup_type": "test",
	}

