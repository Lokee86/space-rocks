extends RefCounted
class_name ProfileContextProvider

const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")

var auth_session_controller


func configure(auth_session_controller_ref) -> void:
	auth_session_controller = auth_session_controller_ref


func context_for_mode(mode: String) -> Dictionary:
	if mode == PregameMenuMode.MULTIPLAYER:
		return _multiplayer_context()
	return _guest_context()


func _guest_context() -> Dictionary:
	return {
		"play_mode": PregameMenuMode.SINGLE_PLAYER,
		"identity_kind": "guest",
		"callsign": "Guest",
		"activity_status": "OFFLINE",
	}


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
