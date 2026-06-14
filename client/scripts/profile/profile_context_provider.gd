extends RefCounted
class_name ProfileContextProvider

const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")

var auth_session_controller
var selected_single_player_identity_kind := "guest"
var selected_local_profile_id := ""
var selected_local_display_name := "Guest"


func configure(auth_session_controller_ref) -> void:
	auth_session_controller = auth_session_controller_ref


func context_for_mode(mode: String) -> Dictionary:
	if mode == PregameMenuMode.MULTIPLAYER:
		return _multiplayer_context()
	return _single_player_context()


func select_guest_profile() -> void:
	selected_single_player_identity_kind = "guest"
	selected_local_profile_id = ""
	selected_local_display_name = "Guest"


func select_local_profile(local_profile_id: String, display_name: String) -> void:
	if local_profile_id.is_empty():
		return

	selected_single_player_identity_kind = "local_profile"
	selected_local_profile_id = local_profile_id
	selected_local_display_name = display_name
	if selected_local_display_name.is_empty():
		selected_local_display_name = "Pilot"


func _guest_context() -> Dictionary:
	return {
		"play_mode": PregameMenuMode.SINGLE_PLAYER,
		"identity_kind": "guest",
		"callsign": "Guest",
		"activity_status": "OFFLINE",
	}


func _single_player_context() -> Dictionary:
	if selected_single_player_identity_kind == "local_profile":
		return {
			"play_mode": PregameMenuMode.SINGLE_PLAYER,
			"identity_kind": "local_profile",
			"local_profile_id": selected_local_profile_id,
			"callsign": selected_local_display_name,
			"activity_status": "LOCAL",
		}
	return _guest_context()


func _multiplayer_context() -> Dictionary:
	var callsign := "Guest"
	var session = _session()
	if session == null or !session.is_signed_in():
		return _guest_context()

	callsign = str(session.display_name)
	if callsign.is_empty():
		callsign = "Pilot"

	return {
		"play_mode": PregameMenuMode.MULTIPLAYER,
		"identity_kind": "authenticated_account",
		"callsign": callsign,
		"activity_status": "ACTIVE",
	}


func _session():
	if auth_session_controller == null or !auth_session_controller.has_method("get_session"):
		return null
	return auth_session_controller.get_session()
