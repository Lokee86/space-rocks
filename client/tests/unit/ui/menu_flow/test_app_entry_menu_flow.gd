extends GutTest

const GameScene := preload("res://scenes/game.tscn")


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


func _force_signed_out(game) -> void:
	game.auth_session_controller.get_session().clear()
	game.auth_session_controller.auth_state_changed.emit()
	await get_tree().process_frame


func _force_signed_in(game) -> void:
	game.auth_session_controller.get_session().set_signed_in("test-token", {"id": 7, "display_name": "Ada"})
	game.auth_session_controller.auth_state_changed.emit()
	await get_tree().process_frame
