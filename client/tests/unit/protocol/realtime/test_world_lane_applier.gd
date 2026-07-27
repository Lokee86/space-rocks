extends GutTest

const WorldLaneApplier := preload("res://scripts/protocol/realtime/world_lane_applier.gd")
const WorldLaneState := preload("res://scripts/protocol/realtime/world_lane_state.gd")
const BaselineTracker := preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const CompactLanePacket := preload("res://scripts/protocol/realtime/compact_lane_packet.gd")


func test_world_lane_state_preserves_team_id_for_ship_rendering() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.apply_full_lane({
		"ships": [{"id": "player-1", "team_id": "team_3"}],
	})

	assert_eq(world_lane_state.ships["player-1"]["team_id"], "team_3")


func test_world_full_replaces_lane_and_removes_missing_entities_by_ownership() -> void:
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
	assert_eq(world_lane_state.ships["ship-1"]["x"], 1.1)
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 0.5)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 0.7)
	assert_eq(world_lane_state.pickups["pickup-1"]["x"], 0.9)
	assert_false(baseline_tracker.needs_resync(LaneMetadata.LANE_WORLD))


func test_world_delta_updates_creates_and_deletes_entities_by_ownership() -> void:
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
			"bullets": [_bullet_packet("bullet-1", 5, 6)],
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
			"bullet_creates": [_bullet_packet("bullet-2", 15, 16)],
			"bullet_updates": [{"id": "bullet-1", "x": 99, "rotation": 0.0}],
			"bullet_deletes": ["bullet-2"],
			"asteroid_creates": [],
			"asteroid_updates": [],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [],
			"pickup_deletes": [],
		}
	)

	assert_true(applied)
	assert_eq(world_lane_state.ships["ship-1"]["x"], 1.1)
	assert_eq(world_lane_state.ships["ship-2"]["x"], 3.0)
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 9.9)
	assert_eq(world_lane_state.bullets["bullet-1"]["y"], 0.6)
	assert_eq(world_lane_state.bullets["bullet-1"]["rotation"], 0.0)
	assert_eq(world_lane_state.bullets["bullet-1"]["owner_id"], "ship-1")
	assert_eq(world_lane_state.bullets["bullet-1"]["weapon_id"], "bullet")
	assert_eq(world_lane_state.bullets["bullet-1"]["projectile_type"], "bullet")
	assert_false(world_lane_state.bullets.has("bullet-2"))


func test_world_delta_treats_missing_sparse_sections_as_empty_noop() -> void:
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
			"bullets": [_bullet_packet("bullet-1", 5, 6)],
			"asteroids": [_asteroid_packet("asteroid-1", 7, 8)],
			"pickups": [_pickup_packet("pickup-1", 9, 10)],
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
		}
	)

	assert_true(applied)
	assert_eq(world_lane_state.ships["ship-1"]["x"], 1.0)
	assert_eq(world_lane_state.ships["ship-1"]["y"], 2.0)
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 0.5)
	assert_eq(world_lane_state.bullets["bullet-1"]["y"], 0.6)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 0.7)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["y"], 0.8)
	assert_eq(world_lane_state.pickups["pickup-1"]["x"], 0.9)
	assert_eq(world_lane_state.pickups["pickup-1"]["y"], 1.0)


func test_world_delta_missing_entities_leave_lane_unchanged_by_ownership() -> void:
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

	assert_eq(world_lane_state.ships["ship-1"]["x"], 1.0)


func test_world_delta_delete_removes_entity_by_ownership() -> void:
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


func test_world_delta_treats_null_arrays_as_empty_and_applies_deletes() -> void:
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
			"bullets": [_bullet_packet("bullet-1", 5, 6)],
			"asteroids": [_asteroid_packet("asteroid-1", 7, 8)],
			"pickups": [_pickup_packet("pickup-1", 9, 10)],
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
			"ship_creates": null,
			"ship_updates": null,
			"ship_deletes": ["ship-1"],
			"bullet_creates": null,
			"bullet_updates": null,
			"bullet_deletes": ["bullet-1"],
			"asteroid_creates": null,
			"asteroid_updates": null,
			"asteroid_deletes": ["asteroid-1"],
			"pickup_creates": null,
			"pickup_updates": null,
			"pickup_deletes": ["pickup-1"],
		}
	)

	assert_true(applied)
	assert_false(world_lane_state.ships.has("ship-1"))
	assert_false(world_lane_state.bullets.has("bullet-1"))
	assert_false(world_lane_state.asteroids.has("asteroid-1"))
	assert_false(world_lane_state.pickups.has("pickup-1"))


func test_world_delta_merges_partial_asteroid_and_pickup_updates_by_ownership() -> void:
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
			"ships": [],
			"bullets": [],
			"asteroids": [_asteroid_packet("asteroid-1", 7, 8)],
			"pickups": [_pickup_packet("pickup-1", 9, 10)],
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
			"ship_creates": [],
			"ship_updates": [],
			"ship_deletes": [],
			"bullet_creates": [],
			"bullet_updates": [],
			"bullet_deletes": [],
			"asteroid_creates": [],
			"asteroid_updates": [{"id": "asteroid-1", "x": 99}],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [{"id": "pickup-1", "x": 88, "age_seconds": 0}],
			"pickup_deletes": [],
		}
	)

	assert_true(applied)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 9.9)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["y"], 0.8)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["size"], 1)
	assert_eq(world_lane_state.pickups["pickup-1"]["x"], 8.8)
	assert_eq(world_lane_state.pickups["pickup-1"]["y"], 1.0)
	assert_eq(world_lane_state.pickups["pickup-1"]["type"], "test")
	assert_eq(world_lane_state.pickups["pickup-1"]["age_seconds"], 0.0)


func test_world_delta_keeps_rehydrated_asteroid_id_stable_across_lane_application() -> void:
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
			"ships": [],
			"bullets": [],
			"asteroids": [{
				"id": "asteroid-1",
				"x": 7,
				"y": 8,
				"velocity_x": 0.0,
				"velocity_y": 0.0,
				"rotation": 0.0,
				"size": 2,
				"health": 90,
				"scale": 1500,
				"variant": 3,
			}],
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
			"ship_creates": [],
			"ship_updates": [],
			"ship_deletes": [],
			"bullet_creates": [],
			"bullet_updates": [],
			"bullet_deletes": [],
			"asteroid_creates": [],
			"asteroid_updates": [{"id": "asteroid-1", "x": 99, "y": 101}],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [],
			"pickup_deletes": [],
		}
	)

	assert_true(applied)
	assert_true(world_lane_state.asteroids.has("asteroid-1"))
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 9.9)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["y"], 10.1)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["size"], 2)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["health"], 90)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["scale"], 1.5)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["variant"], 3)

func test_world_delta_applies_valid_arrays_normally() -> void:
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
			"ship_deletes": [],
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
	assert_eq(world_lane_state.ships["ship-1"]["x"], 1.1)
	assert_eq(world_lane_state.ships["ship-2"]["x"], 3.0)
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 0.5)


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


func test_world_full_preserves_bullet_projectile_type_for_torpedo_presentation() -> void:
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
			"ships": [],
			"bullets": [_bullet_packet("bullet-torpedo", 5, 6, "torpedo")],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	assert_eq(world_lane_state.bullets["bullet-torpedo"]["projectile_type"], "torpedo")
	assert_eq(world_lane_state.bullets["bullet-torpedo"]["weapon_id"], "torpedo")



func test_world_lane_applier_accepts_tuple_expanded_compact_records() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	var expanded := CompactLanePacket.expand_packet({
		"t": "wf",
		"ships": [[1, "v_wing", 10, 20, 30, 100, 50, true, "player", "player-2"]],
		"bullets": [[1, "player-1", 5, 6, 7, "pulse", "laser"]],
		"asteroids": [],
		"pickups": [],
	})
	expanded["baseline_id"] = "baseline-1"
	expanded["sequence"] = 1
	expanded["chunk_index"] = 0
	expanded["chunk_count"] = 1
	expanded["is_final_chunk"] = true

	applier.apply_world_full(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		expanded
	)

	assert_eq(world_lane_state.ships["player-1"]["ship_type"], "v_wing")
	assert_eq(world_lane_state.ships["player-1"]["target_id"], "player-2")
	assert_eq(world_lane_state.bullets["bullet-1"]["owner_id"], "player-1")
	assert_eq(world_lane_state.bullets["bullet-1"]["weapon_id"], "pulse")
	assert_eq(world_lane_state.bullets["bullet-1"]["projectile_type"], "laser")


func test_asteroids_lifecycle_applies_valid_creates_and_deletes() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_asteroid({"id": "asteroid-2", "x": 3, "y": 4, "velocity_x": 0.0, "velocity_y": 0.0, "rotation": 0.0, "size": 1, "health": 100, "scale": 1, "variant": 1})

	var applied := applier.apply_asteroids_lifecycle(
		world_lane_state,
		{
			"type": "asteroids_lifecycle",
			"lane": LaneMetadata.LANE_ASTEROIDS_LIFECYCLE,
			"asteroid_creates": [{"id": "asteroid-1", "x": 10, "y": 20, "velocity_x": 0.0, "velocity_y": 0.0, "rotation": 0.0, "size": 2, "health": 90, "scale": 1500, "variant": 3}],
			"asteroid_deletes": ["asteroid-2"],
		}
	)

	assert_true(applied)
	assert_true(world_lane_state.asteroids.has("asteroid-1"))
	assert_false(world_lane_state.asteroids.has("asteroid-2"))
	assert_eq(world_lane_state.asteroids["asteroid-1"]["variant"], 3)


func test_bullets_lifecycle_applies_valid_creates_and_deletes() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_bullet({"id": "bullet-2", "owner_id": "player-1", "x": 3, "y": 4, "velocity_x": 0.0, "velocity_y": 0.0, "rotation": 0, "lifespan_seconds": 1.0, "weapon_id": "bullet", "projectile_type": "bullet"})

	var applied := applier.apply_bullets_lifecycle(
		world_lane_state,
		{
			"type": "bullets_lifecycle",
			"lane": LaneMetadata.LANE_BULLETS_LIFECYCLE,
			"bullet_creates": [{"id": "bullet-1", "owner_id": "player-1", "x": 10, "y": 20, "velocity_x": 0.0, "velocity_y": 0.0, "rotation": 30, "lifespan_seconds": 1.0, "weapon_id": "torpedo", "projectile_type": "torpedo"}],
			"bullet_deletes": ["bullet-2"],
		}
	)

	assert_true(applied)
	assert_true(world_lane_state.bullets.has("bullet-1"))
	assert_false(world_lane_state.bullets.has("bullet-2"))
	assert_eq(world_lane_state.bullets["bullet-1"]["projectile_type"], "torpedo")


func test_bullets_lifecycle_applies_buffered_hot_update_after_create_and_preserves_projectile_type() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_BULLETS, {"sequence": 1, "bullet_updates": [{"id": "bullet-1", "x": 55, "y": 66, "rotation": 30}]})
	assert_false(world_lane_state.bullets.has("bullet-1"))
	assert_true(world_lane_state.pending_bullet_updates.has("bullet-1"))

	var applied := applier.apply_bullets_lifecycle(
		world_lane_state,
		{
			"type": "bullets_lifecycle",
			"lane": LaneMetadata.LANE_BULLETS_LIFECYCLE,
			"bullet_creates": [{"id": "bullet-1", "owner_id": "player-1", "x": 10, "y": 20, "velocity_x": 0.0, "velocity_y": 0.0, "rotation": 0, "lifespan_seconds": 1.0, "weapon_id": "torpedo", "projectile_type": "torpedo"}],
			"bullet_deletes": [],
		}
	)

	assert_true(applied)
	assert_true(world_lane_state.bullets.has("bullet-1"))
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 5.5)
	assert_eq(world_lane_state.bullets["bullet-1"]["y"], 6.6)
	assert_eq(world_lane_state.bullets["bullet-1"]["rotation"], 0.03)
	assert_eq(world_lane_state.bullets["bullet-1"]["projectile_type"], "torpedo")
	assert_false(world_lane_state.pending_bullet_updates.has("bullet-1"))


func test_asteroids_lifecycle_applies_buffered_hot_update_after_create() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_ASTEROIDS, {"sequence": 1, "asteroid_updates": [{"id": "asteroid-1", "x": 55, "y": 66, "rotation": 30}]})
	assert_false(world_lane_state.asteroids.has("asteroid-1"))
	assert_true(world_lane_state.pending_asteroid_updates.has("asteroid-1"))

	var applied := applier.apply_asteroids_lifecycle(
		world_lane_state,
		{
			"type": "asteroids_lifecycle",
			"lane": LaneMetadata.LANE_ASTEROIDS_LIFECYCLE,
			"asteroid_creates": [{"id": "asteroid-1", "x": 10, "y": 20, "velocity_x": 0.0, "velocity_y": 0.0, "rotation": 0.0, "size": 2, "health": 90, "scale": 1500, "variant": 3}],
			"asteroid_deletes": [],
		}
	)

	assert_true(applied)
	assert_true(world_lane_state.asteroids.has("asteroid-1"))
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 5.5)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["y"], 6.6)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["rotation"], 30)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["variant"], 3)
	assert_eq(world_lane_state.asteroid_dirty_sources["asteroid-1"], "lifecycle_create")
	assert_false(world_lane_state.pending_asteroid_updates.has("asteroid-1"))


func test_asteroids_lifecycle_delete_clears_pending_update_and_blocks_late_hot_update() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_ASTEROIDS, {"sequence": 1, "asteroid_updates": [{"id": "asteroid-1", "x": 55, "y": 66}]})
	assert_true(world_lane_state.pending_asteroid_updates.has("asteroid-1"))

	assert_true(applier.apply_asteroids_lifecycle(world_lane_state, {"type": "asteroids_lifecycle", "lane": LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, "asteroid_creates": [], "asteroid_deletes": ["asteroid-1"]}))
	assert_false(world_lane_state.pending_asteroid_updates.has("asteroid-1"))
	assert_true(world_lane_state.deleted_asteroid_ids.has("asteroid-1"))

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_ASTEROIDS, {"sequence": 2, "asteroid_updates": [{"id": "asteroid-1", "x": 77, "y": 88}]})
	assert_false(world_lane_state.pending_asteroid_updates.has("asteroid-1"))
	assert_false(world_lane_state.asteroids.has("asteroid-1"))


func test_bullets_lifecycle_delete_removes_bullet_and_clears_pending_update() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_BULLETS, {"sequence": 1, "bullet_updates": [{"id": "bullet-1", "x": 55, "y": 66, "rotation": 30}]})
	var applied := applier.apply_bullets_lifecycle(
		world_lane_state,
		{
			"type": "bullets_lifecycle",
			"lane": LaneMetadata.LANE_BULLETS_LIFECYCLE,
			"bullet_creates": [{"id": "bullet-1", "owner_id": "player-1", "x": 10, "y": 20, "velocity_x": 0.0, "velocity_y": 0.0, "rotation": 0, "lifespan_seconds": 1.0, "weapon_id": "torpedo", "projectile_type": "torpedo"}],
			"bullet_deletes": ["bullet-1"],
		}
	)

	assert_true(applied)
	assert_false(world_lane_state.bullets.has("bullet-1"))
	assert_false(world_lane_state.pending_bullet_updates.has("bullet-1"))


func test_asteroids_lifecycle_delete_removes_asteroid() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()

	assert_true(applier.apply_asteroids_lifecycle(world_lane_state, {"type": "asteroids_lifecycle", "lane": LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, "asteroid_creates": [{"id": "asteroid-1", "x": 10, "y": 20, "velocity_x": 0.0, "velocity_y": 0.0, "rotation": 0.0, "size": 2, "health": 90, "scale": 1500, "variant": 3}], "asteroid_deletes": []}))
	assert_true(applier.apply_asteroids_lifecycle(world_lane_state, {"type": "asteroids_lifecycle", "lane": LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, "asteroid_creates": [], "asteroid_deletes": ["asteroid-1"]}))

	assert_false(world_lane_state.asteroids.has("asteroid-1"))


func test_world_lane_state_merges_bullet_updates_without_dropping_omitted_fields() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_bullet({
		"id": "bullet-1",
		"owner_id": "ship-1",
		"x": 5,
		"y": 6,
		"rotation": 7,
		"weapon_id": "pulse",
		"projectile_type": "laser",
	})

	world_lane_state.merge_bullet_update({"id": "bullet-1", "x": 99})

	assert_eq(world_lane_state.bullets["bullet-1"]["owner_id"], "ship-1")
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 99)
	assert_eq(world_lane_state.bullets["bullet-1"]["y"], 6)
	assert_eq(world_lane_state.bullets["bullet-1"]["rotation"], 7)
	assert_eq(world_lane_state.bullets["bullet-1"]["weapon_id"], "pulse")
	assert_eq(world_lane_state.bullets["bullet-1"]["projectile_type"], "laser")


func test_world_lane_state_merges_bullet_updates_with_provided_fields() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_bullet({
		"id": "bullet-1",
		"owner_id": "ship-1",
		"x": 5,
		"y": 6,
		"rotation": 7,
		"weapon_id": "pulse",
		"projectile_type": "laser",
	})

	world_lane_state.merge_bullet_update({"id": "bullet-1", "x": 99, "y": 100, "rotation": 101})

	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 99)
	assert_eq(world_lane_state.bullets["bullet-1"]["y"], 100)
	assert_eq(world_lane_state.bullets["bullet-1"]["rotation"], 101)


func test_world_lane_state_ignores_unknown_bullet_ids_for_merge() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.merge_bullet_update({"id": "bullet-unknown", "x": 99})

	assert_false(world_lane_state.bullets.has("bullet-unknown"))


func test_world_lane_state_ignores_bullet_updates_without_id() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_bullet({
		"id": "bullet-1",
		"owner_id": "ship-1",
		"x": 5,
		"y": 6,
		"rotation": 7,
		"weapon_id": "pulse",
		"projectile_type": "laser",
	})

	world_lane_state.merge_bullet_update({"x": 99})

	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 5)
	assert_eq(world_lane_state.bullets["bullet-1"]["owner_id"], "ship-1")
	assert_eq(world_lane_state.bullets["bullet-1"]["weapon_id"], "pulse")
	assert_eq(world_lane_state.bullets["bullet-1"]["projectile_type"], "laser")


func test_world_lane_state_merges_bullet_updates_with_zero_rotation_and_preserved_projectile_type() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_bullet({
		"id": "bullet-1",
		"owner_id": "ship-1",
		"x": 5,
		"y": 6,
		"rotation": 7,
		"weapon_id": "pulse",
		"projectile_type": "laser",
	})

	world_lane_state.merge_bullet_update({"id": "bullet-1", "rotation": 0})

	assert_eq(world_lane_state.bullets["bullet-1"]["rotation"], 0)
	assert_eq(world_lane_state.bullets["bullet-1"]["projectile_type"], "laser")


func test_world_lane_state_merges_ship_updates_without_dropping_omitted_fields() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_ship({
		"id": "ship-1",
		"x": 10,
		"y": 20,
		"rotation": 30,
		"velocity_x": 1.5,
		"velocity_y": 2.5,
		"thrusting": true,
		"health": 90,
		"shields": 15,
		"ship_type": "v_wing",
		"target_kind": "player",
		"target_id": "player-2",
	})

	world_lane_state.merge_ship_update({"id": "ship-1", "x": 99})

	assert_eq(world_lane_state.ships["ship-1"]["x"], 99)
	assert_eq(world_lane_state.ships["ship-1"]["y"], 20)
	assert_eq(world_lane_state.ships["ship-1"]["rotation"], 30)
	assert_eq(world_lane_state.ships["ship-1"]["health"], 90)
	assert_eq(world_lane_state.ships["ship-1"]["shields"], 15)
	assert_eq(world_lane_state.ships["ship-1"]["ship_type"], "v_wing")
	assert_eq(world_lane_state.ships["ship-1"]["target_kind"], "player")
	assert_eq(world_lane_state.ships["ship-1"]["target_id"], "player-2")


func test_world_lane_state_merges_ship_updates_with_provided_fields() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_ship({
		"id": "ship-1",
		"x": 10,
		"y": 20,
		"rotation": 30,
		"velocity_x": 1.5,
		"velocity_y": 2.5,
		"thrusting": false,
		"health": 90,
		"shields": 15,
		"ship_type": "v_wing",
		"target_kind": "player",
		"target_id": "player-2",
	})

	world_lane_state.merge_ship_update({"id": "ship-1", "x": 99, "y": 101, "rotation": 0.0, "thrusting": false})

	assert_eq(world_lane_state.ships["ship-1"]["x"], 99)
	assert_eq(world_lane_state.ships["ship-1"]["y"], 101)
	assert_eq(world_lane_state.ships["ship-1"]["rotation"], 0.0)
	assert_eq(world_lane_state.ships["ship-1"]["thrusting"], false)


func test_world_lane_state_ignores_unknown_ship_ids_for_merge() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.merge_ship_update({"id": "ship-unknown", "x": 99})

	assert_false(world_lane_state.ships.has("ship-unknown"))


func test_world_lane_state_ignores_ship_updates_without_id() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_ship({
		"id": "ship-1",
		"x": 10,
		"y": 20,
		"rotation": 30,
		"velocity_x": 1.5,
		"velocity_y": 2.5,
		"thrusting": false,
		"health": 90,
		"shields": 15,
		"ship_type": "v_wing",
		"target_kind": "player",
		"target_id": "player-2",
	})

	world_lane_state.merge_ship_update({"x": 99})

	assert_eq(world_lane_state.ships["ship-1"]["x"], 10)
	assert_eq(world_lane_state.ships["ship-1"]["ship_type"], "v_wing")


func test_world_lane_state_creates_full_ship_records() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_ship({
		"id": "ship-1",
		"x": 10,
		"y": 20,
		"rotation": 30,
		"velocity_x": 1.5,
		"velocity_y": 2.5,
		"thrusting": true,
		"health": 90,
		"shields": 15,
		"ship_type": "v_wing",
		"target_kind": "player",
		"target_id": "player-2",
	})

	assert_eq(world_lane_state.ships["ship-1"]["x"], 10)
	assert_eq(world_lane_state.ships["ship-1"]["y"], 20)
	assert_eq(world_lane_state.ships["ship-1"]["rotation"], 30)
	assert_eq(world_lane_state.ships["ship-1"]["thrusting"], true)
	assert_eq(world_lane_state.ships["ship-1"]["ship_type"], "v_wing")
	assert_eq(world_lane_state.ships["ship-1"]["target_kind"], "player")
	assert_eq(world_lane_state.ships["ship-1"]["target_id"], "player-2")


func test_world_lane_state_deletes_ship_records() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_ship({
		"id": "ship-1",
		"x": 10,
		"y": 20,
		"rotation": 30,
		"velocity_x": 1.5,
		"velocity_y": 2.5,
		"thrusting": true,
		"health": 90,
		"shields": 15,
		"ship_type": "v_wing",
		"target_kind": "player",
		"target_id": "player-2",
	})

	world_lane_state.delete_ship("ship-1")

	assert_false(world_lane_state.ships.has("ship-1"))
func test_world_lane_state_merges_asteroid_updates_without_dropping_omitted_fields() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_asteroid({
		"id": "asteroid-1",
		"x": 10,
		"y": 20,
		"velocity_x": 1.5,
		"velocity_y": 2.5,
		"rotation": 30,
		"size": 4,
		"health": 90,
		"scale": 2,
		"variant": "rock",
	})

	world_lane_state.merge_asteroid_update({"id": "asteroid-1", "x": 99})

	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 99)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["y"], 20)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["size"], 4)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["health"], 90)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["scale"], 2)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["variant"], "rock")


func test_world_lane_state_merges_asteroid_updates_with_provided_fields() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_asteroid({
		"id": "asteroid-1",
		"x": 10,
		"y": 20,
		"velocity_x": 1.5,
		"velocity_y": 2.5,
		"rotation": 30,
		"size": 4,
		"health": 90,
		"scale": 2,
		"variant": "rock",
	})

	world_lane_state.merge_asteroid_update({"id": "asteroid-1", "x": 99, "y": 101, "size": 7})

	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 99)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["y"], 101)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["size"], 7)


func test_world_lane_state_applies_zero_value_asteroid_updates() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_asteroid({
		"id": "asteroid-1",
		"x": 10,
		"y": 20,
		"velocity_x": 1.5,
		"velocity_y": 2.5,
		"rotation": 30,
		"size": 4,
		"health": 90,
		"scale": 2,
		"variant": "rock",
	})

	world_lane_state.merge_asteroid_update({"id": "asteroid-1", "rotation": 0, "scale": 0})

	assert_eq(world_lane_state.asteroids["asteroid-1"]["rotation"], 0)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["scale"], 0)


func test_world_lane_state_ignores_unknown_asteroid_ids_for_merge() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.merge_asteroid_update({"id": "asteroid-unknown", "x": 99})

	assert_false(world_lane_state.asteroids.has("asteroid-unknown"))


func test_world_lane_state_merges_pickup_updates_without_dropping_omitted_fields() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_pickup({
		"id": "pickup-1",
		"type": "shield",
		"pickup_class": "powerup",
		"x": 10,
		"y": 20,
		"health": 90,
		"age_seconds": 4,
		"lifespan_seconds": 12,
	})

	world_lane_state.merge_pickup_update({"id": "pickup-1", "x": 99})

	assert_eq(world_lane_state.pickups["pickup-1"]["x"], 99)
	assert_eq(world_lane_state.pickups["pickup-1"]["y"], 20)
	assert_eq(world_lane_state.pickups["pickup-1"]["type"], "shield")
	assert_eq(world_lane_state.pickups["pickup-1"]["pickup_class"], "powerup")
	assert_eq(world_lane_state.pickups["pickup-1"]["health"], 90)
	assert_eq(world_lane_state.pickups["pickup-1"]["age_seconds"], 4)
	assert_eq(world_lane_state.pickups["pickup-1"]["lifespan_seconds"], 12)


func test_world_lane_state_merges_pickup_updates_with_provided_fields() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_pickup({
		"id": "pickup-1",
		"type": "shield",
		"pickup_class": "powerup",
		"x": 10,
		"y": 20,
		"health": 90,
		"age_seconds": 4,
		"lifespan_seconds": 12,
	})

	world_lane_state.merge_pickup_update({"id": "pickup-1", "x": 99, "y": 101, "age_seconds": 0})

	assert_eq(world_lane_state.pickups["pickup-1"]["x"], 99)
	assert_eq(world_lane_state.pickups["pickup-1"]["y"], 101)
	assert_eq(world_lane_state.pickups["pickup-1"]["age_seconds"], 0)


func test_world_lane_state_applies_zero_value_pickup_updates() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_pickup({
		"id": "pickup-1",
		"type": "shield",
		"pickup_class": "powerup",
		"x": 10,
		"y": 20,
		"health": 90,
		"age_seconds": 4,
		"lifespan_seconds": 12,
	})

	world_lane_state.merge_pickup_update({"id": "pickup-1", "health": 0, "lifespan_seconds": 0})

	assert_eq(world_lane_state.pickups["pickup-1"]["health"], 0)
	assert_eq(world_lane_state.pickups["pickup-1"]["lifespan_seconds"], 0)


func test_world_lane_state_ignores_unknown_pickup_ids_for_merge() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.merge_pickup_update({"id": "pickup-unknown", "x": 99})

	assert_false(world_lane_state.pickups.has("pickup-unknown"))


func test_world_lane_state_preserves_ship_target_fields() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_ship({
		"id": "ship-1",
		"x": 10,
		"y": 20,
		"rotation": 0.0,
		"velocity_x": 0.0,
		"velocity_y": 0.0,
		"thrusting": false,
		"health": 100,
		"shields": 0,
		"ship_type": "v_wing",
		"target_kind": "player",
		"target_id": "player-2",
	})

	assert_eq(world_lane_state.ships["ship-1"]["ship_type"], "v_wing")
	assert_eq(world_lane_state.ships["ship-1"]["target_kind"], "player")
	assert_eq(world_lane_state.ships["ship-1"]["target_id"], "player-2")


func test_bullet_tombstones_are_bounded_and_recreated_ids_retain_capacity() -> void:
	var world_lane_state := WorldLaneState.new()

	for index in range(WorldLaneState.DELETED_BULLET_ID_CAP + 1):
		world_lane_state.delete_bullet("bullet-%d" % index)

	assert_eq(world_lane_state.deleted_bullet_ids.size(), WorldLaneState.DELETED_BULLET_ID_CAP)
	assert_false(world_lane_state.deleted_bullet_ids.has("bullet-0"))
	assert_true(world_lane_state.deleted_bullet_ids.has("bullet-%d" % WorldLaneState.DELETED_BULLET_ID_CAP))

	var retained_id := "bullet-%d" % WorldLaneState.DELETED_BULLET_ID_CAP
	world_lane_state.upsert_bullet(_bullet_packet(retained_id, 5, 6))
	assert_false(world_lane_state.deleted_bullet_ids.has(retained_id))
	world_lane_state.delete_bullet(retained_id)
	assert_true(world_lane_state.deleted_bullet_ids.has(retained_id))
	assert_eq(world_lane_state.deleted_bullet_ids.size(), WorldLaneState.DELETED_BULLET_ID_CAP)


func test_pending_bullet_updates_are_bounded_in_insertion_order() -> void:
	var world_lane_state := WorldLaneState.new()

	for index in range(WorldLaneState.PENDING_BULLET_UPDATE_CAP + 1):
		world_lane_state.merge_or_buffer_bullet_update({"id": "bullet-%d" % index, "x": index})

	assert_eq(world_lane_state.pending_bullet_updates.size(), WorldLaneState.PENDING_BULLET_UPDATE_CAP)
	assert_false(world_lane_state.pending_bullet_updates.has("bullet-0"))
	assert_true(world_lane_state.pending_bullet_updates.has("bullet-%d" % WorldLaneState.PENDING_BULLET_UPDATE_CAP))

	world_lane_state.merge_or_buffer_bullet_update({"id": "bullet-1", "x": 1000})
	world_lane_state.merge_or_buffer_bullet_update({"id": "bullet-after", "x": 2000})

	assert_false(world_lane_state.pending_bullet_updates.has("bullet-1"))
	assert_true(world_lane_state.pending_bullet_updates.has("bullet-2"))
	assert_true(world_lane_state.pending_bullet_updates.has("bullet-after"))
	assert_eq(world_lane_state.pending_bullet_updates.size(), WorldLaneState.PENDING_BULLET_UPDATE_CAP)

	world_lane_state.clear_pending_bullet_updates()
	assert_true(world_lane_state.pending_bullet_updates.is_empty())
	assert_true(world_lane_state._pending_bullet_update_order.is_empty())


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


static func _bullet_packet(id: String, x: int, y: int, projectile_type: String = "bullet") -> Dictionary:
	return {
		"id": id,
		"x": x,
		"y": y,
		"velocity_x": 0.0,
		"velocity_y": 0.0,
		"rotation": 0.0,
		"owner_id": "ship-1",
		"lifespan_seconds": 1.0,
		"weapon_id": projectile_type,
		"projectile_type": projectile_type,
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
		"type": "test",
	}














func test_asteroid_delta_updates_existing_asteroid_only() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_asteroid({
		"id": "asteroid-1",
		"x": 1,
		"y": 2,
		"size": 3,
		"health": 4,
		"scale": 5,
		"variant": 6,
	})

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"asteroid_updates": [{"id": "asteroid-1", "x": 9, "y": 10}],
	})

	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 0.9)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["y"], 1.0)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["size"], 3)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["variant"], 6)


func test_asteroid_delta_rejects_stale_sequence_and_accepts_gaps() -> void:
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
			"ships": [],
			"bullets": [],
			"asteroids": [_asteroid_packet("asteroid-1", 10, 10)],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 10,
		"asteroid_updates": [{"id": "asteroid-1", "x": 20}],
	})
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 2.0)

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 9,
		"asteroid_updates": [{"id": "asteroid-1", "x": 5}],
	})
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 2.0)

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 11,
		"asteroid_updates": [{"id": "asteroid-1", "x": 30}],
	})
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 3.0)


func test_asteroid_delta_ignores_missing_and_string_sequences() -> void:
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
			"ships": [],
			"bullets": [],
			"asteroids": [_asteroid_packet("asteroid-1", 10, 10)],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 12,
		"asteroid_updates": [{"id": "asteroid-1", "x": 20}],
	})
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 2.0)

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"asteroid_updates": [{"id": "asteroid-1", "x": 50}],
	})
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 2.0)

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": "not-a-number",
		"asteroid_updates": [{"id": "asteroid-1", "x": 60}],
	})
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 2.0)


func test_asteroid_delta_ignores_unknown_asteroid_updates() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_asteroid({"id": "asteroid-1", "x": 1, "y": 2, "size": 3, "health": 4, "scale": 5, "variant": 6})

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"asteroid_updates": [{"id": "asteroid-unknown", "x": 9, "y": 10}],
	})

	assert_false(world_lane_state.asteroids.has("asteroid-unknown"))
	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 1)


func test_asteroid_delta_ignores_missing_asteroid_updates() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_asteroid({"id": "asteroid-1", "x": 1, "y": 2, "size": 3, "health": 4, "scale": 5, "variant": 6})

	applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"asteroid_updates": [{"x": 9, "y": 10}],
	})

	assert_eq(world_lane_state.asteroids["asteroid-1"]["x"], 1)
	assert_eq(world_lane_state.asteroids["asteroid-1"]["y"], 2)

func test_bullet_delta_updates_existing_bullet_only() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_bullet({
		"id": "bullet-1",
		"owner_id": "ship-1",
		"x": 1,
		"y": 2,
		"rotation": 3,
		"weapon_id": "pulse",
		"projectile_type": "laser",
	})

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"bullet_updates": [{"id": "bullet-1", "x": 9, "y": 10, "rotation": 11}],
	})

	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 0.9)
	assert_eq(world_lane_state.bullets["bullet-1"]["y"], 1.0)
	assert_eq(world_lane_state.bullets["bullet-1"]["rotation"], 0.011)
	assert_eq(world_lane_state.bullets["bullet-1"]["weapon_id"], "pulse")


func test_bullet_delta_rejects_stale_sequence_and_accepts_gaps() -> void:
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
			"ships": [],
			"bullets": [_bullet_packet("bullet-1", 10, 10)],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 10,
		"bullet_updates": [{"id": "bullet-1", "x": 20}],
	})
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 2.0)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 9,
		"bullet_updates": [{"id": "bullet-1", "x": 5}],
	})
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 2.0)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 12,
		"bullet_updates": [{"id": "bullet-1", "x": 40}],
	})
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 4.0)


func test_bullet_delta_ignores_missing_and_string_sequences() -> void:
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
			"ships": [],
			"bullets": [_bullet_packet("bullet-1", 10, 10)],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 12,
		"bullet_updates": [{"id": "bullet-1", "x": 20}],
	})
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 2.0)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"bullet_updates": [{"id": "bullet-1", "x": 50}],
	})
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 2.0)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": "not-a-number",
		"bullet_updates": [{"id": "bullet-1", "x": 60}],
	})
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 2.0)


func test_bullet_delta_ignores_unknown_bullet_updates() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_bullet({"id": "bullet-1", "owner_id": "ship-1", "x": 1, "y": 2, "rotation": 3, "weapon_id": "pulse", "projectile_type": "laser"})

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"bullet_updates": [{"id": "bullet-unknown", "x": 9, "y": 10, "rotation": 11}],
	})

	assert_false(world_lane_state.bullets.has("bullet-unknown"))
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 1)


func test_bullet_delta_ignores_missing_bullet_updates() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_bullet({"id": "bullet-1", "owner_id": "ship-1", "x": 1, "y": 2, "rotation": 3, "weapon_id": "pulse", "projectile_type": "laser"})

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"bullet_updates": [{"x": 9, "y": 10, "rotation": 11}],
	})

	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 1)
	assert_eq(world_lane_state.bullets["bullet-1"]["y"], 2)


func test_bullet_delta_buffers_unknown_bullet_update_without_creating_bullet() -> void:
	var applier := WorldLaneApplier.new()
	var world_lane_state := WorldLaneState.new()

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"bullet_updates": [{"id": "bullet-1", "x": 30, "y": 40, "rotation": 50}],
	})

	assert_false(world_lane_state.bullets.has("bullet-1"))
	assert_true(world_lane_state.pending_bullet_updates.has("bullet-1"))


func test_world_lane_state_marks_dirty_bullet_on_create_and_update() -> void:
	var world_lane_state := WorldLaneState.new()

	world_lane_state.upsert_bullet(_bullet_packet("bullet-1", 10, 20))

	assert_true(world_lane_state.dirty_bullet_ids.has("bullet-1"))
	assert_false(world_lane_state.removed_bullet_ids.has("bullet-1"))

	world_lane_state.clear_bullet_change_sets()
	world_lane_state.merge_bullet_update({"id": "bullet-1", "x": 30, "y": 40})

	assert_true(world_lane_state.dirty_bullet_ids.has("bullet-1"))
	assert_false(world_lane_state.removed_bullet_ids.has("bullet-1"))


func test_world_lane_state_marks_removed_bullet_on_delete() -> void:
	var world_lane_state := WorldLaneState.new()

	world_lane_state.upsert_bullet(_bullet_packet("bullet-1", 10, 20))
	world_lane_state.clear_bullet_change_sets()
	world_lane_state.delete_bullet("bullet-1")

	assert_true(world_lane_state.removed_bullet_ids.has("bullet-1"))
	assert_false(world_lane_state.dirty_bullet_ids.has("bullet-1"))


func test_world_lane_state_bullet_full_replace_requires_full_sync() -> void:
	var world_lane_state := WorldLaneState.new()

	world_lane_state.replace_bullets([_bullet_packet("bullet-1", 10, 20)])

	assert_true(world_lane_state.bullet_full_sync_required)
	assert_false(world_lane_state.dirty_bullet_ids.has("bullet-1"))
	assert_false(world_lane_state.removed_bullet_ids.has("bullet-1"))


func test_world_lane_state_marks_dirty_asteroid_on_create_and_update() -> void:
	var world_lane_state := WorldLaneState.new()

	world_lane_state.upsert_asteroid(_asteroid_packet("asteroid-1", 10, 20))

	assert_true(world_lane_state.dirty_asteroid_ids.has("asteroid-1"))
	assert_false(world_lane_state.removed_asteroid_ids.has("asteroid-1"))

	world_lane_state.clear_asteroid_change_sets()
	world_lane_state.merge_asteroid_update({"id": "asteroid-1", "x": 30, "y": 40})

	assert_true(world_lane_state.dirty_asteroid_ids.has("asteroid-1"))
	assert_false(world_lane_state.removed_asteroid_ids.has("asteroid-1"))


func test_world_lane_state_marks_removed_asteroid_on_delete() -> void:
	var world_lane_state := WorldLaneState.new()

	world_lane_state.upsert_asteroid(_asteroid_packet("asteroid-1", 10, 20))
	world_lane_state.clear_asteroid_change_sets()
	world_lane_state.delete_asteroid("asteroid-1")

	assert_true(world_lane_state.removed_asteroid_ids.has("asteroid-1"))
	assert_false(world_lane_state.dirty_asteroid_ids.has("asteroid-1"))


func test_world_lane_state_asteroid_full_replace_requires_full_sync() -> void:
	var world_lane_state := WorldLaneState.new()

	world_lane_state.replace_asteroids([_asteroid_packet("asteroid-1", 10, 20)])

	assert_true(world_lane_state.asteroid_full_sync_required)
	assert_true(world_lane_state.dirty_asteroid_ids.is_empty())
	assert_true(world_lane_state.removed_asteroid_ids.is_empty())


func test_bullet_create_applies_pending_bullet_update() -> void:
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
			"ships": [],
			"bullets": [],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"bullet_updates": [{"id": "bullet-1", "x": 30, "y": 40, "rotation": 50}],
	})

	var applied := applier.apply_world_delta(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 2,
			"ship_creates": [],
			"ship_updates": [],
			"ship_deletes": [],
			"bullet_creates": [{"id": "bullet-1", "x": 10, "y": 20, "rotation": 25, "velocity_x": 0.0, "velocity_y": 0.0, "owner_id": "ship-1", "lifespan_seconds": 1.0, "weapon_id": "pulse", "projectile_type": "laser"}],
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
	assert_true(world_lane_state.bullets.has("bullet-1"))
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 3.0)
	assert_eq(world_lane_state.bullets["bullet-1"]["y"], 4.0)
	assert_eq(world_lane_state.bullets["bullet-1"]["rotation"], 0.05)
	assert_false(world_lane_state.pending_bullet_updates.has("bullet-1"))

func test_latest_pending_bullet_update_wins_before_create() -> void:
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
			"ships": [],
			"bullets": [],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"bullet_updates": [{"id": "bullet-1", "x": 30, "y": 40, "rotation": 50}],
	})
	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 2,
		"bullet_updates": [{"id": "bullet-1", "x": 60, "y": 70, "rotation": 80}],
	})

	var applied := applier.apply_world_delta(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 2,
			"ship_creates": [],
			"ship_updates": [],
			"ship_deletes": [],
			"bullet_creates": [{"id": "bullet-1", "x": 10, "y": 20, "rotation": 25, "velocity_x": 0.0, "velocity_y": 0.0, "owner_id": "ship-1", "lifespan_seconds": 1.0, "weapon_id": "pulse", "projectile_type": "laser"}],
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
	assert_true(world_lane_state.bullets.has("bullet-1"))
	assert_eq(world_lane_state.bullets["bullet-1"]["x"], 6.0)
	assert_eq(world_lane_state.bullets["bullet-1"]["y"], 7.0)
	assert_eq(world_lane_state.bullets["bullet-1"]["rotation"], 0.08)
	assert_false(world_lane_state.pending_bullet_updates.has("bullet-1"))

func test_bullet_delete_clears_pending_bullet_update() -> void:
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
			"ships": [],
			"bullets": [],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"bullet_updates": [{"id": "bullet-1", "x": 30, "y": 40, "rotation": 50}],
	})

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
			"bullet_deletes": ["bullet-1"],
			"asteroid_creates": [],
			"asteroid_updates": [],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [],
			"pickup_deletes": [],
		}
	)

	assert_false(world_lane_state.pending_bullet_updates.has("bullet-1"))


func test_world_full_clears_pending_bullet_updates() -> void:
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
			"ships": [],
			"bullets": [],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"bullet_updates": [{"id": "bullet-1", "x": 30, "y": 40, "rotation": 50}],
	})
	baseline_tracker.mark_lane_unsynced(LaneMetadata.LANE_WORLD)

	applier.apply_world_full(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-2",
			"sequence": 3,
			"ships": [],
			"bullets": [],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	assert_true(world_lane_state.pending_bullet_updates.is_empty())


func test_bullet_delta_after_delete_does_not_buffer_pending_update() -> void:
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
			"ships": [],
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
			"bullet_creates": [{"id": "bullet-1", "x": 10, "y": 20, "rotation": 25, "velocity_x": 0.0, "velocity_y": 0.0, "owner_id": "ship-1", "lifespan_seconds": 1.0, "weapon_id": "pulse", "projectile_type": "laser"}],
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

	applier.apply_world_delta(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 3,
			"ship_creates": [],
			"ship_updates": [],
			"ship_deletes": [],
			"bullet_creates": [],
			"bullet_updates": [],
			"bullet_deletes": ["bullet-1"],
			"asteroid_creates": [],
			"asteroid_updates": [],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [],
			"pickup_deletes": [],
		}
	)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"bullet_updates": [{"id": "bullet-1", "x": 30, "y": 40, "rotation": 50}],
	})

	assert_false(world_lane_state.bullets.has("bullet-1"))
	assert_false(world_lane_state.pending_bullet_updates.has("bullet-1"))


func test_world_full_clears_deleted_bullet_tombstones() -> void:
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
			"ships": [],
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
			"bullet_creates": [{"id": "bullet-1", "x": 10, "y": 20, "rotation": 25, "velocity_x": 0.0, "velocity_y": 0.0, "owner_id": "ship-1", "lifespan_seconds": 1.0, "weapon_id": "pulse", "projectile_type": "laser"}],
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

	applier.apply_world_delta(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-1",
			"sequence": 3,
			"ship_creates": [],
			"ship_updates": [],
			"ship_deletes": [],
			"bullet_creates": [],
			"bullet_updates": [],
			"bullet_deletes": ["bullet-1"],
			"asteroid_creates": [],
			"asteroid_updates": [],
			"asteroid_deletes": [],
			"pickup_creates": [],
			"pickup_updates": [],
			"pickup_deletes": [],
		}
	)
	baseline_tracker.mark_lane_unsynced(LaneMetadata.LANE_WORLD)

	applier.apply_world_full(
		world_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_WORLD,
		{
			"baseline_id": "baseline-2",
			"sequence": 4,
			"ships": [],
			"bullets": [],
			"asteroids": [],
			"pickups": [],
			"is_final_chunk": true,
		}
	)

	applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_WORLD, {
		"sequence": 1,
		"bullet_updates": [{"id": "bullet-1", "x": 30, "y": 40, "rotation": 50}],
	})

	assert_false(world_lane_state.bullets.has("bullet-1"))
	assert_true(world_lane_state.pending_bullet_updates.has("bullet-1"))


func test_hot_lane_accepts_distinct_chunks_and_rejects_duplicates() -> void:
	var state := WorldLaneState.new()
	assert_true(state.accept_asteroid_delta_sequence(4, 1, 2))
	assert_true(state.accept_asteroid_delta_sequence(4, 0, 2))
	assert_false(state.accept_asteroid_delta_sequence(4, 1, 2))

func test_hot_lane_higher_sequence_resets_chunks_and_lower_stays_rejected() -> void:
	var state := WorldLaneState.new()
	assert_true(state.accept_bullet_delta_sequence(4, 0, 2))
	assert_true(state.accept_bullet_delta_sequence(5, 1, 2))
	assert_true(state.accept_bullet_delta_sequence(5, 0, 2))
	assert_false(state.accept_bullet_delta_sequence(4, 1, 2))

func test_hot_lane_tracking_is_independent_and_clear_world_resets_it() -> void:
	var state := WorldLaneState.new()
	assert_true(state.accept_asteroid_delta_sequence(2, 0, 2))
	assert_true(state.accept_bullet_delta_sequence(2, 0, 2))
	assert_false(state.accept_asteroid_delta_sequence(2, 0, 2))
	assert_false(state.accept_bullet_delta_sequence(2, 0, 2))
	state.clear_world()
	assert_true(state.accept_asteroid_delta_sequence(2, 0, 2))
	assert_true(state.accept_bullet_delta_sequence(2, 0, 2))

func test_hot_lane_unchunked_defaults_remain_accepted() -> void:
	var state := WorldLaneState.new()
	assert_true(state.accept_asteroid_delta_sequence(1))
	assert_true(state.accept_bullet_delta_sequence(1))


func test_asteroid_hot_lane_sequence_accepts_only_finite_non_negative_integers() -> void:
	var state := WorldLaneState.new()
	assert_true(state.accept_asteroid_delta_sequence(0))
	assert_true(state.accept_asteroid_delta_sequence(3.0))
	assert_eq(state.latest_asteroid_delta_sequence, 3)

	for sequence in [3.5, -1, "3", true, NAN, INF, -INF]:
		assert_false(state.accept_asteroid_delta_sequence(sequence))
		assert_eq(state.latest_asteroid_delta_sequence, 3)


func test_bullet_hot_lane_sequence_accepts_only_finite_non_negative_integers() -> void:
	var state := WorldLaneState.new()
	assert_true(state.accept_bullet_delta_sequence(0))
	assert_true(state.accept_bullet_delta_sequence(3.0))
	assert_eq(state.latest_bullet_delta_sequence, 3)

	for sequence in [3.5, -1, "3", false, NAN, INF, -INF]:
		assert_false(state.accept_bullet_delta_sequence(sequence))
		assert_eq(state.latest_bullet_delta_sequence, 3)


func test_asteroid_hot_lane_applies_distinct_chunks_and_rejects_duplicate_replay() -> void:
	var applier := WorldLaneApplier.new()
	var state := WorldLaneState.new()
	state.upsert_asteroid({"id": "asteroid-1", "x": 1, "y": 2})
	state.upsert_asteroid({"id": "asteroid-2", "x": 3, "y": 4})

	applier.apply_asteroid_delta(state, LaneMetadata.LANE_WORLD, {
		"sequence": 7,
		"chunk_index": 0,
		"chunk_count": 2,
		"asteroid_updates": [{"id": "asteroid-1", "x": 10, "y": 20}],
	})
	applier.apply_asteroid_delta(state, LaneMetadata.LANE_WORLD, {
		"sequence": 7,
		"chunk_index": 1,
		"chunk_count": 2,
		"asteroid_updates": [{"id": "asteroid-2", "x": 30, "y": 40}],
	})
	applier.apply_asteroid_delta(state, LaneMetadata.LANE_WORLD, {
		"sequence": 7,
		"chunk_index": 0,
		"chunk_count": 2,
		"asteroid_updates": [{"id": "asteroid-1", "x": 99, "y": 99}],
	})

	assert_eq(state.asteroids["asteroid-1"]["x"], 1.0)
	assert_eq(state.asteroids["asteroid-1"]["y"], 2.0)
	assert_eq(state.asteroids["asteroid-2"]["x"], 3.0)
	assert_eq(state.asteroids["asteroid-2"]["y"], 4.0)


func test_hot_lane_rejects_inconsistent_and_invalid_chunk_metadata() -> void:
	var state := WorldLaneState.new()
	assert_true(state.accept_asteroid_delta_sequence(7, 0, 2))
	assert_false(state.accept_asteroid_delta_sequence(7, 1, 3))
	assert_false(state.accept_asteroid_delta_sequence(7, 2, 2))
	assert_false(state.accept_asteroid_delta_sequence(7, -1, 2))
	assert_false(state.accept_asteroid_delta_sequence(7, 0, 0))
	assert_false(state.accept_asteroid_delta_sequence(7, 0.5, 2))
	assert_false(state.accept_asteroid_delta_sequence(7, 0, 2.5))


func test_asteroids_lifecycle_rejects_invalid_payload_shape() -> void:
	var applier := WorldLaneApplier.new()
	var state := WorldLaneState.new()

	assert_false(applier.apply_asteroids_lifecycle(state, {"asteroid_creates": "invalid"}))
	assert_false(applier.apply_asteroids_lifecycle(state, {"asteroid_creates": ["invalid"]}))


func test_bullets_lifecycle_rejects_invalid_payload_shape() -> void:
	var applier := WorldLaneApplier.new()
	var state := WorldLaneState.new()

	assert_false(applier.apply_bullets_lifecycle(state, {"bullet_deletes": "invalid"}))
	assert_false(applier.apply_bullets_lifecycle(state, {"bullet_creates": ["invalid"]}))
