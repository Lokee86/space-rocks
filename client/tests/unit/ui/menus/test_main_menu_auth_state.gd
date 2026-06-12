extends GutTest

const MainMenuScene := preload("res://scenes/ui/main_menu.tscn")


func test_show_signed_out_updates_auth_display() -> void:
	var menu := await _create_menu()

	menu.show_signed_out()

	assert_eq(menu.login_status_label.text, "Not Signed In")
	assert_false(menu.logout_button.visible)
	assert_false(menu.sign_in_label.visible)
	assert_true(menu.multiplayer_label.visible)


func test_show_signed_in_updates_auth_display() -> void:
	var menu := await _create_menu()

	menu.show_signed_in("Ada Lovelace")

	assert_eq(menu.login_status_label.text, "Ada Lovelace")
	assert_true(menu.logout_button.visible)
	assert_false(menu.sign_in_label.visible)
	assert_true(menu.multiplayer_label.visible)


func test_single_player_button_emits_single_player_requested() -> void:
	var menu := await _create_menu()

	watch_signals(menu)

	(menu.get_node("%SinglePlayerButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(menu, "single_player_requested")


func test_multiplayer_button_emits_multiplayer_requested_when_signed_out() -> void:
	var menu := await _create_menu()

	watch_signals(menu)

	(menu.get_node("%MultiplayerButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(menu, "multiplayer_requested")
	assert_null(menu.get("multiplayer_dialog"))


func test_multiplayer_button_emits_multiplayer_requested_when_signed_in() -> void:
	var menu := await _create_menu()

	watch_signals(menu)
	menu.show_signed_in("Ada Lovelace")
	(menu.get_node("%MultiplayerButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(menu, "multiplayer_requested")
	assert_null(menu.get("multiplayer_dialog"))


func test_logout_button_emits_logout_requested() -> void:
	var menu := await _create_menu()

	watch_signals(menu)

	(menu.get_node("%LogoutButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(menu, "logout_requested")


func _create_menu() -> Control:
	var menu := MainMenuScene.instantiate()
	add_child_autofree(menu)
	await get_tree().process_frame
	return menu
