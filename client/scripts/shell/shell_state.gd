extends RefCounted

const Constants := preload("res://scripts/generated/constants/constants.gd")

var current_state := Constants.SHELL_STATE_MAIN_MENU


func set_state(value: String) -> void:
	current_state = value


func is_state(value: String) -> bool:
	return current_state == value


func current() -> String:
	return current_state

