extends RefCounted

const Constants := preload("res://scripts/constants/constants.gd")

var request_type := Constants.BOOT_REQUEST_NONE
var join_room_code := ""


func request_single_player() -> void:
	request_type = Constants.BOOT_REQUEST_SINGLE_PLAYER
	join_room_code = ""


func request_create_room() -> void:
	request_type = Constants.BOOT_REQUEST_CREATE_ROOM
	join_room_code = ""


func request_join_room(room_code: String) -> void:
	request_type = Constants.BOOT_REQUEST_JOIN_ROOM
	join_room_code = room_code


func has_request() -> bool:
	return request_type != Constants.BOOT_REQUEST_NONE


func consume_request() -> Dictionary:
	var request := {
		"type": request_type,
		"room_code": join_room_code,
	}
	clear()
	return request


func clear() -> void:
	request_type = Constants.BOOT_REQUEST_NONE
	join_room_code = ""
