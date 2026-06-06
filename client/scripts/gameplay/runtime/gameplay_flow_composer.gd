extends RefCounted
class_name GameplayFlowComposer

const GameplayEventLifecycleFlowScript = preload("res://scripts/gameplay/events/gameplay_event_lifecycle_flow.gd")
const GameplayAliveRestoreFlowScript = preload("res://scripts/gameplay/respawn/gameplay_alive_restore_flow.gd")
const GameplayTargetingContextScript = preload("res://scripts/gameplay/targeting/gameplay_targeting_context.gd")
const ServerHitboxOverlayFlowScript = preload("res://scripts/gameplay/debug/server_hitbox_overlay_flow.gd")
const GameplayProcessFlowScript = preload("res://scripts/gameplay/runtime/gameplay_process_flow.gd")

var event_lifecycle_flow
var alive_restore_flow
var targeting_context
var pointer_position_provider
var input_context
var devtools_context
var runtime_tick_flow
var spectate_context
var gameplay_state_apply_flow
var gameplay_process_flow
var server_hitbox_overlay_flow


func configure(
	connection_service_ref,
	game_owner_ref: Node2D,
	player_ref: Player,
	hud_flow_ref,
	menu_flow_ref,
	runtime_context_ref,
	spectate_menu_state_ref = null,
	input_context_ref = null,
	devtools_context_ref = null,
	gameplay_state_apply_flow_ref = null,
	gameplay_process_flow_ref = null
) -> void:
	event_lifecycle_flow = GameplayEventLifecycleFlowScript.new()
	event_lifecycle_flow.configure(
		game_owner_ref,
		hud_flow_ref.hud if hud_flow_ref != null else null,
		hud_flow_ref,
		menu_flow_ref,
		player_ref,
		Callable(runtime_context_ref.world_sync, "visual_position_for_server_position")
	)

	alive_restore_flow = GameplayAliveRestoreFlowScript.new()
	alive_restore_flow.configure(
		runtime_context_ref.world_sync,
		runtime_context_ref.respawn_flow,
		hud_flow_ref,
		menu_flow_ref,
		player_ref
	)

	pointer_position_provider = GameplayPointerPositionProvider.new()
	pointer_position_provider.configure(
		game_owner_ref,
		Callable(runtime_context_ref.world_sync, "server_position_for_visual_position")
	)

	targeting_context = GameplayTargetingContextScript.new()
	targeting_context.configure(
		connection_service_ref,
		runtime_context_ref.world_sync.target_source(),
		Callable(pointer_position_provider, "mouse_visual_position"),
		Callable(pointer_position_provider, "server_position_for_visual_position")
	)

	devtools_context = devtools_context_ref
	if devtools_context == null:
		devtools_context = GameplayDevtoolsContext.new()
		devtools_context.configure(connection_service_ref)

	input_context = input_context_ref
	if input_context == null:
		input_context = GameplayInputContext.new()
		input_context.configure(
			connection_service_ref,
			player_ref,
			menu_flow_ref,
			game_owner_ref,
			devtools_context,
			Callable(runtime_context_ref, "request_respawn"),
			targeting_context,
			Callable(runtime_context_ref.world_sync, "remote_player_nodes")
		)

	gameplay_state_apply_flow = gameplay_state_apply_flow_ref
	if gameplay_state_apply_flow == null:
		gameplay_state_apply_flow = GameplayStateApplyFlow.new()
		gameplay_state_apply_flow.configure(
			input_context,
			devtools_context,
			hud_flow_ref,
			runtime_context_ref.world_sync,
			event_lifecycle_flow,
			alive_restore_flow
		)

	server_hitbox_overlay_flow = ServerHitboxOverlayFlowScript.new()
	server_hitbox_overlay_flow.configure(game_owner_ref, runtime_context_ref.world_sync)

	runtime_tick_flow = GameplayRuntimeTickFlow.new()
	runtime_tick_flow.configure(hud_flow_ref)

	spectate_context = GameplaySpectateContext.new()
	spectate_context.configure(menu_flow_ref, null, runtime_context_ref.world_sync)
	if spectate_menu_state_ref != null:
		spectate_context.configure_menu_state(spectate_menu_state_ref)

	input_context.configure_spectate_routes(
		Callable(spectate_context, "request_open_spectate_menu"),
		Callable(spectate_context, "request_cycle_target")
	)

	gameplay_process_flow = gameplay_process_flow_ref
	if gameplay_process_flow == null:
		gameplay_process_flow = GameplayProcessFlowScript.new()
		gameplay_process_flow.configure(
			runtime_context_ref,
			server_hitbox_overlay_flow,
			runtime_tick_flow,
			devtools_context,
			input_context,
			spectate_context
		)


func apply_gameplay_state(state: Dictionary, has_received_state: bool) -> GameplayStateApplyResult:
	if gameplay_state_apply_flow == null:
		return GameplayStateApplyResult.new()

	var result: GameplayStateApplyResult = gameplay_state_apply_flow.apply_state(state, has_received_state)
	if server_hitbox_overlay_flow != null:
		server_hitbox_overlay_flow.apply_gameplay_state(state)
	return result


func apply_devtools_gameplay_state(state: Dictionary) -> void:
	if devtools_context == null:
		return
	devtools_context.apply_gameplay_state(state)


func apply_devtools_debug_status_packet(packet: Dictionary) -> void:
	if devtools_context == null:
		return
	devtools_context.apply_debug_status_packet(packet)


func apply_debug_shape_catalog_packet(packet: Dictionary) -> void:
	if server_hitbox_overlay_flow == null:
		return
	server_hitbox_overlay_flow.apply_debug_shape_catalog_packet(packet)


func handle_unhandled_input(event: InputEvent, has_received_state: bool) -> bool:
	if input_context == null:
		return false
	return input_context.handle_unhandled_input(event, has_received_state)


func configure_placement_request_route(route: Callable) -> void:
	if devtools_context == null:
		return
	devtools_context.configure_placement_request_route(route)


func handle_placement_result(result: Dictionary) -> void:
	if devtools_context == null:
		return
	devtools_context.handle_placement_result(result)


func refresh_devtools_spawn_player_slots(max_players: int) -> void:
	if devtools_context == null:
		return
	devtools_context.refresh_spawn_player_slots(max_players)


func process(delta: float, has_received_state: bool) -> void:
	if gameplay_process_flow == null:
		return
	gameplay_process_flow.process(delta, has_received_state)


func reset() -> void:
	if input_context != null:
		input_context.reset()
	if event_lifecycle_flow != null:
		event_lifecycle_flow.reset()
	if alive_restore_flow != null:
		alive_restore_flow.reset()
	if runtime_tick_flow != null:
		runtime_tick_flow.reset()
	if spectate_context != null:
		spectate_context.reset()
	if server_hitbox_overlay_flow != null:
		server_hitbox_overlay_flow.reset()
