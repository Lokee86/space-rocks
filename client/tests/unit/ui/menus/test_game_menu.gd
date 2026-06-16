extends GutTest

const GameMenuScene := preload("res://scenes/ui/dialogs/game_menu.tscn")
const Constants := preload("res://scripts/generated/constants/constants.gd")


func test_menu_button_emits_menu_requested() -> void:
	var menu := await _create_menu()

	watch_signals(menu)
	(_menu_button(menu) as BaseButton).emit_signal("pressed")

	assert_signal_emitted(menu, "menu_requested")


func test_multiplayer_game_over_does_not_set_primary_action_to_lobby() -> void:
	var menu := await _create_menu()

	menu.configure_for_state(Constants.SESSION_MODE_MULTIPLAYER, true, Constants.ROOM_STATE_GAME_OVER, false)

	assert_ne(menu.primary_action, Constants.GAME_MENU_PRIMARY_ACTION_LOBBY)
	assert_eq(menu.primary_action, Constants.GAME_MENU_PRIMARY_ACTION_WAITING)
	assert_true(_primary_button(menu).disabled)


func test_multiplayer_game_over_with_spectate_targets_sets_spectate_and_enables_it() -> void:
	var menu := await _create_menu()

	menu.configure_for_state(Constants.SESSION_MODE_MULTIPLAYER, true, Constants.ROOM_STATE_GAME_OVER, true)

	assert_eq(menu.primary_action, Constants.GAME_MENU_PRIMARY_ACTION_SPECTATE)
	assert_false(_primary_button(menu).disabled)


func test_multiplayer_game_over_without_spectate_targets_disables_primary_button() -> void:
	var menu := await _create_menu()

	menu.configure_for_state(Constants.SESSION_MODE_MULTIPLAYER, true, Constants.ROOM_STATE_GAME_OVER, false)

	assert_eq(menu.primary_action, Constants.GAME_MENU_PRIMARY_ACTION_WAITING)
	assert_true(_primary_button(menu).disabled)


func test_multiplayer_not_game_over_uses_resume_and_enables_it() -> void:
	var menu := await _create_menu()

	menu.configure_for_state(Constants.SESSION_MODE_MULTIPLAYER, false, Constants.ROOM_STATE_GAME_OVER, false)

	assert_eq(menu.primary_action, Constants.GAME_MENU_PRIMARY_ACTION_RESUME)
	assert_false(_primary_button(menu).disabled)


func _create_menu() -> Control:
	var menu := GameMenuScene.instantiate()
	add_child_autofree(menu)
	await get_tree().process_frame
	return menu


func _menu_button(menu: Control) -> BaseButton:
	return menu.find_child("MenuButton", true, false) as BaseButton


func _primary_button(menu: Control) -> BaseButton:
	return menu.find_child("PrimaryActionButton", true, false) as BaseButton
