extends RefCounted
class_name GameplayDevtoolsContext

const DevConnectionService := preload("res://scripts/devtools/dev_connection_service.gd")
const DevtoolsHotkeyFlow := preload("res://scripts/devtools/devtools_hotkey_flow.gd")
const DevtoolsPlayerTargetModel := preload("res://scripts/devtools/devtools_player_target_model.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

var debug_flow
var devtools_window_controller
var dev_connection_service
var connection_service
var hotkey_flow
var has_received_gameplay_state := false
var placement_request_route: Callable
var local_player_id := ""
var target_model


func configure(connection_service_ref) -> void:
	connection_service = connection_service_ref
	dev_connection_service = DevConnectionService.new()
	dev_connection_service.configure(connection_service_ref)
	debug_flow = GameplayDebugFlow.new()
	debug_flow.configure(connection_service_ref)
	hotkey_flow = DevtoolsHotkeyFlow.new()
	hotkey_flow.configure(
		Callable(self, "request_respawn_local_player"),
		Callable(self, "request_placement_action")
	)
	devtools_window_controller = DevtoolsWindowController.new()
	target_model = DevtoolsPlayerTargetModel.new()
	_connect_window_controller_signals()


func reset() -> void:
	if debug_flow != null:
		debug_flow.reset()


func process(has_received_state: bool) -> void:
	has_received_gameplay_state = has_received_state
	if Input.is_action_just_pressed("DevToggle0"):
		toggle_devtools_window()
	if hotkey_flow != null:
		hotkey_flow.process(has_received_state)
	if debug_flow != null:
		debug_flow.process(has_received_state)


func toggle_devtools_window() -> void:
	if devtools_window_controller != null:
		devtools_window_controller.toggle_window()


func apply_debug_status(status: Dictionary) -> void:
	if devtools_window_controller != null:
		devtools_window_controller.apply_debug_status(status)


func apply_gameplay_state(state: Dictionary) -> void:
	apply_debug_status(state.get("debug_status", {}))
	if target_model == null:
		return

	target_model.apply_gameplay_state(state)
	if devtools_window_controller != null:
		devtools_window_controller.configure_kill_player_routing(connection_service, target_model.self_id)
		devtools_window_controller.refresh_debug_player_targets(
			target_model.target_rows(),
			target_model.invincible_target_rows(),
			target_model.infinite_lives_target_rows(),
			target_model.player_frozen_target_rows()
		)


func _connect_window_controller_signals() -> void:
	if !devtools_window_controller.toggle_invincible_requested.is_connected(request_toggle_invincible):
		devtools_window_controller.toggle_invincible_requested.connect(request_toggle_invincible)
	if !devtools_window_controller.toggle_infinite_lives_requested.is_connected(request_toggle_infinite_lives):
		devtools_window_controller.toggle_infinite_lives_requested.connect(request_toggle_infinite_lives)
	if !devtools_window_controller.toggle_freeze_world_requested.is_connected(request_toggle_freeze_world):
		devtools_window_controller.toggle_freeze_world_requested.connect(request_toggle_freeze_world)
	if !devtools_window_controller.toggle_freeze_player_requested.is_connected(request_toggle_freeze_player):
		devtools_window_controller.toggle_freeze_player_requested.connect(request_toggle_freeze_player)
	if !devtools_window_controller.placement_action_requested.is_connected(request_placement_action):
		devtools_window_controller.placement_action_requested.connect(request_placement_action)
	if !devtools_window_controller.respawn_player_requested.is_connected(request_respawn_player):
		devtools_window_controller.respawn_player_requested.connect(request_respawn_player)


func request_toggle_invincible(target_player_id: String = "") -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_invincible(target_player_id)


func request_toggle_infinite_lives(target_player_id: String = "") -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_infinite_lives(target_player_id)


func request_toggle_freeze_world() -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_freeze_world()


func request_toggle_freeze_player(target_player_id: String = "") -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_freeze_player(target_player_id)


func configure_local_player_id(player_id: String) -> void:
	local_player_id = player_id


func request_respawn_player(target_player_id: String) -> void:
	if target_player_id == "":
		ClientLogger.game_warn("GameplayDevtoolsContext: respawn request ignored, target_player_id is empty")
		return
	if !has_received_gameplay_state:
		return
	if dev_connection_service == null || !dev_connection_service.is_configured():
		ClientLogger.game_warn("GameplayDevtoolsContext: respawn request ignored, dev_connection_service is unavailable")
		return
	dev_connection_service.send_respawn_player(target_player_id)


func request_respawn_local_player() -> void:
	if local_player_id == "":
		ClientLogger.game_warn("GameplayDevtoolsContext: local respawn request ignored, local_player_id is empty")
		return
	request_respawn_player(local_player_id)


func configure_placement_request_route(route: Callable) -> void:
	placement_request_route = route


func request_placement_action(action_name: StringName, placement_context: Dictionary = {}) -> void:
	if !has_received_gameplay_state:
		return
	if placement_request_route.is_null():
		return
	placement_request_route.call(action_name, placement_context)


func handle_placement_result(result: Dictionary) -> void:
	if result.is_empty():
		return
	var action_name := StringName(result.get("action_name", StringName()))
	if action_name.is_empty():
		return
	if dev_connection_service == null || !dev_connection_service.is_configured():
		return
	dev_connection_service.send_spawn_from_placement_result(result)
