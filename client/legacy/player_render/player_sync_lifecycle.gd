extends RefCounted
class_name PlayerSyncLifecycle

const Constants = preload("res://scripts/constants/constants.gd")
const PLAYER_SCENE := preload("res://scenes/player.tscn")

var owner_node: Node2D
var local_player: Player
var player_nodes := {}
var initialized_players := {}
var apply_local_player_hue_callback: Callable
var apply_remote_player_hue_callback: Callable


func configure(
	game_owner: Node2D,
	player: Player,
	local_hue_callback: Callable = Callable(),
	remote_hue_callback: Callable = Callable()
) -> void:
	owner_node = game_owner
	local_player = player
	apply_local_player_hue_callback = local_hue_callback
	apply_remote_player_hue_callback = remote_hue_callback
	local_player.z_index = Constants.LOCAL_PLAYER_Z_INDEX


func reset() -> void:
	if local_player != null:
		local_player.hide()

	for player_node in player_nodes.values():
		if player_node != local_player:
			player_node.queue_free()

	player_nodes.clear()
	initialized_players.clear()


func get_or_create_player_node(self_id: String, player_id: String) -> Player:
	if player_nodes.has(player_id):
		return player_nodes[player_id]

	if player_id == self_id:
		local_player.visible = true
		local_player.z_index = Constants.LOCAL_PLAYER_Z_INDEX
		if apply_local_player_hue_callback.is_valid():
			apply_local_player_hue_callback.call(local_player)
		player_nodes[player_id] = local_player
		return local_player

	var remote_player := PLAYER_SCENE.instantiate() as Player
	remote_player.z_index = Constants.REMOTE_PLAYER_Z_INDEX
	if apply_remote_player_hue_callback.is_valid():
		apply_remote_player_hue_callback.call(player_id, remote_player)
	owner_node.add_child(remote_player)
	player_nodes[player_id] = remote_player

	return remote_player


func has_player_node(player_id: String) -> bool:
	return player_nodes.has(player_id)


func get_player_ids() -> Array:
	return player_nodes.keys()


func get_player_node(player_id: String) -> Player:
	return player_nodes[player_id]


func get_remote_player_nodes(current_self_id: String) -> Dictionary:
	var remotes := {}
	for player_id in player_nodes.keys():
		if player_id == current_self_id:
			continue

		var player_node = player_nodes[player_id]
		if !is_instance_valid(player_node):
			continue

		remotes[player_id] = player_node

	return remotes


func mark_initialized(player_id: String) -> void:
	initialized_players[player_id] = true


func is_initialized(player_id: String) -> bool:
	return initialized_players.has(player_id)


func erase_player(player_id: String) -> void:
	player_nodes.erase(player_id)
	initialized_players.erase(player_id)


func remove_missing(server_players: Dictionary, self_id: String) -> Array:
	var removed_player_ids: Array = []
	for player_id in player_nodes.keys():
		if server_players.has(player_id):
			continue

		if remove_player_node(self_id, player_id):
			removed_player_ids.append(player_id)

	return removed_player_ids


func remove_player_node(self_id: String, player_id: String) -> bool:
	if !player_nodes.has(player_id):
		if player_id == self_id and local_player != null:
			local_player.stop_transient_effects()
			local_player.visible = false
		return false

	var player_node = player_nodes[player_id]
	if player_node == local_player:
		local_player.stop_transient_effects()
		local_player.visible = false
	else:
		if player_node.has_method("stop_transient_effects"):
			player_node.stop_transient_effects()
		player_node.queue_free()

	erase_player(player_id)
	return true
