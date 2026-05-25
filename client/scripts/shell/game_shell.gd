extends Node2D

const ShellState := preload("res://scripts/shell/shell_state.gd")
const ClientConnectionService := preload("res://scripts/networking/client_connection_service.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const LobbyFlow := preload("res://scripts/lobby/lobby_flow.gd")
const MultiplayerLobbyScene := preload("res://scenes/ui/dialogs/multiplayer_lobby.tscn")

const MULTIPLAYER_WS_URL := "ws://localhost:8080/ws"
const BOOT_REQUEST_NONE := "none"
const BOOT_REQUEST_SINGLE_PLAYER := "single_player"
const BOOT_REQUEST_CREATE_ROOM := "create_room"
const BOOT_REQUEST_JOIN_ROOM := "join_room"

@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground
@onready var canvas_layer: CanvasLayer = $CanvasLayer
@onready var main_menu: Control = $CanvasLayer/MainMenu

var shell_state: ShellState
var connection_service: ClientConnectionService
var lobby_flow: LobbyFlow
var multiplayer_lobby: Control
var pending_boot_request := BOOT_REQUEST_NONE
var pending_join_room_code := ""


func _ready() -> void:
	shell_state = ShellState.new()
	connection_service = ClientConnectionService.new()
	lobby_flow = LobbyFlow.new()
	add_child(connection_service)
	_connect_connection_service()
	_connect_main_menu()
	print("V2 game shell booted: %s" % shell_state.current())


func _connect_connection_service() -> void:
	_connect_connection_signal("connected", Callable(self, "_on_connection_connected"))
	_connect_connection_signal("closed", Callable(self, "_on_connection_closed"))
	_connect_connection_signal("packet_parse_failed", Callable(self, "_on_packet_parse_failed"))
	_connect_connection_signal("room_snapshot_received", Callable(self, "_on_room_snapshot_received"))
	_connect_connection_signal("room_state_changed", Callable(self, "_on_room_state_changed"))
	_connect_connection_signal("room_error_received", Callable(self, "_on_room_error_received"))
	_connect_connection_signal("gameplay_state_received", Callable(self, "_on_gameplay_state_received"))
	_connect_connection_signal("unknown_packet_received", Callable(self, "_on_unknown_packet_received"))


func _connect_connection_signal(signal_name: StringName, handler: Callable) -> void:
	if connection_service.has_signal(signal_name):
		connection_service.connect(signal_name, handler)


func _connect_main_menu() -> void:
	if main_menu.has_signal("single_player_pressed"):
		var single_player_callable := Callable(self, "_on_single_player_pressed")
		main_menu.connect("single_player_pressed", single_player_callable)

	if main_menu.has_signal("multiplayer_create_requested"):
		var create_callable := Callable(self, "_on_multiplayer_create_requested")
		main_menu.connect("multiplayer_create_requested", create_callable)

	if main_menu.has_signal("multiplayer_join_requested"):
		var join_callable := Callable(self, "_on_multiplayer_join_requested")
		main_menu.connect("multiplayer_join_requested", join_callable)


func _on_single_player_pressed() -> void:
	pending_boot_request = BOOT_REQUEST_SINGLE_PLAYER
	pending_join_room_code = ""
	_connect_to_game_server("single player")


func _on_multiplayer_create_requested() -> void:
	pending_boot_request = BOOT_REQUEST_CREATE_ROOM
	pending_join_room_code = ""
	_connect_to_game_server("multiplayer create")


func _on_multiplayer_join_requested(room_code: String) -> void:
	var stripped_room_code := room_code.strip_edges()
	if stripped_room_code.is_empty():
		_log_v2_status("V2 multiplayer join rejected: empty room code")
		return
	pending_boot_request = BOOT_REQUEST_JOIN_ROOM
	pending_join_room_code = stripped_room_code
	_connect_to_game_server("multiplayer join: %s" % stripped_room_code)


func _connect_to_game_server(reason: String) -> void:
	if connection_service.is_server_connected():
		_log_v2_status("V2 already connected for %s" % reason)
		_send_pending_boot_request()
		return
	shell_state.set_state(ShellState.CONNECTING)
	var result := connection_service.connect_to_server(MULTIPLAYER_WS_URL)
	_log_v2_status("V2 connecting to server for %s: %s" % [reason, error_string(result)])


func _send_pending_boot_request() -> void:
	if pending_boot_request == BOOT_REQUEST_NONE:
		return

	if pending_boot_request == BOOT_REQUEST_SINGLE_PLAYER:
		connection_service.send_start_single_player_request()
		_log_v2_status("V2 sent single player request")
	elif pending_boot_request == BOOT_REQUEST_CREATE_ROOM:
		connection_service.send_create_room_request()
		_log_v2_status("V2 sent create room request")
	elif pending_boot_request == BOOT_REQUEST_JOIN_ROOM:
		connection_service.send_join_room_request(pending_join_room_code)
		_log_v2_status("V2 sent join room request: %s" % pending_join_room_code)

	pending_boot_request = BOOT_REQUEST_NONE
	pending_join_room_code = ""


func _on_connection_connected() -> void:
	_log_v2_status("V2 connection connected")
	_send_pending_boot_request()


func _on_connection_closed() -> void:
	_log_v2_status("V2 connection closed")


func _on_packet_parse_failed(text: String) -> void:
	_log_v2_status("V2 packet parse failed: %s" % text)


func _on_room_snapshot_received(_packet: Dictionary) -> void:
	shell_state.set_state(ShellState.LOBBY)
	var summary := lobby_flow.apply_room_snapshot(_packet)
	_log_v2_status("V2 lobby updated: %s" % summary)
	_show_multiplayer_lobby()


func _show_multiplayer_lobby() -> void:
	if main_menu != null:
		main_menu.hide()
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		multiplayer_lobby = MultiplayerLobbyScene.instantiate()
		canvas_layer.add_child(multiplayer_lobby)
		_connect_multiplayer_lobby_signals()

	var state = lobby_flow.current_state()
	multiplayer_lobby.apply_lobby_state(
		state.room_code,
		state.room_state,
		state.local_member_id,
		state.max_players,
		state.members
	)
	if multiplayer_lobby.has_method("set_start_enabled"):
		multiplayer_lobby.set_start_enabled(state.can_start_game())
	multiplayer_lobby.show()


func _connect_multiplayer_lobby_signals() -> void:
	_connect_lobby_signal("ready_requested", Callable(self, "_on_lobby_ready_requested"))
	_connect_lobby_signal("start_game_requested", Callable(self, "_on_lobby_start_game_requested"))
	_connect_lobby_signal("leave_requested", Callable(self, "_on_lobby_leave_requested"))


func _connect_lobby_signal(signal_name: StringName, handler: Callable) -> void:
	if multiplayer_lobby.has_signal(signal_name) && !multiplayer_lobby.is_connected(signal_name, handler):
		multiplayer_lobby.connect(signal_name, handler)


func _on_lobby_ready_requested(ready: bool) -> void:
	connection_service.send_set_ready_request(ready)
	_log_v2_status("V2 lobby ready requested: %s" % ready)


func _on_lobby_start_game_requested() -> void:
	connection_service.send_start_game_request()
	_log_v2_status("V2 lobby start game requested")


func _on_lobby_leave_requested() -> void:
	connection_service.send_leave_room_request()
	_log_v2_status("V2 lobby leave requested")
	_return_to_main_menu_from_lobby()


func _return_to_main_menu_from_lobby() -> void:
	if lobby_flow != null:
		lobby_flow.clear()
	if multiplayer_lobby != null && is_instance_valid(multiplayer_lobby):
		multiplayer_lobby.queue_free()
		multiplayer_lobby = null
	if main_menu != null:
		main_menu.show()
	if shell_state != null:
		shell_state.set_state(ShellState.MAIN_MENU)
	pending_boot_request = BOOT_REQUEST_NONE
	pending_join_room_code = ""


func _on_room_state_changed(_packet: Dictionary) -> void:
	_log_v2_status("V2 room state changed")


func _on_room_error_received(packet: Dictionary) -> void:
	var error_code := str(packet.get("error_code", ""))
	var message := str(packet.get("message", ""))
	_log_v2_status("V2 room error received: code=%s message=%s" % [error_code, message])


func _on_gameplay_state_received(_packet: Dictionary) -> void:
	_log_v2_status("V2 gameplay state received")


func _on_unknown_packet_received(_packet: Dictionary) -> void:
	_log_v2_status("V2 unknown packet received")


func _log_v2_status(message: String) -> void:
	print(message)
