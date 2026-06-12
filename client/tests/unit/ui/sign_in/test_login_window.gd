extends GutTest

const LoginWindowScene := preload("res://scenes/ui/dialogs/login_window.tscn")


func test_manual_login_controls_are_disabled() -> void:
	var window := await _create_window()

	assert_false((window.get_node("%EmailInput") as LineEdit).editable)
	assert_false((window.get_node("%PasswordInput") as LineEdit).editable)
	assert_true((window.get_node("%SignInButton") as BaseButton).disabled)


func test_google_login_is_disabled() -> void:
	var window := await _create_window()

	assert_true((window.get_node("%GoogleLoginButton") as BaseButton).disabled)


func test_back_button_emits_back_requested() -> void:
	var window := await _create_window()

	watch_signals(window)
	(window.get_node("%BackButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(window, "back_requested")


func test_discord_button_emits_discord_login_requested() -> void:
	var window := await _create_window()

	watch_signals(window)
	(window.get_node("%DiscordLoginButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(window, "discord_login_requested")


func _create_window() -> Control:
	var window := LoginWindowScene.instantiate()
	add_child_autofree(window)
	await get_tree().process_frame
	return window
