extends RefCounted

signal boot_request_sent

const PendingBootRequest := preload("res://scripts/shell/pending_boot_request.gd")
const Constants := preload("res://scripts/constants/constants.gd")

var connection_service
var pending_boot_request: PendingBootRequest
var websocket_url := ""
var logger: Callable


func _init(
	connection_service_ref,
	websocket_url_value: String,
	logger_callable: Callable
) -> void:
	connection_service = connection_service_ref
	websocket_url = websocket_url_value
	logger = logger_callable
	pending_boot_request = PendingBootRequest.new()


func request_single_player() -> void:
	pending_boot_request.request_single_player()


func request_create_room() -> void:
	pending_boot_request.request_create_room()


func request_join_room(room_code: String) -> void:
	pending_boot_request.request_join_room(room_code)


func connect_to_game_server(reason: String) -> String:
	if connection_service.is_server_connected():
		_log("V2 already connected for %s" % reason)
		send_pending_boot_request()
		return Constants.CONNECT_RESULT_ALREADY_CONNECTED

	var result = connection_service.connect_to_server(websocket_url)
	_log("V2 connecting to server for %s: %s" % [reason, error_string(result)])
	if result == OK:
		return Constants.CONNECT_RESULT_STARTED_CONNECTING
	return Constants.CONNECT_RESULT_FAILED


func send_pending_boot_request() -> void:
	if !pending_boot_request.has_request():
		return

	var request := pending_boot_request.consume_request()
	var request_type := str(request.get("type", Constants.BOOT_REQUEST_NONE))
	var room_code := str(request.get("room_code", ""))
	var request_sent := true

	if request_type == Constants.BOOT_REQUEST_SINGLE_PLAYER:
		connection_service.send_start_single_player_request()
		_log("V2 sent single player request")
	elif request_type == Constants.BOOT_REQUEST_CREATE_ROOM:
		connection_service.send_create_room_request()
		_log("V2 sent create room request")
	elif request_type == Constants.BOOT_REQUEST_JOIN_ROOM:
		connection_service.send_join_room_request(room_code)
		_log("V2 sent join room request: %s" % room_code)
	else:
		request_sent = false

	if request_sent:
		boot_request_sent.emit()


func clear() -> void:
	pending_boot_request.clear()


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)
