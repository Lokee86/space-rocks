extends GutTest

const LobbyReturnFlow := preload("res://scripts/lobby/lobby_return_flow.gd")


class FakeLobbyFlow:
	extends RefCounted

	var clear_calls := 0

	func clear() -> void:
		clear_calls += 1


class FakeMultiplayerLobbyPresenter:
	extends RefCounted

	var clear_lobby_calls := 0

	func clear_lobby() -> void:
		clear_lobby_calls += 1


class Probe:
	extends RefCounted

	var calls := 0

	func mark_called() -> void:
		calls += 1


func test_return_after_leave_clears_lobby_and_calls_cleanup_and_destination() -> void:
	var lobby_flow := FakeLobbyFlow.new()
	var presenter := FakeMultiplayerLobbyPresenter.new()
	var main_menu := Control.new()
	var cleanup_probe := Probe.new()
	var destination_probe := Probe.new()
	var flow := LobbyReturnFlow.new(lobby_flow, presenter, main_menu, Callable(cleanup_probe, "mark_called"))

	add_child_autofree(main_menu)
	main_menu.hide()
	flow.configure_return_destination(Callable(destination_probe, "mark_called"))

	flow.return_after_leave()

	assert_eq(lobby_flow.clear_calls, 1)
	assert_eq(presenter.clear_lobby_calls, 1)
	assert_eq(cleanup_probe.calls, 1)
	assert_eq(destination_probe.calls, 1)
	assert_false(main_menu.visible)


func test_return_after_leave_falls_back_to_main_menu_without_destination() -> void:
	var lobby_flow := FakeLobbyFlow.new()
	var presenter := FakeMultiplayerLobbyPresenter.new()
	var main_menu := Control.new()
	var flow := LobbyReturnFlow.new(lobby_flow, presenter, main_menu, Callable())

	add_child_autofree(main_menu)
	main_menu.hide()

	flow.return_after_leave()

	assert_true(main_menu.visible)


func test_return_to_main_menu_wrapper_uses_return_after_leave() -> void:
	var lobby_flow := FakeLobbyFlow.new()
	var presenter := FakeMultiplayerLobbyPresenter.new()
	var main_menu := Control.new()
	var destination_probe := Probe.new()
	var flow := LobbyReturnFlow.new(lobby_flow, presenter, main_menu, Callable())

	add_child_autofree(main_menu)
	main_menu.hide()
	flow.configure_return_destination(Callable(destination_probe, "mark_called"))

	flow.return_to_main_menu()

	assert_eq(destination_probe.calls, 1)
