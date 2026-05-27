extends RefCounted
class_name GameplayInputContext

const GameplayInputFlow = preload("res://scripts/gameplay/input/gameplay_input_flow.gd")
const GameplayPauseInputFlow = preload("res://scripts/gameplay/input/gameplay_pause_input_flow.gd")
const GameplayDevtoolsContext = preload("res://scripts/devtools/gameplay_devtools_context.gd")

var input_flow
var pause_input_flow
var devtools_context
var respawn_request_route: Callable
var open_spectate_menu_route: Callable
var cycle_spectate_target_route: Callable


func configure(
	connection_service_ref,
	player_ref,
	menu_flow_ref,
	respawn_request_route_ref: Callable
) -> void:
	input_flow = GameplayInputFlow.new()
	input_flow.configure(connection_service_ref, player_ref, menu_flow_ref)
	pause_input_flow = GameplayPauseInputFlow.new()
	pause_input_flow.configure(menu_flow_ref)
	devtools_context = GameplayDevtoolsContext.new()
	devtools_context.configure(connection_service_ref)
	respawn_request_route = respawn_request_route_ref


func configure_spectate_routes(
	open_spectate_menu_route_ref: Callable,
	cycle_spectate_target_route_ref: Callable
) -> void:
	open_spectate_menu_route = open_spectate_menu_route_ref
	cycle_spectate_target_route = cycle_spectate_target_route_ref


func reset() -> void:
	if input_flow != null:
		input_flow.reset()
	if pause_input_flow != null:
		pause_input_flow.reset()
	if devtools_context != null:
		devtools_context.reset()


func mark_gameplay_state_received() -> void:
	if input_flow != null:
		input_flow.mark_gameplay_state_received()


func process(has_received_state: bool) -> void:
	if pause_input_flow != null:
		pause_input_flow.process(has_received_state)
	if devtools_context != null:
		devtools_context.process(has_received_state)
	if Input.is_action_just_pressed("Respawn") && !respawn_request_route.is_null():
		respawn_request_route.call(has_received_state)
	if input_flow != null:
		input_flow.process()
	if Input.is_action_just_pressed("OpenMenu") && !open_spectate_menu_route.is_null():
		open_spectate_menu_route.call()
	if Input.is_action_just_pressed("SwitchCamera") && !cycle_spectate_target_route.is_null():
		cycle_spectate_target_route.call()
