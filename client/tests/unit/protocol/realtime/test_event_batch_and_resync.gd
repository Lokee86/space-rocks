extends GutTest

const CompactLanePacket := preload("res://scripts/protocol/realtime/compact_lane_packet.gd")
const EventBatchApplier := preload("res://scripts/protocol/realtime/event_batch_applier.gd")
const PresentationAdapter := preload("res://scripts/protocol/realtime/presentation_adapter.gd")
const WorldLaneState := preload("res://scripts/protocol/realtime/world_lane_state.gd")
const OverlayLaneState := preload("res://scripts/protocol/realtime/overlay_lane_state.gd")
const SessionLaneState := preload("res://scripts/protocol/realtime/session_lane_state.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const ResyncState := preload("res://scripts/protocol/realtime/resync_state.gd")
const BaselineTracker := preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const RealtimePresentationState := preload("res://scripts/networking/realtime/realtime_presentation_state.gd")
const RealtimePacketPipeline := preload("res://scripts/networking/realtime/realtime_packet_pipeline.gd")


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

	func apply_world_lane_state(world_lane_state) -> void:
		last_world_lane_state = world_lane_state

	func apply_overlay_lane_state(overlay_lane_state) -> void:
		last_overlay_lane_state = overlay_lane_state

	func apply_session_lane_state(session_lane_state, self_id: String) -> void:
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


class FakeLocalLifecycleFlow:
	var world_lane_state = null
	var session_lane_state = null
	var self_id := ""

	func apply_lane_state(received_world_lane_state, received_session_lane_state, received_self_id: String) -> void:
		world_lane_state = received_world_lane_state
		session_lane_state = received_session_lane_state
		self_id = received_self_id


func test_event_batch_applies_events_once() -> void:
	var applier := EventBatchApplier.new()
	var sink := FakeEventSink.new()

	var applied := applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "presentation-event-1", "type": "spark", "payload": {"value": 1}},
			],
		},
		sink
	)

	assert_true(applied)
	assert_eq(sink.handled_events.size(), 1)
	assert_eq(sink.handled_events[0]["type"], "spark")


func test_compact_event_batch_tuple_expansion_still_applies_and_dedupes() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "eb",
		"q": 11,
		"ms": 123,
		"bid": "event-batch-11",
		"ev": [
			["bb", 1, 123, 568],
			["shd", 2, 1, 2, 3500, 30, 40],
		],
	})

	var applier := EventBatchApplier.new()
	var sink := FakeEventSink.new()

	var applied := applier.apply_event_batch(expanded, sink)
	var reapplied := applier.apply_event_batch(expanded, sink)

	assert_true(applied)
	assert_false(reapplied)
	assert_eq(sink.handled_events.size(), 2)
	assert_eq(sink.handled_events[0]["type"], "bullet_blast")
	assert_eq(sink.handled_events[0]["packet"]["event_id"], "presentation-event-1")
	assert_eq(sink.handled_events[0]["packet"]["x"], 12.3)
	assert_eq(sink.handled_events[0]["packet"]["y"], 56.8)
	assert_eq(sink.handled_events[1]["type"], "ship_death")
	assert_eq(sink.handled_events[1]["packet"]["player_id"], "player-1")
	assert_eq(sink.handled_events[1]["packet"]["respawn_delay"], 3.5)


func test_compact_event_batch_expands_before_application_and_dedupes() -> void:
	var expanded := CompactLanePacket.expand_packet({
		"t": "eb",
		"q": 11,
		"ms": 123,
		"bid": "event-batch-11",
		"ev": [
			{"ei": "presentation-event-1", "t": "bb", "x": 123, "y": 568},
			{"ei": "presentation-event-2", "t": "shd", "pid": "player-1", "lv": 2, "rd": 3500, "x": 30, "y": 40},
		],
	})

	assert_eq(expanded["type"], "event_batch")
	assert_eq(expanded["batch_id"], "event-batch-11")
	assert_eq(expanded["events"][0]["event_id"], "presentation-event-1")
	assert_eq(expanded["events"][0]["type"], "bullet_blast")
	assert_eq(expanded["events"][1]["event_id"], "presentation-event-2")
	assert_eq(expanded["events"][1]["type"], "ship_death")
	assert_eq(expanded["events"][1]["player_id"], "player-1")
	assert_eq(expanded["events"][1]["lives"], 2)
	assert_eq(expanded["events"][1]["respawn_delay"], 3500)

	var applier := EventBatchApplier.new()
	var sink := FakeEventSink.new()

	var first_applied := applier.apply_event_batch(expanded, sink)
	var second_applied := applier.apply_event_batch(expanded, sink)

	assert_true(first_applied)
	assert_false(second_applied)
	assert_eq(sink.handled_events.size(), 2)
	assert_eq(sink.handled_events[0]["type"], "bullet_blast")
	assert_eq(sink.handled_events[0]["packet"]["x"], 12.3)
	assert_eq(sink.handled_events[0]["packet"]["y"], 56.8)
	assert_eq(sink.handled_events[1]["type"], "ship_death")
	assert_eq(sink.handled_events[1]["packet"]["respawn_delay"], 3.5)
	assert_eq(applier.get_applied_events()[0]["x"], 12.3)


func test_presentation_adapter_forwards_applied_event_batch_once_to_event_flow() -> void:
	var applier := EventBatchApplier.new()
	var presentation_state := {
		"world_lane_state": WorldLaneState.new(),
		"overlay_lane_state": OverlayLaneState.new(),
		"session_lane_state": SessionLaneState.new(),
		"event_batch_applier": applier,
	}
	var presentation_adapter := PresentationAdapter.new()
	var world_sync := FakePresentationTarget.new()
	var hud_flow := FakePresentationTarget.new()
	var event_flow := FakeEventFlow.new()

	presentation_state["overlay_lane_state"].self_id = "player-1"

	applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "presentation-event-1", "type": "bullet_blast", "payload": {"value": 1}},
			],
		},
		null
	)

	presentation_adapter.fanout_lane_states(presentation_state, world_sync, hud_flow, event_flow)

	assert_eq(event_flow.apply_server_events_call_count, 1)
	assert_eq(event_flow.received_event_count, 1)
	assert_eq(event_flow.received_event_types[0], "bullet_blast")


func test_presentation_adapter_forwards_decoded_session_state_to_local_lifecycle_flow() -> void:
	var presentation_state := RealtimePresentationState.new()
	presentation_state.world_lane_state = WorldLaneState.new()
	presentation_state.overlay_lane_state = OverlayLaneState.new()
	presentation_state.session_lane_state = SessionLaneState.new()
	presentation_state.event_batch_applier = EventBatchApplier.new()
	presentation_state.overlay_lane_state.self_id = "player-1"
	presentation_state.session_lane_state.player_sessions = {
		"player-1": {"respawn_cooldown": 3500},
	}

	var presentation_adapter := PresentationAdapter.new()
	var world_sync := FakePresentationTarget.new()
	var hud_flow := FakePresentationTarget.new()
	var event_flow := FakeEventFlow.new()
	var local_lifecycle_flow := FakeLocalLifecycleFlow.new()

	presentation_adapter.fanout_lane_states(
		presentation_state,
		world_sync,
		hud_flow,
		event_flow,
		local_lifecycle_flow
	)

	assert_eq(local_lifecycle_flow.world_lane_state, presentation_state.world_lane_state)
	assert_eq(local_lifecycle_flow.session_lane_state.player_sessions["player-1"]["respawn_cooldown"], 3.5)
	assert_eq(local_lifecycle_flow.self_id, "player-1")


func test_pipeline_event_batch_and_resync_packets_update_router_state() -> void:
	var pipeline := RealtimePacketPipeline.new()
	pipeline.begin_match("match-1")
	var callback_state := {"count": 0}
	pipeline.gameplay_packet_applied.connect(func(_packet: Dictionary) -> void:
		callback_state.count += 1
	)

	pipeline.apply_packet({
		"type": "event_batch",
		"match_id": "match-1",
		"batch_id": "batch-11",
		"events": [
			{"event_id": "presentation-event-1", "type": "bullet_blast", "payload": {"value": 1}},
		],
	})
	pipeline.apply_packet({"type": "resync_request", "match_id": "match-1", "lane": "world", "reason": "missing_baseline"})
	pipeline.apply_packet({"type": "resync_request", "match_id": "match-1", "lane": "overlay", "reason": "wrong_baseline"})
	pipeline.apply_packet({"type": "resync_required", "match_id": "match-1", "lane": "overlay", "reason": "wrong_baseline"})

	assert_eq(callback_state.count, 4)
	assert_eq(pipeline.get_presentation_state().event_batch_applier.get_applied_events().size(), 1)
	assert_true(pipeline.get_router().resync_state.needs_resync(LaneMetadata.LANE_WORLD))
	assert_eq(pipeline.get_router().resync_state.get_reason(LaneMetadata.LANE_WORLD), ResyncState.REASON_MISSING_BASELINE)
	assert_true(pipeline.get_router().resync_state.needs_resync(LaneMetadata.LANE_OVERLAY))
	assert_eq(pipeline.get_router().resync_state.get_reason(LaneMetadata.LANE_OVERLAY), ResyncState.REASON_WRONG_BASELINE)


func test_resync_required_marks_pending_without_emitting_request() -> void:
	var pipeline := RealtimePacketPipeline.new()
	pipeline.begin_match("match-1")
	var emitted := {"count": 0}
	pipeline.resync_request_required.connect(func(_lane, _baseline_id, _sequence, _reason) -> void:
		emitted.count += 1
	)

	pipeline.apply_packet({"type": "resync_request", "match_id": "match-1", "lane": "world", "reason": "missing_baseline"})
	assert_eq(emitted.count, 0)
	pipeline.apply_packet({"type": "resync_required", "match_id": "match-1", "lane": "world", "reason": "missing_baseline"})

	assert_true(pipeline.get_router().baseline_tracker.needs_resync(LaneMetadata.LANE_WORLD))
	assert_true(pipeline.get_router().resync_state.needs_resync(LaneMetadata.LANE_WORLD))
	assert_eq(pipeline.get_router().resync_state.get_reason(LaneMetadata.LANE_WORLD), ResyncState.REASON_MISSING_BASELINE)
	assert_eq(emitted.count, 0)


func test_delayed_resync_required_does_not_regress_recovered_lane() -> void:
	var pipeline := RealtimePacketPipeline.new()
	pipeline.begin_match("match-1")
	var emitted := {"count": 0}
	pipeline.resync_request_required.connect(func(_lane, _baseline_id, _sequence, _reason) -> void:
		emitted.count += 1
	)
	pipeline.apply_packet({"type": "world_full", "match_id": "match-1", "lane": "world", "baseline_id": "baseline-1", "sequence": 1, "snapshot_id": "snapshot-1", "ships": [], "bullets": [], "asteroids": [], "pickups": [], "is_final_chunk": true})
	pipeline.apply_packet({"type": "world_delta", "match_id": "match-1", "lane": "world", "baseline_id": "wrong-baseline", "sequence": 2, "ships": [], "bullets": [], "asteroids": [], "pickups": []})
	assert_eq(emitted.count, 1)
	pipeline.apply_packet({"type": "world_full", "match_id": "match-1", "lane": "world", "baseline_id": "baseline-2", "sequence": 2, "snapshot_id": "snapshot-2", "ships": [], "bullets": [], "asteroids": [], "pickups": [], "is_final_chunk": true})
	assert_true(pipeline.get_router().baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))
	assert_false(pipeline.get_router().baseline_tracker.needs_resync(LaneMetadata.LANE_WORLD))
	assert_false(pipeline.get_router().resync_state.needs_resync(LaneMetadata.LANE_WORLD))
	pipeline.apply_packet({"type": "resync_required", "match_id": "match-1", "lane": "world", "baseline_id": "baseline-1", "sequence": 1, "reason": "wrong_baseline"})
	assert_true(pipeline.get_router().baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))
	assert_false(pipeline.get_router().baseline_tracker.needs_resync(LaneMetadata.LANE_WORLD))
	assert_false(pipeline.get_router().resync_state.needs_resync(LaneMetadata.LANE_WORLD))
	assert_eq(emitted.count, 1)


func test_presentation_adapter_handles_null_overlay_cooldowns() -> void:
	var presentation_adapter := PresentationAdapter.new()
	var presentation_state := RealtimePresentationState.new()
	presentation_state.world_lane_state = WorldLaneState.new()
	presentation_state.overlay_lane_state = OverlayLaneState.new()
	presentation_state.session_lane_state = SessionLaneState.new()
	presentation_state.event_batch_applier = EventBatchApplier.new()
	presentation_state.overlay_lane_state.respawn_cooldown = null
	presentation_state.overlay_lane_state.primary_cooldown_remaining = null
	presentation_state.overlay_lane_state.secondary_cooldown_remaining = null

	var world_sync := FakePresentationTarget.new()
	var hud_flow := FakePresentationTarget.new()
	var event_flow := FakeEventFlow.new()

	presentation_adapter.fanout_lane_states(presentation_state, world_sync, hud_flow, event_flow)

	assert_not_null(world_sync.last_world_lane_state)
	assert_not_null(hud_flow.last_overlay_lane_state)
	assert_eq(hud_flow.last_overlay_lane_state.respawn_cooldown, null)
	assert_eq(hud_flow.last_overlay_lane_state.primary_cooldown_remaining, null)
	assert_eq(hud_flow.last_overlay_lane_state.secondary_cooldown_remaining, null)
	assert_eq(event_flow.apply_server_events_call_count, 0)


func test_repeated_batch_id_still_applies_unseen_event_ids() -> void:
	var applier := EventBatchApplier.new()
	var sink := FakeEventSink.new()

	applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "presentation-event-1", "type": "spark", "payload": {"value": 1}},
			],
		},
		sink
	)
	var applied := applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "presentation-event-2", "type": "spark", "payload": {"value": 2}},
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
				{"event_id": "presentation-event-1", "type": "spark", "payload": {"value": 1}},
			],
		},
		sink
	)
	var applied := applier.apply_event_batch(
		{
			"batch_id": "batch-2",
			"events": [
				{"event_id": "presentation-event-1", "type": "spark", "payload": {"value": 2}},
			],
		},
		sink
	)

	assert_false(applied)
	assert_eq(sink.handled_events.size(), 1)



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






