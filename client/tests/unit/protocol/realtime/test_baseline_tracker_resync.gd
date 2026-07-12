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
	tracker.record_full_chunk("world", "baseline-2", 20, null, 0, 3, false)
	var requests: Array = []
	tracker.resync_required.connect(func(_lane, _baseline, _sequence, reason): requests.append(reason))
	tracker.record_full_chunk("world", "baseline-2", 20, null, 0, 3, true)
	assert_false(tracker.is_lane_synced("world"))
	assert_true(tracker.needs_resync("world"))
	assert_eq(requests, [ResyncState.REASON_STALE_OR_INVALID_SEQUENCE])

func test_pipeline_reset_rebinds_request_propagation() -> void:
	var pipeline := RealtimePacketPipeline.new()
	pipeline.begin_match("match-1")
	var requests: Array = []
	pipeline.resync_request_required.connect(func(lane, baseline, sequence, reason): requests.append([lane, baseline, sequence, reason]))
	pipeline.apply_packet({"type": "world_delta", "match_id": "match-1", "baseline_id": "baseline-1", "sequence": 1})
	pipeline.reset()
	pipeline.begin_match("match-1")
	pipeline.apply_packet({"type": "world_delta", "match_id": "match-1", "baseline_id": "baseline-1", "sequence": 1})
	assert_eq(requests.size(), 2)

func test_full_chunk_count_change_is_rejected_without_state_mutation() -> void:
	var tracker := _tracker_with_baseline()
	tracker.record_delta("world", "baseline-2", 11)
	assert_true(tracker.record_full_chunk("world", "baseline-2", 20, null, 0, 2, false))
	var before := tracker.get_lane_state("world")
	assert_false(tracker.record_full_chunk("world", "baseline-2", 20, null, 1, 3, false))
	assert_eq(tracker.get_lane_state("world"), before)

func test_generated_resync_builder_has_exact_contract_fields() -> void:
	assert_eq(Packets.resync_request_packet("match-1", "world", "baseline-1", 10, "wrong_baseline"), {"type": "resync_request", "match_id": "match-1", "lane": "world", "baseline_id": "baseline-1", "sequence": 10, "reason": "wrong_baseline"})

func test_invalid_sequence_forms_request_stale_resync_without_state_mutation() -> void:
	for invalid_sequence in [true, "10", null, -1, 1.5, INF]:
		var tracker := _tracker_with_baseline()
		var before := tracker.get_lane_state("world")
		var requests: Array = []
		tracker.resync_required.connect(func(_lane, _baseline, _sequence, reason): requests.append(reason))
		assert_false(tracker.record_delta("world", "baseline-1", invalid_sequence))
		assert_eq(requests, [ResyncState.REASON_STALE_OR_INVALID_SEQUENCE])
		assert_eq(tracker.get_active_baseline_id("world"), before.baseline_id)
		assert_eq(tracker.get_last_accepted_sequence("world"), before.sequence)
		assert_false(tracker.is_lane_synced("world"))
		assert_true(tracker.needs_resync("world"))

func test_invalid_baseline_forms_request_wrong_baseline_without_state_mutation() -> void:
	for invalid_baseline in [true, "", null, 10]:
		var tracker := _tracker_with_baseline()
		var before := tracker.get_lane_state("world")
		var requests: Array = []
		tracker.resync_required.connect(func(_lane, _baseline, _sequence, reason): requests.append(reason))
		assert_false(tracker.record_delta("world", invalid_baseline, 11))
		assert_eq(requests, [ResyncState.REASON_WRONG_BASELINE])
		assert_eq(tracker.get_active_baseline_id("world"), before.baseline_id)
		assert_eq(tracker.get_last_accepted_sequence("world"), before.sequence)
		assert_false(tracker.is_lane_synced("world"))
		assert_true(tracker.needs_resync("world"))

func test_invalid_full_chunk_metadata_requests_resync_without_state_mutation() -> void:
	for chunk_metadata in [[true, 1, true], [0, 0, true], [1.5, 2, false], [0, 2, "true"]]:
		var tracker := _tracker_with_baseline()
		var before := tracker.get_lane_state("world")
		var requests: Array = []
		tracker.resync_required.connect(func(_lane, _baseline, _sequence, reason): requests.append(reason))
		assert_false(tracker.record_full_packet("world", "baseline-2", 20, null, chunk_metadata[0], chunk_metadata[1], chunk_metadata[2]))
		assert_eq(requests, [ResyncState.REASON_STALE_OR_INVALID_SEQUENCE])
		assert_eq(tracker.get_active_baseline_id("world"), before.baseline_id)
		assert_eq(tracker.get_last_accepted_sequence("world"), before.sequence)
		assert_false(tracker.is_lane_synced("world"))
		assert_true(tracker.needs_resync("world"))

func test_integral_float_sequence_is_normalized_and_accepted() -> void:
	var tracker := BaselineTracker.new()
	assert_true(tracker.record_full_packet("world", "baseline-1", 10.0))
	assert_eq(tracker.get_last_accepted_sequence("world"), 10)

func test_repeated_malformed_metadata_emits_one_recovery_request() -> void:
	var tracker := _tracker_with_baseline()
	var requests: Array = []
	tracker.resync_required.connect(func(_lane, _baseline, _sequence, reason): requests.append(reason))
	tracker.record_delta("world", "baseline-1", "bad")
	tracker.record_delta("world", "baseline-1", null)
	tracker.record_full_packet("world", "", 12)
	assert_eq(requests, [ResyncState.REASON_STALE_OR_INVALID_SEQUENCE])

func test_large_integer_sequence_is_preserved_without_float_precision_loss() -> void:
	var tracker := BaselineTracker.new()
	var large_sequence := 9007199254740993
	assert_true(tracker.record_full_packet("world", "baseline-1", large_sequence))
	assert_eq(tracker.get_last_accepted_sequence("world"), large_sequence)

func test_unsupported_baseline_lane_is_rejected_without_state() -> void:
	var tracker := BaselineTracker.new()
	assert_false(tracker.record_full_packet("asteroids", "baseline-1", 1))
	assert_false(tracker.is_lane_synced("asteroids"))

func test_router_records_malformed_sequence_as_stale_recovery() -> void:
	var router := preload("res://scripts/protocol/realtime/realtime_router.gd").new()
	var reasons: Array = []
	router.resync_request_required.connect(func(_lane, _baseline, _sequence, reason): reasons.append(reason))
	router.route_lane_packet({"type": "world_delta", "lane": "world", "baseline_id": "baseline-1", "sequence": "bad"})
	assert_eq(reasons, [ResyncState.REASON_STALE_OR_INVALID_SEQUENCE])
	assert_true(router.resync_state.needs_resync("world"))
	assert_eq(router.resync_state.get_reason("world"), ResyncState.REASON_STALE_OR_INVALID_SEQUENCE)
