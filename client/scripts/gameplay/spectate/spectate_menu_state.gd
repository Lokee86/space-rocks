extends RefCounted

const ClientLogger := preload("res://scripts/logging/logger.gd")

# Owns spectate availability/menu state. This is the future home for outside
# spectating rules; it does not render UI or send packets.

var self_id := ""
var current_target_id := ""
var player_lifecycle := {}


func apply_gameplay_state(state: Dictionary) -> void:
	self_id = str(state.get("self_id", ""))
	var lifecycle = state.get("player_lifecycle", {})
	if lifecycle is Dictionary:
		player_lifecycle = lifecycle
	else:
		player_lifecycle = {}


func reset() -> void:
	self_id = ""
	current_target_id = ""
	player_lifecycle.clear()


func spectate_target_ids() -> Array:
	var target_ids := []
	for player_id in player_lifecycle.keys():
		if player_id == self_id:
			continue

		var lifecycle_value := str(player_lifecycle[player_id])
		if lifecycle_value == "Dead" || lifecycle_value == "GameOver":
			continue

		target_ids.append(player_id)
	return target_ids


func current_target() -> String:
	var target_ids := spectate_target_ids()
	if current_target_id in target_ids:
		return current_target_id
	if target_ids.is_empty():
		return ""
	return str(target_ids[0])


func begin_spectating() -> String:
	current_target_id = current_target()
	return current_target_id


func cycle_next_target() -> String:
	var target_ids := spectate_target_ids()
	if target_ids.is_empty():
		current_target_id = ""
		return current_target_id

	var current_index := target_ids.find(current_target_id)
	if current_index == -1:
		current_target_id = str(target_ids[0])
		return current_target_id

	current_target_id = str(target_ids[(current_index + 1) % target_ids.size()])
	return current_target_id


func has_spectate_targets() -> bool:
	var result := !self_id.is_empty() && !spectate_target_ids().is_empty()
	_log_has_spectate_targets(result)
	return result


func _log_has_spectate_targets(result: bool) -> void:
	ClientLogger.shell_debug(
		"Spectate menu state trace: self_id=%s lifecycle_keys=%s has_spectate_targets=%s"
		% [self_id, player_lifecycle.keys(), result]
	)

