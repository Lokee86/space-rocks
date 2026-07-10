extends GutTest

const GameplaySessionController := preload("res://scripts/session/gameplay_session_controller.gd")


class FakeConnectionService:
	extends Node

	var close_calls := 0
	var events: Array[String] = []

	func close_gracefully() -> void:
		close_calls += 1
		events.append("close_gracefully_started")
		await get_tree().process_frame
		events.append("close_gracefully_finished")


class FakeSessionContext:
	extends RefCounted

	var clear_calls := 0

	func clear() -> void:
		clear_calls += 1


class FakeShellBootFlow:
	extends RefCounted

	var clear_calls := 0

	func clear() -> void:
		clear_calls += 1


class ReplayProbe:
	extends RefCounted

	var events: Array[String] = []

	func mark_replay_requested() -> void:
		events.append("replay_requested")


class FakePresentationState:
	extends RefCounted

	var world_lane_state = null
	var overlay_lane_state = null
	var session_lane_state = null
	var event_batch_applier = null


class FakePresentationBridge:
	extends RefCounted

	var active_calls := 0
	var flush_calls := 0
	var reset_calls := 0
	var handled_packets: Array = []

	func activate() -> void:
		active_calls += 1

	func handle_gameplay_packet(packet: Dictionary) -> void:
		handled_packets.append(packet)

	func flush_pending() -> bool:
		flush_calls += 1
		return true

	func reset() -> void:
		reset_calls += 1


class FakeRealtimePacketPipeline:
	extends RefCounted

	var ready := true
	var presentation_state := FakePresentationState.new()

	func is_gameplay_ready() -> bool:
		return ready

	func get_presentation_state():
		return presentation_state


class FakeRuntimeContext:
	extends RefCounted

	var world_sync := {"world": true}


class FakeGameplayShellFlow:
	extends RefCounted

	var runtime_context := FakeRuntimeContext.new()


class FakeGameplayComposition:
	extends RefCounted

	var gameplay_shell_flow := FakeGameplayShellFlow.new()
	var gameplay_hud_flow := {"hud": true}
	var process_count := 0
	var events: Array = []
	var required_lane_baselines_synced_values: Array = []
	var devtools_states: Array = []
	var restore_calls := 0
	var reset_calls := 0

	func set_required_lane_baselines_synced(value: bool) -> void:
		required_lane_baselines_synced_values.append(value)

	func process(_delta: float, _required_lane_baselines_synced: bool) -> void:
		process_count += 1
		if events is Array:
			events.append("process")

	func apply_devtools_gameplay_state(state: Dictionary) -> void:
		devtools_states.append(state)

	func handle_devtools_input(_event: InputEvent) -> bool:
		events.append("devtools_input")
		return true

	func handle_gameplay_input(_event: InputEvent) -> bool:
		events.append("gameplay_input")
		return true

	func restore_alive_presentation_from_realtime_state(_presentation_state) -> void:
		restore_calls += 1

	func reset() -> void:
		reset_calls += 1


class FakeNotReadyGameplayReadiness:
	extends RefCounted

	func is_gameplay_ready() -> bool:
		return false



func _make_controller(pipeline_ready := true) -> Dictionary:
	var controller := GameplaySessionController.new()
	var connection_service := FakeConnectionService.new()
	var pipeline := FakeRealtimePacketPipeline.new()
	pipeline.ready = pipeline_ready
	var presentation_bridge := FakePresentationBridge.new()
	var composition := FakeGameplayComposition.new()
	var session_context := FakeSessionContext.new()
	var shell_boot_flow := FakeShellBootFlow.new()
	add_child_autofree(connection_service)
	add_child_autofree(controller)
	controller.connection_service = connection_service
	controller.realtime_packet_pipeline = pipeline
	controller.presentation_bridge = presentation_bridge
	controller.gameplay_composition = composition
	controller.session_context = session_context
	controller.shell_boot_flow = shell_boot_flow
	controller.logger = Callable()
	return {
		"controller": controller,
		"connection_service": connection_service,
		"pipeline": pipeline,
		"presentation_bridge": presentation_bridge,
		"gameplay_composition": composition,
		"session_context": session_context,
		"shell_boot_flow": shell_boot_flow,
	}


func test_replay_waits_for_graceful_close_before_emitting_replay_requested() -> void:
	var controller := GameplaySessionController.new()
	var connection_service := FakeConnectionService.new()
	var session_context := FakeSessionContext.new()
	var shell_boot_flow := FakeShellBootFlow.new()
	var replay_probe := ReplayProbe.new()

	add_child_autofree(connection_service)
	add_child_autofree(controller)
	controller.connection_service = connection_service
	controller.session_context = session_context
	controller.shell_boot_flow = shell_boot_flow
	controller.logger = Callable()
	connection_service.events = replay_probe.events
	controller.replay_requested.connect(Callable(replay_probe, "mark_replay_requested"))

	await controller._on_gameplay_replay_requested()

	assert_eq(connection_service.close_calls, 1)
	assert_eq(session_context.clear_calls, 1)
	assert_eq(shell_boot_flow.clear_calls, 1)
	assert_eq(replay_probe.events, ["close_gracefully_started", "close_gracefully_finished", "replay_requested"])


func test_first_ready_gameplay_packet_schedules_and_performs_initial_presentation_fanout() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_bridge: FakePresentationBridge = setup["presentation_bridge"]

	controller.accepts_gameplay_packets = true

	controller.handle_gameplay_packet({"type": "world_delta"})

	assert_eq(presentation_bridge.handled_packets, [{"type": "world_delta"}])
	assert_eq(presentation_bridge.flush_calls, 0)
	controller._process(0.016)
	assert_eq(presentation_bridge.flush_calls, 1)


func test_later_gameplay_packet_can_delegate_presentation_again_after_flush() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_bridge: FakePresentationBridge = setup["presentation_bridge"]

	controller.accepts_gameplay_packets = true

	controller.handle_gameplay_packet({"type": "world_delta"})
	controller._process(0.016)

	assert_eq(presentation_bridge.handled_packets.size(), 1)
	assert_eq(presentation_bridge.flush_calls, 1)

	controller.handle_gameplay_packet({"type": "world_delta"})

	assert_eq(presentation_bridge.handled_packets.size(), 2)
	assert_eq(presentation_bridge.handled_packets[1], {"type": "world_delta"})


func test_gameplay_packet_is_delegated_before_readiness_and_flush_is_gated() -> void:
	var setup = _make_controller(false)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_bridge: FakePresentationBridge = setup["presentation_bridge"]
	var gameplay_composition: FakeGameplayComposition = setup["gameplay_composition"]

	controller.accepts_gameplay_packets = true
	controller.gameplay_composition = gameplay_composition
	controller.handle_gameplay_packet({"type": "world_delta"})
	controller._process(0.016)

	assert_eq(presentation_bridge.handled_packets, [{"type": "world_delta"}])
	assert_eq(presentation_bridge.flush_calls, 1)
	assert_eq(gameplay_composition.process_count, 1)


func test_flush_runs_before_gameplay_process() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_bridge: FakePresentationBridge = setup["presentation_bridge"]
	var gameplay_composition := FakeGameplayComposition.new()

	controller.accepts_gameplay_packets = true
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = null
	gameplay_composition.events = []

	controller._process(0.016)

	assert_eq(presentation_bridge.flush_calls, 1)
	assert_eq(gameplay_composition.process_count, 1)


func test_begin_accepting_gameplay_packets_activates_presentation_bridge() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_bridge: FakePresentationBridge = setup["presentation_bridge"]

	controller.begin_accepting_gameplay_packets()

	assert_true(controller.accepts_gameplay_packets)
	assert_eq(presentation_bridge.active_calls, 1)


func test_reset_delegates_to_presentation_bridge() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_bridge: FakePresentationBridge = setup["presentation_bridge"]
	var gameplay_composition: FakeGameplayComposition = setup["gameplay_composition"]

	controller.accepts_gameplay_packets = true
	controller.reset()

	assert_false(controller.accepts_gameplay_packets)
	assert_eq(presentation_bridge.reset_calls, 1)
	assert_eq(gameplay_composition.reset_calls, 1)




func test_input_ignored_when_gameplay_packets_are_not_accepted() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var gameplay_composition := FakeGameplayComposition.new()
	var events: Array = []
	var input_event := InputEventMouseButton.new()
	var viewport := controller.get_viewport()

	controller.accepts_gameplay_packets = false
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = null
	gameplay_composition.events = events

	controller._input(input_event)

	assert_true(events.is_empty())
	assert_eq(viewport.is_input_handled() if viewport.has_method("is_input_handled") else false, false)


func test_unhandled_input_ignored_when_gameplay_packets_are_not_accepted() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var gameplay_composition := FakeGameplayComposition.new()
	var events: Array = []
	var input_event := InputEventMouseButton.new()

	controller.accepts_gameplay_packets = false
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = null
	gameplay_composition.events = events

	controller._unhandled_input(input_event)

	assert_true(events.is_empty())
