extends GutTest

const RealtimePacketPipeline := preload("res://scripts/networking/realtime/realtime_packet_pipeline.gd")


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
