extends RefCounted
class_name SpectateTargets

const PlayerLifecycleState = preload("res://scripts/gameplay/session/player_lifecycle_state.gd")


static func select_target(
	local_player_id: String,
	current_target_id: String,
	remote_player_positions: Dictionary,
	player_lifecycle := {}
) -> String:
	var target_ids := _valid_target_ids(local_player_id, remote_player_positions, player_lifecycle)
	if target_ids.is_empty():
		return ""
	if target_ids.has(current_target_id):
		return current_target_id

	return target_ids[0]


static func cycle_target(
	local_player_id: String,
	current_target_id: String,
	remote_player_positions: Dictionary,
	player_lifecycle := {}
) -> String:
	var target_ids := _valid_target_ids(local_player_id, remote_player_positions, player_lifecycle)
	if target_ids.is_empty():
		return ""

	var current_index := target_ids.find(current_target_id)
	if current_index == -1:
		return target_ids[0]

	return target_ids[(current_index + 1) % target_ids.size()]


static func _valid_target_ids(
	local_player_id: String,
	remote_player_positions: Dictionary,
	player_lifecycle: Dictionary
) -> Array[String]:
	var target_ids: Array[String] = []
	for player_id in remote_player_positions.keys():
		var target_id := str(player_id)
		if target_id == "" || target_id == local_player_id:
			continue
		if !PlayerLifecycleState.is_active(player_lifecycle.get(target_id, "")):
			continue
		target_ids.append(target_id)

	target_ids.sort()
	return target_ids
