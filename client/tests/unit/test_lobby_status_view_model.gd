extends GutTest

const Constants := preload("res://scripts/constants/constants.gd")
const LobbyStatusViewModel := preload("res://scripts/ui/lobby/lobby_status_view_model.gd")


func test_status_text_uses_local_player_id_for_owner_identity() -> void:
	var status := LobbyStatusViewModel.status_text(
		Constants.ROOM_STATE_LOBBY,
		"session-1",
		"Player-1",
		"Player-1",
		[],
		true
	)

	assert_eq(status, Constants.STATUS_READY_TO_START)
