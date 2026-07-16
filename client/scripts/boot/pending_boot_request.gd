extends RefCounted

const Constants := preload("res://scripts/generated/constants/constants.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")

var request_type := Constants.BOOT_REQUEST_NONE
var join_room_code := ""
var local_profile_id := ""
var _operation_trace_factory: Callable
var trace_id := ""


func _init(operation_trace_factory: Callable = Callable()) -> void:
	_operation_trace_factory = operation_trace_factory


func request_single_player(local_profile_id_value := "") -> void:
	request_type = Constants.BOOT_REQUEST_SINGLE_PLAYER
	join_room_code = ""
	local_profile_id = local_profile_id_value
	trace_id = _new_trace_id("start_single_player")


func request_create_room() -> void:
	request_type = Constants.BOOT_REQUEST_CREATE_ROOM
	join_room_code = ""
	local_profile_id = ""
	trace_id = _new_trace_id("create_room")


func request_join_room(room_code: String) -> void:
	request_type = Constants.BOOT_REQUEST_JOIN_ROOM
	join_room_code = room_code
	local_profile_id = ""
	trace_id = _new_trace_id("join_room")


func has_request() -> bool:
	return request_type != Constants.BOOT_REQUEST_NONE


func current_type() -> String:
	return request_type


func current_trace_id() -> String:
	return trace_id


func is_single_player_request() -> bool:
	return request_type == Constants.BOOT_REQUEST_SINGLE_PLAYER


func is_multiplayer_request() -> bool:
	return request_type == Constants.BOOT_REQUEST_CREATE_ROOM || request_type == Constants.BOOT_REQUEST_JOIN_ROOM


func consume_request() -> Dictionary:
	var request := {
		"type": request_type,
		"room_code": join_room_code,
		"local_profile_id": local_profile_id,
		"trace_id": trace_id,
	}
	clear()
	return request


func clear() -> void:
	request_type = Constants.BOOT_REQUEST_NONE
	join_room_code = ""
	local_profile_id = ""
	trace_id = ""


func _new_trace_id(operation_name: String) -> String:
	return ClientOperationTrace.create(operation_name, _operation_trace_factory).trace_id()