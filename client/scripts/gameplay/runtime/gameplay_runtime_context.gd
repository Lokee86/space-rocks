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

