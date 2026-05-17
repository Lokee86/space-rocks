extends RefCounted
class_name WorldSync

signal bullet_spawned

const Constants = preload("res://scripts/constants.gd")
const Packets = preload("res://scripts/packets.gd")
const PLAYER_SCENE := preload("res://scenes/player.tscn")
const BULLET_SCENE := preload("res://scenes/bullet.tscn")
const ASTEROID_SCENE := preload("res://scenes/asteroid.tscn")
const ASTEROID_Z_INDEX := 10
const BULLET_Z_INDEX := 20
const PLAYER_Z_INDEX := 30

var owner_node: Node2D
var local_player: Player
var bullets_layer: Node2D
var asteroids_layer: Node2D
var player_nodes := {}
var bullet_nodes := {}
var asteroid_nodes := {}
var initialized_players := {}
var initialized_bullets := {}
var initialized_asteroids := {}
var target_player_positions := {}
var target_player_rotations := {}
var target_bullet_positions := {}
var target_bullet_rotations := {}
var target_asteroid_positions := {}


func configure(
	game_owner: Node2D,
	player: Player,
	bullets: Node2D,
	asteroids: Node2D
) -> void:
	owner_node = game_owner
	local_player = player
	bullets_layer = bullets
	asteroids_layer = asteroids

	asteroids_layer.z_index = ASTEROID_Z_INDEX
	bullets_layer.z_index = BULLET_Z_INDEX
	local_player.z_index = PLAYER_Z_INDEX


func apply_state(
	self_id: String,
	server_players: Dictionary,
	server_bullets: Dictionary,
	server_asteroids: Dictionary,
	play_new_bullet_sounds: bool
) -> void:
	_remove_missing_players(server_players, self_id)
	_remove_missing_bullets(server_bullets)
	_remove_missing_asteroids(server_asteroids)
	_apply_bullets(server_bullets, play_new_bullet_sounds)
	_apply_asteroids(server_asteroids)
	_apply_players(self_id, server_players)


func interpolate(delta: float) -> void:
	var weight := 1.0 - exp(-Constants.PLAYER_INTERPOLATION_SPEED * delta)
	for player_id in player_nodes.keys():
		if !target_player_positions.has(player_id):
			continue

		var player_node = player_nodes[player_id]
		player_node.position = player_node.position.lerp(target_player_positions[player_id], weight)
		player_node.rotation = lerp_angle(player_node.rotation, target_player_rotations[player_id], weight)

	for bullet_id in bullet_nodes.keys():
		if !target_bullet_positions.has(bullet_id):
			continue

		var bullet_node = bullet_nodes[bullet_id]
		bullet_node.global_position = bullet_node.global_position.lerp(
			target_bullet_positions[bullet_id],
			weight
		)
		bullet_node.rotation = lerp_angle(bullet_node.rotation, target_bullet_rotations[bullet_id], weight)

	for asteroid_id in asteroid_nodes.keys():
		if !target_asteroid_positions.has(asteroid_id):
			continue

		var asteroid_node = asteroid_nodes[asteroid_id]
		asteroid_node.global_position = asteroid_node.global_position.lerp(
			target_asteroid_positions[asteroid_id],
			weight
		)


func _apply_players(self_id: String, server_players: Dictionary) -> void:
	for player_id in server_players.keys():
		var state: Dictionary = server_players[player_id]
		var player_node = _get_player_node(self_id, player_id)
		var server_position := Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])
		var server_rotation: float = state[Packets.FIELD_ROTATION]

		target_player_positions[player_id] = server_position
		target_player_rotations[player_id] = server_rotation

		if !initialized_players.has(player_id):
			initialized_players[player_id] = true
			player_node.position = server_position
			player_node.rotation = server_rotation


func _get_player_node(self_id: String, player_id: String):
	if player_nodes.has(player_id):
		return player_nodes[player_id]

	if player_id == self_id:
		local_player.visible = true
		local_player.z_index = PLAYER_Z_INDEX
		player_nodes[player_id] = local_player
		return local_player

	var remote_player = PLAYER_SCENE.instantiate()
	remote_player.z_index = PLAYER_Z_INDEX
	owner_node.add_child(remote_player)
	player_nodes[player_id] = remote_player

	return remote_player


func _remove_missing_players(server_players: Dictionary, self_id: String) -> void:
	for player_id in player_nodes.keys():
		if server_players.has(player_id):
			continue

		_remove_player_node(self_id, player_id)


func _remove_player_node(self_id: String, player_id: String) -> void:
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


func _apply_bullets(server_bullets: Dictionary, play_new_bullet_sounds: bool) -> void:
	for bullet_id in server_bullets.keys():
		var state: Dictionary = server_bullets[bullet_id]
		var is_new_bullet := !bullet_nodes.has(bullet_id)
		var bullet_node = _get_bullet_node(bullet_id)
		var server_position := Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])
		var server_rotation: float = state[Packets.FIELD_ROTATION]

		target_bullet_positions[bullet_id] = server_position
		target_bullet_rotations[bullet_id] = server_rotation

		if !initialized_bullets.has(bullet_id):
			initialized_bullets[bullet_id] = true
			bullet_node.global_position = server_position
			bullet_node.rotation = server_rotation

		if is_new_bullet && play_new_bullet_sounds:
			bullet_spawned.emit()


func _get_bullet_node(bullet_id):
	if bullet_nodes.has(bullet_id):
		return bullet_nodes[bullet_id]

	var bullet_node = BULLET_SCENE.instantiate()
	bullets_layer.add_child(bullet_node)
	bullet_nodes[bullet_id] = bullet_node

	return bullet_node


func _remove_missing_bullets(server_bullets: Dictionary) -> void:
	for bullet_id in bullet_nodes.keys():
		if server_bullets.has(bullet_id):
			continue

		bullet_nodes[bullet_id].queue_free()
		bullet_nodes.erase(bullet_id)
		initialized_bullets.erase(bullet_id)
		target_bullet_positions.erase(bullet_id)
		target_bullet_rotations.erase(bullet_id)


func _apply_asteroids(server_asteroids: Dictionary) -> void:
	for asteroid_id in server_asteroids.keys():
		var state: Dictionary = server_asteroids[asteroid_id]
		var asteroid_node = _get_asteroid_node(asteroid_id)
		var server_position := Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])

		target_asteroid_positions[asteroid_id] = server_position

		if !initialized_asteroids.has(asteroid_id):
			initialized_asteroids[asteroid_id] = true
			asteroid_node.global_position = server_position
			asteroid_node.scale = Vector2.ONE * float(state[Packets.FIELD_SIZE]) * Constants.ASTEROID_SIZE_SCALE
			asteroid_node.set_asteroid_variant(state[Packets.FIELD_VARIANT])


func _get_asteroid_node(asteroid_id):
	if asteroid_nodes.has(asteroid_id):
		return asteroid_nodes[asteroid_id]

	var asteroid_node = ASTEROID_SCENE.instantiate()
	asteroids_layer.add_child(asteroid_node)
	asteroid_nodes[asteroid_id] = asteroid_node

	return asteroid_node


func _remove_missing_asteroids(server_asteroids: Dictionary) -> void:
	for asteroid_id in asteroid_nodes.keys():
		if server_asteroids.has(asteroid_id):
			continue

		asteroid_nodes[asteroid_id].queue_free()
		asteroid_nodes.erase(asteroid_id)
		initialized_asteroids.erase(asteroid_id)
		target_asteroid_positions.erase(asteroid_id)
