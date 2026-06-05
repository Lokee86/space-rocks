extends RefCounted

var player_render_api
var asteroid_sync
var bullet_sync
var pickup_sync
var current_self_id := ""


func configure(player_render_api_ref, asteroid_sync_ref, bullet_sync_ref, pickup_sync_ref) -> void:
	player_render_api = player_render_api_ref
	asteroid_sync = asteroid_sync_ref
	bullet_sync = bullet_sync_ref
	pickup_sync = pickup_sync_ref


func set_current_self_id(self_id: String) -> void:
	current_self_id = self_id


func player_positions() -> Dictionary:
	var positions := {}
	if player_render_api == null:
		return positions

	if current_self_id != "":
		positions[current_self_id] = {
			"visual_position": player_render_api.visual_position(),
			"server_position": player_render_api.server_position()
		}

	var remote_positions: Dictionary = player_render_api.get_remote_player_visual_positions(current_self_id)
	for player_id in remote_positions.keys():
		var visual_position = remote_positions[player_id]
		positions[player_id] = {
			"visual_position": visual_position,
			"server_position": visual_position
		}

	return positions


func asteroid_positions() -> Dictionary:
	if asteroid_sync == null:
		return {}
	return asteroid_sync.asteroid_target_positions()


func bullet_positions() -> Dictionary:
	if bullet_sync == null:
		return {}
	return bullet_sync.bullet_target_positions()


func pickup_positions() -> Dictionary:
	if pickup_sync == null:
		return {}
	return pickup_sync.pickup_target_positions()
