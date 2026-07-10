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


class FakeNotReadyGameplayReadiness:
	extends RefCounted

	func is_gameplay_ready() -> bool:
		return false


class FakePresentationAdapter:
	extends RefCounted

	var fanout_count := 0
	var events: Array = []
	var marked_fanned_out := 0

	func can_fanout() -> bool:
		return true

	func fanout_lane_states(_presentation_state, _world_sync, _gameplay_hud_flow, _event_lifecycle_flow) -> void:
		fanout_count += 1
		if events is Array:
			events.append("fanout")

	func mark_fanned_out() -> void:
		marked_fanned_out += 1


func _make_controller(pipeline_ready := true) -> Dictionary:
	var controller := GameplaySessionController.new()
	var connection_service := FakeConnectionService.new()
	var pipeline := FakeRealtimePacketPipeline.new()
	pipeline.ready = pipeline_ready
	var session_context := FakeSessionContext.new()
	var shell_boot_flow := FakeShellBootFlow.new()
	add_child_autofree(connection_service)
	add_child_autofree(controller)
	controller.configure(
		connection_service,
		pipeline,
		Node2D.new(),
		Node.new(),
		Node2D.new(),
		Node2D.new(),
		Node2D.new(),
		Control.new(),
		Control.new(),
		Control.new(),
		session_context,
		shell_boot_flow,
		Callable()
	)
	return {
		"controller": controller,
		"connection_service": connection_service,
		"pipeline": pipeline,
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
	var presentation_adapter := FakePresentationAdapter.new()
	var gameplay_composition := FakeGameplayComposition.new()
	var events: Array = []

	controller.accepts_gameplay_packets = true
	controller.gameplay_presentation_adapter = presentation_adapter
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = null
	controller.gameplay_presentation_adapter.events = events
	gameplay_composition.events = events

	controller.handle_gameplay_packet({"type": "world_delta"})

	assert_eq(presentation_adapter.fanout_count, 0)
	controller._process(0.016)
	assert_eq(presentation_adapter.fanout_count, 1)
	assert_eq(events, ["fanout", "process"])


func test_later_gameplay_packet_can_mark_presentation_dirty_again_after_fanout() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_adapter := FakePresentationAdapter.new()
	var gameplay_composition := FakeGameplayComposition.new()
	var events: Array = []

	controller.accepts_gameplay_packets = true
	controller.gameplay_presentation_adapter = presentation_adapter
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = null
	controller.gameplay_presentation_adapter.events = events
	gameplay_composition.events = events

	controller.handle_gameplay_packet({"type": "world_delta"})
	controller._process(0.016)

	assert_eq(presentation_adapter.fanout_count, 1)

	controller._process(0.016)

	assert_eq(presentation_adapter.fanout_count, 1)

	controller.handle_gameplay_packet({"type": "world_delta"})
	controller._process(0.016)

	assert_eq(presentation_adapter.fanout_count, 2)


func test_deferred_presentation_does_not_fan_out_when_gameplay_is_not_ready() -> void:
	var setup = _make_controller(false)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_adapter := FakePresentationAdapter.new()
	var gameplay_composition := FakeGameplayComposition.new()
	var events: Array = []

	controller.accepts_gameplay_packets = true
	controller.gameplay_presentation_adapter = presentation_adapter
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = null
	controller.gameplay_presentation_adapter.events = events
	gameplay_composition.events = events

	controller.handle_gameplay_packet({"type": "world_delta"})
	controller._process(0.016)

	assert_eq(presentation_adapter.fanout_count, 0)


func test_dirty_presentation_fanout_runs_before_gameplay_process() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_adapter := FakePresentationAdapter.new()
	var gameplay_composition := FakeGameplayComposition.new()
	var events: Array = []

	controller.accepts_gameplay_packets = true
	controller.gameplay_presentation_adapter = presentation_adapter
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = null
	presentation_adapter.events = events
	gameplay_composition.events = events

	controller.handle_gameplay_packet({"type": "world_delta"})
	controller._process(0.016)

	assert_eq(events, ["fanout", "process"])


func test_input_ignored_when_gameplay_packets_are_not_accepted() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_adapter := FakePresentationAdapter.new()
	var gameplay_composition := FakeGameplayComposition.new()
	var events: Array = []
	var input_event := InputEventMouseButton.new()
	var viewport := controller.get_viewport()

	controller.accepts_gameplay_packets = false
	controller.gameplay_presentation_adapter = presentation_adapter
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = null
	presentation_adapter.events = events
	gameplay_composition.events = events

	controller._input(input_event)

	assert_true(events.is_empty())
	assert_eq(viewport.has_method("is_input_handled") ? viewport.is_input_handled() : false, false)


func test_unhandled_input_ignored_when_gameplay_packets_are_not_accepted() -> void:
	var setup = _make_controller(true)
	var controller: GameplaySessionController = setup["controller"]
	var presentation_adapter := FakePresentationAdapter.new()
	var gameplay_composition := FakeGameplayComposition.new()
	var events: Array = []
	var input_event := InputEventMouseButton.new()

	controller.accepts_gameplay_packets = false
	controller.gameplay_presentation_adapter = presentation_adapter
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = null
	presentation_adapter.events = events
	gameplay_composition.events = events

	controller._unhandled_input(input_event)

	assert_true(events.is_empty())
