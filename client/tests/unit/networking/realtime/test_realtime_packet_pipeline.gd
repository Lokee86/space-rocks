extends GutTest

const RealtimePacketPipeline := preload("res://scripts/networking/realtime/realtime_packet_pipeline.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")


func test_valid_realtime_packet_mutates_router_state_before_signal_callback() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var callback_state := {"state_seen": false, "count": 0}

	pipeline.gameplay_packet_applied.connect(func(packet: Dictionary) -> void:
		callback_state.count += 1
		assert_true(pipeline.get_router().baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))
		assert_true(pipeline.get_router().world_lane_state.ships.is_empty())
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
	assert_true(pipeline.get_router().baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))


func test_invalid_or_unsupported_packets_do_not_emit_gameplay_packet_applied() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var callback_state := {"count": 0}

	pipeline.gameplay_packet_applied.connect(func(_packet: Dictionary) -> void:
		callback_state.count += 1
	)

	pipeline.apply_packet({"type": "not_real"})
	pipeline.apply_packet({"foo": "bar"})

	assert_eq(callback_state.count, 0)




func test_reset_replaces_router_and_clears_previous_state() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var old_router := pipeline.get_router()

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
	assert_true(old_router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))
	assert_eq(pipeline.get_readiness(), old_router.get_gameplay_readiness())

	pipeline.reset()

	var new_router := pipeline.get_router()
	assert_ne(new_router, old_router)
	assert_false(new_router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))
	assert_false(new_router.is_presentable())
	assert_eq(pipeline.get_readiness(), new_router.get_gameplay_readiness())
	assert_false(pipeline.get_readiness() == old_router.get_gameplay_readiness())
	assert_false(pipeline.get_router().baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))
