extends GutTest

const LobbySessionState := preload("res://scripts/lobby/lobby_session_state.gd")


func test_is_local_owner_uses_local_player_id() -> void:
	var state := LobbySessionState.new()
	state.local_player_id = "Player-1"
	state.owner_id = "Player-1"

	assert_true(state.is_local_owner())
