extends GutTest

const GameplaySessionState := preload("res://scripts/gameplay/session/gameplay_session_state.gd")


func test_blank_room_state_can_process_gameplay_packets() -> void:
	assert_true(GameplaySessionState.can_process_gameplay_packets(""))


func test_in_game_can_process_gameplay_packets() -> void:
	assert_true(GameplaySessionState.can_process_gameplay_packets("InGame"))


func test_game_over_can_process_gameplay_packets() -> void:
	assert_true(GameplaySessionState.can_process_gameplay_packets("GameOver"))


func test_lobby_non_game_state_cannot_process_gameplay_packets() -> void:
	assert_false(GameplaySessionState.can_process_gameplay_packets("Lobby"))


func test_multiplayer_room_game_over_counts_as_game_over() -> void:
	assert_true(GameplaySessionState.is_game_over("Multiplayer", "GameOver", false))


func test_single_player_room_game_over_alone_does_not_count_as_game_over() -> void:
	assert_false(GameplaySessionState.is_game_over("SinglePlayer", "GameOver", false))


func test_single_player_hud_game_over_counts_as_game_over() -> void:
	assert_true(GameplaySessionState.is_game_over("SinglePlayer", "GameOver", true))
