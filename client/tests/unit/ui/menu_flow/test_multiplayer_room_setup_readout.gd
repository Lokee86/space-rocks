extends GutTest

const RoomSetupScene := preload("res://scenes/ui/transmission_displays/multiplayer_room_setup_readout.tscn")


func test_multiplayer_defaults_to_arcade_survival_ffa() -> void:
	var setup := await _create_setup()

	assert_eq(setup.current_config(), {
		"preset_id": "arcade_survival",
		"starting_lives": 3,
		"infinite_lives": false,
		"target_score": 0,
		"target_kills": 0,
		"team_structure": "ffa",
		"team_assignment_mode": "",
		"team_count": 0,
		"max_players": 8,
	})
	assert_false((setup.get_node("%TargetScoreRow") as Control).visible)
	assert_false((setup.get_node("%CustomTargetScoreRow") as Control).visible)
	assert_false((setup.get_node("%AssignmentRow") as Control).visible)
	assert_false((setup.get_node("%TeamCountRow") as Control).visible)


func test_score_attack_targets_use_25000_increments_through_150000() -> void:
	var setup := await _create_setup()
	var game_mode_select = setup.get_node("%GameModeSelect")
	game_mode_select.select_value("score_attack")
	game_mode_select.emit_signal("item_selected", game_mode_select.selected)
	var target_score_select = setup.get_node("%TargetScoreSelect")
	var values := []
	for index in range(target_score_select.item_count):
		values.append(target_score_select.get_item_metadata(index))

	assert_eq(values, [25000, 50000, 75000, 100000, 125000, 150000, "custom"])


func test_score_attack_exposes_target_score_and_emits_mode_config() -> void:
	var setup := await _create_setup()
	var game_mode_select = setup.get_node("%GameModeSelect")
	var target_score_select = setup.get_node("%TargetScoreSelect")
	var lives_select = setup.get_node("%LivesSelect")
	game_mode_select.select_value("score_attack")
	game_mode_select.emit_signal("item_selected", game_mode_select.selected)
	target_score_select.select_value(50000)
	target_score_select.emit_signal("item_selected", target_score_select.selected)
	lives_select.select_value(5)
	watch_signals(setup)

	(setup.get_node("%CreateButton") as BaseButton).emit_signal("pressed")

	assert_true((setup.get_node("%TargetScoreRow") as Control).visible)
	assert_false((setup.get_node("%CustomTargetScoreRow") as Control).visible)
	assert_signal_emitted_with_parameters(setup, "create_requested", [{
		"preset_id": "score_attack",
		"starting_lives": 5,
		"infinite_lives": false,
		"target_score": 50000,
		"target_kills": 0,
		"team_structure": "ffa",
		"team_assignment_mode": "",
		"team_count": 0,
		"max_players": 8,
	}])


func test_custom_score_target_accepts_commas() -> void:
	var setup := await _create_setup()
	var game_mode_select = setup.get_node("%GameModeSelect")
	var target_score_select = setup.get_node("%TargetScoreSelect")
	var custom_target_input := setup.get_node("%CustomTargetScoreInput") as LineEdit
	game_mode_select.select_value("score_attack")
	game_mode_select.emit_signal("item_selected", game_mode_select.selected)
	target_score_select.select_value("custom")
	target_score_select.emit_signal("item_selected", target_score_select.selected)
	custom_target_input.text = "137,500"
	custom_target_input.emit_signal("text_changed", custom_target_input.text)

	assert_true((setup.get_node("%CustomTargetScoreRow") as Control).visible)
	assert_false((setup.get_node("%CreateButton") as BaseButton).disabled)
	assert_eq(setup.current_config().target_score, 137500)


func test_custom_score_target_blocks_create_until_positive_integer_entered() -> void:
	var setup := await _create_setup()
	var game_mode_select = setup.get_node("%GameModeSelect")
	var target_score_select = setup.get_node("%TargetScoreSelect")
	var custom_target_input := setup.get_node("%CustomTargetScoreInput") as LineEdit
	var create_button := setup.get_node("%CreateButton") as BaseButton
	game_mode_select.select_value("score_attack")
	game_mode_select.emit_signal("item_selected", game_mode_select.selected)
	target_score_select.select_value("custom")
	target_score_select.emit_signal("item_selected", target_score_select.selected)
	watch_signals(setup)

	assert_true(create_button.disabled)
	create_button.emit_signal("pressed")
	assert_signal_not_emitted(setup, "create_requested")

	custom_target_input.text = "25001"
	custom_target_input.emit_signal("text_changed", custom_target_input.text)
	assert_false(create_button.disabled)


func test_deathmatch_forces_ffa_infinite_respawns_and_kill_target() -> void:
	var setup := await _create_setup()
	var game_mode_select = setup.get_node("%GameModeSelect")
	var target_select = setup.get_node("%TargetScoreSelect")
	game_mode_select.select_value("deathmatch")
	game_mode_select.emit_signal("item_selected", game_mode_select.selected)

	var target_values := []
	for index in range(target_select.item_count):
		target_values.append(target_select.get_item_metadata(index))
	target_select.select_value(25)

	assert_eq(target_values, [5, 10, 15, 25, 50, "custom"])
	assert_eq((setup.get_node("Content/Fields/TargetScoreRow/Row/Label") as Label).text, "KILL TARGET")
	assert_false((setup.get_node("Content/Fields/LivesRow") as Control).visible)
	assert_false((setup.get_node("%TeamStructureRow") as Control).visible)
	assert_eq(setup.current_config(), {
		"preset_id": "deathmatch",
		"starting_lives": 0,
		"infinite_lives": true,
		"target_score": 0,
		"target_kills": 25,
		"team_structure": "ffa",
		"team_assignment_mode": "",
		"team_count": 0,
		"max_players": 8,
	})


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
		"target_score": 25000,
		"target_kills": 0,
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
