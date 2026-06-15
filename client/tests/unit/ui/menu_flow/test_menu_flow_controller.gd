extends GutTest

const MenuFlowController := preload("res://scripts/ui/menu_flow/menu_flow_controller.gd")
const MenuRoute := preload("res://scripts/ui/menu_flow/menu_route.gd")
const AuthSession := preload("res://scripts/auth/auth_session.gd")
const ProfileStatsProvider := preload("res://scripts/profile/profile_stats_provider.gd")
const PlayerDataProfileApiClient := preload("res://scripts/profile/player_data_profile_api_client.gd")
const ApiRequestResult := preload("res://scripts/api/api_request_result.gd")


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


class JoinProbe:
	extends RefCounted

	var calls := 0
	var last_room_code := ""

	func mark_join(room_code: String) -> void:
		calls += 1
		last_room_code = room_code


class FakeAuthSessionController:
	extends RefCounted

	var session := AuthSession.new()

	func get_session():
		return session


class FakeProfileStatsProvider:
	extends RefCounted

	var load_calls := 0
	var last_context := {}
	var profile := {
		"callsign": "Guest",
		"activity_status": "OFFLINE",
		"identity_kind": "guest",
		"stats": {
			"total_score": 123,
			"high_score": 99,
			"ship_deaths": 7,
			"games_played": 8,
			"wins": 3,
		},
	}

	func load_profile(context: Dictionary):
		load_calls += 1
		last_context = context.duplicate(true)
		return profile.duplicate(true)


class FakePlayerDataProfileApiClient:
	extends RefCounted

	var call_count := 0
	var last_play_mode := ""
	var last_identity_kind := ""
	var last_local_profile_id := ""
	var last_token := ""
	var result: ApiRequestResult = ApiRequestResult.success(200, {
		"profile": {
			"callsign": "Guest",
			"activity_status": "OFFLINE",
			"identity_kind": "guest",
			"stats": {
				"total_score": 50,
				"high_score": 50,
				"ship_deaths": 2,
				"games_played": 1,
				"wins": 1,
			},
		}
	})

	func load_profile(play_mode: String, identity_kind: String, local_profile_id := "", token := ""):
		call_count += 1
		last_play_mode = play_mode
		last_identity_kind = identity_kind
		last_local_profile_id = local_profile_id
		last_token = token
		return result


class FakePregameMenuFlow:
	extends RefCounted

	func configure(_pregame_menu, _return_to_main_menu, _start_single_player, _create_room, _show_join_dialog, _logout, _clear_for_room_transition, _profile_context_provider, _profile_flow, _transmission_flow) -> void:
		pass

	func show_single_player() -> void:
		pass


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
	assert_true((pregame_menu.get_node_or_null("%CallsignLabel") as Label).text.contains("Guest"))


func test_show_multiplayer_pregame_routes_and_instantiates_menu() -> void:
	var controller := await _create_controller(_signed_in_auth_session_controller("Ada"))

	controller.show_multiplayer_pregame()
	await get_tree().process_frame

	var pregame_menu := controller.get_pregame_menu()
	assert_eq(controller.get_current_route(), MenuRoute.PREGAME_MENU)
	assert_false(controller.main_menu.visible)
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")
	assert_true((pregame_menu.get_node_or_null("%CallsignLabel") as Label).text.contains("Ada"))


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


func test_profile_button_mounts_profile_readout_in_single_player_pregame() -> void:
	var fake_profile_stats_provider := FakeProfileStatsProvider.new()
	var controller := await _create_controller(null, fake_profile_stats_provider)

	controller.show_single_player_pregame()
	await get_tree().process_frame

	var pregame_menu := controller.get_pregame_menu()
	(pregame_menu.get_node_or_null("%ProfileButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var screen_display := pregame_menu.find_child("ScreenDisplay", true, false) as Control
	assert_not_null(screen_display)
	assert_eq(controller.get_current_route(), MenuRoute.PREGAME_MENU)
	assert_eq(screen_display.get_child_count(), 1)
	assert_eq(screen_display.get_child(0).name, "ProfileReadout")
	assert_eq(fake_profile_stats_provider.load_calls, 1)
	assert_eq(fake_profile_stats_provider.last_context.get("identity_kind", ""), "guest")


func test_profile_button_mounts_profile_readout_with_injected_stats_provider() -> void:
	var fake_profile_stats_provider := FakeProfileStatsProvider.new()
	fake_profile_stats_provider.profile = {
		"callsign": "Ada",
		"activity_status": "ACTIVE",
		"identity_kind": "authenticated_account",
		"stats": {
			"total_score": 123,
			"high_score": 99,
			"ship_deaths": 7,
			"games_played": 8,
			"wins": 3,
		},
	}
	var controller := await _create_controller(_signed_in_auth_session_controller("Ada"), fake_profile_stats_provider)

	controller.show_multiplayer_pregame()
	await get_tree().process_frame

	var pregame_menu := controller.get_pregame_menu()
	(pregame_menu.get_node_or_null("%ProfileButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var screen_display := pregame_menu.find_child("ScreenDisplay", true, false) as Control
	var profile_readout := screen_display.get_child(0) as Control

	assert_not_null(profile_readout)
	assert_eq(profile_readout.name, "ProfileReadout")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/CallsignLabel") as Label).text, "CALLSIGN: Ada")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/ActivityLabel") as Label).text, "STATUS: ACTIVE")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/TotalScoreContainer/VBoxContainer/TotalScoreValueLabel") as Label).text, "123")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/HighScoreContainer/VBoxContainer/HighScoreValueLabel") as Label).text, "99")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/MissionsContainer/VBoxContainer/MissionsValueLabel") as Label).text, "8")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/WinsContainer/VBoxContainer/WinsValueLabel") as Label).text, "3")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/ShipLossesContainer/VBoxContainer/ShipLossesValueLabel") as Label).text, "7")
	assert_eq(fake_profile_stats_provider.load_calls, 1)
	assert_eq(fake_profile_stats_provider.last_context.get("identity_kind", ""), "authenticated_account")


func test_profile_button_uses_profile_api_for_guest_stats() -> void:
	var profile_stats_provider := ProfileStatsProvider.new()
	var api_client := FakePlayerDataProfileApiClient.new()
	profile_stats_provider.configure(null, api_client)

	var controller := await _create_controller(null, profile_stats_provider)
	controller.show_single_player_pregame()
	await get_tree().process_frame

	var pregame_menu := controller.get_pregame_menu()
	(pregame_menu.get_node_or_null("%ProfileButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	var screen_display := pregame_menu.find_child("ScreenDisplay", true, false) as Control
	var profile_readout := screen_display.get_child(0) as Control

	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/TotalScoreContainer/VBoxContainer/TotalScoreValueLabel") as Label).text, "50")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/HighScoreContainer/VBoxContainer/HighScoreValueLabel") as Label).text, "50")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/ShipLossesContainer/VBoxContainer/ShipLossesValueLabel") as Label).text, "2")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/MissionsContainer/VBoxContainer/MissionsValueLabel") as Label).text, "1")
	assert_eq((profile_readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/WinsContainer/VBoxContainer/WinsValueLabel") as Label).text, "1")
	assert_eq(api_client.call_count, 1)
	assert_eq(api_client.last_play_mode, "single_player")
	assert_eq(api_client.last_identity_kind, "guest")
	assert_eq(api_client.last_local_profile_id, "")
	assert_eq(api_client.last_token, "")


func test_back_clears_profile_readout_before_returning_main_menu() -> void:
	var fake_profile_stats_provider := FakeProfileStatsProvider.new()
	var controller := await _create_controller(null, fake_profile_stats_provider)

	controller.show_single_player_pregame()
	await get_tree().process_frame

	var pregame_menu := controller.get_pregame_menu()
	var screen_display := pregame_menu.find_child("ScreenDisplay", true, false) as Control

	(pregame_menu.get_node_or_null("%ProfileButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame
	await get_tree().process_frame

	assert_not_null(screen_display)
	assert_eq(screen_display.get_child_count(), 1)
	assert_eq(screen_display.get_child(0).name, "ProfileReadout")

	(pregame_menu.get_node_or_null("%BackButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame

	assert_eq(controller.get_current_route(), MenuRoute.PREGAME_MENU)
	assert_eq(screen_display.get_child_count(), 0)
	assert_eq(fake_profile_stats_provider.load_calls, 1)

	(pregame_menu.get_node_or_null("%BackButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame

	assert_eq(controller.get_current_route(), MenuRoute.MAIN_MENU)


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


func test_clear_for_gameplay_keeps_selected_local_profile_for_replay() -> void:
	var controller := await _create_controller()
	controller.profile_context_provider.select_local_profile("local-profile-replay", "ACE")
	controller.pregame_menu = Control.new()
	add_child_autofree(controller.pregame_menu)
	controller.pregame_menu_flow = FakePregameMenuFlow.new()

	controller.show_single_player_pregame()
	await get_tree().process_frame

	controller.clear_for_gameplay()
	await get_tree().process_frame

	var context := controller.get_single_player_context()
	assert_eq(context.get("identity_kind", ""), "local_profile")
	assert_eq(context.get("local_profile_id", ""), "local-profile-replay")
	assert_eq(context.get("callsign", ""), "ACE")


func test_clear_for_gameplay_uses_guest_when_no_local_profile_selected() -> void:
	var controller := await _create_controller()
	controller.pregame_menu = Control.new()
	add_child_autofree(controller.pregame_menu)
	controller.pregame_menu_flow = FakePregameMenuFlow.new()

	controller.show_single_player_pregame()
	await get_tree().process_frame

	controller.clear_for_gameplay()
	await get_tree().process_frame

	var context := controller.get_single_player_context()
	assert_eq(context.get("identity_kind", ""), "guest")
	assert_eq(context.get("callsign", ""), "Guest")


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


func test_show_join_dialog_routes_and_instantiates_dialog() -> void:
	var controller := await _create_controller()

	controller.show_join_dialog()
	await get_tree().process_frame

	assert_eq(controller.get_current_route(), MenuRoute.JOIN_DIALOG)
	assert_false(controller.main_menu.visible)
	assert_not_null(controller.get_join_dialog())


func test_show_join_dialog_clears_existing_sign_in_screen() -> void:
	var controller := await _create_controller()

	controller.show_sign_in_screen()
	await get_tree().process_frame
	var old_sign_in := controller.get_sign_in_screen()

	controller.show_join_dialog()
	await get_tree().process_frame

	assert_false(is_instance_valid(old_sign_in))
	assert_null(controller.get_sign_in_screen())
	assert_not_null(controller.get_join_dialog())
	assert_eq(controller.get_current_route(), MenuRoute.JOIN_DIALOG)


func test_show_multiplayer_pregame_clears_join_dialog() -> void:
	var controller := await _create_controller()

	controller.show_join_dialog()
	await get_tree().process_frame
	var old_join_dialog := controller.get_join_dialog()

	controller.show_multiplayer_pregame()
	await get_tree().process_frame

	assert_false(is_instance_valid(old_join_dialog))
	assert_null(controller.get_join_dialog())
	var pregame_menu := controller.get_pregame_menu()
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")


func test_join_dialog_cancel_returns_to_multiplayer_pregame() -> void:
	var controller := await _create_controller()

	controller.show_multiplayer_pregame()
	await get_tree().process_frame
	controller.show_join_dialog()
	await get_tree().process_frame
	var join_dialog := controller.get_join_dialog()
	(join_dialog.get_node_or_null("%CancelButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame

	assert_eq(controller.get_current_route(), MenuRoute.PREGAME_MENU)
	assert_null(controller.get_join_dialog())
	var pregame_menu := controller.get_pregame_menu()
	assert_not_null(pregame_menu)
	assert_eq((pregame_menu.get_node_or_null("%ModeLabel") as Label).text, "MULTIPLAYER")


func test_clear_for_room_transition_clears_pregame_sign_in_join_and_hides_main() -> void:
	var controller := await _create_controller()
	var pregame_menu := Control.new()
	var sign_in_screen := Control.new()
	var join_dialog := Control.new()

	add_child_autofree(pregame_menu)
	add_child_autofree(sign_in_screen)
	add_child_autofree(join_dialog)

	controller.pregame_menu = pregame_menu
	controller.sign_in_screen = sign_in_screen
	controller.join_dialog = join_dialog
	controller.main_menu.show()
	controller.current_route = MenuRoute.JOIN_DIALOG

	controller.clear_for_room_transition()
	await get_tree().process_frame

	assert_null(controller.get_pregame_menu())
	assert_null(controller.get_sign_in_screen())
	assert_null(controller.get_join_dialog())
	assert_false(controller.main_menu.visible)
	assert_eq(controller.get_current_route(), "")
	assert_false(is_instance_valid(pregame_menu))
	assert_false(is_instance_valid(sign_in_screen))
	assert_false(is_instance_valid(join_dialog))


func test_join_dialog_valid_code_calls_join_callback_and_clears_ui() -> void:
	var controller := await _create_controller()
	var join_probe := JoinProbe.new()

	controller.configure(
		controller.canvas_layer,
		controller.main_menu,
		Callable(),
		Callable(),
		Callable(),
		Callable(join_probe, "mark_join"))
	controller.show_multiplayer_pregame()
	await get_tree().process_frame
	controller.show_join_dialog()
	await get_tree().process_frame
	(controller.get_join_dialog().get_node_or_null("%RoomCodeInput") as LineEdit).text = " ROOM42 "
	(controller.get_join_dialog().get_node_or_null("%JoinButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame

	assert_eq(join_probe.calls, 1)
	assert_eq(join_probe.last_room_code, "ROOM42")
	assert_null(controller.get_join_dialog())
	assert_null(controller.get_pregame_menu())
	assert_false(controller.main_menu.visible)


func test_join_dialog_empty_code_shows_status_and_stays_open() -> void:
	var controller := await _create_controller()

	controller.show_multiplayer_pregame()
	await get_tree().process_frame
	controller.show_join_dialog()
	await get_tree().process_frame
	(controller.get_join_dialog().get_node_or_null("%RoomCodeInput") as LineEdit).text = "   "
	(controller.get_join_dialog().get_node_or_null("%JoinButton") as BaseButton).emit_signal("pressed")
	await get_tree().process_frame

	assert_eq((controller.get_join_dialog().get_node_or_null("%StatusLabel") as Label).text, "Must enter an ID to join.")
	assert_not_null(controller.get_join_dialog())
	assert_eq(controller.get_current_route(), MenuRoute.JOIN_DIALOG)
	assert_not_null(controller.get_pregame_menu())
	assert_false(controller.main_menu.visible)


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


func _create_controller(auth_session_controller = null, profile_stats_provider = null) -> MenuFlowController:
	var canvas_layer := CanvasLayer.new()
	var main_menu := Control.new()
	var controller := MenuFlowController.new()

	add_child_autofree(canvas_layer)
	add_child_autofree(main_menu)

	controller.configure(canvas_layer, main_menu, Callable(), Callable(), Callable(), Callable(), Callable(), auth_session_controller, profile_stats_provider)
	await get_tree().process_frame
	return controller


func _signed_in_auth_session_controller(display_name: String) -> FakeAuthSessionController:
	var auth_session_controller := FakeAuthSessionController.new()
	auth_session_controller.session.set_signed_in("bearer-token", {
		"id": 7,
		"display_name": display_name,
	})
	return auth_session_controller
