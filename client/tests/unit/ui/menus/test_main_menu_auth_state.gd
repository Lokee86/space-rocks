extends GutTest

const MainMenuScene := preload("res://scenes/ui/main_menu.tscn")


func test_show_signed_out_updates_auth_display() -> void:
	var menu := await _create_menu()

	menu.show_signed_out()

	assert_eq(menu.login_status_label.text, "Not Signed In")
	assert_false(menu.logout_button.visible)
	assert_true(menu.sign_in_label.visible)
	assert_false(menu.multiplayer_label.visible)


func test_show_signed_in_updates_auth_display() -> void:
	var menu := await _create_menu()

	menu.show_signed_in("Ada Lovelace")

	assert_eq(menu.login_status_label.text, "Ada Lovelace")
	assert_true(menu.logout_button.visible)
	assert_false(menu.sign_in_label.visible)
	assert_true(menu.multiplayer_label.visible)


func test_multiplayer_button_emits_sign_in_requested_when_signed_out() -> void:
	var menu := await _create_menu()
	watch_signals(menu)

	menu.multiplayer_button.emit_signal("pressed")

	assert_signal_emitted(menu, "sign_in_requested")
	assert_eq(menu.login_status_label.text, "Not Signed In")
	assert_null(menu.multiplayer_dialog)


func test_multiplayer_button_opens_multiplayer_dialog_when_signed_in() -> void:
	var menu := await _create_menu()
	var sign_in_requested := false

	menu.sign_in_requested.connect(func() -> void:
		sign_in_requested = true
	)

	menu.show_signed_in("Ada Lovelace")
	menu.multiplayer_button.emit_signal("pressed")

	assert_false(sign_in_requested)
	assert_not_null(menu.multiplayer_dialog)
	assert_true(is_instance_valid(menu.multiplayer_dialog))


func _create_menu() -> Control:
	var menu := MainMenuScene.instantiate()
	add_child_autofree(menu)
	await get_tree().process_frame
	return menu
