extends RefCounted

const BOOT_REQUEST_NONE := "none"
const BOOT_REQUEST_SINGLE_PLAYER := "single_player"
const BOOT_REQUEST_CREATE_ROOM := "create_room"
const BOOT_REQUEST_JOIN_ROOM := "join_room"

var request_type := BOOT_REQUEST_NONE
var join_room_code := ""


func request_single_player() -> void:
	request_type = BOOT_REQUEST_SINGLE_PLAYER
	join_room_code = ""


func request_create_room() -> void:
	request_type = BOOT_REQUEST_CREATE_ROOM
	join_room_code = ""


func request_join_room(room_code: String) -> void:
	request_type = BOOT_REQUEST_JOIN_ROOM
	join_room_code = room_code


func has_request() -> bool:
	return request_type != BOOT_REQUEST_NONE


func consume_request() -> Dictionary:
	var request := {
		"type": request_type,
		"room_code": join_room_code,
	}
	clear()
	return request


func clear() -> void:
	request_type = BOOT_REQUEST_NONE
	join_room_code = ""
