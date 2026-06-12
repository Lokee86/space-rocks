extends GutTest

const MenuFlowController := preload("res://scripts/ui/menu_flow/menu_flow_controller.gd")
const MenuRoute := preload("res://scripts/ui/menu_flow/menu_route.gd")


class StartSinglePlayerProbe:
	extends RefCounted

	var calls := 0

	func mark_called() -> void:
		calls += 1


class Probe:
	extends RefCounted

	var calls := 0

	func mark_called() -> void:
		calls += 1


func test_configure_starts_on_main_menu() -> void:
	var canvas_layer := CanvasLayer.new()
	var main_menu := Control.new()
	var controller := MenuFlowController.new()

	add_child_autofree(canvas_layer)
	add_child_autofree(main_menu)

	controller.configure(canvas_layer, main_menu)

	assert_eq(controller.get_current_route(), MenuRoute.MAIN_MENU)
	assert_true(main_menu.visible)


func test_show_single_player_pregame_routes_and_instantiates_menu() -> void:
	var controller := await _create_controller()

	controller.show_single_player_pregame()
	await get_tree().process_frame

	var pregame_menu := controller.get_pregame_menu()
	assert_eq(controller.get_current_route(), MenuRoute.PREGAME_MENU)
	assert_false(controller.main_menu.visible)
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "SINGLE PLAYER")


func test_show_multiplayer_pregame_routes_and_instantiates_menu() -> void:
	var controller := await _create_controller()

	controller.show_multiplayer_pregame()
	await get_tree().process_frame

	var pregame_menu := controller.get_pregame_menu()
	assert_eq(controller.get_current_route(), MenuRoute.PREGAME_MENU)
	assert_false(controller.main_menu.visible)
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")


func test_pregame_back_returns_to_main_menu() -> void:
	var controller := await _create_controller()

	controller.show_single_player_pregame()
	await get_tree().process_frame
	var old_pregame_menu := controller.get_pregame_menu()
	(old_pregame_menu.get_node_or_null("%BackButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame

	assert_eq(controller.get_current_route(), MenuRoute.MAIN_MENU)
	assert_true(controller.main_menu.visible)
	assert_null(controller.get_pregame_menu())
	assert_false(is_instance_valid(old_pregame_menu))


func test_clear_for_gameplay_removes_pregame_and_keeps_main_menu_hidden() -> void:
	var controller := await _create_controller()

	controller.show_single_player_pregame()
	await get_tree().process_frame
	var old_pregame_menu := controller.get_pregame_menu()

	controller.clear_for_gameplay()
	await get_tree().process_frame

	assert_null(controller.get_pregame_menu())
	assert_false(is_instance_valid(old_pregame_menu))
	assert_false(controller.main_menu.visible)
	assert_eq(controller.get_current_route(), "")


func test_show_sign_in_screen_routes_and_instantiates_login_window() -> void:
	var controller := await _create_controller()

	controller.show_sign_in_screen()
	await get_tree().process_frame

	assert_eq(controller.get_current_route(), MenuRoute.SIGN_IN_SCREEN)
	assert_false(controller.main_menu.visible)
	assert_not_null(controller.get_sign_in_screen())


func test_sign_in_back_returns_to_main_menu() -> void:
	var controller := await _create_controller()

	controller.show_sign_in_screen()
	await get_tree().process_frame
	var sign_in_screen := controller.get_sign_in_screen()
	(sign_in_screen.get_node_or_null("%BackButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame

	assert_eq(controller.get_current_route(), MenuRoute.MAIN_MENU)
	assert_true(controller.main_menu.visible)
	assert_null(controller.get_sign_in_screen())
	assert_false(is_instance_valid(sign_in_screen))


func test_show_sign_in_screen_clears_pregame_menu() -> void:
	var controller := await _create_controller()

	controller.show_single_player_pregame()
	await get_tree().process_frame
	var old_pregame := controller.get_pregame_menu()

	controller.show_sign_in_screen()
	await get_tree().process_frame

	assert_false(is_instance_valid(old_pregame))
	assert_null(controller.get_pregame_menu())
	assert_not_null(controller.get_sign_in_screen())


func test_show_multiplayer_pregame_clears_sign_in_screen() -> void:
	var controller := await _create_controller()

	controller.show_sign_in_screen()
	await get_tree().process_frame
	var old_sign_in := controller.get_sign_in_screen()

	controller.show_multiplayer_pregame()
	await get_tree().process_frame

	assert_false(is_instance_valid(old_sign_in))
	assert_null(controller.get_sign_in_screen())
	var pregame_menu := controller.get_pregame_menu()
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")


func test_sign_in_discord_button_calls_injected_callback() -> void:
	var controller := await _create_controller()
	var probe := Probe.new()

	controller.configure(controller.canvas_layer, controller.main_menu, Callable(), Callable(probe, "mark_called"))
	controller.show_sign_in_screen()
	await get_tree().process_frame

	(controller.get_sign_in_screen().get_node_or_null("%DiscordLoginButton") as BaseButton).emit_signal("pressed")

	assert_eq(probe.calls, 1)


func test_sign_in_discord_button_reaches_auth_callback() -> void:
	var canvas_layer := CanvasLayer.new()
	var main_menu := Control.new()
	var controller := MenuFlowController.new()
	var probe := Probe.new()

	add_child_autofree(canvas_layer)
	add_child_autofree(main_menu)

	controller.configure(canvas_layer, main_menu, Callable(), Callable(probe, "mark_called"))
	controller.show_sign_in_screen()
	await get_tree().process_frame

	(controller.get_sign_in_screen().get_node_or_null("%DiscordLoginButton") as BaseButton).emit_signal("pressed")

	assert_eq(probe.calls, 1)


func test_play_endless_from_single_player_pregame_calls_start_single_player_callback() -> void:
	var controller := await _create_controller()
	var start_probe := StartSinglePlayerProbe.new()

	controller.configure(controller.canvas_layer, controller.main_menu, Callable(start_probe, "mark_called"))
	controller.show_single_player_pregame()
	await get_tree().process_frame

	var pregame_menu := controller.get_pregame_menu()
	(pregame_menu.get_node_or_null("%EndlessCreateButton") as BaseButton).emit_signal("pressed")

	assert_eq(start_probe.calls, 1)


func test_play_endless_from_multiplayer_pregame_does_not_call_start_single_player_callback() -> void:
	var controller := await _create_controller()
	var start_probe := StartSinglePlayerProbe.new()

	controller.configure(controller.canvas_layer, controller.main_menu, Callable(start_probe, "mark_called"))
	controller.show_multiplayer_pregame()
	await get_tree().process_frame

	var pregame_menu := controller.get_pregame_menu()
	(pregame_menu.get_node_or_null("%EndlessCreateButton") as BaseButton).emit_signal("pressed")

	assert_eq(start_probe.calls, 0)


func _create_controller() -> MenuFlowController:
	var canvas_layer := CanvasLayer.new()
	var main_menu := Control.new()
	var controller := MenuFlowController.new()

	add_child_autofree(canvas_layer)
	add_child_autofree(main_menu)

	controller.configure(canvas_layer, main_menu)
	await get_tree().process_frame
	return controller
