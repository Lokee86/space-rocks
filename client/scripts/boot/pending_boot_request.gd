extends RefCounted

const Constants := preload("res://scripts/generated/constants/constants.gd")

var request_type := Constants.BOOT_REQUEST_NONE
var join_room_code := ""
var local_profile_id := ""


func request_single_player(local_profile_id_value := "") -> void:
	request_type = Constants.BOOT_REQUEST_SINGLE_PLAYER
	join_room_code = ""
	local_profile_id = local_profile_id_value


func request_create_room() -> void:
	request_type = Constants.BOOT_REQUEST_CREATE_ROOM
	join_room_code = ""
	local_profile_id = ""


func request_join_room(room_code: String) -> void:
	request_type = Constants.BOOT_REQUEST_JOIN_ROOM
	join_room_code = room_code
	local_profile_id = ""


func has_request() -> bool:
	return request_type != Constants.BOOT_REQUEST_NONE


func current_type() -> String:
	return request_type


func is_single_player_request() -> bool:
	return request_type == Constants.BOOT_REQUEST_SINGLE_PLAYER


func is_multiplayer_request() -> bool:
	return request_type == Constants.BOOT_REQUEST_CREATE_ROOM || request_type == Constants.BOOT_REQUEST_JOIN_ROOM


func consume_request() -> Dictionary:
	var request := {
		"type": request_type,
		"room_code": join_room_code,
		"local_profile_id": local_profile_id,
	}
	clear()
	return request


func clear() -> void:
	request_type = Constants.BOOT_REQUEST_NONE
	join_room_code = ""
	local_profile_id = ""

