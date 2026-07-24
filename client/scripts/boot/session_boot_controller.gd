extends Node


const ShellBootFlow := preload("res://scripts/boot/shell_boot_flow.gd")
const SessionNetworkTargetScript := preload("res://scripts/boot/session_network_target.gd")

const Constants := preload("res://scripts/generated/constants/constants.gd")

var connection_service
var shell_boot_flow
var session_context
var websocket_url := Constants.MULTIPLAYER_WS_URL
var websocket_url_override := ""
var logger: Callable


func configure(logger_callable: Callable) -> void:
	logger = logger_callable


func _ready() -> void:
	session_context = ClientSessionContext.new()
	connection_service = ClientConnectionService.new()
	add_child(connection_service)
	shell_boot_flow = ShellBootFlow.new(connection_service, websocket_url, logger)


func set_websocket_url_override(url: String) -> void:
	websocket_url_override = url.strip_edges()


func request_single_player(local_profile_id := "") -> void:
	session_context.request_single_player()
	shell_boot_flow.request_single_player(local_profile_id)
	shell_boot_flow.set_websocket_url(_websocket_url_for_mode(Constants.SESSION_MODE_SINGLE_PLAYER))
	shell_boot_flow.connect_to_game_server("single player")


func request_loadout_options(local_profile_id: String, play_mode: String, mode_id: String) -> void:
	shell_boot_flow.request_loadout_options(local_profile_id, play_mode, mode_id)
	shell_boot_flow.set_websocket_url(_websocket_url_for_mode(play_mode))
	if connection_service.is_server_connected():
		if play_mode == Constants.SESSION_MODE_SINGLE_PLAYER or connection_service.is_websocket_auth_authenticated():
			shell_boot_flow.send_pending_loadout_request()
		return
	shell_boot_flow.connect_to_game_server("%s loadout" % play_mode)


func request_create_room(config: Dictionary = {}) -> void:
	session_context.request_multiplayer()
	shell_boot_flow.request_create_room(config)
	shell_boot_flow.set_websocket_url(_websocket_url_for_mode(Constants.SESSION_MODE_MULTIPLAYER))
	shell_boot_flow.connect_to_game_server("multiplayer create")


func request_join_room(room_code: String) -> void:
	session_context.request_multiplayer()
	shell_boot_flow.request_join_room(room_code)
	shell_boot_flow.set_websocket_url(_websocket_url_for_mode(Constants.SESSION_MODE_MULTIPLAYER))
	shell_boot_flow.connect_to_game_server("multiplayer join: %s" % room_code)


func _websocket_url_for_mode(mode: String) -> String:
	if !websocket_url_override.is_empty():
		return websocket_url_override
	return SessionNetworkTargetScript.websocket_url_for_mode(mode)


func get_connection_service():
	return connection_service


func get_shell_boot_flow():
	return shell_boot_flow


func get_session_context():
	return session_context

