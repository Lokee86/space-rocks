extends RefCounted
class_name DevtoolsStateContext

var has_gameplay_readiness := false
var player_dev_label_mode := ""
var local_player_id := ""
var game_target_kind := ""
var game_target_id := ""
var game_target_player_id := ""


func set_has_gameplay_readiness(value: bool) -> void:
	has_gameplay_readiness = value


func set_has_received_gameplay_state(value: bool) -> void:
	set_has_gameplay_readiness(value)


func has_gameplay_state() -> bool:
	return has_gameplay_readiness


func set_local_player_id(player_id: String) -> void:
	local_player_id = player_id


func get_local_player_id() -> String:
	return local_player_id


func set_game_target(kind: String, id: String) -> void:
	game_target_kind = kind
	game_target_id = id
	if game_target_kind == "player":
		game_target_player_id = game_target_id
	else:
		game_target_player_id = ""


func get_game_target_kind() -> String:
	return game_target_kind


func get_game_target_id() -> String:
	return game_target_id


func get_game_target_player_id() -> String:
	return game_target_player_id


func set_player_dev_label_mode(mode: String) -> void:
	player_dev_label_mode = mode


func get_player_dev_label_mode() -> String:
	return player_dev_label_mode


func reset_game_target() -> void:
	game_target_kind = ""
	game_target_id = ""
	game_target_player_id = ""
