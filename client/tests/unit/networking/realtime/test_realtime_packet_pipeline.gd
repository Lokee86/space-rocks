extends GutTest

const RealtimePacketPipeline := preload("res://scripts/networking/realtime/realtime_packet_pipeline.gd")
const ResyncState := preload("res://scripts/protocol/realtime/resync_state.gd")


func test_valid_realtime_packet_mutates_router_state_before_signal_callback() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var callback_state := {"state_seen": false, "count": 0}

	assert_false(pipeline.is_gameplay_ready())
	var presentation_state: Variant = pipeline.get_presentation_state()
	assert_true(presentation_state != null)
	var presentation_state_identity: Variant = presentation_state

	pipeline.gameplay_packet_applied.connect(func(_packet: Dictionary) -> void:
		callback_state.count += 1
		assert_false(pipeline.is_gameplay_ready())
		assert_true(presentation_state.world_lane_state != null)
		assert_true(presentation_state.world_lane_state.ships.is_empty())
		callback_state.state_seen = true
	)

	pipeline.apply_packet({
		"type": "world_full",
		"baseline_id": "world-baseline-1",
		"sequence": 1,
		"snapshot_id": "world-snapshot-1",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})

	assert_true(callback_state.state_seen)
	assert_eq(callback_state.count, 1)
	assert_false(pipeline.is_gameplay_ready())
	assert_true(presentation_state.world_lane_state != null)
	assert_true(presentation_state.world_lane_state.ships.is_empty())
	assert_true(pipeline.get_presentation_state() == presentation_state_identity)


func test_invalid_or_unsupported_packets_do_not_emit_gameplay_packet_applied() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var callback_state := {"count": 0}
	pipeline.gameplay_packet_applied.connect(func(_packet: Dictionary) -> void:
		callback_state.count += 1
	)

	pipeline.apply_packet({"type": "not_real"})
	pipeline.apply_packet({"foo": "bar"})

	assert_eq(callback_state.count, 0)


func test_event_batch_and_resync_packets_route_through_explicit_pipeline_handlers() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var callback_state := {"count": 0}
	pipeline.gameplay_packet_applied.connect(func(_packet: Dictionary) -> void:
		callback_state.count += 1
	)

	pipeline.apply_packet({
		"type": "event_batch",
		"batch_id": "batch-1",
		"events": [
			{"event_id": "presentation-event-1", "type": "spark", "payload": {"value": 1}},
		],
	})
	pipeline.apply_packet({"type": "resync_request", "lane": "world"})
	pipeline.apply_packet({"type": "resync_required", "lane": "overlay"})

	assert_eq(callback_state.count, 3)
	assert_eq(pipeline.get_presentation_state().event_batch_applier.get_applied_events().size(), 1)
	assert_true(pipeline.get_router().resync_state.needs_resync("world"))
	assert_eq(pipeline.get_router().resync_state.get_reason("world"), ResyncState.REASON_MISSING_BASELINE)
	assert_true(pipeline.get_router().resync_state.needs_resync("overlay"))
	assert_eq(pipeline.get_router().resync_state.get_reason("overlay"), ResyncState.REASON_WRONG_BASELINE)


func test_reset_clears_lifecycle_pending_state_before_matching_baseline_arrives() -> void:
	var pipeline := RealtimePacketPipeline.new()

	pipeline.apply_packet({
		"type": "asteroids_lifecycle",
		"lane": "asteroids.lifecycle",
		"sequence": 1,
		"baseline_id": "world-baseline-2",
		"asteroid_creates": [{"id": "asteroid-pre-reset", "x": 10, "y": 20}],
		"asteroid_deletes": [],
	})
	pipeline.reset()
	pipeline.apply_packet({
		"type": "world_full",
		"baseline_id": "world-baseline-2",
		"sequence": 1,
		"snapshot_id": "world-snapshot-2",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})

	assert_false(pipeline.get_router().world_lane_state.asteroids.has("asteroid-pre-reset"))


func test_reset_clears_lifecycle_applied_sequence_state() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var world_full := {
		"type": "world_full",
		"baseline_id": "world-baseline-1",
		"sequence": 1,
		"snapshot_id": "world-snapshot-1",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	}
	var bullet_lifecycle := {
		"type": "bullets_lifecycle",
		"lane": "bullets.lifecycle",
		"sequence": 5,
		"baseline_id": "world-baseline-1",
		"bullet_creates": [{"id": "bullet-reset", "owner_id": "player-1", "x": 10, "y": 20}],
		"bullet_deletes": [],
	}

	pipeline.apply_packet(world_full)
	pipeline.apply_packet(bullet_lifecycle)
	assert_true(pipeline.get_router().world_lane_state.bullets.has("bullet-reset"))

	pipeline.reset()
	pipeline.apply_packet(world_full)
	pipeline.apply_packet(bullet_lifecycle)

	assert_true(pipeline.get_router().world_lane_state.bullets.has("bullet-reset"))


func test_reset_preserves_presentation_state_identity_and_clears_stale_state() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var presentation_state := pipeline.get_presentation_state()
	var world_lane_state: Variant = presentation_state.world_lane_state
	var overlay_lane_state: Variant = presentation_state.overlay_lane_state
	var session_lane_state: Variant = presentation_state.session_lane_state
	var event_batch_applier: Variant = presentation_state.event_batch_applier

	pipeline.apply_packet({
		"type": "world_full",
		"baseline_id": "world-baseline-1",
		"sequence": 1,
		"snapshot_id": "world-snapshot-1",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})
	assert_false(pipeline.is_gameplay_ready())
	assert_true(presentation_state.world_lane_state != null)

	pipeline.reset()

	assert_true(pipeline.get_presentation_state() == presentation_state)
	assert_false(pipeline.is_gameplay_ready())
	assert_true(presentation_state.world_lane_state != null)
	assert_true(presentation_state.overlay_lane_state != null)
	assert_true(presentation_state.session_lane_state != null)
	assert_true(presentation_state.event_batch_applier != null)
	assert_false(presentation_state.world_lane_state == world_lane_state)
	assert_false(presentation_state.overlay_lane_state == overlay_lane_state)
	assert_false(presentation_state.session_lane_state == session_lane_state)
	assert_false(presentation_state.event_batch_applier == event_batch_applier)
