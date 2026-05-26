extends Node2D

const ShellState := preload("res://scripts/shell/shell_state.gd")
const ClientSessionContext := preload("res://scripts/shell/client_session_context.gd")
const ClientConnectionService := preload("res://scripts/networking/client_connection_service.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const LobbyFlow := preload("res://scripts/lobby/lobby_flow.gd")
const LobbyNetworkActions := preload("res://scripts/shell/lobby_network_actions.gd")
const LobbyPacketReader := preload("res://scripts/lobby/lobby_packet_reader.gd")
const LobbyReturnFlow := preload("res://scripts/shell/lobby_return_flow.gd")
const LobbyShellFlow := preload("res://scripts/shell/lobby_shell_flow.gd")
const GameplayShellFlow := preload("res://scripts/shell/gameplay_shell_flow.gd")
const GameplayHudFlow := preload("res://scripts/shell/gameplay_hud_flow.gd")
const MultiplayerDialogStatusPresenter := preload("res://scripts/shell/multiplayer_dialog_status_presenter.gd")
const MultiplayerLobbyPresenter := preload("res://scripts/shell/multiplayer_lobby_presenter.gd")
const RoomSnapshotShellState := preload("res://scripts/shell/room_snapshot_shell_state.gd")
const ShellBootFlow := preload("res://scripts/shell/shell_boot_flow.gd")
const ShellConstants := preload("res://scripts/shell/constants.gd")

@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground
@onready var canvas_layer: CanvasLayer = $CanvasLayer
@onready var main_menu: Control = $CanvasLayer/MainMenu
@onready var hud: Control = $CanvasLayer/HUD
@onready var game_over_sound: AudioStreamPlayer = $CanvasLayer/HUD/CenterContainer/GameOverContainer/GameOverSound
@onready var player: Player = $Player
@onready var bullets: Node2D = $Bullets
@onready var asteroids: Node2D = $Asteroids

var shell_state: ShellState
var session_context: ClientSessionContext
var connection_service: ClientConnectionService
var lobby_flow: LobbyFlow
var lobby_network_actions: LobbyNetworkActions
var lobby_return_flow: LobbyReturnFlow
var multiplayer_dialog_status_presenter: MultiplayerDialogStatusPresenter
var multiplayer_lobby_presenter: MultiplayerLobbyPresenter
var shell_boot_flow: ShellBootFlow
var lobby_shell_flow: LobbyShellFlow
var gameplay_shell_flow: GameplayShellFlow
var gameplay_hud_flow: GameplayHudFlow


func _ready() -> void:
	shell_state = ShellState.new()
	session_context = ClientSessionContext.new()
	connection_service = ClientConnectionService.new()
	lobby_flow = LobbyFlow.new()
	lobby_network_actions = LobbyNetworkActions.new(connection_service, Callable(self, "_log_v2_status"))
	multiplayer_dialog_status_presenter = MultiplayerDialogStatusPresenter.new()
	multiplayer_lobby_presenter = MultiplayerLobbyPresenter.new()
	lobby_return_flow = LobbyReturnFlow.new(
		lobby_flow,
		multiplayer_lobby_presenter,
		main_menu,
		Callable(self, "_on_lobby_returned_to_main_menu")
	)
	shell_boot_flow = ShellBootFlow.new(
		connection_service,
		ShellConstants.MULTIPLAYER_WS_URL,
		Callable(self, "_log_v2_status")
	)
	lobby_shell_flow = LobbyShellFlow.new(
		lobby_flow,
		session_context,
		lobby_network_actions,
		lobby_return_flow,
		multiplayer_lobby_presenter,
		main_menu,
		canvas_layer,
		Callable(self, "_log_v2_status")
	)
	gameplay_hud_flow = GameplayHudFlow.new()
	gameplay_hud_flow.configure(hud)
	gameplay_shell_flow = GameplayShellFlow.new()
	gameplay_shell_flow.configure(
		connection_service,
		self,
		player,
		bullets,
		asteroids,
		gameplay_hud_flow,
		game_over_sound
	)
	gameplay_shell_flow.gameplay_started.connect(Callable(self, "_on_gameplay_started"))
	add_child(connection_service)
	_connect_connection_service()
	_connect_main_menu()
	print("V2 game shell booted: %s" % shell_state.current())


func _process(delta: float) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.process(delta)


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
	var connect_result := shell_boot_flow.connect_to_game_server(reason)
	if connect_result == ShellBootFlow.CONNECT_RESULT_STARTED_CONNECTING:
		shell_state.set_state(ShellState.CONNECTING)


func _on_connection_connected() -> void:
	_log_v2_status("V2 connection connected")
	shell_boot_flow.send_pending_boot_request()


func _on_connection_closed() -> void:
	_log_v2_status("V2 connection closed")


func _on_packet_parse_failed(text: String) -> void:
	_log_v2_status("V2 packet parse failed: %s" % text)


func _on_room_snapshot_received(_packet: Dictionary) -> void:
	var room_state := LobbyPacketReader.room_state(_packet)
	shell_state.set_state(RoomSnapshotShellState.from_room_state(room_state))
	lobby_shell_flow.apply_room_snapshot(_packet)


func _on_lobby_returned_to_main_menu() -> void:
	if shell_state != null:
		shell_state.set_state(ShellState.MAIN_MENU)
	if gameplay_shell_flow != null:
		gameplay_shell_flow.reset()
	main_menu.show()
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
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_gameplay_state(_packet)


func _on_gameplay_started() -> void:
	main_menu.hide()


func _on_unknown_packet_received(_packet: Dictionary) -> void:
	_log_v2_status("V2 unknown packet received")


func _log_v2_status(message: String) -> void:
	ClientLogger.shell_info(message)
