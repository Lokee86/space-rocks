extends GutTest

const GameplayMenuFlow := preload("res://scripts/shell/gameplay_menu_flow.gd")
const GameMenuScene := preload("res://scenes/ui/dialogs/game_menu.tscn")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const SpectateMenuState := preload("res://scripts/gameplay/spectate/spectate_menu_state.gd")


class FakeSessionContext extends ClientSessionContext:
	pass

func test_menu_requested_emits_return_to_pregame_with_active_session_mode() -> void:
	var hud := _create_hud_with_menu()
	var session_context := FakeSessionContext.new()
	session_context.active_mode = Constants.SESSION_MODE_MULTIPLAYER
	var flow := GameplayMenuFlow.new()
	var menu := hud.get_node("CenterContainer/GameOverContainer/MarginContainer2/GameMenu") as GameMenu

	add_child_autofree(hud)
	flow.configure(hud, null, null, session_context)
	watch_signals(flow)

	menu.menu_requested.emit()

	assert_signal_emitted_with_parameters(
		flow,
		"return_to_pregame_requested",
		[Constants.SESSION_MODE_MULTIPLAYER]
	)


func test_multiplayer_game_over_shows_spectate_action_then_cycle_prompt() -> void:
	var hud := _create_hud_with_menu()
	var session_context := FakeSessionContext.new()
	session_context.active_mode = Constants.SESSION_MODE_MULTIPLAYER
	var flow := GameplayMenuFlow.new()
	var menu := hud.get_node("CenterContainer/GameOverContainer/MarginContainer2/GameMenu") as GameMenu
	var cycle_view := hud.get_node("CenterContainer/GameOverContainer/MarginContainer2/CycleView") as Control
	var spectate_state := SpectateMenuState.new()

	add_child_autofree(hud)
	flow.configure(hud, null, null, session_context)
	flow.configure_spectate_menu_state(spectate_state)
	spectate_state.apply_gameplay_state({
		"overlay": {"self_id": "player-1"},
		"world": {"ships": {"player-2": {}}},
		"session": {"player_lifecycle": {"player-2": "active"}},
	})

	flow.set_game_over()

	assert_true(menu.visible)
	assert_false(cycle_view.visible)
	assert_eq(menu.primary_action, Constants.GAME_MENU_PRIMARY_ACTION_SPECTATE)
	assert_false(menu.primary_action_button.disabled)

	menu.spectate_requested.emit()

	assert_false(menu.visible)
	assert_true(cycle_view.visible)


func test_open_menu_uses_overlay_game_menu_when_match_over_overlay_enabled() -> void:
	var hud := _create_hud_with_menu()
	var overlay_parent := Control.new()
	var session_context := FakeSessionContext.new()
	session_context.active_mode = Constants.SESSION_MODE_SINGLE_PLAYER
	var flow := GameplayMenuFlow.new()
	var menu := hud.get_node("CenterContainer/GameOverContainer/MarginContainer2/GameMenu") as GameMenu

	add_child_autofree(hud)
	add_child_autofree(overlay_parent)
	flow.configure(hud, null, null, session_context)
	flow.configure_overlay_parent(overlay_parent)
	flow.set_match_over_overlay_enabled(true)

	Input.action_press("OpenMenu")
	assert_true(flow.handle_open_menu_pressed(true))
	Input.action_release("OpenMenu")

	assert_eq(overlay_parent.get_child_count(), 1)
	var overlay_menu := overlay_parent.get_child(0) as GameMenu
	assert_true(overlay_menu.visible)
	assert_false(menu.visible)
	assert_eq(overlay_menu.primary_action, Constants.GAME_MENU_PRIMARY_ACTION_RESUME)
	assert_true(overlay_menu.primary_action_button.disabled)


func test_open_menu_keeps_embedded_hud_menu_behavior_when_overlay_disabled() -> void:
	var hud := _create_hud_with_menu()
	var overlay_parent := Control.new()
	var session_context := FakeSessionContext.new()
	session_context.active_mode = Constants.SESSION_MODE_SINGLE_PLAYER
	var flow := GameplayMenuFlow.new()
	var menu := hud.get_node("CenterContainer/GameOverContainer/MarginContainer2/GameMenu") as GameMenu

	add_child_autofree(hud)
	add_child_autofree(overlay_parent)
	flow.configure(hud, null, null, session_context)
	flow.configure_overlay_parent(overlay_parent)

	Input.action_press("OpenMenu")
	assert_true(flow.handle_open_menu_pressed(true))
	Input.action_release("OpenMenu")

	assert_eq(overlay_parent.get_child_count(), 0)
	assert_true(menu.visible)
	assert_eq(menu.primary_action, Constants.GAME_MENU_PRIMARY_ACTION_RESUME)
	assert_false(menu.primary_action_button.disabled)


func _create_hud_with_menu() -> Control:
	var hud := Control.new()
	var center_container := Control.new()
	center_container.name = "CenterContainer"
	hud.add_child(center_container)

	var game_over_container := Control.new()
	game_over_container.name = "GameOverContainer"
	center_container.add_child(game_over_container)

	var margin_container := Control.new()
	margin_container.name = "MarginContainer"
	game_over_container.add_child(margin_container)

	var margin_container2 := Control.new()
	margin_container2.name = "MarginContainer2"
	game_over_container.add_child(margin_container2)

	var cycle_view := Control.new()
	cycle_view.name = "CycleView"
	margin_container2.add_child(cycle_view)

	var menu := GameMenuScene.instantiate()
	menu.name = "GameMenu"
	margin_container2.add_child(menu)

	return hud
