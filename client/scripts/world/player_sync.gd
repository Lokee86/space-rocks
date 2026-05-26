extends RefCounted
class_name PlayerSync

const Constants = preload("res://scripts/constants/constants.gd")
const PLAYER_SCENE := preload("res://scenes/player.tscn")
const Packets = preload("res://scripts/networking/packets/packets.gd")
const PlayerSyncState = preload("res://scripts/world/player_sync_state.gd")
const VisualSyncPositions = preload("res://scripts/world/visual_sync_positions.gd")
const REMOTE_PLAYER_HUES := [
	Constants.REMOTE_PLAYER_HUE_ZERO,
	Constants.REMOTE_PLAYER_HUE_ONE,
	Constants.REMOTE_PLAYER_HUE_TWO,
	Constants.REMOTE_PLAYER_HUE_THREE,
	Constants.REMOTE_PLAYER_HUE_FOUR,
	Constants.REMOTE_PLAYER_HUE_FIVE,
	Constants.REMOTE_PLAYER_HUE_SIX,
	Constants.REMOTE_PLAYER_HUE_SEVEN,
]

var owner_node: Node2D
var local_player: Player
var player_nodes := {}
var initialized_players := {}
var target_player_positions := {}
var target_player_rotations := {}
var remote_player_visual_positions := {}
var remote_player_hues := {}


func configure(game_owner: Node2D, player: Player) -> void:
	owner_node = game_owner
	local_player = player
	local_player.z_index = Constants.LOCAL_PLAYER_Z_INDEX


func get_player_node(self_id: String, player_id: String):
	if player_nodes.has(player_id):
		return player_nodes[player_id]

	if player_id == self_id:
		local_player.visible = true
		local_player.z_index = Constants.LOCAL_PLAYER_Z_INDEX
		local_player.set_player_hue(Constants.LOCAL_PLAYER_DEFAULT_HUE)
		player_nodes[player_id] = local_player
		return local_player

	var remote_player = PLAYER_SCENE.instantiate()
	remote_player.z_index = Constants.REMOTE_PLAYER_Z_INDEX
	apply_remote_player_hue(player_id, remote_player)
	owner_node.add_child(remote_player)
	player_nodes[player_id] = remote_player

	return remote_player


func apply(
	self_id: String,
	server_players: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2
) -> void:
	for player_id in server_players.keys():
		var state: Dictionary = server_players[player_id]
		var player_node = get_player_node(self_id, player_id)
		var server_position := Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])
		var visual_position := server_position
		var server_rotation: float = state[Packets.FIELD_ROTATION]
		var is_paused := PlayerSyncState.is_paused(state)

		if player_id == self_id:
			visual_position = local_visual_position
		else:
			visual_position = VisualSyncPositions.relative_to_local_visual(
				local_visual_position,
				local_server_position,
				server_position
			)
			correct_remote_visual_copy_mismatch(player_id, player_node, visual_position)
			apply_remote_player_hue(player_id, player_node)

		target_player_positions[player_id] = visual_position
		target_player_rotations[player_id] = server_rotation
		if player_id != self_id:
			player_node.visible = !is_paused

		if !initialized_players.has(player_id):
			initialized_players[player_id] = true
			player_node.position = visual_position
			player_node.rotation = server_rotation


func correct_remote_visual_copy_mismatch(
	player_id: String,
	player_node: Node2D,
	visual_position: Vector2
) -> void:
	# Remote targets are local-relative, but rendered remotes can briefly stay in
	# an old visual copy; snap cache/render state before interpolation crosses it.
	if !initialized_players.has(player_id):
		return
	if !VisualSyncPositions.is_world_copy_mismatch(player_node.position, visual_position):
		return

	player_node.position = visual_position
	target_player_positions[player_id] = visual_position
	remote_player_visual_positions[player_id] = visual_position


func apply_remote_player_hue(player_id: String, remote_player: Player) -> void:
	var remote_hue := remote_hue_for_player(player_id)
	remote_player_hues[player_id] = remote_hue
	remote_player.set_player_hue(remote_hue)


func remove_missing(server_players: Dictionary, self_id: String) -> void:
	for player_id in player_nodes.keys():
		if server_players.has(player_id):
			continue

		remove_player_node(self_id, player_id)


func remove_player_node(self_id: String, player_id: String) -> void:
	if !player_nodes.has(player_id):
		if player_id == self_id:
			local_player.visible = false
		return

	var player_node = player_nodes[player_id]
	if player_node == local_player:
		local_player.visible = false
	else:
		player_node.queue_free()

	player_nodes.erase(player_id)
	initialized_players.erase(player_id)
	target_player_positions.erase(player_id)
	target_player_rotations.erase(player_id)
	remote_player_visual_positions.erase(player_id)
	remote_player_hues.erase(player_id)


func interpolate(weight: float, current_self_id: String) -> void:
	for player_id in player_nodes.keys():
		if !target_player_positions.has(player_id):
			continue

		var player_node = player_nodes[player_id]
		player_node.position = player_node.position.lerp(
			target_player_positions[player_id],
			weight
		)
		player_node.rotation = lerp_angle(
			player_node.rotation,
			target_player_rotations[player_id],
			weight
		)
		if player_id == current_self_id:
			remote_player_visual_positions.erase(player_id)
		else:
			remote_player_visual_positions[player_id] = player_node.position


func get_remote_player_hues(current_self_id: String) -> Dictionary:
	var hues := remote_player_hues.duplicate()
	hues.erase(current_self_id)
	return hues


func get_remote_player_visual_positions(current_self_id: String) -> Dictionary:
	var positions := remote_player_visual_positions.duplicate()
	positions.erase(current_self_id)
	return positions


func remote_hue_for_player(player_id: String) -> float:
	var start_index := player_id_hash(player_id) % REMOTE_PLAYER_HUES.size()
	for offset in range(REMOTE_PLAYER_HUES.size()):
		var hue: float = REMOTE_PLAYER_HUES[(start_index + offset) % REMOTE_PLAYER_HUES.size()]
		if !hues_similar(hue, Constants.LOCAL_PLAYER_DEFAULT_HUE):
			return hue

	return Constants.REMOTE_PLAYER_FALLBACK_HUE


func hues_similar(
	a: float,
	b: float,
	tolerance := Constants.REMOTE_PLAYER_HUE_SIMILARITY_TOLERANCE
) -> bool:
	var distance: float = abs(fposmod(a, 1.0) - fposmod(b, 1.0))
	return min(distance, 1.0 - distance) < tolerance


func player_id_hash(player_id: String) -> int:
	var hash_value: int = 2166136261
	for index in range(player_id.length()):
		hash_value = int((hash_value ^ player_id.unicode_at(index)) * 16777619) & 0x7fffffff

	return hash_value
