extends RefCounted
class_name GameplayShellFlow

const GameplayStatePacketReader = preload("res://scripts/gameplay/state/gameplay_state_packet_reader.gd")
const PlayerPauseStatePacketReader = preload("res://scripts/gameplay/state/player_pause_state_packet_reader.gd")
const PlayerPauseStateTracker = preload("res://scripts/gameplay/state/player_pause_state_tracker.gd")

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
var has_received_state := false


func configure(
	connection_service_ref,
	game_owner: Node2D,
	player_ref: Player,
	bullets: Node2D,
	asteroids: Node2D,
	hud_flow_ref,
	menu_flow_ref
) -> void:
	connection_service = connection_service_ref
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
	runtime_context.configure_world(game_owner, player_ref, bullets, asteroids, player_pause_state_tracker)
	runtime_context.configure_events(
		game_owner,
		hud_flow.hud if hud_flow != null else null,
		hud_flow,
		menu_flow
	)
	runtime_context.configure_respawn(connection_service, hud_flow)
	runtime_tick_flow = GameplayRuntimeTickFlow.new()
	runtime_tick_flow.configure(hud_flow)
	input_context = GameplayInputContext.new()
	input_context.configure(
		connection_service,
		player,
		menu_flow,
		Callable(runtime_context, "request_respawn")
	)
	spectate_context = GameplaySpectateContext.new()
	spectate_context.configure(menu_flow, null, runtime_context.world_sync)
	input_context.configure_spectate_routes(
		Callable(spectate_context, "request_open_spectate_menu"),
		Callable(spectate_context, "request_cycle_target")
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


func apply_gameplay_state(packet: Dictionary) -> void:
	if runtime_context == null:
		return

	var is_first_gameplay_state := !has_received_state
	var state := GameplayStatePacketReader.read(packet)
	if input_context != null:
		input_context.apply_debug_status(state.get("debug_status", {}))
	if input_context != null:
		input_context.mark_gameplay_state_received()
	if hud_flow != null:
		hud_flow.apply_gameplay_state_summary(state)
	runtime_context.apply_world_state(state, has_received_state)
	runtime_context.apply_respawn_alive_restore(state, menu_flow)
	runtime_context.apply_server_events(state)
	has_received_state = true
	if is_first_gameplay_state:
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
	if input_context != null && input_context.has_method("handle_debug_placement_result"):
		input_context.handle_debug_placement_result(result)


func current_camera() -> Camera2D:
	if runtime_context == null:
		return null
	return runtime_context.current_camera()


func remote_player_visual_positions() -> Dictionary:
	if runtime_context == null:
		return {}
	return runtime_context.remote_player_visual_positions()


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	if runtime_context == null:
		return visual_position
	return runtime_context.server_position_for_visual_position(visual_position)


func process(_delta: float) -> void:
	if runtime_context != null:
		runtime_context.process(_delta)
	if runtime_tick_flow != null:
		runtime_tick_flow.process(_delta)
	if input_context != null:
		input_context.process(has_received_state)
	if spectate_context != null:
		spectate_context.process()


func _on_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()


func _on_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()
