extends RefCounted
class_name PlayerSyncTargets

var target_player_positions := {}
var target_player_rotations := {}
var remote_player_visual_positions := {}


func reset() -> void:
	target_player_positions.clear()
	target_player_rotations.clear()
	remote_player_visual_positions.clear()


func set_target_player_state(player_id: String, position: Vector2, rotation: float) -> void:
	target_player_positions[player_id] = position
	target_player_rotations[player_id] = rotation


func has_target_player_position(player_id: String) -> bool:
	return target_player_positions.has(player_id)


func has_target_player_rotation(player_id: String) -> bool:
	return target_player_rotations.has(player_id)


func get_target_player_position(player_id: String) -> Vector2:
	return target_player_positions[player_id]


func get_target_player_rotation(player_id: String) -> float:
	return float(target_player_rotations[player_id])


func erase_player(player_id: String) -> void:
	target_player_positions.erase(player_id)
	target_player_rotations.erase(player_id)
	remote_player_visual_positions.erase(player_id)


func set_remote_player_visual_position(player_id: String, position: Vector2) -> void:
	remote_player_visual_positions[player_id] = position


func erase_remote_player_visual_position(player_id: String) -> void:
	remote_player_visual_positions.erase(player_id)


func get_remote_player_visual_positions_without(current_self_id: String) -> Dictionary:
	var positions := remote_player_visual_positions.duplicate()
	positions.erase(current_self_id)
	return positions


func build_server_hitbox_draw_entries(_current_self_id: String, player_lifecycle: PlayerSyncLifecycle) -> Array:
	var entries: Array = []

	for player_id in target_player_positions.keys():
		if !target_player_rotations.has(player_id):
			continue

		var visual_position: Vector2 = target_player_positions[player_id]
		var rotation: float = float(target_player_rotations[player_id])

		if player_lifecycle.has_player_node(player_id):
			var player_node := player_lifecycle.get_player_node(player_id) as Node2D
			if player_node != null and is_instance_valid(player_node):
				visual_position = player_node.global_position
				rotation = player_node.rotation

		entries.append({
			"kind": "player",
			"id": String(player_id),
			"visual_position": visual_position,
			"rotation": rotation,
			"scale": 1.0,
		})

	return entries
