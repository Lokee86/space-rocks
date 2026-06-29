extends GutTest

const EventBatchApplier := preload("res://scripts/protocol/realtime/event_batch_applier.gd")
const ResyncState := preload("res://scripts/protocol/realtime/resync_state.gd")
const BaselineTracker := preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const PresentationAdapter := preload("res://scripts/protocol/realtime/presentation_adapter.gd")


class FakeEventSink:
	var handled_events: Array = []

	func handle_presentation_event(event_type, payload, event_packet) -> void:
		handled_events.append({
			"type": event_type,
			"payload": payload,
			"packet": event_packet,
		})


func test_event_batch_applies_events_once() -> void:
	var applier := EventBatchApplier.new()
	var sink := FakeEventSink.new()

	var applied := applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "event-1", "type": "spark", "payload": {"value": 1}},
			],
		},
		sink
	)

	assert_true(applied)
	assert_eq(sink.handled_events.size(), 1)
	assert_eq(sink.handled_events[0]["type"], "spark")


func test_duplicate_batch_id_is_suppressed() -> void:
	var applier := EventBatchApplier.new()
	var sink := FakeEventSink.new()

	applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "event-1", "type": "spark", "payload": {"value": 1}},
			],
		},
		sink
	)
	var applied := applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "event-2", "type": "spark", "payload": {"value": 2}},
			],
		},
		sink
	)

	assert_false(applied)
	assert_eq(sink.handled_events.size(), 1)


func test_duplicate_event_id_is_suppressed() -> void:
	var applier := EventBatchApplier.new()
	var sink := FakeEventSink.new()

	applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "event-1", "type": "spark", "payload": {"value": 1}},
			],
		},
		sink
	)
	var applied := applier.apply_event_batch(
		{
			"batch_id": "batch-2",
			"events": [
				{"event_id": "event-1", "type": "spark", "payload": {"value": 2}},
			],
		},
		sink
	)

	assert_false(applied)
	assert_eq(sink.handled_events.size(), 1)


func test_event_batch_does_not_mark_gameplay_ready() -> void:
	var presentation_adapter := PresentationAdapter.new()
	var applier := EventBatchApplier.new()
	var sink := FakeEventSink.new()

	applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "event-1", "type": "spark", "payload": {}},
			],
		},
		sink
	)

	assert_false(presentation_adapter.is_presentable())


func test_wrong_baseline_marks_lane_resync_needed() -> void:
	var resync_state := ResyncState.new()

	resync_state.mark_wrong_baseline(LaneMetadata.LANE_WORLD)

	assert_true(resync_state.needs_resync(LaneMetadata.LANE_WORLD))
	assert_eq(resync_state.get_reason(LaneMetadata.LANE_WORLD), ResyncState.REASON_WRONG_BASELINE)


func test_stale_sequence_is_ignored() -> void:
	var tracker := BaselineTracker.new()

	tracker.record_full_packet(LaneMetadata.LANE_WORLD, "baseline-1", 2, "snapshot-1", 0, 1, true)
	var applied := tracker.record_delta(LaneMetadata.LANE_WORLD, "baseline-1", 1)

	assert_false(applied)
	assert_false(tracker.needs_resync(LaneMetadata.LANE_WORLD))

