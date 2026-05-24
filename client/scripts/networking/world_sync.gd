extends RefCounted
class_name WorldSync

signal bullet_spawned

const Constants = preload("res://scripts/constants/constants.gd")
const Packets = preload("res://scripts/networking/packets.gd")
const PlayerSyncState = preload("res://scripts/networking/player_sync_state.gd")
const VisualSyncPositions = preload("res://scripts/networking/visual_sync_positions.gd")
const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")
const PLAYER_SCENE := preload("res://scenes/player.tscn")
const BULLET_SCENE := preload("res://scenes/bullet.tscn")
const ASTEROID_SCENE := preload("res://scenes/asteroid.tscn")
const ASTEROID_Z_INDEX := 10
const BULLET_Z_INDEX := 20
const REMOTE_PLAYER_Z_INDEX := 30
const LOCAL_PLAYER_Z_INDEX := 31
const LOCAL_PLAYER_DEFAULT_HUE := 0.0
const REMOTE_PLAYER_HUES := [
	0.58,
	0.33,
	0.10,
	0.76,
	0.50,
	0.18,
	0.67,
	0.88,
]

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
var warned_missing_asteroid_scale := {}
var target_player_positions := {}
var target_player_rotations := {}
var target_bullet_positions := {}
var target_bullet_rotations := {}
var target_asteroid_positions := {}
var asteroid_server_positions := {}
var asteroid_visual_positions := {}
var local_server_position := Vector2.ZERO
var local_visual_position := Vector2.ZERO
var has_local_visual_position := false
var remote_player_visual_positions := {}
var remote_player_hues := {}
var current_self_id := ""


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
	local_player.z_index = LOCAL_PLAYER_Z_INDEX


func apply_state(
	self_id: String,
	server_players: Dictionary,
	server_bullets: Dictionary,
	server_asteroids: Dictionary,
	play_new_bullet_sounds: bool
) -> void:
	current_self_id = self_id
	_remove_missing_players(server_players, self_id)
	_remove_missing_bullets(server_bullets)
	_remove_missing_asteroids(server_asteroids)
	_apply_players(self_id, server_players)
	_apply_bullets(server_bullets, play_new_bullet_sounds)
	_apply_asteroids(server_asteroids)


func interpolate(delta: float) -> void:
	var weight := 1.0 - exp(-Constants.PLAYER_INTERPOLATION_SPEED * delta)
	for player_id in player_nodes.keys():
		if !target_player_positions.has(player_id):
			continue

		var player_node = player_nodes[player_id]
		player_node.position = player_node.position.lerp(target_player_positions[player_id], weight)
		player_node.rotation = lerp_angle(player_node.rotation, target_player_rotations[player_id], weight)
		if player_id == current_self_id:
			remote_player_visual_positions.erase(player_id)
		else:
			remote_player_visual_positions[player_id] = player_node.position

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
	if server_players.has(self_id):
		var local_state: Dictionary = server_players[self_id]
		_initialize_local_visual_position(Vector2(local_state[Packets.FIELD_X], local_state[Packets.FIELD_Y]))

	for player_id in server_players.keys():
		var state: Dictionary = server_players[player_id]
		var player_node = _get_player_node(self_id, player_id)
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
			_correct_remote_visual_copy_mismatch(player_id, player_node, visual_position)
			_apply_remote_player_hue(player_id, player_node)

		target_player_positions[player_id] = visual_position
		target_player_rotations[player_id] = server_rotation
		player_node.visible = !is_paused

		if !initialized_players.has(player_id):
			initialized_players[player_id] = true
			player_node.position = visual_position
			player_node.rotation = server_rotation


func _initialize_local_visual_position(server_position: Vector2) -> void:
	var wrapped_server_position := WorldWrapScript.wrap_position(server_position)
	if has_local_visual_position:
		local_visual_position += WorldWrapScript.shortest_delta(local_server_position, wrapped_server_position)
		local_server_position = wrapped_server_position
		return

	local_server_position = wrapped_server_position
	local_visual_position = local_server_position
	has_local_visual_position = true


func _correct_remote_visual_copy_mismatch(
	player_id: String,
	player_node: Node2D,
	visual_position: Vector2
) -> void:
	# Remote targets are local-relative, but rendered remotes can briefly stay in
	# an old visual copy; snap cache/render state before interpolation crosses it.
	if !initialized_players.has(player_id):
		return
	if !_is_world_copy_mismatch(player_node.position, visual_position):
		return

	player_node.position = visual_position
	target_player_positions[player_id] = visual_position
	remote_player_visual_positions[player_id] = visual_position


func _is_world_copy_mismatch(current_position: Vector2, target_position: Vector2) -> bool:
	return VisualSyncPositions.is_world_copy_mismatch(current_position, target_position)


func _get_player_node(self_id: String, player_id: String):
	if player_nodes.has(player_id):
		return player_nodes[player_id]

	if player_id == self_id:
		local_player.visible = true
		local_player.z_index = LOCAL_PLAYER_Z_INDEX
		local_player.set_player_hue(LOCAL_PLAYER_DEFAULT_HUE)
		player_nodes[player_id] = local_player
		return local_player

	var remote_player = PLAYER_SCENE.instantiate()
	remote_player.z_index = REMOTE_PLAYER_Z_INDEX
	_apply_remote_player_hue(player_id, remote_player)
	owner_node.add_child(remote_player)
	player_nodes[player_id] = remote_player

	return remote_player


func _apply_remote_player_hue(player_id: String, remote_player: Player) -> void:
	var remote_hue := _remote_hue_for_player(player_id)
	remote_player_hues[player_id] = remote_hue
	remote_player.set_player_hue(remote_hue)


func _remote_hue_for_player(player_id: String) -> float:
	var start_index := _player_id_hash(player_id) % REMOTE_PLAYER_HUES.size()
	for offset in range(REMOTE_PLAYER_HUES.size()):
		var hue: float = REMOTE_PLAYER_HUES[(start_index + offset) % REMOTE_PLAYER_HUES.size()]
		if !_hues_similar(hue, LOCAL_PLAYER_DEFAULT_HUE):
			return hue

	return 0.5


func _hues_similar(a: float, b: float, tolerance := 0.08) -> bool:
	var distance: float = abs(fposmod(a, 1.0) - fposmod(b, 1.0))
	return min(distance, 1.0 - distance) < tolerance


func _player_id_hash(player_id: String) -> int:
	var hash_value: int = 2166136261
	for index in range(player_id.length()):
		hash_value = int((hash_value ^ player_id.unicode_at(index)) * 16777619) & 0x7fffffff

	return hash_value


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
	remote_player_visual_positions.erase(player_id)
	remote_player_hues.erase(player_id)


func _apply_bullets(server_bullets: Dictionary, play_new_bullet_sounds: bool) -> void:
	for bullet_id in server_bullets.keys():
		var state: Dictionary = server_bullets[bullet_id]
		var is_new_bullet := !bullet_nodes.has(bullet_id)
		var bullet_node = _get_bullet_node(bullet_id)
		var server_position := Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])
		var visual_position := local_visual_position + WorldWrapScript.shortest_delta(
			local_server_position,
			server_position
		)
		var server_rotation: float = state[Packets.FIELD_ROTATION]

		target_bullet_positions[bullet_id] = visual_position
		target_bullet_rotations[bullet_id] = server_rotation

		if !initialized_bullets.has(bullet_id):
			initialized_bullets[bullet_id] = true
			bullet_node.global_position = visual_position
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
		var raw_server_position := Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])
		var visual_position: Vector2

		if asteroid_server_positions.has(asteroid_id):
			visual_position = asteroid_visual_positions[asteroid_id] + WorldWrapScript.shortest_delta(
				asteroid_server_positions[asteroid_id],
				raw_server_position
			)
			target_asteroid_positions[asteroid_id] = visual_position
			asteroid_server_positions[asteroid_id] = raw_server_position
			asteroid_visual_positions[asteroid_id] = visual_position
		else:
			# First-seen asteroid positions may intentionally be outside wrapped world bounds for offscreen spawns.
			visual_position = local_visual_position + WorldWrapScript.shortest_delta(
				local_server_position,
				raw_server_position
			)
			target_asteroid_positions[asteroid_id] = visual_position
			asteroid_server_positions[asteroid_id] = raw_server_position
			asteroid_visual_positions[asteroid_id] = visual_position

		_apply_asteroid_scale(asteroid_id, asteroid_node, state)

		if !initialized_asteroids.has(asteroid_id):
			initialized_asteroids[asteroid_id] = true
			asteroid_node.global_position = visual_position
			asteroid_node.set_asteroid_variant(state[Packets.FIELD_VARIANT])


func _apply_asteroid_scale(asteroid_id: String, asteroid_node: Node2D, state: Dictionary) -> void:
	if state.has(Packets.FIELD_SCALE):
		asteroid_node.scale = Vector2.ONE * float(state[Packets.FIELD_SCALE])
		return

	if warned_missing_asteroid_scale.has(asteroid_id):
		return

	warned_missing_asteroid_scale[asteroid_id] = true
	push_warning("Asteroid state missing scale for %s" % asteroid_id)


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
		warned_missing_asteroid_scale.erase(asteroid_id)
		target_asteroid_positions.erase(asteroid_id)
		asteroid_server_positions.erase(asteroid_id)
		asteroid_visual_positions.erase(asteroid_id)


func get_remote_player_visual_positions() -> Dictionary:
	var positions := remote_player_visual_positions.duplicate()
	positions.erase(current_self_id)
	return positions


func get_remote_player_hues() -> Dictionary:
	var hues := remote_player_hues.duplicate()
	hues.erase(current_self_id)
	return hues


func visual_position_for_server_position(server_position: Vector2) -> Vector2:
	if !has_local_visual_position:
		return server_position

	return local_visual_position + WorldWrapScript.shortest_delta(
		local_server_position,
		server_position
	)
