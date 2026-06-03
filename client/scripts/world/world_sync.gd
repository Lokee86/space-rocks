extends RefCounted

signal bullet_spawned

const Constants = preload("res://scripts/constants/constants.gd")
const AsteroidSyncScript = preload("res://scripts/world/asteroid_sync.gd")
const BulletSyncScript = preload("res://scripts/world/bullet_sync.gd")
const LocalVisualSyncScript = preload("res://scripts/world/local_visual_sync.gd")
const Packets = preload("res://scripts/networking/packets/packets.gd")
const PlayerSyncScript = preload("res://scripts/world/player_sync.gd")

var asteroid_sync
var bullet_sync
var local_visual_sync
var player_sync
var local_player: Player
var current_self_id := ""


func configure(
	game_owner: Node2D,
	player: Player,
	bullets: Node2D,
	asteroids: Node2D,
	pause_state_tracker = null
) -> void:
	asteroid_sync = AsteroidSyncScript.new()
	asteroid_sync.configure(asteroids)
	bullet_sync = BulletSyncScript.new()
	bullet_sync.configure(bullets)
	local_player = player
	bullet_sync.bullet_spawned.connect(func() -> void:
		if local_player != null:
			local_player.play_laser_sound()
		bullet_spawned.emit()
	)
	local_visual_sync = LocalVisualSyncScript.new()
	player_sync = PlayerSyncScript.new()
	player_sync.configure(game_owner, player, pause_state_tracker)

	asteroids.z_index = Constants.ASTEROID_Z_INDEX
	bullets.z_index = Constants.BULLET_Z_INDEX


func reset() -> void:
	current_self_id = ""
	if player_sync != null:
		player_sync.reset()
	if asteroid_sync != null:
		asteroid_sync.reset()
	clear_view_target_player()


func apply_state(
	self_id: String,
	server_players: Dictionary,
	server_bullets: Dictionary,
	server_asteroids: Dictionary,
	play_new_bullet_sounds: bool
) -> void:
	current_self_id = self_id
	player_sync.remove_missing(server_players, self_id)
	bullet_sync.remove_missing(server_bullets)
	asteroid_sync.remove_missing(server_asteroids)
	if server_players.has(self_id):
		var local_state: Dictionary = server_players[self_id]
		local_visual_sync.update_from_server_position(
			Vector2(local_state[Packets.FIELD_X], local_state[Packets.FIELD_Y])
		)
	player_sync.apply(
		self_id,
		server_players,
		local_visual_sync.visual_position(),
		local_visual_sync.server_position()
	)
	bullet_sync.apply(
		server_bullets,
		play_new_bullet_sounds,
		local_visual_sync.visual_position(),
		local_visual_sync.server_position()
	)
	asteroid_sync.apply(
		server_asteroids,
		local_visual_sync.visual_position(),
		local_visual_sync.server_position()
	)


func interpolate(delta: float) -> void:
	var weight := 1.0 - exp(-Constants.PLAYER_INTERPOLATION_SPEED * delta)
	player_sync.interpolate(weight, current_self_id)
	bullet_sync.interpolate(weight)
	asteroid_sync.interpolate(weight)


func get_remote_player_visual_positions() -> Dictionary:
	return player_sync.get_remote_player_visual_positions(current_self_id)


func get_remote_player_hues() -> Dictionary:
	if player_sync == null:
		return {}
	return player_sync.get_remote_player_hues(current_self_id)


func remote_player_nodes() -> Dictionary:
	if player_sync == null:
		return {}
	return player_sync.remote_player_nodes(current_self_id)


func player_nodes() -> Dictionary:
	if player_sync == null:
		return {}
	return player_sync.player_nodes()


func focus_camera_on_player(player_id: String) -> bool:
	if player_sync == null:
		return false
	return player_sync.focus_camera_on_player(player_id)


func set_view_target_player(player_id: String) -> void:
	if player_sync != null:
		player_sync.set_view_target_player(player_id)


func clear_view_target_player() -> void:
	if player_sync != null:
		player_sync.clear_view_target_player()


func visual_position_for_server_position(server_position: Vector2) -> Vector2:
	return local_visual_sync.visual_position_for_server_position(server_position)


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	return local_visual_sync.server_position_for_visual_position(visual_position)


func player_target_positions() -> Dictionary:
	var positions := {}
	if player_sync == null:
		return positions

	if current_self_id != "":
		positions[current_self_id] = {
			"visual_position": local_visual_sync.visual_position(),
			"server_position": local_visual_sync.server_position()
		}

	var remote_positions: Dictionary = player_sync.get_remote_player_visual_positions(current_self_id)
	for player_id in remote_positions.keys():
		var visual_position = remote_positions[player_id]
		positions[player_id] = {
			"visual_position": visual_position,
			"server_position": visual_position
		}

	return positions


func asteroid_target_positions() -> Dictionary:
	if asteroid_sync == null:
		return {}
	return asteroid_sync.asteroid_target_positions()


func bullet_target_positions() -> Dictionary:
	if bullet_sync == null:
		return {}
	return bullet_sync.bullet_target_positions()


func server_hitbox_draw_entries() -> Array:
	var entries: Array = []
	if player_sync == null || asteroid_sync == null || bullet_sync == null:
		return entries

	entries.append_array(player_sync.server_hitbox_draw_entries(current_self_id))
	entries.append_array(asteroid_sync.server_hitbox_draw_entries())
	entries.append_array(bullet_sync.server_hitbox_draw_entries())
	return entries
