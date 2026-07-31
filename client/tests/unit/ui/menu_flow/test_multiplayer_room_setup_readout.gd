extends GutTest

const RoomSetupScene := preload("res://scenes/ui/transmission_displays/multiplayer_room_setup_readout.tscn")


func test_multiplayer_defaults_to_arcade_survival_ffa() -> void:
	var setup := await _create_setup()

	assert_eq(setup.current_config(), {
		"preset_id": "arcade_survival",
		"starting_lives": 3,
		"infinite_lives": false,
		"target_score": 0,
		"team_structure": "ffa",
		"team_assignment_mode": "",
		"team_count": 0,
		"max_players": 8,
	})
	assert_false((setup.get_node("%TargetScoreRow") as Control).visible)
	assert_false((setup.get_node("%AssignmentRow") as Control).visible)
	assert_false((setup.get_node("%TeamCountRow") as Control).visible)


func test_score_attack_exposes_target_score_and_emits_mode_config() -> void:
	var setup := await _create_setup()
	var game_mode_select = setup.get_node("%GameModeSelect")
	var target_score_select = setup.get_node("%TargetScoreSelect")
	var lives_select = setup.get_node("%LivesSelect")
	game_mode_select.select_value("score_attack")
	game_mode_select.emit_signal("item_selected", game_mode_select.selected)
	target_score_select.select_value(2500)
	lives_select.select_value(5)
	watch_signals(setup)

	(setup.get_node("%CreateButton") as BaseButton).emit_signal("pressed")

	assert_true((setup.get_node("%TargetScoreRow") as Control).visible)
	assert_signal_emitted_with_parameters(setup, "create_requested", [{
		"preset_id": "score_attack",
		"starting_lives": 5,
		"infinite_lives": false,
		"target_score": 2500,
		"team_structure": "ffa",
		"team_assignment_mode": "",
		"team_count": 0,
		"max_players": 8,
	}])


func test_auto_balanced_selection_exposes_team_count() -> void:
	var setup := await _create_setup()
	var team_structure_select = setup.get_node("%TeamStructureSelect")
	var team_count_select = setup.get_node("%TeamCountSelect")
	team_structure_select.select_value("auto_balanced")
	team_structure_select.emit_signal("item_selected", team_structure_select.selected)
	team_count_select.select_value(4)

	assert_true((setup.get_node("%TeamCountRow") as Control).visible)
	assert_false((setup.get_node("%AssignmentRow") as Control).visible)
	assert_eq(setup.current_config().team_count, 4)


func test_custom_selection_exposes_assignment_mode() -> void:
	var setup := await _create_setup()
	var team_structure_select = setup.get_node("%TeamStructureSelect")
	var assignment_select = setup.get_node("%AssignmentSelect")
	team_structure_select.select_value("custom")
	team_structure_select.emit_signal("item_selected", team_structure_select.selected)
	assignment_select.select_value("player_selected")

	assert_true((setup.get_node("%AssignmentRow") as Control).visible)
	assert_false((setup.get_node("%TeamCountRow") as Control).visible)
	assert_eq(setup.current_config().team_assignment_mode, "player_selected")


func test_single_player_hides_multiplayer_rows_and_returns_mode_config() -> void:
	var setup := await _create_setup()
	setup.configure_single_player()
	var game_mode_select = setup.get_node("%GameModeSelect")
	var lives_select = setup.get_node("%LivesSelect")
	game_mode_select.select_value("score_attack")
	game_mode_select.emit_signal("item_selected", game_mode_select.selected)
	lives_select.select_value("infinite")

	assert_eq(setup.current_config(), {
		"preset_id": "score_attack",
		"starting_lives": 0,
		"infinite_lives": true,
		"target_score": 1000,
	})
	assert_false((setup.get_node("%TeamStructureRow") as Control).visible)
	assert_false((setup.get_node("%AssignmentRow") as Control).visible)
	assert_false((setup.get_node("%TeamCountRow") as Control).visible)
	assert_false((setup.get_node("%MaxPlayersRow") as Control).visible)
	assert_eq((setup.get_node("%Title") as Label).text, "SINGLE PLAYER CONFIGURATION")


func _create_setup() -> Control:
	var setup := RoomSetupScene.instantiate()
	add_child_autofree(setup)
	await get_tree().process_frame
	return setup
