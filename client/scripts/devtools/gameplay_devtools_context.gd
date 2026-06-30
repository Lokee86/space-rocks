extends RefCounted
class_name GameplayDevtoolsContext

const DevConnectionService := preload("res://scripts/devtools/dev_connection_service.gd")
const DevtoolsDisplayRefreshFlow := preload("res://scripts/devtools/devtools_display_refresh_flow.gd")

var debug_flow
var command_context
var devtools_window_controller
var gameplay_state_context
var display_refresh_flow
var dev_connection_service
var overlay_context
var hotkey_context
var placement_context
var window_action_context
var state_context


func configure(connection_service_ref) -> void:
	dev_connection_service = DevConnectionService.new()
	dev_connection_service.configure(connection_service_ref)
	debug_flow = GameplayDebugFlow.new()
	debug_flow.configure(connection_service_ref)
	devtools_window_controller = DevtoolsWindowController.new()
	display_refresh_flow = DevtoolsDisplayRefreshFlow.new()
	display_refresh_flow.configure(devtools_window_controller)
	state_context = DevtoolsStateContext.new()
	command_context = DevtoolsCommandContext.new()
	command_context.configure(debug_flow, state_context)
	command_context.configure_connection(connection_service_ref)
	command_context.configure_dev_connection(dev_connection_service)
	overlay_context = DevtoolsOverlayContext.new()
	overlay_context.configure(state_context, connection_service_ref)
	gameplay_state_context = DevtoolsGameplayStateContext.new()
	gameplay_state_context.configure(connection_service_ref, devtools_window_controller, display_refresh_flow, state_context, overlay_context)
	placement_context = DevtoolsPlacementContext.new()
	placement_context.configure(state_context, dev_connection_service)
	var hotkey_flow := DevtoolsHotkeyFlow.new()
	hotkey_flow.configure(
		Callable(command_context, "request_respawn_local_player"),
		Callable(placement_context, "request_placement_action")
	)
	hotkey_context = DevtoolsHotkeyContext.new()
	hotkey_context.configure(state_context, overlay_context, hotkey_flow, Callable(self, "toggle_devtools_window"))
	window_action_context = DevtoolsWindowActionContext.new()
	window_action_context.configure(devtools_window_controller, command_context, placement_context, overlay_context)
	window_action_context.connect_signals()


func configure_remote_player_nodes_provider(provider: Callable) -> void:
	if overlay_context != null:
		overlay_context.configure_remote_player_nodes_provider(provider)


func reset() -> void:
	if debug_flow != null:
		debug_flow.reset()
	if display_refresh_flow != null:
		display_refresh_flow.reset()
	if overlay_context != null:
		overlay_context.reset()
	if state_context != null:
		state_context.reset_game_target()


func process(required_lane_baselines_synced: bool) -> void:
	if state_context != null:
		state_context.set_has_lane_baseline_sync(required_lane_baselines_synced)
	if hotkey_context != null:
		hotkey_context.process(required_lane_baselines_synced)
	if command_context != null:
		command_context.process(required_lane_baselines_synced)
	if overlay_context != null:
		overlay_context.process(required_lane_baselines_synced)


func toggle_devtools_window() -> void:
	if devtools_window_controller != null:
		devtools_window_controller.toggle_window()


func apply_debug_status(status: Dictionary) -> void:
	if gameplay_state_context != null:
		gameplay_state_context.apply_debug_status(status)


func apply_debug_status_packet(packet: Dictionary) -> void:
	if gameplay_state_context != null:
		gameplay_state_context.apply_debug_status_packet(packet)


func apply_gameplay_state(state: Dictionary) -> void:
	if gameplay_state_context != null:
		gameplay_state_context.apply_gameplay_state(state)


func refresh_spawn_player_slots(max_players: int) -> void:
	if gameplay_state_context != null:
		gameplay_state_context.refresh_spawn_player_slots(max_players)


func configure_server_hitbox_overlay(overlay_ref) -> void:
	if overlay_context != null:
		overlay_context.configure_server_hitbox_overlay(overlay_ref)


func configure_local_player_id(player_id: String) -> void:
	if state_context != null:
		state_context.set_local_player_id(player_id)


func configure_placement_request_route(route: Callable) -> void:
	if placement_context != null:
		placement_context.configure_placement_request_route(route)


func configure_local_respawn_confirmation_marker(marker: Callable) -> void:
	if command_context != null:
		command_context.configure_local_respawn_confirmation_marker(marker)


func handle_placement_result(result: Dictionary) -> void:
	if self.placement_context != null:
		self.placement_context.handle_placement_result(result)