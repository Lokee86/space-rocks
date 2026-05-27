extends RefCounted
class_name GameplayRuntimeContext

const WorldSyncScript = preload("res://scripts/world/world_sync.gd")
const GameplayEventFlow = preload("res://scripts/shell/gameplay_event_flow.gd")
const GameplayDeathFlow = preload("res://scripts/shell/gameplay_death_flow.gd")
const GameplayRespawnFlow = preload("res://scripts/shell/gameplay_respawn_flow.gd")
const GameplayInputContext = preload("res://scripts/gameplay/input/gameplay_input_context.gd")

var world_sync
var player
var event_flow
var death_flow
var respawn_flow
var input_context
var hud_flow


func configure_world(
	game_owner: Node2D,
	player_ref: Player,
	bullets: Node2D,
	asteroids: Node2D
) -> void:
	player = player_ref
	world_sync = WorldSyncScript.new()
	world_sync.configure(game_owner, player_ref, bullets, asteroids)


func configure_events(
	game_owner: Node2D,
	game_over_sound: AudioStreamPlayer,
	hud_flow_ref,
	menu_flow_ref
) -> void:
	event_flow = GameplayEventFlow.new()
	event_flow.configure(
		game_owner,
		game_over_sound,
		Callable(world_sync, "visual_position_for_server_position")
	)
	death_flow = GameplayDeathFlow.new()
	death_flow.configure(hud_flow_ref, menu_flow_ref, event_flow)
	event_flow.self_death_event.connect(Callable(death_flow, "apply_self_death_event"))


func configure_respawn(connection_service_ref, hud_flow_ref) -> void:
	hud_flow = hud_flow_ref
	respawn_flow = GameplayRespawnFlow.new()
	respawn_flow.configure(connection_service_ref, hud_flow_ref)


func configure_input(connection_service_ref, player_ref, menu_flow_ref) -> void:
	input_context = GameplayInputContext.new()
	input_context.configure(connection_service_ref, player_ref, menu_flow_ref)


func configure_pause_input(menu_flow_ref) -> void:
	if input_context == null:
		input_context = GameplayInputContext.new()
	input_context.configure_pause_input(menu_flow_ref)


func reset() -> void:
	if player != null:
		player.hide()
	if world_sync != null:
		world_sync.reset()
	if event_flow != null:
		event_flow.reset()
	if death_flow != null:
		death_flow.reset()
	if respawn_flow != null:
		respawn_flow.reset()
	if input_context != null:
		input_context.reset()


func process(delta: float) -> void:
	if world_sync != null:
		world_sync.interpolate(delta)


func process_respawn(has_received_state: bool) -> void:
	if respawn_flow != null:
		respawn_flow.process(has_received_state)


func mark_input_gameplay_state_received() -> void:
	if input_context != null:
		input_context.mark_gameplay_state_received()


func process_input(has_received_state: bool) -> void:
	if input_context != null:
		input_context.process(has_received_state)


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


func apply_server_events(state: Dictionary) -> void:
	if event_flow != null:
		event_flow.apply_server_events(state["server_events"], state["self_id"])


func apply_respawn_alive_restore(state: Dictionary, menu_flow_ref) -> void:
	if (
		hud_flow == null
		|| respawn_flow == null
		|| !respawn_flow.should_restore_alive_hud(state, player)
	):
		return

	hud_flow.set_alive()
	if menu_flow_ref != null:
		menu_flow_ref.set_alive()
	respawn_flow.clear_awaiting_confirmation()


func current_camera() -> Camera2D:
	if player == null:
		return null
	return player.get_node_or_null("Camera2D") as Camera2D


func remote_player_visual_positions() -> Dictionary:
	if world_sync == null:
		return {}
	return world_sync.get_remote_player_visual_positions()
