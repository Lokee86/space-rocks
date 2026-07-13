extends GutTest

const LobbySessionState := preload("res://scripts/lobby/lobby_session_state.gd")
const MultiplayerLobby := preload("res://scripts/ui/lobby/multiplayer_lobby.gd")
const MultiplayerLobbyPresenter := preload("res://scripts/lobby/multiplayer_lobby_presenter.gd")

var ready_calls: Array = []
var start_calls := 0
var leave_calls := 0


func before_each() -> void:
	ready_calls = []
	start_calls = 0
	leave_calls = 0


func test_show_lobby_mounts_typed_lobby_and_applies_state() -> void:
	var canvas_layer := CanvasLayer.new()
	add_child_autofree(canvas_layer)
	var presenter := MultiplayerLobbyPresenter.new()
	var state := LobbySessionState.new()
	state.apply_snapshot("ROOM-7", "lobby", "Player-1", "Player-1", 4, [{"player_id": "Player-1", "ready": true}])

	var lobby := presenter.show_lobby(canvas_layer, state, {})

	assert_true(lobby is MultiplayerLobby)
	assert_eq(presenter.current_lobby(), lobby)
	assert_eq(lobby.room_code_label.text, "ROOM-7")
	assert_false(lobby.start_game_button.disabled)
	presenter.clear_lobby()
	await get_tree().process_frame


func test_show_lobby_forwards_all_callback_signals() -> void:
	var canvas_layer := CanvasLayer.new()
	add_child_autofree(canvas_layer)
	var presenter := MultiplayerLobbyPresenter.new()
	var state := LobbySessionState.new()
	var callbacks := {
		"ready_requested": Callable(self, "_on_ready_requested"),
		"start_game_requested": Callable(self, "_on_start_game_requested"),
		"leave_requested": Callable(self, "_on_leave_requested"),
	}

	var lobby := presenter.show_lobby(canvas_layer, state, callbacks)
	lobby.ready_requested.emit(true)
	lobby.start_game_requested.emit()
	lobby.leave_requested.emit()

	assert_eq(ready_calls, [true])
	assert_eq(start_calls, 1)
	assert_eq(leave_calls, 1)
	presenter.clear_lobby()
	await get_tree().process_frame


func _on_ready_requested(ready: bool) -> void:
	ready_calls.append(ready)


func _on_start_game_requested() -> void:
	start_calls += 1


func _on_leave_requested() -> void:
	leave_calls += 1
