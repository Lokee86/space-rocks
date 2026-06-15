extends GutTest

const ProfileContextProvider := preload("res://scripts/profile/profile_context_provider.gd")
const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")


class FakeSession:
	extends RefCounted

	var signed_in := false
	var display_name := ""

	func is_signed_in() -> bool:
		return signed_in


class FakeAuthSessionController:
	extends RefCounted

	var session

	func get_session():
		return session


func test_single_player_context_returns_guest_offline_guest() -> void:
	var provider := ProfileContextProvider.new()

	var context := provider.context_for_mode(PregameMenuMode.SINGLE_PLAYER)

	assert_eq(context.get("play_mode"), PregameMenuMode.SINGLE_PLAYER)
	assert_eq(context.get("callsign"), "Guest")
	assert_eq(context.get("activity_status"), "ACTIVE")
	assert_eq(context.get("identity_kind"), "guest")


func test_single_player_selected_local_profile_returns_active_selected_display_name() -> void:
	var provider := ProfileContextProvider.new()
	provider.select_local_profile("local-profile-1", "ACE")

	var context := provider.context_for_mode(PregameMenuMode.SINGLE_PLAYER)

	assert_eq(context.get("play_mode"), PregameMenuMode.SINGLE_PLAYER)
	assert_eq(context.get("callsign"), "ACE")
	assert_eq(context.get("activity_status"), "ACTIVE")
	assert_eq(context.get("identity_kind"), "local_profile")
	assert_eq(context.get("local_profile_id"), "local-profile-1")


func test_multiplayer_signed_in_returns_display_name_active_authenticated_account() -> void:
	var controller := FakeAuthSessionController.new()
	controller.session = _create_session(true, "Ada")
	var provider := ProfileContextProvider.new()
	provider.configure(controller)

	var context := provider.context_for_mode(PregameMenuMode.MULTIPLAYER)

	assert_eq(context.get("play_mode"), PregameMenuMode.MULTIPLAYER)
	assert_eq(context.get("callsign"), "Ada")
	assert_eq(context.get("activity_status"), "ACTIVE")
	assert_eq(context.get("identity_kind"), "authenticated_account")


func test_multiplayer_signed_in_with_empty_display_name_falls_back_to_pilot() -> void:
	var controller := FakeAuthSessionController.new()
	controller.session = _create_session(true, "")
	var provider := ProfileContextProvider.new()
	provider.configure(controller)

	var context := provider.context_for_mode(PregameMenuMode.MULTIPLAYER)

	assert_eq(context.get("play_mode"), PregameMenuMode.MULTIPLAYER)
	assert_eq(context.get("callsign"), "Pilot")
	assert_eq(context.get("activity_status"), "ACTIVE")
	assert_eq(context.get("identity_kind"), "authenticated_account")


func test_multiplayer_signed_out_falls_back_to_guest_offline_guest() -> void:
	var controller := FakeAuthSessionController.new()
	controller.session = _create_session(false, "Ada")
	var provider := ProfileContextProvider.new()
	provider.configure(controller)

	var context := provider.context_for_mode(PregameMenuMode.MULTIPLAYER)

	assert_eq(context.get("play_mode"), PregameMenuMode.SINGLE_PLAYER)
	assert_eq(context.get("callsign"), "Guest")
	assert_eq(context.get("activity_status"), "OFFLINE")
	assert_eq(context.get("identity_kind"), "guest")


func _create_session(is_signed_in_value: bool, display_name_value: String) -> FakeSession:
	var session := FakeSession.new()
	session.signed_in = is_signed_in_value
	session.display_name = display_name_value
	return session
