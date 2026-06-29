extends GutTest

const GameplayReadiness := preload("res://scripts/protocol/realtime/gameplay_readiness.gd")
const BaselineTracker := preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const EventBatchApplier := preload("res://scripts/protocol/realtime/event_batch_applier.gd")
const ResyncState := preload("res://scripts/protocol/realtime/resync_state.gd")
const GameplayStateFlow := preload("res://scripts/gameplay/state/gameplay_state_flow.gd")


func test_gameplay_is_not_ready_initially() -> void:
	var readiness := GameplayReadiness.new()

	assert_false(readiness.is_gameplay_ready())


func test_world_only_baseline_synced_does_not_mark_gameplay_ready() -> void:
	var readiness := GameplayReadiness.new()
	var tracker := BaselineTracker.new()
	tracker.bind_readiness(readiness)

	tracker.record_full_packet(LaneMetadata.LANE_WORLD, "baseline-1", 1, "snapshot-1", 0, 1, true)

	assert_false(readiness.is_gameplay_ready())


func test_overlay_only_baseline_synced_does_not_mark_gameplay_ready() -> void:
	var readiness := GameplayReadiness.new()
	var tracker := BaselineTracker.new()
	tracker.bind_readiness(readiness)

	tracker.record_full_packet(LaneMetadata.LANE_OVERLAY, "baseline-1", 1, "snapshot-1", 0, 1, true)

	assert_false(readiness.is_gameplay_ready())


func test_session_only_baseline_synced_does_not_mark_gameplay_ready() -> void:
	var readiness := GameplayReadiness.new()
	var tracker := BaselineTracker.new()
	tracker.bind_readiness(readiness)

	tracker.record_full_packet(LaneMetadata.LANE_SESSION, "baseline-1", 1, "snapshot-1", 0, 1, true)

	assert_false(readiness.is_gameplay_ready())


func test_event_batch_does_not_mark_gameplay_ready() -> void:
	var readiness := GameplayReadiness.new()
	var tracker := BaselineTracker.new()
	tracker.bind_readiness(readiness)
	var event_batch_applier := EventBatchApplier.new()
	var sink := _FakeEventSink.new()

	event_batch_applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [{"event_id": "event-1", "type": "spark", "payload": {}}],
		},
		sink
	)

	assert_false(readiness.is_gameplay_ready())


func test_debug_control_packet_does_not_mark_gameplay_ready() -> void:
	var readiness := GameplayReadiness.new()
	var tracker := BaselineTracker.new()
	tracker.bind_readiness(readiness)
	var resync_state := ResyncState.new()

	resync_state.mark_wrong_baseline(LaneMetadata.LANE_WORLD)

	assert_false(readiness.is_gameplay_ready())


func test_world_overlay_session_synced_marks_gameplay_ready() -> void:
	var readiness := GameplayReadiness.new()
	var tracker := BaselineTracker.new()
	tracker.bind_readiness(readiness)

	tracker.record_full_packet(LaneMetadata.LANE_WORLD, "baseline-1", 1, "snapshot-1", 0, 1, true)
	tracker.record_full_packet(LaneMetadata.LANE_OVERLAY, "baseline-1", 1, "snapshot-1", 0, 1, true)
	tracker.record_full_packet(LaneMetadata.LANE_SESSION, "baseline-1", 1, "snapshot-1", 0, 1, true)

	assert_true(readiness.is_gameplay_ready())


func test_wrong_baseline_resync_needed_clears_readiness() -> void:
	var readiness := GameplayReadiness.new()
	var tracker := BaselineTracker.new()
	tracker.bind_readiness(readiness)

	tracker.record_full_packet(LaneMetadata.LANE_WORLD, "baseline-1", 1, "snapshot-1", 0, 1, true)
	tracker.record_full_packet(LaneMetadata.LANE_OVERLAY, "baseline-1", 1, "snapshot-1", 0, 1, true)
	tracker.record_full_packet(LaneMetadata.LANE_SESSION, "baseline-1", 1, "snapshot-1", 0, 1, true)
	assert_true(readiness.is_gameplay_ready())

	tracker.record_delta(LaneMetadata.LANE_WORLD, "baseline-2", 2)

	assert_false(readiness.is_gameplay_ready())


func test_required_lane_baselines_synced_redirects_to_gameplay_readiness() -> void:
	var flow := GameplayStateFlow.new()
	var readiness := GameplayReadiness.new()
	var tracker := BaselineTracker.new()
	tracker.bind_readiness(readiness)
	flow.set_gameplay_readiness(readiness)

	assert_false(flow.is_gameplay_ready())
	tracker.record_full_packet(LaneMetadata.LANE_WORLD, "baseline-1", 1, "snapshot-1", 0, 1, true)
	tracker.record_full_packet(LaneMetadata.LANE_OVERLAY, "baseline-1", 1, "snapshot-1", 0, 1, true)
	tracker.record_full_packet(LaneMetadata.LANE_SESSION, "baseline-1", 1, "snapshot-1", 0, 1, true)
	assert_true(flow.is_gameplay_ready())


class _FakeEventSink:
	func handle_presentation_event(_event_type, _payload, _event_packet) -> void:
		pass

