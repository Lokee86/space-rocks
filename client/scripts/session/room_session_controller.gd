extends RefCounted

const LobbyFlow := preload("res://scripts/lobby/lobby_flow.gd")
const LobbyNetworkActions := preload("res://scripts/lobby/lobby_network_actions.gd")
const LobbyReturnFlow := preload("res://scripts/lobby/lobby_return_flow.gd")
const LobbyShellFlow := preload("res://scripts/lobby/lobby_shell_flow.gd")

const MultiplayerDialogStatusPresenter := preload("res://scripts/lobby/multiplayer_dialog_status_presenter.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

var main_menu: Control
var canvas_layer: CanvasLayer
var session_context
var connection_service
var shell_boot_flow
var client_config_sender: Callable
var latest_room_state := ""
var _current_match_id := ""
var latest_match_result := {}

var lobby_flow
var lobby_network_actions
var lobby_return_flow
var lobby_shell_flow
var multiplayer_lobby_presenter
var multiplayer_dialog_status_presenter
var room_transition_completed: Callable
var room_operation_failed: Callable


func configure(
	main_menu_ref: Control,
	canvas_layer_ref: CanvasLayer,
	session_context_ref,
	connection_service_ref,
	shell_boot_flow_ref
) -> void:
	main_menu = main_menu_ref
	canvas_layer = canvas_layer_ref
	session_context = session_context_ref
	connection_service = connection_service_ref
	shell_boot_flow = shell_boot_flow_ref

	lobby_flow = LobbyFlow.new()
	lobby_network_actions = LobbyNetworkActions.new(connection_service, Callable())
	multiplayer_lobby_presenter = MultiplayerLobbyPresenter.new()
	multiplayer_dialog_status_presenter = MultiplayerDialogStatusPresenter.new()
	lobby_return_flow = LobbyReturnFlow.new(
		lobby_flow,
		multiplayer_lobby_presenter,
		main_menu,
		Callable(self, "_on_lobby_left_room")
	)
	lobby_shell_flow = LobbyShellFlow.new(
		lobby_flow,
		session_context,
		lobby_network_actions,
		lobby_return_flow,
		multiplayer_lobby_presenter,
		main_menu,
		canvas_layer,
		Callable()
	)


func configure_client_config_sender(sender: Callable) -> void:
	client_config_sender = sender


func configure_room_transition_completed(callback: Callable) -> void:
	room_transition_completed = callback


func configure_room_operation_failed(callback: Callable) -> void:
	room_operation_failed = callback


func configure_lobby_leave_return_destination(destination: Callable) -> void:
	if lobby_return_flow != null:
		lobby_return_flow.configure_return_destination(destination)


func handle_room_snapshot(packet: Dictionary) -> void:
	var operation := _active_initial_room_operation()
	if !operation.is_empty() and room_transition_completed.is_valid():
		room_transition_completed.call()
	lobby_shell_flow.apply_room_snapshot(packet)
	var state = lobby_flow.current_state()
	latest_room_state = state.room_state
	_current_match_id = str(packet.get(Packets.FIELD_CURRENT_MATCH_ID, ""))
	_cache_match_result_from_snapshot(packet)
	_clear_room_operation_after_snapshot()
	if state.room_state == Constants.ROOM_STATE_IN_GAME && !client_config_sender.is_null():
		client_config_sender.call()


func handle_room_state_changed(packet: Dictionary) -> void:
	var room_state := str(packet.get(Packets.FIELD_ROOM_STATE, ""))
	if !room_state.is_empty():
		latest_room_state = room_state


func current_room_state() -> String:
	if !latest_room_state.is_empty():
		return latest_room_state
	if lobby_flow == null:
		return ""
	return lobby_flow.current_state().room_state


func current_match_id() -> String:
	return _current_match_id


func current_match_result() -> Dictionary:
	if latest_match_result is Dictionary:
		return latest_match_result
	return {}


func _cache_match_result_from_snapshot(packet: Dictionary) -> void:
	var match_result = packet.get(Packets.FIELD_MATCH_RESULT, null)
	if match_result is Dictionary:
		var match_id := str(match_result.get(Packets.FIELD_MATCH_ID, ""))
		if !match_id.is_empty():
			latest_match_result = match_result
			return
	latest_match_result = {}


func current_max_players() -> int:
	if lobby_flow == null:
		return 0
	return int(lobby_flow.current_state().max_players)


func lobby_state_snapshot() -> Dictionary:
	if lobby_flow == null:
		return {}
	var state = lobby_flow.current_state()
	return {
		"room_code": state.room_code,
		"room_state": state.room_state,
		"local_player_id": state.local_player_id,
		"owner_id": state.owner_id,
		"max_players": state.max_players,
		"members": state.members.duplicate(true),
		"all_members_ready": state.all_members_ready(),
		"can_start_game": state.can_start_game(),
	}


func request_ready(ready: bool) -> void:
	if lobby_network_actions != null:
		lobby_network_actions.send_ready_requested(ready)


func request_add_bot() -> void:
	if lobby_network_actions != null:
		lobby_network_actions.send_add_bot_requested()


func request_start_game() -> void:
	if lobby_network_actions != null:
		lobby_network_actions.send_start_game_requested()


func handle_room_error(packet: Dictionary) -> void:
	var error_code := str(packet.get(Packets.FIELD_ERROR_CODE, ""))
	var packet_trace_id := str(packet.get(Packets.FIELD_TRACE_ID, ""))
	var operation := ""
	var active_trace_id := ""
	if connection_service != null && connection_service.has_method("active_room_operation_type"):
		operation = _canonical_room_operation(str(connection_service.active_room_operation_type()))
		if connection_service.has_method("active_room_operation_trace_id"):
			active_trace_id = str(connection_service.active_room_operation_trace_id())
	var trace_id := packet_trace_id if !packet_trace_id.is_empty() else active_trace_id
	var matches_active_operation := packet_trace_id.is_empty() || packet_trace_id == active_trace_id
	if _is_initial_room_operation(operation):
		ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_ROOM_OPERATION_FAILED,
			"",
			{"trace_id": trace_id, "error_code": error_code},
			{"operation": operation}
		)
		if matches_active_operation:
			if connection_service != null and connection_service.has_method("clear_room_operation_context"):
				connection_service.clear_room_operation_context()
			var friendly_message: String = multiplayer_dialog_status_presenter.friendly_room_error_message(
				error_code,
				str(packet.get(Packets.FIELD_MESSAGE, ""))
			)
			if room_operation_failed.is_valid():
				room_operation_failed.call(operation, friendly_message)
			else:
				multiplayer_dialog_status_presenter.show_status(main_menu, friendly_message)
	if error_code == "removed_by_owner" && lobby_return_flow != null:
		lobby_return_flow.return_after_leave()
	elif !_is_initial_room_operation(operation):
		multiplayer_dialog_status_presenter.show_room_error(main_menu, packet)


func _on_lobby_left_room() -> void:
	_current_match_id = ""
	if session_context != null:
		session_context.clear()
	if shell_boot_flow != null:
		shell_boot_flow.clear()


func _clear_room_operation_after_snapshot() -> void:
	if connection_service == null:
		return
	if !connection_service.has_method("active_room_operation_type"):
		return
	var operation := _canonical_room_operation(str(connection_service.active_room_operation_type()))
	if !_is_initial_room_operation(operation):
		return
	if connection_service.has_method("clear_room_operation_context"):
		connection_service.clear_room_operation_context()


func _active_initial_room_operation() -> String:
	if connection_service == null or !connection_service.has_method("active_room_operation_type"):
		return ""
	var operation := _canonical_room_operation(str(connection_service.active_room_operation_type()))
	return operation if _is_initial_room_operation(operation) else ""


func _canonical_room_operation(operation: String) -> String:
	match operation:
		Constants.BOOT_REQUEST_CREATE_ROOM:
			return "create_room"
		Constants.BOOT_REQUEST_JOIN_ROOM:
			return "join_room"
		Constants.BOOT_REQUEST_SINGLE_PLAYER:
			return "start_single_player"
		"start_single_player":
			return "start_single_player"
		_:
			return ""


func _is_initial_room_operation(operation: String) -> bool:
	return operation == "create_room" || operation == "join_room" || operation == "start_single_player"

