extends RefCounted
class_name WorldSync

signal bullet_spawned

const Constants = preload("res://scripts/constants/constants.gd")
const AsteroidSyncScript = preload("res://scripts/networking/asteroid_sync.gd")
const BulletSyncScript = preload("res://scripts/networking/bullet_sync.gd")
const LocalVisualSyncScript = preload("res://scripts/networking/local_visual_sync.gd")
const Packets = preload("res://scripts/networking/packets.gd")
const PlayerSyncScript = preload("res://scripts/networking/player_sync.gd")
const ASTEROID_Z_INDEX := 10
const BULLET_Z_INDEX := 20

var owner_node: Node2D
var bullets_layer: Node2D
var asteroids_layer: Node2D
var asteroid_sync
var bullet_sync
var local_visual_sync
var player_sync
var current_self_id := ""


func configure(
	game_owner: Node2D,
	player: Player,
	bullets: Node2D,
	asteroids: Node2D
) -> void:
	owner_node = game_owner
	bullets_layer = bullets
	asteroids_layer = asteroids
	asteroid_sync = AsteroidSyncScript.new()
	asteroid_sync.configure(asteroids_layer)
	bullet_sync = BulletSyncScript.new()
	bullet_sync.configure(bullets_layer)
	bullet_sync.bullet_spawned.connect(func() -> void:
		bullet_spawned.emit()
	)
	local_visual_sync = LocalVisualSyncScript.new()
	player_sync = PlayerSyncScript.new()
	player_sync.configure(owner_node, player)

	asteroids_layer.z_index = ASTEROID_Z_INDEX
	bullets_layer.z_index = BULLET_Z_INDEX


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
	return player_sync.get_remote_player_hues(current_self_id)


func visual_position_for_server_position(server_position: Vector2) -> Vector2:
	return local_visual_sync.visual_position_for_server_position(server_position)
