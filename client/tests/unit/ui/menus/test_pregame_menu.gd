extends GutTest

const PregameMenuScene := preload("res://scenes/ui/pregame_menu.tscn")


func test_show_single_player_mode_sets_single_player_labels() -> void:
	var menu := await _create_menu()

	menu.show_single_player_mode()

	assert_eq((menu.get_node_or_null("%ModeLabel") as Label).text, "SINGLE PLAYER")
	assert_true((menu.get_node_or_null("%EndlessLabel") as Control).visible)
	assert_false((menu.get_node_or_null("%CreateLabel") as Control).visible)
	assert_true((menu.get_node_or_null("%CampaignLabel") as Control).visible)
	assert_false((menu.get_node_or_null("%JoinLabel") as Control).visible)
	assert_true((menu.get_node_or_null("%SelectPilotLabel") as Control).visible)
	assert_false((menu.get_node_or_null("%LogoutLabel") as Control).visible)
	assert_false((menu.get_node_or_null("%EndlessCreateButton") as BaseButton).disabled)
	assert_true((menu.get_node_or_null("%CampaignJoinButton") as BaseButton).disabled)
	assert_false((menu.get_node_or_null("%SelectPilotLogoutButton") as BaseButton).disabled)


func test_show_multiplayer_mode_sets_multiplayer_labels() -> void:
	var menu := await _create_menu()

	menu.show_multiplayer_mode()

	assert_eq((menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")
	assert_false((menu.get_node_or_null("%EndlessLabel") as Control).visible)
	assert_true((menu.get_node_or_null("%CreateLabel") as Control).visible)
	assert_false((menu.get_node_or_null("%CampaignLabel") as Control).visible)
	assert_true((menu.get_node_or_null("%JoinLabel") as Control).visible)
	assert_false((menu.get_node_or_null("%SelectPilotLabel") as Control).visible)
	assert_true((menu.get_node_or_null("%LogoutLabel") as Control).visible)
	assert_false((menu.get_node_or_null("%EndlessCreateButton") as BaseButton).disabled)
	assert_false((menu.get_node_or_null("%CampaignJoinButton") as BaseButton).disabled)
	assert_false((menu.get_node_or_null("%SelectPilotLogoutButton") as BaseButton).disabled)


func test_set_callsign_updates_callsign_label() -> void:
	var menu := await _create_menu()

	menu.set_callsign("Ace")

	assert_eq((menu.get_node_or_null("%CallsignLabel") as Label).text, "CALLSIGN:\nAce")


func test_back_button_emits_back_requested() -> void:
	var menu := await _create_menu()

	watch_signals(menu)

	(menu.get_node_or_null("%BackButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(menu, "back_requested")


func test_single_player_future_buttons_are_disabled() -> void:
	var menu := await _create_menu()

	menu.show_single_player_mode()

	assert_true((menu.get_node_or_null("%CampaignJoinButton") as BaseButton).disabled)
	assert_true((menu.get_node_or_null("%LoadoutButton") as BaseButton).disabled)
	assert_true((menu.get_node_or_null("%ProvisionerButton") as BaseButton).disabled)
	assert_true((menu.get_node_or_null("%BuyOrebitsButton") as BaseButton).disabled)

	var rankings_button := menu.get_node_or_null("%RankingsButton") as BaseButton
	if rankings_button != null:
		assert_true(rankings_button.disabled)


func test_play_endless_button_emits_play_endless_requested_in_single_player_mode() -> void:
	var menu := await _create_menu()

	menu.show_single_player_mode()
	watch_signals(menu)

	(menu.get_node_or_null("%EndlessCreateButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(menu, "play_endless_requested")
	assert_signal_not_emitted(menu, "create_game_requested")


func test_play_endless_button_does_not_emit_in_multiplayer_mode() -> void:
	var menu := await _create_menu()

	menu.show_multiplayer_mode()
	watch_signals(menu)

	(menu.get_node_or_null("%EndlessCreateButton") as BaseButton).emit_signal("pressed")

	assert_signal_not_emitted(menu, "play_endless_requested")


func test_create_button_emits_create_game_requested_in_multiplayer_mode() -> void:
	var menu := await _create_menu()

	menu.show_multiplayer_mode()
	watch_signals(menu)

	(menu.get_node_or_null("%EndlessCreateButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(menu, "create_game_requested")


func test_join_button_emits_join_game_requested_in_multiplayer_mode() -> void:
	var menu := await _create_menu()

	menu.show_multiplayer_mode()
	watch_signals(menu)

	(menu.get_node_or_null("%CampaignJoinButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(menu, "join_game_requested")


func test_logout_button_emits_logout_requested_in_multiplayer_mode() -> void:
	var menu := await _create_menu()

	menu.show_multiplayer_mode()
	watch_signals(menu)

	(menu.get_node_or_null("%SelectPilotLogoutButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(menu, "logout_requested")


func test_join_button_does_not_emit_in_single_player_mode() -> void:
	var menu := await _create_menu()

	menu.show_single_player_mode()
	watch_signals(menu)

	(menu.get_node_or_null("%CampaignJoinButton") as BaseButton).emit_signal("pressed")

	assert_signal_not_emitted(menu, "join_game_requested")


func test_logout_button_does_not_emit_in_single_player_mode() -> void:
	var menu := await _create_menu()

	menu.show_single_player_mode()
	watch_signals(menu)

	(menu.get_node_or_null("%SelectPilotLogoutButton") as BaseButton).emit_signal("pressed")

	assert_signal_not_emitted(menu, "logout_requested")


func _create_menu() -> Control:
	var menu := PregameMenuScene.instantiate()
	add_child_autofree(menu)
	await get_tree().process_frame
	return menu
