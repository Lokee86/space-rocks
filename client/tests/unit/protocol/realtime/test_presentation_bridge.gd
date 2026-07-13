extends GutTest

const PresentationBridge := preload("res://scripts/protocol/realtime/presentation_bridge.gd")
const RealtimePacketPipeline := preload("res://scripts/networking/realtime/realtime_packet_pipeline.gd")
const RealtimePresentationState := preload("res://scripts/networking/realtime/realtime_presentation_state.gd")
const PresentationAdapter := preload("res://scripts/protocol/realtime/presentation_adapter.gd")

class FakePacketPipeline extends RealtimePacketPipeline:
	var gameplay_ready := true
	var presentation_state := RealtimePresentationState.new()

	func is_gameplay_ready() -> bool:
		return gameplay_ready

	func get_presentation_state():
		return presentation_state

class FakePresentationAdapter extends PresentationAdapter:
	var calls: Array = []

	func fanout_lane_states(presentation_state: RealtimePresentationState, world_sync_ref: WorldSync = null, gameplay_hud_flow_ref: GameplayHudFlow = null, event_flow_ref: GameplayEventLifecycleFlow = null, local_lifecycle_flow_ref: GameplayLocalLifecycleFlow = null) -> void:
		calls.append({
			"presentation_state": presentation_state,
			"world_sync": world_sync_ref,
			"gameplay_hud_flow": gameplay_hud_flow_ref,
			"event_flow": event_flow_ref,
			"local_lifecycle_flow": local_lifecycle_flow_ref,
		})

class FakeWorldSync extends WorldSync:
	pass

class FakeGameplayHudFlow extends GameplayHudFlow:
	pass

class FakeEventLifecycleFlow extends GameplayEventLifecycleFlow:
	pass

class FakeLocalLifecycleFlow extends GameplayLocalLifecycleFlow:
	pass

class FakeLogger:
	var messages: Array = []

	func record(message: String) -> void:
		messages.append(message)

class FakeGameplayComposition:
	extends GameplayComposition

	var devtools_states: Array = []
	var call_order: Array = []
	var event_lifecycle_flow: GameplayEventLifecycleFlow
	var local_lifecycle_flow: GameplayLocalLifecycleFlow

	func _init() -> void:
		world_sync = FakeWorldSync.new()
		gameplay_hud_flow = FakeGameplayHudFlow.new()
		event_lifecycle_flow = FakeEventLifecycleFlow.new()
		local_lifecycle_flow = FakeLocalLifecycleFlow.new()

	func get_event_lifecycle_flow():
		call_order.append("get_event_lifecycle_flow")
		return event_lifecycle_flow

	func get_local_lifecycle_flow():
		call_order.append("get_local_lifecycle_flow")
		return local_lifecycle_flow

	func get_world_sync() -> WorldSync:
		return world_sync

	func get_gameplay_hud_flow() -> GameplayHudFlow:
		return gameplay_hud_flow

	func apply_devtools_gameplay_state(state: Dictionary) -> void:
		call_order.append("apply_devtools_gameplay_state")
		devtools_states.append(state)

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


	pipeline.gameplay_ready = true

	assert_true(bridge.flush_pending())
	assert_false(bridge.has_pending_presentation())
	assert_eq(presentation_adapter.calls.size(), 1)
	assert_eq(composition.devtools_states.size(), 1)


func test_flush_pending_fanout_order_runs_before_devtools() -> void:
	var fixture := _make_bridge()
	var bridge = fixture.bridge
	var composition = fixture.composition
	var presentation_adapter = fixture.presentation_adapter

	bridge.handle_gameplay_packet({"type": "event_batch", "batch_id": "ordered", "events": [{"type": "alpha"}, {"type": "beta"}]})

	assert_true(bridge.flush_pending())
	assert_eq(presentation_adapter.calls.size(), 1)
	assert_eq(composition.call_order, [
		"get_event_lifecycle_flow",
		"get_local_lifecycle_flow",
		"apply_devtools_gameplay_state",
	])
	assert_eq(presentation_adapter.calls[0]["world_sync"], composition.world_sync)
	assert_eq(presentation_adapter.calls[0]["gameplay_hud_flow"], composition.gameplay_hud_flow)
	assert_eq(presentation_adapter.calls[0]["event_flow"], composition.event_lifecycle_flow)
	assert_eq(presentation_adapter.calls[0]["local_lifecycle_flow"], composition.local_lifecycle_flow)

func test_non_event_packet_flush_passes_local_lifecycle_without_event_flow() -> void:
	var fixture := _make_bridge()
	var bridge = fixture.bridge
	var composition = fixture.composition
	var presentation_adapter = fixture.presentation_adapter

	bridge.handle_gameplay_packet({"type": "world_state", "world": {"ships": {}}})

	assert_true(bridge.flush_pending())
	assert_eq(composition.call_order, ["get_local_lifecycle_flow", "apply_devtools_gameplay_state"])
	assert_eq(presentation_adapter.calls.size(), 1)
	assert_eq(presentation_adapter.calls[0]["local_lifecycle_flow"], composition.local_lifecycle_flow)
	assert_null(presentation_adapter.calls[0]["event_flow"])

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
