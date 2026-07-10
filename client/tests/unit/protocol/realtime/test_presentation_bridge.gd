extends GutTest

const PresentationBridge := preload("res://scripts/protocol/realtime/presentation_bridge.gd")

class FakeLogger:
	var messages: Array = []

	func record(message: String) -> void:
		messages.append(message)

class FakePacketPipeline:
	extends RefCounted

	var gameplay_ready := true
	var presentation_state := {"presentation": true}

	func is_gameplay_ready() -> bool:
		return gameplay_ready

	func get_presentation_state():
		return presentation_state

class FakeGameplayComposition:
	extends RefCounted

	var gameplay_shell_flow := {
		"runtime_context": {"world_sync": "world-sync"}
	}
	var gameplay_hud_flow := "hud-flow"
	var devtools_states: Array = []
	var restore_calls: Array = []
	var call_order: Array = []
	var event_lifecycle_flow := "event-flow"

	func get_event_lifecycle_flow():
		call_order.append("get_event_lifecycle_flow")
		return event_lifecycle_flow

	func apply_devtools_gameplay_state(state: Dictionary) -> void:
		call_order.append("apply_devtools_gameplay_state")
		devtools_states.append(state)

	func restore_alive_presentation_from_realtime_state(presentation_state) -> void:
		call_order.append("restore_alive_presentation_from_realtime_state")
		restore_calls.append(presentation_state)

class FakePresentationAdapter:
	extends RefCounted

	var calls: Array = []

	func fanout_lane_states(presentation_state, world_sync_ref = null, gameplay_hud_flow_ref = null, event_flow_ref = null) -> void:
		calls.append({
			"presentation_state": presentation_state,
			"world_sync": world_sync_ref,
			"gameplay_hud_flow": gameplay_hud_flow_ref,
			"event_flow": event_flow_ref,
		})

func _make_bridge(active := true, ready := true) -> Dictionary:
	var bridge := PresentationBridge.new()
	var pipeline := FakePacketPipeline.new()
	pipeline.gameplay_ready = ready
	var presentation_adapter := FakePresentationAdapter.new()
	var composition := FakeGameplayComposition.new()
	var logger := FakeLogger.new()
	bridge.configure(pipeline, presentation_adapter, composition, Callable(logger, "record"))
	if active:
		bridge.activate()
	return {
		"bridge": bridge,
		"pipeline": pipeline,
		"presentation_adapter": presentation_adapter,
		"composition": composition,
		"logger": logger,
	}

func test_bridge_starts_inactive_without_pending_presentation() -> void:
	var bridge := PresentationBridge.new()

	assert_false(bridge.has_pending_presentation())

func test_mark_pending_is_ignored_until_bridge_is_activated() -> void:
	var bridge := PresentationBridge.new()

	bridge.mark_pending()

	assert_false(bridge.has_pending_presentation())

func test_handle_gameplay_packet_coalesces_multiple_packets_into_single_flush() -> void:
	var fixture := _make_bridge()
	var bridge = fixture.bridge
	var pipeline = fixture.pipeline
	var presentation_adapter = fixture.presentation_adapter
	var composition = fixture.composition

	bridge.handle_gameplay_packet({"type": "event_batch", "batch_id": "first", "events": [{"type": "spawn"}]})
	bridge.handle_gameplay_packet({"type": "event_batch", "batch_id": "second", "events": [{"type": "move"}]})

	assert_true(bridge.has_pending_presentation())
	assert_eq(presentation_adapter.calls.size(), 0)

	assert_true(bridge.flush_pending())
	assert_false(bridge.has_pending_presentation())
	assert_eq(presentation_adapter.calls.size(), 1)
	assert_eq(composition.devtools_states.size(), 1)
	assert_eq(composition.restore_calls.size(), 1)
	assert_true(pipeline.gameplay_ready)

func test_flush_pending_requires_readiness_and_keeps_pending_until_ready() -> void:
	var fixture := _make_bridge(true, false)
	var bridge = fixture.bridge
	var pipeline = fixture.pipeline
	var presentation_adapter = fixture.presentation_adapter
	var composition = fixture.composition

	bridge.handle_gameplay_packet({"type": "event_batch", "batch_id": "blocked", "events": []})

	assert_true(bridge.has_pending_presentation())
	assert_false(bridge.flush_pending())
	assert_true(bridge.has_pending_presentation())
	assert_eq(presentation_adapter.calls.size(), 0)
	assert_eq(composition.devtools_states.size(), 0)
	assert_eq(composition.restore_calls.size(), 0)

	pipeline.gameplay_ready = true

	assert_true(bridge.flush_pending())
	assert_false(bridge.has_pending_presentation())
	assert_eq(presentation_adapter.calls.size(), 1)
	assert_eq(composition.devtools_states.size(), 1)
	assert_eq(composition.restore_calls.size(), 1)

func test_flush_pending_fanout_order_runs_before_devtools_and_restore() -> void:
	var fixture := _make_bridge()
	var bridge = fixture.bridge
	var composition = fixture.composition
	var presentation_adapter = fixture.presentation_adapter

	bridge.handle_gameplay_packet({"type": "event_batch", "batch_id": "ordered", "events": [{"type": "alpha"}, {"type": "beta"}]})

	assert_true(bridge.flush_pending())
	assert_eq(presentation_adapter.calls.size(), 1)
	assert_eq(composition.call_order, [
		"get_event_lifecycle_flow",
		"apply_devtools_gameplay_state",
		"restore_alive_presentation_from_realtime_state",
	])
	assert_eq(presentation_adapter.calls[0]["world_sync"], "world-sync")
	assert_eq(presentation_adapter.calls[0]["gameplay_hud_flow"], "hud-flow")
	assert_eq(presentation_adapter.calls[0]["event_flow"], "event-flow")
	assert_eq(composition.restore_calls[0]["presentation"], true)

func test_deactivate_clears_owned_pending_state() -> void:
	var bridge := PresentationBridge.new()
	bridge.activate()
	bridge.mark_pending()

	bridge.deactivate()

	assert_false(bridge.has_pending_presentation())

func test_reset_clears_activation_and_pending_state() -> void:
	var bridge := PresentationBridge.new()
	bridge.activate()
	bridge.mark_pending()

	bridge.reset()

	assert_false(bridge.has_pending_presentation())
	bridge.mark_pending()
	assert_false(bridge.has_pending_presentation())

func test_reactivate_after_reset_restores_pending_ownership() -> void:
	var bridge := PresentationBridge.new()
	bridge.activate()
	bridge.mark_pending()
	bridge.reset()

	bridge.activate()
	bridge.mark_pending()

	assert_true(bridge.has_pending_presentation())
