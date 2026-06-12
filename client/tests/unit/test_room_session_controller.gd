extends GutTest

const RoomSessionController := preload("res://scripts/session/room_session_controller.gd")


class FakeSessionContext:
	extends RefCounted

	var clear_calls := 0

	func clear() -> void:
		clear_calls += 1

	func activate_requested_mode() -> void:
		pass

	func should_show_multiplayer_lobby(_room_state: String) -> bool:
		return false


class FakeConnectionService:
	extends RefCounted

	func send_set_ready_request(_ready: bool) -> void:
		pass

	func send_start_game_request() -> void:
		pass

	func send_leave_room_request() -> void:
		pass


class FakeShellBootFlow:
	extends RefCounted

	var clear_calls := 0

	func clear() -> void:
		clear_calls += 1


class Probe:
	extends RefCounted

	var calls := 0

	func mark_called() -> void:
		calls += 1


func test_lobby_return_cleanup_clears_session_context_and_shell_boot_flow() -> void:
	var setup := _create_controller()

	setup.controller.lobby_return_flow.return_after_leave()

	assert_eq(setup.session_context.clear_calls, 1)
	assert_eq(setup.shell_boot_flow.clear_calls, 1)


func test_configure_lobby_leave_return_destination_passes_destination_to_lobby_return_flow() -> void:
	var setup := _create_controller()
	var destination_probe := Probe.new()

	setup.controller.configure_lobby_leave_return_destination(Callable(destination_probe, "mark_called"))
	setup.controller.lobby_return_flow.return_after_leave()

	assert_eq(destination_probe.calls, 1)


func _create_controller() -> Dictionary:
	var main_menu := Control.new()
	var canvas_layer := CanvasLayer.new()
	var session_context := FakeSessionContext.new()
	var connection_service := FakeConnectionService.new()
	var shell_boot_flow := FakeShellBootFlow.new()
	var controller := RoomSessionController.new()

	add_child_autofree(main_menu)
	add_child_autofree(canvas_layer)
	main_menu.hide()

	controller.configure(
		main_menu,
		canvas_layer,
		session_context,
		connection_service,
		shell_boot_flow,
		Callable()
	)

	return {
		"controller": controller,
		"main_menu": main_menu,
		"session_context": session_context,
		"shell_boot_flow": shell_boot_flow,
	}
