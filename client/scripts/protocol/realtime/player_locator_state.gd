extends RefCounted
class_name PlayerLocatorState

var player_locators: Dictionary = {}
var sequence := 0
var server_sent_msec := 0
var received_msec := 0

func replace_player_locators(next_player_locators: Dictionary, next_sequence: int, next_server_sent_msec: int) -> bool:
	if next_sequence <= sequence:
		return false
	player_locators = next_player_locators
	sequence = next_sequence
	server_sent_msec = next_server_sent_msec
	received_msec = Time.get_ticks_msec()
	return true

func reset() -> void:
	player_locators = {}
	sequence = 0
	server_sent_msec = 0
	received_msec = 0
