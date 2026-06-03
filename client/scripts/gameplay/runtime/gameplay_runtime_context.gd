extends RefCounted
class_name GameplayRuntimeContext

const WorldSyncScript = preload("res://scripts/world/world_sync.gd")

var world_sync
var player
var respawn_flow


func configure_world(
	game_owner: Node2D,
	player_ref: Player,
	bullets: Node2D,
	asteroids: Node2D,
	pause_state_tracker = null
) -> void:
	player = player_ref
	world_sync = WorldSyncScript.new()
	world_sync.configure(game_owner, player_ref, bullets, asteroids, pause_state_tracker)


func configure_respawn(connection_service_ref, hud_flow_ref) -> void:
	respawn_flow = GameplayRespawnFlow.new()
	respawn_flow.configure(connection_service_ref, hud_flow_ref)


func reset() -> void:
	if player != null:
		player.hide()
	if world_sync != null:
		world_sync.reset()
	if respawn_flow != null:
		respawn_flow.reset()


func process(delta: float) -> void:
	if world_sync != null:
		world_sync.interpolate(delta)


func request_respawn(has_received_state: bool) -> void:
	if respawn_flow != null:
		respawn_flow.request_respawn(has_received_state)


func apply_world_state(state: Dictionary, has_received_state: bool) -> void:
	if world_sync == null:
		return

	world_sync.apply_state(
		state["self_id"],
		state["server_players"],
		state["server_bullets"],
		state["server_asteroids"],
		has_received_state
	)


func current_camera() -> Camera2D:
	if player == null:
		return null
	return player.get_node_or_null("Camera2D") as Camera2D


func remote_player_visual_positions() -> Dictionary:
	if world_sync == null:
		return {}
	return world_sync.get_remote_player_visual_positions()


func remote_player_hues() -> Dictionary:
	if world_sync == null:
		return {}
	return world_sync.get_remote_player_hues()


func remote_player_nodes() -> Dictionary:
	if world_sync == null:
		return {}
	return world_sync.remote_player_nodes()


func server_hitbox_draw_entries() -> Array:
	if world_sync == null:
		return []
	return world_sync.server_hitbox_draw_entries()


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	if world_sync == null:
		return visual_position
	return world_sync.server_position_for_visual_position(visual_position)
