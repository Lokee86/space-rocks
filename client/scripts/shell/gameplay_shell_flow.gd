extends RefCounted
class_name GameplayShellFlow

const GameplayPauseStateFlowScript = preload("res://scripts/gameplay/state/gameplay_pause_state_flow.gd")
const GameplayFlowComposerScript = preload("res://scripts/gameplay/runtime/gameplay_flow_composer.gd")

signal gameplay_started
signal quit_to_main_menu_requested
signal return_to_lobby_requested

var runtime_context
var flow_composer
var gameplay_pause_state_flow
var hud_flow
var menu_flow
var has_received_state := false


func configure(
	connection_service_ref,
	game_owner_ref: Node2D,
	player_ref: Player,
	view_anchor_ref,
	bullets: Node2D,
	asteroids: Node2D,
	pickups: Node2D,
	hud_flow_ref,
	menu_flow_ref,
	spectate_menu_state_ref = null
) -> void:
	hud_flow = hud_flow_ref
	menu_flow = menu_flow_ref
	gameplay_pause_state_flow = GameplayPauseStateFlowScript.new()
	if menu_flow != null:
		menu_flow.configure_lifecycle_routes(
			Callable(self, "_on_quit_to_main_menu_requested"),
			Callable(self, "_on_return_to_lobby_requested")
	)
	runtime_context = GameplayRuntimeContext.new()
	runtime_context.configure_world(game_owner_ref, player_ref, view_anchor_ref, bullets, asteroids, pickups, gameplay_pause_state_flow.tracker())

	runtime_context.configure_respawn(connection_service_ref, hud_flow)

	flow_composer = GameplayFlowComposerScript.new()
	flow_composer.configure(
		connection_service_ref,
		game_owner_ref,
		player_ref,
		hud_flow,
		menu_flow,
		runtime_context,
		spectate_menu_state_ref
	)


func reset() -> void:
	has_received_state = false
	if runtime_context != null:
		runtime_context.reset()
	if hud_flow != null:
		hud_flow.reset()
	if menu_flow != null:
		menu_flow.reset()
	if gameplay_pause_state_flow != null:
		gameplay_pause_state_flow.reset()
	if flow_composer != null:
		flow_composer.reset()


func apply_gameplay_state(state: Dictionary) -> void:
	if flow_composer == null:
		return
	var result: GameplayStateApplyResult = flow_composer.apply_gameplay_state(state, has_received_state)
	has_received_state = result.has_received_state
	if result.started_gameplay:
		gameplay_started.emit()


func apply_player_pause_state_packet(packet: Dictionary) -> void:
	if gameplay_pause_state_flow == null:
		return
	gameplay_pause_state_flow.apply_packet(packet)


func apply_devtools_debug_status_packet(packet: Dictionary) -> void:
	if flow_composer == null:
		return
	flow_composer.apply_devtools_debug_status_packet(packet)


func handle_unhandled_input(event: InputEvent) -> bool:
	if flow_composer == null:
		return false
	return flow_composer.handle_unhandled_input(event, has_received_state)


func apply_devtools_gameplay_state(state: Dictionary) -> void:
	if flow_composer == null:
		return
	flow_composer.apply_devtools_gameplay_state(state)


func configure_devtools_placement_request_route(route: Callable) -> void:
	if flow_composer == null:
		return
	flow_composer.configure_placement_request_route(route)


func handle_devtools_placement_result(result: Dictionary) -> void:
	if flow_composer == null:
		return
	flow_composer.handle_placement_result(result)


func refresh_devtools_spawn_player_slots(max_players: int) -> void:
	if flow_composer == null:
		return
	flow_composer.refresh_devtools_spawn_player_slots(max_players)


func process(_delta: float) -> void:
	if flow_composer == null:
		return
	flow_composer.process(_delta, has_received_state)


func _on_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()


func _on_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()
