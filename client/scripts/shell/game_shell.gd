extends Node2D

const ShellState := preload("res://scripts/shell/shell_state.gd")
const ClientSessionContext := preload("res://scripts/shell/client_session_context.gd")
const ClientConnectionService := preload("res://scripts/networking/client_connection_service.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const LobbyFlow := preload("res://scripts/lobby/lobby_flow.gd")
const LobbyShellFlow := preload("res://scripts/shell/lobby_shell_flow.gd")
const MultiplayerDialogStatusPresenter := preload("res://scripts/shell/multiplayer_dialog_status_presenter.gd")
const MultiplayerLobbyPresenter := preload("res://scripts/shell/multiplayer_lobby_presenter.gd")
const RoomSnapshotShellState := preload("res://scripts/shell/room_snapshot_shell_state.gd")
const ShellBootFlow := preload("res://scripts/shell/shell_boot_flow.gd")

const MULTIPLAYER_WS_URL := "ws://localhost:8080/ws"

@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground
@onready var canvas_layer: CanvasLayer = $CanvasLayer
@onready var main_menu: Control = $CanvasLayer/MainMenu

var shell_state: ShellState
var session_context: ClientSessionContext
var connection_service: ClientConnectionService
var lobby_flow: LobbyFlow
var multiplayer_dialog_status_presenter: MultiplayerDialogStatusPresenter
var multiplayer_lobby_presenter: MultiplayerLobbyPresenter
var shell_boot_flow: ShellBootFlow
var lobby_shell_flow: LobbyShellFlow


func _ready() -> void:
	shell_state = ShellState.new()
	session_context = ClientSessionContext.new()
	connection_service = ClientConnectionService.new()
	lobby_flow = LobbyFlow.new()
	multiplayer_dialog_status_presenter = MultiplayerDialogStatusPresenter.new()
	multiplayer_lobby_presenter = MultiplayerLobbyPresenter.new()
	shell_boot_flow = ShellBootFlow.new(connection_service, MULTIPLAYER_WS_URL, Callable(self, "_log_v2_status"))
	lobby_shell_flow = LobbyShellFlow.new(
		lobby_flow,
		session_context,
		connection_service,
		multiplayer_lobby_presenter,
		main_menu,
		canvas_layer,
		Callable(self, "_log_v2_status"),
		Callable(self, "_on_lobby_returned_to_main_menu")
	)
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
	session_context.request_single_player()
	shell_boot_flow.request_single_player()
	_connect_to_game_server("single player")


func _on_multiplayer_create_requested() -> void:
	session_context.request_multiplayer()
	shell_boot_flow.request_create_room()
	_connect_to_game_server("multiplayer create")


func _on_multiplayer_join_requested(room_code: String) -> void:
	var stripped_room_code := room_code.strip_edges()
	if stripped_room_code.is_empty():
		_log_v2_status("V2 multiplayer join rejected: empty room code")
		multiplayer_dialog_status_presenter.show_status(main_menu, "Must enter an ID to join.")
		return
	session_context.request_multiplayer()
	shell_boot_flow.request_join_room(stripped_room_code)
	_connect_to_game_server("multiplayer join: %s" % stripped_room_code)


func _connect_to_game_server(reason: String) -> void:
	if shell_boot_flow.connect_to_game_server(reason):
		shell_state.set_state(ShellState.CONNECTING)


func _on_connection_connected() -> void:
	_log_v2_status("V2 connection connected")
	shell_boot_flow.send_pending_boot_request()


func _on_connection_closed() -> void:
	_log_v2_status("V2 connection closed")


func _on_packet_parse_failed(text: String) -> void:
	_log_v2_status("V2 packet parse failed: %s" % text)


func _on_room_snapshot_received(_packet: Dictionary) -> void:
	var room_state := str(_packet.get("room_state", ""))
	shell_state.set_state(RoomSnapshotShellState.from_room_state(room_state))
	lobby_shell_flow.apply_room_snapshot(_packet)


func _on_lobby_returned_to_main_menu() -> void:
	if shell_state != null:
		shell_state.set_state(ShellState.MAIN_MENU)
	shell_boot_flow.clear()


func _on_room_state_changed(_packet: Dictionary) -> void:
	_log_v2_status("V2 room state changed")


func _on_room_error_received(packet: Dictionary) -> void:
	var error_code := str(packet.get("error_code", ""))
	var message := str(packet.get("message", ""))
	_log_v2_status("V2 room error received: code=%s message=%s" % [error_code, message])
	multiplayer_dialog_status_presenter.show_room_error(main_menu, packet)


func _on_gameplay_state_received(_packet: Dictionary) -> void:
	_log_v2_status("V2 gameplay state received")


func _on_unknown_packet_received(_packet: Dictionary) -> void:
	_log_v2_status("V2 unknown packet received")


func _log_v2_status(message: String) -> void:
	ClientLogger.shell_info(message)
