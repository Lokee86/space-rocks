extends RefCounted

const MAIN_MENU := "main_menu"
const CONNECTING := "connecting"
const LOBBY := "lobby"
const GAMEPLAY := "gameplay"
const GAME_OVER := "game_over"
const RETURNING_TO_LOBBY := "returning_to_lobby"
const DISCONNECTED := "disconnected"
const ERROR := "error"

var current_state := MAIN_MENU


func set_state(value: String) -> void:
	current_state = value


func is_state(value: String) -> bool:
	return current_state == value


func current() -> String:
	return current_state
