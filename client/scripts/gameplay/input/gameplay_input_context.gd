extends RefCounted
class_name GameplayInputContext

const TargetRequestFlow = preload("res://scripts/gameplay/input/target_request_flow.gd")
const MouseActionFlow = preload("res://scripts/gameplay/input/mouse_action_flow.gd")

var input_flow
var pause_input_flow
var devtools_context
var target_request_flow
var mouse_action_flow
var remote_player_nodes_provider: Callable


func apply_debug_status(status: Dictionary) -> void:
	if devtools_context != null:
		devtools_context.apply_debug_status(status)


func apply_gameplay_state(state: Dictionary) -> void:
	if devtools_context != null:
		devtools_context.apply_gameplay_state(state)


func refresh_debug_spawn_player_slots(max_players: int) -> void:
	if devtools_context != null:
		devtools_context.refresh_spawn_player_slots(max_players)
var respawn_request_route: Callable
var open_spectate_menu_route: Callable
var cycle_spectate_target_route: Callable


func configure(
	connection_service_ref,
	player_ref,
	menu_flow_ref,
	game_owner_ref,
	respawn_request_route_ref: Callable,
	target_visual_candidates_provider_ref: Callable = Callable(),
	mouse_visual_position_provider_ref: Callable = Callable(),
	server_position_converter_ref: Callable = Callable(),
	remote_player_nodes_provider_ref: Callable = Callable()
) -> void:
	input_flow = GameplayInputFlow.new()
	input_flow.configure(connection_service_ref, player_ref, menu_flow_ref)
	pause_input_flow = GameplayPauseInputFlow.new()
	pause_input_flow.configure(menu_flow_ref)
	devtools_context = GameplayDevtoolsContext.new()
	devtools_context.configure(connection_service_ref)
	remote_player_nodes_provider = remote_player_nodes_provider_ref
	if devtools_context.has_method("configure_remote_player_nodes_provider"):
		devtools_context.configure_remote_player_nodes_provider(remote_player_nodes_provider)
	if game_owner_ref != null && devtools_context.has_method("configure_server_hitbox_overlay"):
		devtools_context.configure_server_hitbox_overlay(game_owner_ref.get_node_or_null("ServerHitboxOverlay"))
	target_request_flow = TargetRequestFlow.new()
	target_request_flow.configure(
		connection_service_ref,
		target_visual_candidates_provider_ref,
		mouse_visual_position_provider_ref,
		server_position_converter_ref
	)
	mouse_action_flow = MouseActionFlow.new()
	mouse_action_flow.configure(target_request_flow)
	respawn_request_route = respawn_request_route_ref


func configure_spectate_routes(
	open_spectate_menu_route_ref: Callable,
	cycle_spectate_target_route_ref: Callable
) -> void:
	open_spectate_menu_route = open_spectate_menu_route_ref
	cycle_spectate_target_route = cycle_spectate_target_route_ref


func configure_debug_placement_route(route: Callable) -> void:
	if devtools_context != null:
		devtools_context.configure_placement_request_route(route)


func handle_debug_placement_result(result: Dictionary) -> void:
	if devtools_context != null:
		devtools_context.handle_placement_result(result)


func reset() -> void:
	if input_flow != null:
		input_flow.reset()
	if pause_input_flow != null:
		pause_input_flow.reset()
	if devtools_context != null:
		devtools_context.reset()
	if mouse_action_flow != null:
		mouse_action_flow.clear_pending_context()


func mark_gameplay_state_received() -> void:
	if input_flow != null:
		input_flow.mark_gameplay_state_received()

func handle_unhandled_input(event: InputEvent, has_received_state: bool) -> bool:
	if !has_received_state:
		return false
	if mouse_action_flow == null:
		return false
	return mouse_action_flow.handle_input_event(event)


func process(has_received_state: bool) -> void:
	var open_menu_consumed := false
	if pause_input_flow != null:
		open_menu_consumed = pause_input_flow.process(has_received_state)
	if devtools_context != null:
		devtools_context.process(has_received_state)
	if Input.is_action_just_pressed("Respawn") && !respawn_request_route.is_null():
		respawn_request_route.call(has_received_state)
	if input_flow != null:
		input_flow.process()
	if !open_menu_consumed && Input.is_action_just_pressed("OpenMenu") && !open_spectate_menu_route.is_null():
		open_spectate_menu_route.call()
	if Input.is_action_just_pressed("SwitchCamera") && !cycle_spectate_target_route.is_null():
		cycle_spectate_target_route.call()
