extends GutTest

const LifecycleLaneGate := preload("res://scripts/protocol/realtime/lifecycle_lane_gate.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")


func _packet(sequence, baseline_id = "world-baseline-1") -> Dictionary:
	return {"sequence": sequence, "baseline_id": baseline_id}


func test_matching_active_world_baseline_returns_apply() -> void:
	var gate := LifecycleLaneGate.new()
	var decision := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(1), true, "world-baseline-1")

	assert_eq(decision.status, LifecycleLaneGate.DECISION_APPLY)


func test_unsynced_world_returns_queue() -> void:
	var gate := LifecycleLaneGate.new()
	var decision := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(1), false, "world-baseline-1")

	assert_eq(decision.status, LifecycleLaneGate.DECISION_QUEUE)


func test_different_active_world_baseline_returns_queue() -> void:
	var gate := LifecycleLaneGate.new()
	var decision := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(1), true, "world-baseline-2")

	assert_eq(decision.status, LifecycleLaneGate.DECISION_QUEUE)


func test_missing_sequence_is_rejected() -> void:
	var gate := LifecycleLaneGate.new()
	var decision := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, {"baseline_id": "world-baseline-1"}, true, "world-baseline-1")

	assert_eq(decision.status, LifecycleLaneGate.DECISION_REJECT)


func test_non_integer_sequences_are_rejected() -> void:
	for sequence in ["1", 1.5, true]:
		var gate := LifecycleLaneGate.new()
		var decision := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(sequence), true, "world-baseline-1")

		assert_eq(decision.status, LifecycleLaneGate.DECISION_REJECT)


func test_missing_and_empty_baseline_ids_are_rejected() -> void:
	var gate := LifecycleLaneGate.new()
	var missing := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, {"sequence": 1}, true, "world-baseline-1")
	var empty := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(2, ""), true, "world-baseline-1")

	assert_eq(missing.status, LifecycleLaneGate.DECISION_REJECT)
	assert_eq(empty.status, LifecycleLaneGate.DECISION_REJECT)


func test_unsupported_lane_is_rejected() -> void:
	var gate := LifecycleLaneGate.new()
	var decision := gate.submit(LaneMetadata.LANE_WORLD, _packet(1), true, "world-baseline-1")

	assert_eq(decision.status, LifecycleLaneGate.DECISION_REJECT)


func test_mark_applied_rejects_the_same_sequence() -> void:
	var gate := LifecycleLaneGate.new()
	gate.mark_applied(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, 3)
	var decision := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(3), true, "world-baseline-1")

	assert_eq(decision.status, LifecycleLaneGate.DECISION_REJECT)


func test_lower_sequence_is_rejected() -> void:
	var gate := LifecycleLaneGate.new()
	gate.mark_applied(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, 3)
	var decision := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(2), true, "world-baseline-1")

	assert_eq(decision.status, LifecycleLaneGate.DECISION_REJECT)


func test_higher_non_contiguous_sequence_is_accepted() -> void:
	var gate := LifecycleLaneGate.new()
	gate.mark_applied(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, 3)
	var decision := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(7), true, "world-baseline-1")

	assert_eq(decision.status, LifecycleLaneGate.DECISION_APPLY)


func test_asteroid_and_bullet_latest_sequences_are_independent() -> void:
	var gate := LifecycleLaneGate.new()
	gate.mark_applied(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, 5)
	var bullet_decision := gate.submit(LaneMetadata.LANE_BULLETS_LIFECYCLE, _packet(1), true, "world-baseline-1")
	var asteroid_decision := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(1), true, "world-baseline-1")

	assert_eq(bullet_decision.status, LifecycleLaneGate.DECISION_APPLY)
	assert_eq(asteroid_decision.status, LifecycleLaneGate.DECISION_REJECT)


func test_queued_packets_are_returned_only_for_matching_baseline() -> void:
	var gate := LifecycleLaneGate.new()
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(1, "world-baseline-1"), false, "world-baseline-2")
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(2, "world-baseline-2"), false, "world-baseline-2")

	var baseline_one := gate.take_pending_for_baseline("world-baseline-1")
	var baseline_two := gate.take_pending_for_baseline("world-baseline-2")

	assert_eq(baseline_one.size(), 1)
	assert_eq(baseline_one[0].sequence, 1)
	assert_eq(baseline_two.size(), 1)
	assert_eq(baseline_two[0].sequence, 2)


func test_asteroid_packets_drain_in_ascending_sequence_order() -> void:
	var gate := LifecycleLaneGate.new()
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(5), false, "world-baseline-2")
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(2), false, "world-baseline-2")
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(9), false, "world-baseline-2")

	var drained := gate.take_pending_for_baseline("world-baseline-2")

	assert_eq([drained[0].sequence, drained[1].sequence, drained[2].sequence], [2, 5, 9])


func test_bullet_packets_drain_in_ascending_sequence_order() -> void:
	var gate := LifecycleLaneGate.new()
	gate.submit(LaneMetadata.LANE_BULLETS_LIFECYCLE, _packet(6), false, "world-baseline-2")
	gate.submit(LaneMetadata.LANE_BULLETS_LIFECYCLE, _packet(1), false, "world-baseline-2")
	gate.submit(LaneMetadata.LANE_BULLETS_LIFECYCLE, _packet(4), false, "world-baseline-2")

	var drained := gate.take_pending_for_baseline("world-baseline-2")

	assert_eq([drained[0].sequence, drained[1].sequence, drained[2].sequence], [1, 4, 6])


func test_different_lifecycle_lanes_are_not_ordered_against_each_other() -> void:
	var gate := LifecycleLaneGate.new()
	gate.submit(LaneMetadata.LANE_BULLETS_LIFECYCLE, _packet(1), false, "world-baseline-2")
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(9), false, "world-baseline-2")

	var drained := gate.take_pending_for_baseline("world-baseline-2")

	assert_eq(drained[0].lane, LaneMetadata.LANE_ASTEROIDS_LIFECYCLE)
	assert_eq(drained[1].lane, LaneMetadata.LANE_BULLETS_LIFECYCLE)


func test_already_pending_lane_sequence_is_rejected_as_duplicate() -> void:
	var gate := LifecycleLaneGate.new()
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(3), false, "world-baseline-2")
	var duplicate := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(3, "world-baseline-3"), false, "world-baseline-3")

	assert_eq(duplicate.status, LifecycleLaneGate.DECISION_REJECT)


func test_draining_removes_pending_duplicate_tracking() -> void:
	var gate := LifecycleLaneGate.new()
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(3), false, "world-baseline-2")
	gate.take_pending_for_baseline("world-baseline-2")
	var resubmitted := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(3), false, "world-baseline-2")

	assert_eq(resubmitted.status, LifecycleLaneGate.DECISION_QUEUE)


func test_activating_baseline_three_discards_older_numeric_baselines() -> void:
	var gate := LifecycleLaneGate.new()
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(1, "world-baseline-1"), false, "world-baseline-0")
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(2, "world-baseline-2"), false, "world-baseline-0")
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(3, "world-baseline-3"), false, "world-baseline-0")
	gate.discard_obsolete_baselines("world-baseline-3")

	assert_eq(gate.take_pending_for_baseline("world-baseline-1").size(), 0)
	assert_eq(gate.take_pending_for_baseline("world-baseline-2").size(), 0)
	assert_eq(gate.take_pending_for_baseline("world-baseline-3").size(), 1)


func test_higher_numeric_baseline_remains_queued() -> void:
	var gate := LifecycleLaneGate.new()
	gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(4, "world-baseline-4"), false, "world-baseline-0")
	gate.discard_obsolete_baselines("world-baseline-3")

	assert_eq(gate.take_pending_for_baseline("world-baseline-4").size(), 1)


func test_unparseable_baseline_remains_bounded_not_obsolete() -> void:
	var gate := LifecycleLaneGate.new()
	for index in range(5):
		gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(index + 1, "pending-" + str(index)), false, "world-baseline-3")
	gate.discard_obsolete_baselines("world-baseline-3")

	assert_eq(gate.take_pending_for_baseline("pending-0").size(), 0)
	assert_eq(gate.take_pending_for_baseline("pending-4").size(), 1)


func test_per_lane_packet_overflow_discards_oldest_packet() -> void:
	var gate := LifecycleLaneGate.new()
	for sequence in range(1, 10):
		gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(sequence), false, "world-baseline-2")

	var drained := gate.take_pending_for_baseline("world-baseline-2")

	assert_eq(drained.size(), 8)
	assert_eq(drained[0].sequence, 2)
	assert_eq(drained[7].sequence, 9)


func test_baseline_bucket_overflow_discards_oldest_non_active_bucket() -> void:
	var gate := LifecycleLaneGate.new()
	for baseline_number in range(1, 6):
		var baseline_id := "world-baseline-" + str(baseline_number)
		gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(baseline_number, baseline_id), false, "world-baseline-5")

	assert_eq(gate.take_pending_for_baseline("world-baseline-1").size(), 0)
	assert_eq(gate.take_pending_for_baseline("world-baseline-5").size(), 1)


func test_reset_clears_applied_sequence_and_all_pending_state() -> void:
	var gate := LifecycleLaneGate.new()
	gate.mark_applied(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, 5)
	gate.submit(LaneMetadata.LANE_BULLETS_LIFECYCLE, _packet(7), false, "world-baseline-2")
	gate.reset()

	var accepted := gate.submit(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, _packet(1), true, "world-baseline-1")
	var pending := gate.take_pending_for_baseline("world-baseline-2")

	assert_eq(accepted.status, LifecycleLaneGate.DECISION_APPLY)
	assert_eq(pending.size(), 0)
