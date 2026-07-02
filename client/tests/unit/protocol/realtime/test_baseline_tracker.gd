extends GutTest

const BaselineTracker := preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")


func test_record_delta_preserves_last_full_baseline_and_tracks_snapshot_id() -> void:
	var tracker := BaselineTracker.new()

	tracker.record_full_packet(LaneMetadata.LANE_WORLD, "world-baseline-1", 1, "world-baseline-1", 0, 1, true)
	var applied := tracker.record_delta(LaneMetadata.LANE_WORLD, "world-baseline-1", 2, "world-snapshot-2")

	assert_true(applied)
	var state := tracker.get_lane_state(LaneMetadata.LANE_WORLD)
	assert_eq(state.baseline_id, "world-baseline-1")
	assert_eq(state.sequence, 2)
	assert_eq(state.snapshot_id, "world-snapshot-2")
	assert_false(tracker.needs_resync(LaneMetadata.LANE_WORLD))


func test_record_delta_rejects_mismatched_baseline_and_marks_resync_needed() -> void:
	var tracker := BaselineTracker.new()

	tracker.record_full_packet(LaneMetadata.LANE_WORLD, "world-baseline-1", 1, "world-baseline-1", 0, 1, true)
	var applied := tracker.record_delta(LaneMetadata.LANE_WORLD, "world-baseline-2", 2, "world-snapshot-2")

	assert_false(applied)
	assert_true(tracker.needs_resync(LaneMetadata.LANE_WORLD))


func test_record_delta_keeps_baseline_anchored_while_tracking_snapshot_id() -> void:
	var tracker := BaselineTracker.new()

	tracker.record_full_packet("world", "world-baseline-1", 1, "world-baseline-1", 0, 1, true)
	var applied := tracker.record_delta("world", "world-baseline-1", 2, "world-snapshot-2")

	assert_true(applied)
	var state := tracker.get_lane_state("world")
	assert_eq(state.baseline_id, "world-baseline-1")
	assert_eq(state.sequence, 2)
	assert_eq(state.snapshot_id, "world-snapshot-2")
	assert_false(tracker.needs_resync("world"))


func test_record_delta_marks_resync_when_baseline_changes() -> void:
	var tracker := BaselineTracker.new()

	tracker.record_full_packet("world", "world-baseline-1", 1, "world-baseline-1", 0, 1, true)
	var applied := tracker.record_delta("world", "world-baseline-2", 2, "world-snapshot-2")

	assert_false(applied)
	assert_true(tracker.needs_resync("world"))
