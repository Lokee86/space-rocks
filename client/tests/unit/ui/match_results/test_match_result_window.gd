extends GutTest

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
