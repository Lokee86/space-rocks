extends Node
class_name LocalAlphaMatchSmoke

const SessionBootControllerScript := preload("res://scripts/boot/session_boot_controller.gd")
const SessionNetworkControllerScript := preload("res://scripts/session/session_network_controller.gd")
const GameplayDebugFlowScript := preload("res://scripts/devtools/gameplay_debug_flow.gd")
const DevtoolsTargetResolverScript := preload("res://scripts/devtools/devtools_target_resolver.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")

const CONDITION_ATTEMPTS := 150
const RETRY_SECONDS := 0.1

var _latest_room_snapshot: Dictionary = {}
var _room_error: Dictionary = {}
var _tooling_results: Array[Dictionary] = []
var _tooling_error: Dictionary = {}
var _realtime_ready := false
var _connection_closed := false


func run(profile_id: String, score: int) -> String:
	_reset_state()
	var boot_controller = SessionBootControllerScript.new()
	boot_controller.configure(Callable())
	add_child(boot_controller)
	await get_tree().process_frame

	var connection_service = boot_controller.get_connection_service()
	_connect_signals(connection_service)
	var session_network_controller = SessionNetworkControllerScript.new()
	session_network_controller.configure(
		connection_service,
		boot_controller.get_shell_boot_flow(),
		{}
	)
	session_network_controller.connect_connection_signals()

	boot_controller.request_single_player(profile_id)
	if !await _wait_until(func() -> bool: return _room_state() == Constants.ROOM_STATE_IN_GAME):
		_close(connection_service, boot_controller)
		return _failure("single-player room did not enter in_game")

	var player_id := str(_latest_room_snapshot.get(Packets.FIELD_LOCAL_PLAYER_ID, ""))
	var match_id := str(_latest_room_snapshot.get(Packets.FIELD_CURRENT_MATCH_ID, ""))
	if player_id.is_empty() || match_id.is_empty():
		_close(connection_service, boot_controller)
		return "in-game room snapshot did not identify the player and match"
	if !await _wait_until(func() -> bool: return _realtime_ready):
		_close(connection_service, boot_controller)
		return _failure("realtime tooling lane did not become ready")

	var debug_flow = GameplayDebugFlowScript.new()
	debug_flow.configure(connection_service)
	var target_scope := DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER
	if !await _run_tooling_command(
		func() -> void: debug_flow.set_score(target_scope, player_id, score),
		Packets.TYPE_DEBUG_SET_SCORE
	):
		_close(connection_service, boot_controller)
		return _failure("could not set deterministic smoke score")
	if !await _run_tooling_command(
		func() -> void: debug_flow.set_lives(target_scope, player_id, 0),
		Packets.TYPE_DEBUG_SET_LIVES
	):
		_close(connection_service, boot_controller)
		return _failure("could not exhaust smoke player lives")
	if !await _run_tooling_command(
		func() -> void: debug_flow.kill_player(target_scope, player_id),
		Packets.TYPE_DEBUG_KILL_PLAYER
	):
		_close(connection_service, boot_controller)
		return _failure("could not defeat smoke player")

	if !await _wait_until(func() -> bool: return _room_state() == Constants.ROOM_STATE_GAME_OVER):
		_close(connection_service, boot_controller)
		return _failure("single-player match did not reach game_over")
	var result_error := _validate_result(match_id, player_id, score)
	_close(connection_service, boot_controller)
	await get_tree().create_timer(0.2).timeout
	return result_error


func _validate_result(match_id: String, player_id: String, score: int) -> String:
	var result = _latest_room_snapshot.get(Packets.FIELD_MATCH_RESULT, {})
	if typeof(result) != TYPE_DICTIONARY:
		return "game-over snapshot did not contain a match result"
	if str(result.get(Packets.FIELD_MATCH_ID, "")) != match_id:
		return "game-over result did not match the active match"
	var result_players = result.get(Packets.FIELD_PLAYERS, [])
	if typeof(result_players) != TYPE_ARRAY || result_players.size() != 1:
		return "game-over result did not contain one single-player result"
	var player_result: Dictionary = result_players[0]
	if str(player_result.get("game_player_id", "")) != player_id \
			|| int(player_result.get(Packets.FIELD_SCORE, -1)) != score:
		return "game-over result did not preserve the deterministic player score"
	return ""


func _run_tooling_command(send_command: Callable, expected_command_type: String) -> bool:
	var result_count := _tooling_results.size()
	send_command.call()
	if !await _wait_until(
		func() -> bool:
			return _tooling_results.size() > result_count || !_tooling_error.is_empty()
	):
		return false
	if !_tooling_error.is_empty():
		return false
	var result: Dictionary = _tooling_results.back()
	return str(result.get(Packets.FIELD_COMMAND_TYPE, "")) == expected_command_type \
			&& bool(result.get(Packets.FIELD_APPLIED, false))


func _wait_until(condition: Callable) -> bool:
	for _attempt in CONDITION_ATTEMPTS:
		if bool(condition.call()):
			return true
		if _connection_closed || !_room_error.is_empty():
			return false
		await get_tree().create_timer(RETRY_SECONDS).timeout
	return false


func _connect_signals(connection_service) -> void:
	connection_service.room_snapshot_received.connect(_on_room_snapshot_received)
	connection_service.room_error_received.connect(_on_room_error_received)
	connection_service.realtime_transport_ready.connect(_on_realtime_transport_ready)
	connection_service.tooling_command_result_received.connect(_on_tooling_command_result_received)
	connection_service.tooling_error_received.connect(_on_tooling_error_received)
	connection_service.closed.connect(_on_connection_closed)


func _reset_state() -> void:
	_latest_room_snapshot = {}
	_room_error = {}
	_tooling_results.clear()
	_tooling_error = {}
	_realtime_ready = false
	_connection_closed = false


func _room_state() -> String:
	return str(_latest_room_snapshot.get(Packets.FIELD_ROOM_STATE, ""))


func _failure(message: String) -> String:
	if !_room_error.is_empty():
		return "%s: %s" % [message, str(_room_error)]
	if !_tooling_error.is_empty():
		return "%s: %s" % [message, str(_tooling_error)]
	if _connection_closed:
		return "%s: connection closed" % message
	return message


func _close(connection_service, boot_controller) -> void:
	if connection_service != null:
		connection_service.begin_graceful_close()
	if boot_controller != null:
		boot_controller.queue_free()


func _on_room_snapshot_received(packet: Dictionary) -> void:
	_latest_room_snapshot = packet.duplicate(true)


func _on_room_error_received(packet: Dictionary) -> void:
	_room_error = packet.duplicate(true)


func _on_realtime_transport_ready() -> void:
	_realtime_ready = true


func _on_tooling_command_result_received(packet: Dictionary) -> void:
	_tooling_results.append(packet.duplicate(true))


func _on_tooling_error_received(packet: Dictionary) -> void:
	_tooling_error = packet.duplicate(true)


func _on_connection_closed() -> void:
	_connection_closed = true
