extends GutTest

const EventBatchApplier := preload("res://scripts/protocol/realtime/event_batch_applier.gd")
const ResyncState := preload("res://scripts/protocol/realtime/resync_state.gd")
const BaselineTracker := preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const PresentationAdapter := preload("res://scripts/protocol/realtime/presentation_adapter.gd")
const GameplayReadiness := preload("res://scripts/protocol/realtime/gameplay_readiness.gd")


class FakeEventSink:
	var handled_events: Array = []

	func handle_presentation_event(event_type, payload, event_packet) -> void:
		handled_events.append({
			"type": event_type,
			"payload": payload,
			"packet": event_packet,
		})


class FakePresentationTarget:
	var last_world_lane_state = null
	var last_overlay_lane_state = null
	var last_session_lane_state = null

	func apply_world_lane_state(world_sync, world_lane_state, self_id: String) -> void:
		last_world_lane_state = world_lane_state

	func apply_overlay_lane_state(hud_flow, overlay_lane_state) -> void:
		last_overlay_lane_state = overlay_lane_state

	func apply_session_lane_state(hud_flow, session_lane_state, self_id: String) -> void:
		last_session_lane_state = session_lane_state


class FakeEventFlow:
	var apply_server_events_call_count := 0
	var received_event_count := 0
	var received_event_types: Array = []

	func apply_server_events(events: Array, self_id: String) -> void:
		apply_server_events_call_count += 1
		received_event_count += events.size()
		for event in events:
			received_event_types.append(str(event.get("type", "")))


class FakeRouter:
	var world_lane_state = null
	var overlay_lane_state = null
	var session_lane_state = null
	var event_batch_applier = null


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


func test_presentation_adapter_forwards_applied_event_batch_once_to_event_flow() -> void:
	var applier := EventBatchApplier.new()
	var router := FakeRouter.new()
	var presentation_adapter := PresentationAdapter.new()
	var readiness := GameplayReadiness.new()
	readiness.mark_world_baseline_synced()
	readiness.mark_overlay_baseline_synced()
	readiness.mark_session_baseline_synced()
	var world_sync := FakePresentationTarget.new()
	var hud_flow := FakePresentationTarget.new()
	var event_flow := FakeEventFlow.new()

	router.world_lane_state = {}
	router.overlay_lane_state = {}
	router.session_lane_state = {}
	router.event_batch_applier = applier
	presentation_adapter.bind_gameplay_readiness(readiness)

	applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "event-1", "type": "bullet_blast", "payload": {"value": 1}},
			],
		},
		null
	)

	presentation_adapter.fanout_lane_states(router, world_sync, hud_flow, event_flow)
	presentation_adapter.fanout_lane_states(router, world_sync, hud_flow, event_flow)

	assert_eq(event_flow.apply_server_events_call_count, 1)
	assert_eq(event_flow.received_event_count, 1)
	assert_eq(event_flow.received_event_types[0], "bullet_blast")


func test_repeated_batch_id_still_applies_unseen_event_ids() -> void:
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

	assert_true(applied)
	assert_eq(sink.handled_events.size(), 2)
	assert_eq(sink.handled_events[1]["type"], "spark")


func test_repeated_batch_id_skips_missing_event_id_defensively() -> void:
	var applier := EventBatchApplier.new()
	var sink := FakeEventSink.new()

	applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"type": "spark", "payload": {"value": 1}},
			],
		},
		sink
	)
	var applied := applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"type": "spark", "payload": {"value": 2}},
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

