extends RefCounted

const Constants = preload("res://scripts/generated/constants/constants.gd")
const AsteroidSyncScript = preload("res://scripts/world/asteroid_sync.gd")
const BulletSyncScript = preload("res://scripts/world/bullet_sync.gd")
const PickupSyncScript = preload("res://scripts/world/pickup_sync.gd")
const PlayerRenderApiScript = preload("res://scripts/world/player_render/player_render_api.gd")
const TargetPositionSource = preload("res://scripts/gameplay/targeting/target_position_source.gd")

var asteroid_sync
var bullet_sync
var pickup_sync
var player_render_api
var target_position_source
var view_anchor: Node2D
var local_player: Player
var current_self_id := ""


func configure(
	game_owner: Node2D,
	player: Player,
	view_anchor_ref: Node2D,
	bullets: Node2D,
	asteroids: Node2D,
	pickups: Node2D,
	pause_state_tracker = null
) -> void:
	asteroid_sync = AsteroidSyncScript.new()
	asteroid_sync.configure(asteroids)
	bullet_sync = BulletSyncScript.new()
	bullet_sync.configure(bullets)
	pickup_sync = PickupSyncScript.new()
	pickup_sync.configure(pickups)
	local_player = player
	player_render_api = PlayerRenderApiScript.new()
	view_anchor = view_anchor_ref
	player_render_api.configure(game_owner, player, view_anchor_ref, pause_state_tracker)
	target_position_source = TargetPositionSource.new()
	target_position_source.configure(player_render_api, asteroid_sync, bullet_sync, pickup_sync)

	asteroids.z_index = Constants.ASTEROID_Z_INDEX
	pickups.z_index = Constants.PICKUP_Z_INDEX
	bullets.z_index = Constants.BULLET_Z_INDEX


func reset() -> void:
	current_self_id = ""
	if player_render_api != null:
		player_render_api.reset()
	if asteroid_sync != null:
		asteroid_sync.reset()
	if pickup_sync != null:
		pickup_sync.reset()
	clear_view_target_player()


func apply_state(
	self_id: String,
	server_players: Dictionary,
	server_bullets: Dictionary,
	server_asteroids: Dictionary,
	server_pickups: Dictionary = {}
) -> void:
	current_self_id = self_id

	if target_position_source != null:
		target_position_source.set_current_self_id(self_id)
	player_render_api.remove_missing(server_players, self_id)
	bullet_sync.remove_missing(server_bullets)
	asteroid_sync.remove_missing(server_asteroids)
	pickup_sync.remove_missing(server_pickups)
	player_render_api.apply_state(self_id, server_players)
	bullet_sync.apply(
		server_bullets,
		player_render_api.visual_position(),
		player_render_api.server_position()
	)
	asteroid_sync.apply(
		server_asteroids,
		player_render_api.visual_position(),
		player_render_api.server_position()
	)
	pickup_sync.apply(
		server_pickups,
		player_render_api.visual_position(),
		player_render_api.server_position()
	)


func interpolate(delta: float) -> void:
	var weight := 1.0 - exp(-Constants.PLAYER_INTERPOLATION_SPEED * delta)
	player_render_api.interpolate(weight, current_self_id)
	bullet_sync.interpolate(weight)
	asteroid_sync.interpolate(weight)
	pickup_sync.interpolate(weight)


func get_remote_player_visual_positions() -> Dictionary:
	if player_render_api == null:
		return {}
	return player_render_api.get_remote_player_visual_positions(current_self_id)


func get_remote_player_hues() -> Dictionary:
	if player_render_api == null:
		return {}
	return player_render_api.get_remote_player_hues(current_self_id)


func remote_player_nodes() -> Dictionary:
	if player_render_api == null:
		return {}
	return player_render_api.remote_player_nodes(current_self_id)


func player_nodes() -> Dictionary:
	if player_render_api == null:
		return {}
	return player_render_api.player_nodes()


func focus_camera_on_player(player_id: String) -> bool:
	if player_render_api == null:
		return false
	return player_render_api.focus_camera_on_player(player_id)


func set_view_target_player(player_id: String) -> void:
	if player_render_api != null:
		player_render_api.set_view_target_player(player_id)


func clear_view_target_player() -> void:
	if player_render_api != null:
		player_render_api.clear_view_target_player()


func visual_position_for_server_position(server_position: Vector2) -> Vector2:
	return player_render_api.visual_position_for_server_position(server_position)


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	return player_render_api.server_position_for_visual_position(visual_position)


func target_source():
	return target_position_source
