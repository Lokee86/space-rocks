extends GutTest

const MatchResultWindowScene := preload("res://scenes/ui/dialogs/match_result_window.tscn")


func test_apply_rows_renders_player_score_row_without_kills_label() -> void:
	var window := MatchResultWindowScene.instantiate()
	add_child_autofree(window)

	window.apply_rows([
		{
			"game_player_id": "player-1",
			"ship_deaths": 2,
			"score": 450,
			"won": true,
		}
	])

	await get_tree().process_frame

	var score_container := window.get_node("%ScoreContainer")
	assert_eq(score_container.get_child_count(), 1)

	var row := score_container.get_child(0)
	assert_eq((row.get_node("%PlayerIDLabel") as Label).text, "player-1")
	assert_eq((row.get_node("%GameDeathsLabel") as Label).text, "2")
	assert_eq((row.get_node("%GameScoreLabel") as Label).text, "450")
	assert_null(row.get_node_or_null("%GameKillsLabel"))

const MatchResultWindowScene := preload("res://scenes/ui/dialogs/match_result_window.tscn")


func test_lobby_replay_button_emits_lobby_replay_requested() -> void:
	var window := await _create_window()

	watch_signals(window)
	(window.get_node("%LobbyReplayButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(window, "lobby_replay_requested")


func test_menu_button_emits_menu_requested() -> void:
	var window := await _create_window()

	watch_signals(window)
	(window.get_node("%MenuButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(window, "menu_requested")


func test_quit_button_emits_quit_requested() -> void:
	var window := await _create_window()

	watch_signals(window)
	(window.get_node("%QuitButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(window, "quit_requested")


func _create_window() -> Control:
	var window := MatchResultWindowScene.instantiate()
	add_child_autofree(window)
	await get_tree().process_frame
	return window
