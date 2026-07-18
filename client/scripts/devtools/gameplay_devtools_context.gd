extends RefCounted
class_name GameplayDevtoolsContext

const ClientLogger := preload("res://scripts/logging/logger.gd")
const DevConnectionService := preload("res://scripts/devtools/dev_connection_service.gd")
const DevtoolsDisplayRefreshFlow := preload("res://scripts/devtools/devtools_display_refresh_flow.gd")
const MeasurementOverlaySnapshot := preload("res://scripts/devtools/measurement/client_measurement_overlay_snapshot.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

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
var measurement_coordinator
var measurement_overlay_snapshot
var _devtools_enabled_event_emitted := false


func configure(connection_service_ref, operation_trace_factory: Callable = Callable()) -> void:
	dev_connection_service = DevConnectionService.new()
	dev_connection_service.configure(connection_service_ref, operation_trace_factory)
	debug_flow = GameplayDebugFlow.new()
	debug_flow.configure(connection_service_ref, operation_trace_factory)
	devtools_window_controller = DevtoolsWindowController.new()
	display_refresh_flow = DevtoolsDisplayRefreshFlow.new()
	display_refresh_flow.configure(devtools_window_controller)
	state_context = DevtoolsStateContext.new()
	command_context = DevtoolsCommandContext.new()
	command_context.configure(debug_flow, state_context, operation_trace_factory)
	command_context.configure_connection(connection_service_ref)
	command_context.configure_dev_connection(dev_connection_service)
	devtools_window_controller.configure_kill_player_request_route(
		Callable(command_context, "request_kill_player")
	)
	overlay_context = DevtoolsOverlayContext.new()
	overlay_context.configure(state_context, connection_service_ref)
	gameplay_state_context = DevtoolsGameplayStateContext.new()
	gameplay_state_context.configure(connection_service_ref, devtools_window_controller, display_refresh_flow, state_context, overlay_context)
	placement_context = DevtoolsPlacementContext.new()
	placement_context.configure(state_context, dev_connection_service, operation_trace_factory)
	var hotkey_flow := DevtoolsHotkeyFlow.new()
	hotkey_flow.configure(
		Callable(command_context, "request_respawn_local_player"),
		Callable(placement_context, "request_placement_action")
	)
	hotkey_context = DevtoolsHotkeyContext.new()
	hotkey_context.configure(
		state_context,
		overlay_context,
		hotkey_flow,
		Callable(self, "toggle_devtools_window"),
		Callable(self, "toggle_measurement")
	)
	window_action_context = DevtoolsWindowActionContext.new()
	window_action_context.configure(devtools_window_controller, command_context, placement_context, overlay_context)
	window_action_context.connect_signals()
	if (
		connection_service_ref != null
		and dev_connection_service != null
		and debug_flow != null
		and command_context != null
		and placement_context != null
		and window_action_context != null
		and !_devtools_enabled_event_emitted
	):
		_devtools_enabled_event_emitted = true
		ClientLogger.emit_canonical(ObservabilityContract.EVENT_DEVTOOLS_ENABLED)


func configure_remote_player_nodes_provider(provider: Callable) -> void:
	if overlay_context != null:
		overlay_context.configure_remote_player_nodes_provider(provider)


func get_world_telemetry_context():
	if overlay_context == null or not overlay_context.has_method("get_world_telemetry_context"):
		return null
	return overlay_context.get_world_telemetry_context()


func reset() -> void:
	if debug_flow != null:
		debug_flow.reset()
	if display_refresh_flow != null:
		display_refresh_flow.reset()
	if overlay_context != null:
		overlay_context.reset()
	if measurement_overlay_snapshot != null:
		measurement_overlay_snapshot.reset()
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


func configure_measurement(coordinator_ref, client_measurement_context_ref = null) -> void:
	measurement_coordinator = coordinator_ref
	measurement_overlay_snapshot = MeasurementOverlaySnapshot.new()
	measurement_overlay_snapshot.configure(measurement_coordinator, client_measurement_context_ref)
	if overlay_context != null and overlay_context.has_method("configure_measurement_snapshot_provider"):
		overlay_context.configure_measurement_snapshot_provider(
			Callable(measurement_overlay_snapshot, "snapshot")
		)
	if devtools_window_controller != null:
		var start_route := Callable(self, "_on_measurement_start_requested")
		if devtools_window_controller.has_signal("measurement_start_requested") and !devtools_window_controller.is_connected("measurement_start_requested", start_route):
			devtools_window_controller.connect("measurement_start_requested", start_route)
		var stop_route := Callable(self, "_on_measurement_stop_requested")
		if devtools_window_controller.has_signal("measurement_stop_requested") and !devtools_window_controller.is_connected("measurement_stop_requested", stop_route):
			devtools_window_controller.connect("measurement_stop_requested", stop_route)
		var reset_route := Callable(self, "_on_measurement_reset_requested")
		if devtools_window_controller.has_signal("measurement_reset_requested") and !devtools_window_controller.is_connected("measurement_reset_requested", reset_route):
			devtools_window_controller.connect("measurement_reset_requested", reset_route)
	if measurement_coordinator != null:
		var state_route := Callable(self, "_on_measurement_state_changed")
		if measurement_coordinator.has_signal("state_changed") and !measurement_coordinator.is_connected("state_changed", state_route):
			measurement_coordinator.connect("state_changed", state_route)
		var error_route := Callable(self, "_on_measurement_error_received")
		if measurement_coordinator.has_signal("error_received") and !measurement_coordinator.is_connected("error_received", error_route):
			measurement_coordinator.connect("error_received", error_route)
		_refresh_measurement_state()


func toggle_measurement() -> void:
	if measurement_coordinator == null:
		return
	var state: Dictionary = measurement_coordinator.get_state()
	var pending_request_ids: Dictionary = state.get("pending_request_ids", {})
	if str(state.get("active_run_id", "")) != "":
		stop_measurement()
	elif pending_request_ids.has("start"):
		_refresh_measurement_state()
	else:
		start_measurement()


func start_measurement(scenario_label: String = "") -> String:
	if measurement_coordinator == null:
		return ""
	var request_id: String = measurement_coordinator.start(scenario_label)
	_refresh_measurement_state()
	return request_id


func stop_measurement() -> String:
	if measurement_coordinator == null:
		return ""
	var request_id: String = measurement_coordinator.stop()
	_refresh_measurement_state()
	return request_id


func reset_measurement() -> String:
	if measurement_coordinator == null:
		return ""
	var request_id: String = measurement_coordinator.reset()
	_refresh_measurement_state()
	return request_id


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


func request_kill_player(target_scope: String = "", target_player_id: String = "") -> void:
	if command_context != null:
		command_context.request_kill_player(target_scope, target_player_id)


func handle_placement_result(result: Dictionary) -> void:
	if self.placement_context != null:
		self.placement_context.handle_placement_result(result)


func _on_measurement_start_requested(scenario_label: String) -> void:
	start_measurement(scenario_label)


func _on_measurement_stop_requested() -> void:
	stop_measurement()


func _on_measurement_reset_requested() -> void:
	reset_measurement()


func _on_measurement_state_changed(_state: Dictionary) -> void:
	_refresh_measurement_state()


func _on_measurement_error_received(_error: Dictionary) -> void:
	_refresh_measurement_state()


func _refresh_measurement_state() -> void:
	if devtools_window_controller == null or measurement_coordinator == null:
		return
	devtools_window_controller.refresh_measurement_state(measurement_coordinator.get_state())
