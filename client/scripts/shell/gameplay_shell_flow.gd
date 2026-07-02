extends RefCounted
class_name GameplayShellFlow

const GameplayPauseStateFlowScript = preload("res://scripts/gameplay/state/gameplay_pause_state_flow.gd")
const GameplayFlowComposerScript = preload("res://scripts/gameplay/runtime/gameplay_flow_composer.gd")

signal gameplay_started
signal quit_to_main_menu_requested
signal return_to_pregame_requested(session_mode: String)
signal return_to_lobby_requested

var runtime_context
var flow_composer
var gameplay_pause_state_flow
var hud_flow
var menu_flow
var match_end_flow
var has_received_lane_baselines_synced := false


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
	spectate_menu_state_ref = null,
	match_end_flow_ref = null
) -> void:
	hud_flow = hud_flow_ref
	menu_flow = menu_flow_ref
	match_end_flow = match_end_flow_ref
	gameplay_pause_state_flow = GameplayPauseStateFlowScript.new()
	if menu_flow != null:
		menu_flow.configure_lifecycle_routes(
			Callable(self, "_on_quit_to_main_menu_requested"),
			Callable(self, "_on_return_to_lobby_requested")
		)
		var return_to_pregame_callable := Callable(self, "_on_return_to_pregame_requested")
		if menu_flow.has_signal("return_to_pregame_requested") && !menu_flow.return_to_pregame_requested.is_connected(return_to_pregame_callable):
			menu_flow.return_to_pregame_requested.connect(return_to_pregame_callable)
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
		spectate_menu_state_ref,
		null,
		null,
		null,
		match_end_flow
	)


func reset() -> void:
	has_received_lane_baselines_synced = false
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


func set_required_lane_baselines_synced(value: bool) -> void:
	has_received_lane_baselines_synced = value

func get_event_lifecycle_flow():
	if flow_composer == null:
		return null
	return flow_composer.get_event_lifecycle_flow()

func apply_player_pause_state_packet(packet: Dictionary) -> void:
	if gameplay_pause_state_flow == null:
		return
	gameplay_pause_state_flow.apply_packet(packet)


func apply_devtools_debug_status_packet(packet: Dictionary) -> void:
	if flow_composer == null:
		return
	flow_composer.apply_devtools_debug_status_packet(packet)


func apply_debug_shape_catalog_packet(packet: Dictionary) -> void:
	if flow_composer == null:
		return
	flow_composer.apply_debug_shape_catalog_packet(packet)


func handle_unhandled_input(event: InputEvent) -> bool:
	if flow_composer == null:
		return false
	return flow_composer.handle_unhandled_input(event, has_received_lane_baselines_synced)


func apply_devtools_gameplay_state(state: Dictionary) -> void:
	if flow_composer == null:
		return
	flow_composer.apply_devtools_gameplay_state(state)


func restore_alive_presentation_from_lane_state(world_lane_state, session_lane_state, self_id: String) -> void:
	if flow_composer == null:
		return
	flow_composer.restore_alive_presentation_from_lane_state(world_lane_state, session_lane_state, self_id)


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
	flow_composer.process(_delta, has_received_lane_baselines_synced)


func _on_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()


func _on_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()


func _on_return_to_pregame_requested(session_mode: String) -> void:
	return_to_pregame_requested.emit(session_mode)
