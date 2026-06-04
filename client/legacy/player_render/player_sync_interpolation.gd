extends RefCounted
class_name PlayerSyncInterpolation

const VisualSyncPositions = preload("res://legacy/player_render/visual_sync_positions.gd")


func interpolate_player_nodes(
	weight: float,
	current_self_id: String,
	player_lifecycle: PlayerSyncLifecycle,
	player_targets: PlayerSyncTargets,
	_view_target_player_id: String,
	_local_player: Player
) -> void:
	for player_id in player_lifecycle.get_player_ids():
		if !player_targets.has_target_player_position(player_id):
			continue

		var player_node = player_lifecycle.get_player_node(player_id)
		player_node.position = player_node.position.lerp(
			player_targets.get_target_player_position(player_id),
			weight
		)
		player_node.rotation = lerp_angle(
			player_node.rotation,
			player_targets.get_target_player_rotation(player_id),
			weight
		)
		if player_id == current_self_id:
			player_targets.erase_remote_player_visual_position(player_id)
		else:
			player_targets.set_remote_player_visual_position(player_id, player_node.position)

func correct_remote_visual_copy_mismatch(
	player_id: String,
	player_node: Node2D,
	visual_position: Vector2,
	player_lifecycle: PlayerSyncLifecycle,
	player_targets: PlayerSyncTargets
) -> void:
	# Remote targets are local-relative, but rendered remotes can briefly stay in
	# an old visual copy; snap cache/render state before interpolation crosses it.
	if !player_lifecycle.is_initialized(player_id):
		return
	if !VisualSyncPositions.is_world_copy_mismatch(player_node.position, visual_position):
		return

	player_node.position = visual_position
	player_targets.set_target_player_state(
		player_id,
		visual_position,
		player_targets.get_target_player_rotation(player_id)
	)
	player_targets.set_remote_player_visual_position(player_id, visual_position)
