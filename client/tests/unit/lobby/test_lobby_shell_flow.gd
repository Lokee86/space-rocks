extends GutTest

const LobbyShellFlow := preload("res://scripts/lobby/lobby_shell_flow.gd")


class FakeLobbyFlow:
	extends RefCounted

	func apply_room_snapshot(_packet: Dictionary) -> String:
		return ""

	func current_state():
		return {
			"room_state": "",
		}


class FakeSessionContext:
	extends RefCounted

	func activate_requested_mode() -> void:
		pass

	func should_show_multiplayer_lobby(_room_state) -> bool:
		return false


class FakeLobbyNetworkActions:
	extends RefCounted

	var leave_calls := 0

	func send_ready_requested(_ready: bool) -> void:
		pass

	func send_start_game_requested() -> void:
		pass

	func send_leave_requested() -> void:
		leave_calls += 1


class FakeLobbyReturnFlow:
	extends RefCounted

	var return_after_leave_calls := 0

	func return_after_leave() -> void:
		return_after_leave_calls += 1


class FakeMultiplayerLobbyPresenter:
	extends RefCounted

	func clear_lobby() -> void:
		pass

	func show_lobby(_canvas_layer, _state, _callbacks) -> void:
		pass


func test_lobby_leave_sends_leave_and_returns_after_leave() -> void:
	var lobby_flow := FakeLobbyFlow.new()
	var session_context := FakeSessionContext.new()
	var network_actions := FakeLobbyNetworkActions.new()
	var return_flow := FakeLobbyReturnFlow.new()
	var presenter := FakeMultiplayerLobbyPresenter.new()
	var main_menu := Control.new()
	var canvas_layer := CanvasLayer.new()
	var flow := LobbyShellFlow.new(
		lobby_flow,
		session_context,
		network_actions,
		return_flow,
		presenter,
		main_menu,
		canvas_layer,
		Callable())

	add_child_autofree(main_menu)
	add_child_autofree(canvas_layer)

	flow._on_lobby_leave_requested()

	assert_eq(network_actions.leave_calls, 1)
	assert_eq(return_flow.return_after_leave_calls, 1)
