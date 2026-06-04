extends RefCounted

const Constants := preload("res://scripts/generated/constants/constants.gd")

var requested_mode := Constants.SESSION_MODE_NONE
var active_mode := Constants.SESSION_MODE_NONE


func clear() -> void:
	requested_mode = Constants.SESSION_MODE_NONE
	active_mode = Constants.SESSION_MODE_NONE


func request_single_player() -> void:
	requested_mode = Constants.SESSION_MODE_SINGLE_PLAYER


func request_multiplayer() -> void:
	requested_mode = Constants.SESSION_MODE_MULTIPLAYER


func activate_requested_mode() -> void:
	if requested_mode != Constants.SESSION_MODE_NONE:
		active_mode = requested_mode


func is_single_player() -> bool:
	return active_mode == Constants.SESSION_MODE_SINGLE_PLAYER


func is_multiplayer() -> bool:
	return active_mode == Constants.SESSION_MODE_MULTIPLAYER


func should_show_multiplayer_lobby(room_state: String) -> bool:
	return active_mode == Constants.SESSION_MODE_MULTIPLAYER && room_state == Constants.ROOM_STATE_LOBBY

