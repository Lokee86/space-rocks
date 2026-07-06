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
	assert_eq(
		replay_probe.events,
		["close_gracefully_started", "close_gracefully_finished", "replay_requested"]
	)

class FakeGameplayReadiness:
	extends RefCounted

	func is_gameplay_ready() -> bool:
		return true


class FakeGameplayStateFlow:
	extends RefCounted

	var ready := true
	var set_ready_calls: Array = []

	func is_gameplay_ready() -> bool:
		return ready

	func set_gameplay_readiness(_readiness) -> void:
		pass


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

	func restore_alive_presentation_from_realtime_router(_router) -> void:
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

	func bind_gameplay_readiness(_readiness) -> void:
		pass

	func can_fanout() -> bool:
		return true

	func fanout_lane_states(_router, _world_sync, _gameplay_hud_flow, _event_lifecycle_flow) -> void:
		fanout_count += 1
		if events is Array:
			events.append("fanout")

	func mark_fanned_out() -> void:
		marked_fanned_out += 1


func test_multiple_gameplay_packets_before_process_cause_only_one_presentation_fanout() -> void:
	var controller := GameplaySessionController.new()
	var presentation_adapter := FakePresentationAdapter.new()
	var gameplay_composition := FakeGameplayComposition.new()
	var gameplay_state_flow := FakeGameplayStateFlow.new()
	var gameplay_readiness := FakeGameplayReadiness.new()

	add_child_autofree(controller)
	controller.accepts_gameplay_packets = true
	controller.gameplay_presentation_adapter = presentation_adapter
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = gameplay_state_flow
	controller._gameplay_readiness = gameplay_readiness
	controller.gameplay_realtime_router = {"router": true}

	controller.handle_gameplay_packet({"type": "world_delta"})
	controller.handle_gameplay_packet({"type": "world_delta"})

	assert_eq(presentation_adapter.fanout_count, 0)

	controller._process(0.016)

	assert_eq(presentation_adapter.fanout_count, 1)

	controller._process(0.016)

	assert_eq(presentation_adapter.fanout_count, 1)

func test_later_gameplay_packet_can_mark_presentation_dirty_again_after_fanout() -> void:
	var controller := GameplaySessionController.new()
	var presentation_adapter := FakePresentationAdapter.new()
	var gameplay_composition := FakeGameplayComposition.new()
	var gameplay_state_flow := FakeGameplayStateFlow.new()
	var gameplay_readiness := FakeGameplayReadiness.new()

	add_child_autofree(controller)
	controller.accepts_gameplay_packets = true
	controller.gameplay_presentation_adapter = presentation_adapter
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = gameplay_state_flow
	controller._gameplay_readiness = gameplay_readiness
	controller.gameplay_realtime_router = {"router": true}

	controller.handle_gameplay_packet({"type": "world_delta"})
	controller._process(0.016)

	assert_eq(presentation_adapter.fanout_count, 1)

	controller._process(0.016)

	assert_eq(presentation_adapter.fanout_count, 1)

	controller.handle_gameplay_packet({"type": "world_delta"})
	controller._process(0.016)

	assert_eq(presentation_adapter.fanout_count, 2)

func test_deferred_presentation_does_not_fan_out_when_gameplay_is_not_ready() -> void:
	var controller := GameplaySessionController.new()
	var presentation_adapter := FakePresentationAdapter.new()
	var gameplay_composition := FakeGameplayComposition.new()
	var gameplay_state_flow := FakeGameplayStateFlow.new()
	var gameplay_readiness := FakeNotReadyGameplayReadiness.new()

	add_child_autofree(controller)
	controller.accepts_gameplay_packets = true
	controller.gameplay_presentation_adapter = presentation_adapter
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = gameplay_state_flow
	controller._gameplay_readiness = gameplay_readiness
	controller.gameplay_realtime_router = {"router": true}

	controller.handle_gameplay_packet({"type": "world_delta"})
	controller._process(0.016)

	assert_eq(presentation_adapter.fanout_count, 0)




func test_dirty_presentation_fanout_runs_before_gameplay_process() -> void:
	var controller := GameplaySessionController.new()
	var presentation_adapter := FakePresentationAdapter.new()
	var gameplay_composition := FakeGameplayComposition.new()
	var gameplay_state_flow := FakeGameplayStateFlow.new()
	var gameplay_readiness := FakeGameplayReadiness.new()
	var events: Array = []

	add_child_autofree(controller)
	controller.accepts_gameplay_packets = true
	controller.gameplay_presentation_adapter = presentation_adapter
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_state_flow = gameplay_state_flow
	controller._gameplay_readiness = gameplay_readiness
	controller.gameplay_realtime_router = {"router": true}
	presentation_adapter.events = events
	gameplay_composition.events = events

	controller.handle_gameplay_packet({"type": "world_delta"})
	controller._process(0.016)

	assert_eq(events, ["fanout", "process"])



