extends Node

const ClientConnectionService := preload("res://scripts/networking/client_connection_service.gd")
const ShellBootFlow := preload("res://scripts/boot/shell_boot_flow.gd")
const SessionNetworkTarget := preload("res://scripts/boot/session_network_target.gd")
const ClientSessionContext := preload("res://scripts/session/client_session_context.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")

var connection_service
var shell_boot_flow
var session_context
var websocket_url := Constants.MULTIPLAYER_WS_URL
var logger: Callable


func configure(logger_callable: Callable) -> void:
	logger = logger_callable


func _ready() -> void:
	session_context = ClientSessionContext.new()
	connection_service = ClientConnectionService.new()
	add_child(connection_service)
	shell_boot_flow = ShellBootFlow.new(connection_service, websocket_url, logger)


func request_single_player(local_profile_id := "") -> void:
	session_context.request_single_player()
	shell_boot_flow.request_single_player(local_profile_id)
	shell_boot_flow.set_websocket_url(SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_SINGLE_PLAYER))
	shell_boot_flow.connect_to_game_server("single player")


func request_create_room() -> void:
	session_context.request_multiplayer()
	shell_boot_flow.request_create_room()
	shell_boot_flow.set_websocket_url(SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_MULTIPLAYER))
	shell_boot_flow.connect_to_game_server("multiplayer create")


func request_join_room(room_code: String) -> void:
	session_context.request_multiplayer()
	shell_boot_flow.request_join_room(room_code)
	shell_boot_flow.set_websocket_url(SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_MULTIPLAYER))
	shell_boot_flow.connect_to_game_server("multiplayer join: %s" % room_code)


func get_connection_service():
	return connection_service


func get_shell_boot_flow():
	return shell_boot_flow


func get_session_context():
	return session_context

