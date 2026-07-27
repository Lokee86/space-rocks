extends GutTest

const MatchResultWindowScene := preload("res://scenes/ui/dialogs/match_result_window.tscn")
const PlayerScoreRow := preload("res://scripts/ui/match_results/player_score_row.gd")


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
	var score_rows: Array[PlayerScoreRow] = []
	for child in score_container.get_children():
		var child_row := child as PlayerScoreRow
		if child_row != null:
			score_rows.append(child_row)

	assert_eq(score_rows.size(), 1)

	var row: PlayerScoreRow = score_rows[0]
	assert_eq((row.get_node("%PlayerIDLabel") as Label).text, "player-1")
	assert_eq((row.get_node("%GameDeathsLabel") as Label).text, "2")
	assert_eq((row.get_node("%GameScoreLabel") as Label).text, "450")
	assert_null(row.get_node_or_null("%GameKillsLabel"))


func test_ffa_rows_are_sorted_by_score_descending() -> void:
	var window := await _create_window()
	window.apply_rows([
		{"game_player_id": "low", "score": 10, "ship_deaths": 0},
		{"game_player_id": "high", "score": 300, "ship_deaths": 2},
		{"game_player_id": "middle", "score": 125, "ship_deaths": 1},
	])
	await get_tree().process_frame

	var ids := []
	for child in (window.get_node("%ScoreContainer") as Container).get_children():
		if child is PlayerScoreRow:
			ids.append((child.get_node("%PlayerIDLabel") as Label).text)
	assert_eq(ids, ["high", "middle", "low"])


func test_team_rows_are_grouped_by_team_score_and_subsorted() -> void:
	var window := await _create_window()
	window.apply_rows([
		{"game_player_id": "team1-low", "team_id": "team_1", "score": 50, "ship_deaths": 0},
		{"game_player_id": "team2-high", "team_id": "team_2", "score": 400, "ship_deaths": 1},
		{"game_player_id": "team1-high", "team_id": "team_1", "score": 250, "ship_deaths": 2},
	])
	await get_tree().process_frame

	var children := (window.get_node("%ScoreContainer") as Container).get_children()
	var team_swatch := children[0].get_child(0) as ColorRect
	assert_eq(team_swatch.size_flags_vertical, Control.SIZE_SHRINK_CENTER)
	assert_almost_eq(team_swatch.size.x, team_swatch.size.y, 0.001)
	assert_eq((children[0].get_child(1) as Label).text, "TEAM 2  —  400")
	assert_eq((children[1].get_node("%PlayerIDLabel") as Label).text, "team2-high")
	assert_eq((children[2].get_child(1) as Label).text, "TEAM 1  —  300")
	assert_eq((children[3].get_node("%PlayerIDLabel") as Label).text, "team1-high")
	assert_eq((children[4].get_node("%PlayerIDLabel") as Label).text, "team1-low")
	assert_true(window.find_child("ResultsScroll", true, false) is ScrollContainer)


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


func test_single_player_replay_button_disables_when_unavailable() -> void:
	var window := await _create_window()
	window.configure_for_mode("single_player")
	window.set_replay_available(false)

	assert_true((window.get_node("%LobbyReplayButton") as BaseButton).disabled)


func test_multiplayer_replay_button_remains_enabled_when_unavailable() -> void:
	var window := await _create_window()
	window.configure_for_mode("multiplayer")
	window.set_replay_available(false)

	assert_false((window.get_node("%LobbyReplayButton") as BaseButton).disabled)


func _create_window() -> Control:
	var window := MatchResultWindowScene.instantiate()
	add_child_autofree(window)
	await get_tree().process_frame
	return window
