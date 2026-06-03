extends RefCounted
class_name GameplayShellFlow

const PlayerPauseStatePacketReader = preload("res://scripts/gameplay/state/player_pause_state_packet_reader.gd")
const PlayerPauseStateTracker = preload("res://scripts/gameplay/state/player_pause_state_tracker.gd")
const GameplayStateApplyFlowScript = preload("res://scripts/gameplay/state/gameplay_state_apply_flow.gd")
const GameplayProcessFlowScript = preload("res://scripts/gameplay/runtime/gameplay_process_flow.gd")
const ServerHitboxOverlayFlowScript = preload("res://scripts/gameplay/debug/server_hitbox_overlay_flow.gd")

signal gameplay_started
signal quit_to_main_menu_requested
signal return_to_lobby_requested

var player
var runtime_context
var connection_service
var hud_flow
var menu_flow
var input_context
var runtime_tick_flow
var spectate_context
var player_pause_state_tracker
var gameplay_state_apply_flow
var gameplay_process_flow
var server_hitbox_overlay_flow
var has_received_state := false
var game_owner: Node2D


func configure(
	connection_service_ref,
	game_owner_ref: Node2D,
	player_ref: Player,
	bullets: Node2D,
	asteroids: Node2D,
	hud_flow_ref,
	menu_flow_ref
) -> void:
	connection_service = connection_service_ref
	self.game_owner = game_owner_ref
	player = player_ref
	hud_flow = hud_flow_ref
	menu_flow = menu_flow_ref
	player_pause_state_tracker = PlayerPauseStateTracker.new()
	if menu_flow != null:
		menu_flow.configure_lifecycle_routes(
			Callable(self, "_on_quit_to_main_menu_requested"),
			Callable(self, "_on_return_to_lobby_requested")
	)
	runtime_context = GameplayRuntimeContext.new()
	runtime_context.configure_world(game_owner_ref, player_ref, bullets, asteroids, player_pause_state_tracker)
	runtime_context.configure_events(
		game_owner_ref,
		hud_flow.hud if hud_flow != null else null,
		hud_flow,
		menu_flow
	)
	runtime_context.configure_respawn(connection_service, hud_flow)
	input_context = GameplayInputContext.new()
	input_context.configure(
		connection_service,
		player,
		menu_flow,
		game_owner_ref,
		Callable(runtime_context, "request_respawn"),
		Callable(runtime_context, "target_visual_candidates"),
		Callable(self, "mouse_visual_position"),
		Callable(self, "server_position_for_visual_position"),
		Callable(runtime_context, "remote_player_nodes")
	)
	gameplay_state_apply_flow = GameplayStateApplyFlowScript.new()
	gameplay_state_apply_flow.configure(input_context, hud_flow, runtime_context, menu_flow)
	server_hitbox_overlay_flow = ServerHitboxOverlayFlowScript.new()
	server_hitbox_overlay_flow.configure(game_owner_ref, runtime_context)
	runtime_tick_flow = GameplayRuntimeTickFlow.new()
	runtime_tick_flow.configure(hud_flow)
	spectate_context = GameplaySpectateContext.new()
	spectate_context.configure(menu_flow, null, runtime_context.world_sync)
	input_context.configure_spectate_routes(
		Callable(spectate_context, "request_open_spectate_menu"),
		Callable(spectate_context, "request_cycle_target")
	)
	gameplay_process_flow = GameplayProcessFlowScript.new()
	gameplay_process_flow.configure(
		runtime_context,
		server_hitbox_overlay_flow,
		runtime_tick_flow,
		input_context,
		spectate_context
	)


func reset() -> void:
	has_received_state = false
	if runtime_context != null:
		runtime_context.reset()
	if hud_flow != null:
		hud_flow.reset()
	if menu_flow != null && menu_flow.has_method("reset"):
		menu_flow.reset()
	if input_context != null:
		input_context.reset()
	if runtime_tick_flow != null:
		runtime_tick_flow.reset()
	if spectate_context != null:
		spectate_context.reset()
	if player_pause_state_tracker != null:
		player_pause_state_tracker.reset()
	if server_hitbox_overlay_flow != null:
		server_hitbox_overlay_flow.reset()


func apply_gameplay_state(state: Dictionary) -> void:
	_apply_gameplay_state(state)


func _apply_gameplay_state(state: Dictionary) -> void:
	if gameplay_state_apply_flow == null:
		return

	var result: GameplayStateApplyResult = gameplay_state_apply_flow.apply_state(state, has_received_state)
	has_received_state = result.has_received_state
	if result.started_gameplay:
		gameplay_started.emit()


func apply_player_pause_state_packet(packet: Dictionary) -> void:
	if !PlayerPauseStatePacketReader.is_player_pause_state(packet):
		return
	var state := PlayerPauseStatePacketReader.read(packet)
	player_pause_state_tracker.apply_state(state)


func configure_spectate_menu_state(spectate_menu_state_ref) -> void:
	if spectate_context != null:
		spectate_context.configure_menu_state(spectate_menu_state_ref)


func configure_debug_placement_route(route: Callable) -> void:
	if input_context != null:
		input_context.configure_debug_placement_route(route)


func handle_debug_placement_result(result: Dictionary) -> void:
	if input_context != null:
		input_context.handle_debug_placement_result(result)


func refresh_debug_spawn_player_slots(max_players: int) -> void:
	if input_context != null:
		input_context.refresh_debug_spawn_player_slots(max_players)


func current_camera() -> Camera2D:
	if runtime_context == null:
		return null
	return runtime_context.current_camera()


func remote_player_visual_positions() -> Dictionary:
	if runtime_context == null:
		return {}
	return runtime_context.remote_player_visual_positions()


func remote_player_hues() -> Dictionary:
	if runtime_context == null:
		return {}
	return runtime_context.remote_player_hues()


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	if runtime_context == null:
		return visual_position
	return runtime_context.server_position_for_visual_position(visual_position)


func mouse_visual_position() -> Vector2:
	if runtime_context == null:
		return Vector2.ZERO
	if current_camera() == null:
		return Vector2.ZERO
	if game_owner == null:
		return Vector2.ZERO
	return game_owner.get_global_mouse_position()

func handle_unhandled_input(event: InputEvent) -> bool:
	if input_context == null:
		return false
	return input_context.handle_unhandled_input(event, has_received_state)


func process(_delta: float) -> void:
	if gameplay_process_flow == null:
		return
	gameplay_process_flow.process(_delta, has_received_state)


func _on_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()


func _on_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()
