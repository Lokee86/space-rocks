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


func test_multiplayer_button_routes_to_multiplayer_pregame() -> void:
	var game := await _create_game()
	var main_menu := game.get_node("%MainMenu") as Control
	var multiplayer_button := main_menu.get_node("%MultiplayerButton") as BaseButton

	multiplayer_button.emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	assert_false(main_menu.visible)
	var pregame_menu := _find_pregame_menu(game.get_node("CanvasLayer") as CanvasLayer)
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
