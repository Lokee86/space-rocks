extends GutTest

const BaselineTracker := preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const ResyncState := preload("res://scripts/protocol/realtime/resync_state.gd")
const RealtimePacketPipeline := preload("res://scripts/networking/realtime/realtime_packet_pipeline.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")

func _tracker_with_baseline() -> BaselineTracker:
	var tracker := BaselineTracker.new()
	tracker.record_full_packet("world", "baseline-1", 10)
	return tracker

func test_missing_baseline_delta_emits_one_request_with_empty_context() -> void:
	var tracker := BaselineTracker.new()
	var requests: Array = []
	tracker.resync_required.connect(func(lane, baseline, sequence, reason): requests.append([lane, baseline, sequence, reason]))
	tracker.record_delta("world", "baseline-1", 1)
	tracker.record_delta("world", "baseline-1", 2)
	assert_eq(requests, [["world", "", null, ResyncState.REASON_MISSING_BASELINE]])

func test_wrong_baseline_delta_preserves_previous_context_and_dedupes() -> void:
	var tracker := _tracker_with_baseline()
	var requests: Array = []
	tracker.resync_required.connect(func(lane, baseline, sequence, reason): requests.append([lane, baseline, sequence, reason]))
	tracker.record_delta("world", "baseline-2", 11)
	tracker.record_delta("world", "baseline-3", 12)
	assert_eq(requests, [["world", "baseline-1", 10, ResyncState.REASON_WRONG_BASELINE]])

func test_stale_duplicate_delta_is_silent() -> void:
	var tracker := _tracker_with_baseline()
	var requests: Array = []
	tracker.resync_required.connect(func(_lane, _baseline, _sequence, reason): requests.append(reason))
	assert_false(tracker.record_delta("world", "baseline-1", 10))
	assert_false(tracker.record_delta("world", "baseline-1", 9))
	assert_eq(requests.size(), 0)

func test_changed_baseline_single_full_recovers_and_clears_pending() -> void:
	var tracker := _tracker_with_baseline()
	tracker.record_delta("world", "baseline-2", 11)
	tracker.record_full_packet("world", "baseline-2", 20, null, 0, 1, true)
	assert_true(tracker.is_lane_synced("world"))
	assert_false(tracker.needs_resync("world"))

func test_changed_baseline_chunked_recovery_accepts_chunks_until_final() -> void:
	var tracker := _tracker_with_baseline()
	tracker.record_delta("world", "baseline-2", 11)
	tracker.record_full_chunk("world", "baseline-2", 20, null, 0, 2, false)
	tracker.record_full_chunk("world", "baseline-2", 20, null, 1, 2, true)
	assert_true(tracker.is_lane_synced("world"))
	assert_false(tracker.needs_resync("world"))

func test_duplicate_and_out_of_order_chunks_do_not_complete_recovery() -> void:
	var tracker := _tracker_with_baseline()
	tracker.record_delta("world", "baseline-2", 11)
	tracker.record_full_chunk("world", "baseline-2", 20, null, 0, 3, false)
	tracker.record_full_chunk("world", "baseline-2", 20, null, 0, 3, true)
	tracker.record_full_chunk("world", "baseline-2", 20, null, 1, 3, true)
	assert_false(tracker.is_lane_synced("world"))
	tracker.record_full_chunk("world", "baseline-2", 20, null, 1, 3, false)
	tracker.record_full_chunk("world", "baseline-2", 20, null, 2, 3, true)
	assert_true(tracker.is_lane_synced("world"))
	assert_false(tracker.needs_resync("world"))

func test_pipeline_reset_rebinds_request_propagation() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var requests: Array = []
	pipeline.resync_request_required.connect(func(lane, baseline, sequence, reason): requests.append([lane, baseline, sequence, reason]))
	pipeline.apply_packet({"type": "world_delta", "baseline_id": "baseline-1", "sequence": 1})
	pipeline.reset()
	pipeline.apply_packet({"type": "world_delta", "baseline_id": "baseline-1", "sequence": 1})
	assert_eq(requests.size(), 2)

func test_full_chunk_count_change_is_rejected_without_state_mutation() -> void:
	var tracker := _tracker_with_baseline()
	tracker.record_delta("world", "baseline-2", 11)
	assert_true(tracker.record_full_chunk("world", "baseline-2", 20, null, 0, 2, false))
	var before := tracker.get_lane_state("world")
	assert_false(tracker.record_full_chunk("world", "baseline-2", 20, null, 1, 3, true))
	assert_eq(tracker.get_lane_state("world"), before)

func test_generated_resync_builder_has_exact_contract_fields() -> void:
	assert_eq(Packets.resync_request_packet("world", "baseline-1", 10, "wrong_baseline"), {"type": "resync_request", "lane": "world", "baseline_id": "baseline-1", "sequence": 10, "reason": "wrong_baseline"})
