extends GutTest

const GameScene := preload("res://scenes/game.tscn")
const Constants := preload("res://scripts/generated/constants/constants.gd")


func test_single_player_button_routes_to_single_player_pregame() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var single_player_button := main_menu.get_node("%SinglePlayerButton") as BaseButton

	single_player_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	assert_false(main_menu.visible)
	var pregame_menu := _find_pregame_menu(game.get_node("CanvasLayer") as CanvasLayer)
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "SINGLE PLAYER")


func test_single_player_pregame_play_endless_starts_game() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var single_player_button := main_menu.get_node("%SinglePlayerButton") as BaseButton

	single_player_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var canvas_layer := game.get_node("CanvasLayer") as CanvasLayer
	var pregame_menu := _find_pregame_menu(canvas_layer)
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "SINGLE PLAYER")

	(pregame_menu.get_node("%EndlessCreateButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame
	await get_tree().process_frame

	assert_null(_find_pregame_menu(canvas_layer))
	assert_false(main_menu.visible)


func test_multiplayer_button_routes_signed_out_to_login_window() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var multiplayer_button := main_menu.get_node("%MultiplayerButton") as BaseButton

	await _force_signed_out(game)
	multiplayer_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	assert_false(main_menu.visible)
	var canvas_layer := game.get_node("CanvasLayer") as CanvasLayer
	assert_not_null(_find_login_window(canvas_layer))
	assert_null(_find_pregame_menu(canvas_layer))


func test_multiplayer_button_routes_signed_in_to_multiplayer_pregame() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var multiplayer_button := main_menu.get_node("%MultiplayerButton") as BaseButton

	await _force_signed_in(game)
	multiplayer_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var canvas_layer := game.get_node("CanvasLayer") as CanvasLayer
	var pregame_menu := _find_pregame_menu(canvas_layer)
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")
	assert_null(_find_login_window(canvas_layer))


func test_successful_auth_from_login_window_routes_to_multiplayer_pregame() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var multiplayer_button := main_menu.get_node("%MultiplayerButton") as BaseButton

	await _force_signed_out(game)
	multiplayer_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var canvas_layer := game.get_node("CanvasLayer") as CanvasLayer
	assert_not_null(_find_login_window(canvas_layer))

	await _force_signed_in(game)
	await get_tree().process_frame
	await get_tree().process_frame

	var pregame_menu := _find_pregame_menu(canvas_layer)
	assert_null(_find_login_window(canvas_layer))
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")


func test_multiplayer_pregame_create_clears_menu_and_requests_create_room() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var multiplayer_button := main_menu.get_node("%MultiplayerButton") as BaseButton

	await _force_signed_in(game)
	multiplayer_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var canvas_layer := game.get_node("CanvasLayer") as CanvasLayer
	var pregame_menu := _find_pregame_menu(canvas_layer)
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")

	(pregame_menu.get_node_or_null("%EndlessCreateButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	assert_null(_find_pregame_menu(canvas_layer))
	assert_false(main_menu.visible)
	var shell_boot_flow = game.session_boot_controller.get_shell_boot_flow()
	assert_eq(shell_boot_flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)
	assert_true(shell_boot_flow.pending_request_is_multiplayer())
	assert_false(shell_boot_flow.pending_request_is_single_player())


func test_multiplayer_pregame_join_opens_join_dialog() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var multiplayer_button := main_menu.get_node("%MultiplayerButton") as BaseButton

	await _force_signed_in(game)
	multiplayer_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var canvas_layer := game.get_node("CanvasLayer") as CanvasLayer
	var pregame_menu := _find_pregame_menu(canvas_layer)
	assert_not_null(pregame_menu)

	(pregame_menu.get_node_or_null("%CampaignJoinButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	assert_not_null(_find_join_dialog(canvas_layer))
	assert_not_null(_find_pregame_menu(canvas_layer))


func test_join_dialog_valid_code_clears_menu_and_requests_join_room() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var multiplayer_button := main_menu.get_node("%MultiplayerButton") as BaseButton

	await _force_signed_in(game)
	multiplayer_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var canvas_layer := game.get_node("CanvasLayer") as CanvasLayer
	var pregame_menu := _find_pregame_menu(canvas_layer)
	assert_not_null(pregame_menu)

	(pregame_menu.get_node_or_null("%CampaignJoinButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var join_dialog := _find_join_dialog(canvas_layer)
	assert_not_null(join_dialog)
	(join_dialog.get_node_or_null("%RoomCodeInput") as LineEdit).text = " ABCD "
	(join_dialog.get_node_or_null("%JoinButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	assert_null(_find_join_dialog(canvas_layer))
	assert_null(_find_pregame_menu(canvas_layer))
	assert_false(main_menu.visible)
	var shell_boot_flow = game.session_boot_controller.get_shell_boot_flow()
	assert_eq(shell_boot_flow.pending_request_type(), Constants.BOOT_REQUEST_JOIN_ROOM)
	assert_true(shell_boot_flow.pending_request_is_multiplayer())
	assert_false(shell_boot_flow.pending_request_is_single_player())


func test_join_dialog_cancel_returns_to_multiplayer_pregame() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var multiplayer_button := main_menu.get_node("%MultiplayerButton") as BaseButton

	await _force_signed_in(game)
	multiplayer_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var canvas_layer := game.get_node("CanvasLayer") as CanvasLayer
	var pregame_menu := _find_pregame_menu(canvas_layer)
	assert_not_null(pregame_menu)

	(pregame_menu.get_node_or_null("%CampaignJoinButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var join_dialog := _find_join_dialog(canvas_layer)
	assert_not_null(join_dialog)
	(join_dialog.get_node_or_null("%CancelButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	assert_null(_find_join_dialog(canvas_layer))
	var returned_pregame_menu := _find_pregame_menu(canvas_layer)
	assert_not_null(returned_pregame_menu)
	assert_eq((returned_pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")


func test_multiplayer_pregame_logout_returns_to_main_menu_signed_out() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var multiplayer_button := main_menu.get_node("%MultiplayerButton") as BaseButton

	await _force_signed_in(game)
	multiplayer_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var canvas_layer := game.get_node("CanvasLayer") as CanvasLayer
	var pregame_menu := _find_pregame_menu(canvas_layer)
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")

	(pregame_menu.get_node_or_null("%SelectPilotLogoutButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	assert_true(main_menu.visible)
	assert_null(_find_pregame_menu(canvas_layer))
	assert_false(game.auth_session_controller.get_session().is_signed_in())
	assert_false((main_menu.get_node_or_null("%LogoutButton") as BaseButton).visible)
	assert_eq((main_menu.find_child("LoginStatusLabel", true, false) as Label).text, "Not Signed In")


func test_lobby_leave_return_destination_opens_multiplayer_pregame() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var canvas_layer := game.get_node("CanvasLayer") as CanvasLayer

	await _force_signed_in(game)
	assert_true(main_menu.visible)

	game.room_session_controller.lobby_return_flow.return_after_leave()
	await get_tree().process_frame
	await get_tree().process_frame

	assert_false(main_menu.visible)
	var pregame_menu := _find_pregame_menu(canvas_layer)
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")


func _create_game() -> Control:
	var game := GameScene.instantiate()
	add_child_autofree(game)
	await get_tree().process_frame
	await get_tree().process_frame
	return game


func _find_pregame_menu(canvas_layer: CanvasLayer) -> Control:
	if canvas_layer == null:
		return null

	for child in canvas_layer.get_children():
		if child is Control and child.name == "PregameMenu":
			return child

	return null


func _find_login_window(canvas_layer: CanvasLayer) -> Control:
	if canvas_layer == null:
		return null

	for child in canvas_layer.get_children():
		if child is Control and child.name == "LoginWindow":
			return child

	return null


func _find_join_dialog(canvas_layer: CanvasLayer) -> Control:
	if canvas_layer == null:
		return null

	for child in canvas_layer.get_children():
		if child is Control and child.name == "JoinDialog":
			return child

	return null


func _force_signed_out(game) -> void:
	game.auth_session_controller.get_session().clear()
	game.auth_session_controller.auth_state_changed.emit()
	await get_tree().process_frame


func _force_signed_in(game) -> void:
	game.auth_session_controller.get_session().set_signed_in("test-token", {"id": 7, "display_name": "Ada"})
	game.auth_session_controller.auth_state_changed.emit()
	await get_tree().process_frame
