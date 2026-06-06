extends RefCounted
class_name DevtoolsGameplayStateContext

const DebugStatusPacketReader = preload("res://scripts/devtools/debug_status_packet_reader.gd")

var connection_service
var devtools_window_controller
var display_refresh_flow
var state_context
var overlay_context


func configure(connection_service_ref, devtools_window_controller_ref, display_refresh_flow_ref, state_context_ref, overlay_context_ref) -> void:
	connection_service = connection_service_ref
	devtools_window_controller = devtools_window_controller_ref
	display_refresh_flow = display_refresh_flow_ref
	state_context = state_context_ref
	overlay_context = overlay_context_ref


func apply_debug_status(status: Dictionary) -> void:
	if devtools_window_controller != null:
		devtools_window_controller.apply_debug_status(status)


func apply_debug_status_packet(packet: Dictionary) -> void:
	var state = DebugStatusPacketReader.read(packet)
	if display_refresh_flow != null:
		display_refresh_flow.apply_debug_status_packet(state)
		return

	if devtools_window_controller != null:
		devtools_window_controller.apply_debug_status(state.get("debug_status", {}))


func apply_gameplay_state(state: Dictionary) -> void:
	if display_refresh_flow != null:
		display_refresh_flow.refresh_gameplay_state(state)
		if state_context != null:
			state_context.set_local_player_id(display_refresh_flow.local_player_id())
			state_context.set_game_target(display_refresh_flow.game_target_kind(), display_refresh_flow.game_target_id())
	if devtools_window_controller != null:
		devtools_window_controller.configure_kill_player_routing(
			connection_service,
			state_context.get_local_player_id() if state_context != null else "",
			state_context.get_game_target_kind() if state_context != null else "",
			state_context.get_game_target_id() if state_context != null else ""
		)
	if overlay_context != null:
		overlay_context.apply_gameplay_state(state)


func refresh_spawn_player_slots(max_players: int) -> void:
	if display_refresh_flow == null:
		return
	display_refresh_flow.refresh_spawn_player_slots(max_players)
