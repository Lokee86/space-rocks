extends RefCounted

const LobbyFlow := preload("res://scripts/lobby/lobby_flow.gd")
const LobbyNetworkActions := preload("res://scripts/lobby/lobby_network_actions.gd")
const LobbyReturnFlow := preload("res://scripts/lobby/lobby_return_flow.gd")
const LobbyShellFlow := preload("res://scripts/lobby/lobby_shell_flow.gd")
const MultiplayerLobbyPresenter := preload("res://scripts/lobby/multiplayer_lobby_presenter.gd")
const MultiplayerDialogStatusPresenter := preload("res://scripts/lobby/multiplayer_dialog_status_presenter.gd")
const Constants := preload("res://scripts/constants/constants.gd")
const Packets := preload("res://scripts/networking/packets/packets.gd")

var main_menu: Control
var canvas_layer: CanvasLayer
var session_context
var connection_service
var shell_boot_flow
var client_config_sender: Callable
var logger: Callable
var latest_room_state := ""

var lobby_flow
var lobby_network_actions
var lobby_return_flow
var lobby_shell_flow
var multiplayer_lobby_presenter
var multiplayer_dialog_status_presenter


func configure(
	main_menu_ref: Control,
	canvas_layer_ref: CanvasLayer,
	session_context_ref,
	connection_service_ref,
	shell_boot_flow_ref,
	logger_callable: Callable
) -> void:
	main_menu = main_menu_ref
	canvas_layer = canvas_layer_ref
	session_context = session_context_ref
	connection_service = connection_service_ref
	shell_boot_flow = shell_boot_flow_ref
	logger = logger_callable

	lobby_flow = LobbyFlow.new()
	lobby_network_actions = LobbyNetworkActions.new(connection_service, logger)
	multiplayer_lobby_presenter = MultiplayerLobbyPresenter.new()
	multiplayer_dialog_status_presenter = MultiplayerDialogStatusPresenter.new()
	lobby_return_flow = LobbyReturnFlow.new(
		lobby_flow,
		multiplayer_lobby_presenter,
		main_menu,
		Callable(self, "_on_lobby_returned_to_main_menu")
	)
	lobby_shell_flow = LobbyShellFlow.new(
		lobby_flow,
		session_context,
		lobby_network_actions,
		lobby_return_flow,
		multiplayer_lobby_presenter,
		main_menu,
		canvas_layer,
		logger
	)


func configure_client_config_sender(sender: Callable) -> void:
	client_config_sender = sender


func handle_room_snapshot(packet: Dictionary) -> void:
	lobby_shell_flow.apply_room_snapshot(packet)
	var state = lobby_flow.current_state()
	latest_room_state = state.room_state
	if state.room_state == Constants.ROOM_STATE_IN_GAME && !client_config_sender.is_null():
		client_config_sender.call()


func handle_room_state_changed(packet: Dictionary) -> void:
	var room_state := str(packet.get(Packets.FIELD_ROOM_STATE, ""))
	if !room_state.is_empty():
		latest_room_state = room_state
	_log("Room state changed: %s" % latest_room_state)


func current_room_state() -> String:
	if !latest_room_state.is_empty():
		return latest_room_state
	if lobby_flow == null:
		return ""
	return lobby_flow.current_state().room_state


func handle_room_error(packet: Dictionary) -> void:
	var error_code := str(packet.get("error_code", ""))
	var message := str(packet.get("message", ""))
	_log("Room error received: code=%s message=%s" % [error_code, message])
	multiplayer_dialog_status_presenter.show_room_error(main_menu, packet)


func _on_lobby_returned_to_main_menu() -> void:
	if shell_boot_flow != null:
		shell_boot_flow.clear()


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)
