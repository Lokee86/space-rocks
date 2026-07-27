extends GutTest

const RoomSetupScene := preload("res://scenes/ui/transmission_displays/multiplayer_room_setup_readout.tscn")


func test_defaults_to_ffa_with_eight_players() -> void:
	var setup := await _create_setup()

	assert_eq(setup.current_config(), {
		"team_structure": "ffa",
		"team_assignment_mode": "",
		"team_count": 0,
		"max_players": 8,
	})
	assert_false((setup.get_node("%AssignmentRow") as Control).visible)
	assert_false((setup.get_node("%TeamCountRow") as Control).visible)


func test_auto_balanced_selection_exposes_team_count_and_emits_config() -> void:
	var setup := await _create_setup()
	var mode_select = setup.get_node("%ModeSelect")
	var team_count_select = setup.get_node("%TeamCountSelect")
	mode_select.select_value("auto_balanced")
	mode_select.emit_signal("item_selected", mode_select.selected)
	team_count_select.select_value(4)
	watch_signals(setup)

	(setup.get_node("%CreateButton") as BaseButton).emit_signal("pressed")

	assert_true((setup.get_node("%TeamCountRow") as Control).visible)
	assert_false((setup.get_node("%AssignmentRow") as Control).visible)
	assert_signal_emitted_with_parameters(setup, "create_requested", [{
		"team_structure": "auto_balanced",
		"team_assignment_mode": "",
		"team_count": 4,
		"max_players": 8,
	}])


func test_custom_selection_exposes_assignment_mode() -> void:
	var setup := await _create_setup()
	var mode_select = setup.get_node("%ModeSelect")
	var assignment_select = setup.get_node("%AssignmentSelect")
	mode_select.select_value("custom")
	mode_select.emit_signal("item_selected", mode_select.selected)
	assignment_select.select_value("player_selected")

	assert_true((setup.get_node("%AssignmentRow") as Control).visible)
	assert_false((setup.get_node("%TeamCountRow") as Control).visible)
	assert_eq(setup.current_config().team_assignment_mode, "player_selected")


func _create_setup() -> Control:
	var setup := RoomSetupScene.instantiate()
	add_child_autofree(setup)
	await get_tree().process_frame
	return setup
