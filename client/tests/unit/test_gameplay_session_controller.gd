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
