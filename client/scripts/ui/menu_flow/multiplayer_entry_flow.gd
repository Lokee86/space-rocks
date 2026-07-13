class_name MultiplayerEntryFlow
extends RefCounted

const MenuRoute := preload("res://scripts/ui/menu_flow/menu_route.gd")

var menu_flow_controller: MenuFlowController
var auth_session_controller: AuthSessionController


func configure(menu_flow_controller_ref: MenuFlowController = null, auth_session_controller_ref: AuthSessionController = null) -> void:
	menu_flow_controller = menu_flow_controller_ref
	auth_session_controller = auth_session_controller_ref


func request_multiplayer() -> void:
	if _is_signed_in():
		menu_flow_controller.show_multiplayer_pregame()
	else:
		menu_flow_controller.show_sign_in_screen()


func handle_auth_state_changed() -> void:
	if menu_flow_controller == null:
		return

	if menu_flow_controller.get_current_route() != MenuRoute.SIGN_IN_SCREEN:
		return

	if _is_signed_in():
		menu_flow_controller.show_multiplayer_pregame()


func _is_signed_in() -> bool:
	if auth_session_controller == null:
		return false

	var session: AuthSession = auth_session_controller.get_session()
	if session == null:
		return false

	return session.is_signed_in()
