extends GutTest

const MatchResultsFlow := preload("res://scripts/ui/match_results/match_results_flow.gd")


func test_show_results_mounts_window() -> void:
	var mount_parent := Control.new()
	var flow := MatchResultsFlow.new()

	add_child_autofree(mount_parent)
	flow.configure(mount_parent)

	var window := flow.show_results("single_player", [])

	assert_not_null(window)
	assert_eq(mount_parent.get_child_count(), 1)
	assert_eq(mount_parent.get_child(0), window)


func test_show_results_twice_clears_old_window() -> void:
	var mount_parent := Control.new()
	var flow := MatchResultsFlow.new()

	add_child_autofree(mount_parent)
	flow.configure(mount_parent)

	var first_window := flow.show_results("single_player", [])
	var second_window := flow.show_results("single_player", [])

	assert_false(is_instance_valid(first_window))
	assert_not_null(second_window)
	assert_eq(mount_parent.get_child_count(), 1)
	assert_eq(mount_parent.get_child(0), second_window)


func test_single_player_lobby_replay_emits_replay_requested() -> void:
	var flow := await _create_flow("single_player")

	watch_signals(flow)
	(flow.window.get_node("%LobbyReplayButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(flow, "replay_requested")
	assert_signal_not_emitted(flow, "return_to_lobby_requested")


func test_multiplayer_lobby_replay_emits_return_to_lobby_requested() -> void:
	var flow := await _create_flow("multiplayer")

	watch_signals(flow)
	(flow.window.get_node("%LobbyReplayButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(flow, "return_to_lobby_requested")
	assert_signal_not_emitted(flow, "replay_requested")


func test_menu_emits_return_to_pregame_requested() -> void:
	var flow := await _create_flow("single_player")

	watch_signals(flow)
	(flow.window.get_node("%MenuButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(flow, "return_to_pregame_requested")


func test_quit_emits_quit_to_main_menu_requested() -> void:
	var flow := await _create_flow("single_player")

	watch_signals(flow)
	(flow.window.get_node("%QuitButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(flow, "quit_to_main_menu_requested")


func _create_flow(session_mode: String) -> MatchResultsFlow:
	var mount_parent := Control.new()
	var flow := MatchResultsFlow.new()

	add_child_autofree(mount_parent)
	flow.configure(mount_parent)
	flow.show_results(session_mode, [])
	await get_tree().process_frame
	return flow
