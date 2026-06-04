extends RefCounted
class_name GameplayDevtoolsContext

const DevConnectionService := preload("res://scripts/devtools/dev_connection_service.gd")
const DevtoolsCommandContext := preload("res://scripts/devtools/context/devtools_command_context.gd")
const DevtoolsDisplayRefreshFlow := preload("res://scripts/devtools/devtools_display_refresh_flow.gd")
const DevtoolsHotkeyContext := preload("res://scripts/devtools/context/devtools_hotkey_context.gd")
const DevtoolsGameplayStateContext := preload("res://scripts/devtools/context/devtools_gameplay_state_context.gd")
const DevtoolsOverlayContext := preload("res://scripts/devtools/context/devtools_overlay_context.gd")
const DevtoolsPlacementContext := preload("res://scripts/devtools/context/devtools_placement_context.gd")
const DevtoolsWindowActionContext := preload("res://scripts/devtools/context/devtools_window_action_context.gd")
const DevtoolsStateContext := preload("res://scripts/devtools/context/devtools_state_context.gd")

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
var connection_service
var hotkey_flow


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
	display_refresh_flow = DevtoolsDisplayRefreshFlow.new()
	display_refresh_flow.configure(devtools_window_controller)
	state_context = DevtoolsStateContext.new()
	command_context = DevtoolsCommandContext.new()
	command_context.configure(debug_flow, state_context)
	command_context.configure_connection(connection_service_ref)
	overlay_context = DevtoolsOverlayContext.new()
	overlay_context.configure(state_context, connection_service_ref)
	gameplay_state_context = DevtoolsGameplayStateContext.new()
	gameplay_state_context.configure(connection_service_ref, devtools_window_controller, display_refresh_flow, state_context, overlay_context)
	hotkey_context = DevtoolsHotkeyContext.new()
	hotkey_context.configure(state_context, overlay_context, hotkey_flow, Callable(self, "toggle_devtools_window"))
	window_action_context = DevtoolsWindowActionContext.new()
	window_action_context.configure(devtools_window_controller, self)
	window_action_context.connect_signals()
	placement_context = DevtoolsPlacementContext.new()
	placement_context.configure(state_context, dev_connection_service)


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


func process(has_received_state: bool) -> void:
	if state_context != null:
		state_context.set_has_received_gameplay_state(has_received_state)
	if hotkey_context != null:
		hotkey_context.process(has_received_state)
	if command_context != null:
		command_context.process(has_received_state)
	if overlay_context != null:
		overlay_context.process(has_received_state)


func toggle_devtools_window() -> void:
	if devtools_window_controller != null:
		devtools_window_controller.toggle_window()


func apply_debug_status(status: Dictionary) -> void:
	if gameplay_state_context != null:
		gameplay_state_context.apply_debug_status(status)


func apply_gameplay_state(state: Dictionary) -> void:
	if gameplay_state_context != null:
		gameplay_state_context.apply_gameplay_state(state)


func refresh_spawn_player_slots(max_players: int) -> void:
	if gameplay_state_context != null:
		gameplay_state_context.refresh_spawn_player_slots(max_players)


func configure_server_hitbox_overlay(overlay_ref) -> void:
	if overlay_context != null:
		overlay_context.configure_server_hitbox_overlay(overlay_ref)


func request_toggle_invincible(target_scope: String = "", target_player_id: String = "") -> void:
	if command_context != null:
		command_context.request_toggle_invincible(target_scope, target_player_id)


func request_toggle_infinite_lives(target_scope: String = "", target_player_id: String = "") -> void:
	if command_context != null:
		command_context.request_toggle_infinite_lives(target_scope, target_player_id)


func request_toggle_freeze_world(freeze_target: String = "") -> void:
	if command_context != null:
		command_context.request_toggle_freeze_world(freeze_target)


func request_toggle_freeze_player(target_scope: String = "", target_player_id: String = "") -> void:
	if command_context != null:
		command_context.request_toggle_freeze_player(target_scope, target_player_id)


func _get_player_dev_labels_context():
	return overlay_context.get_player_dev_labels_context() if overlay_context != null else null


func _get_world_telemetry_context():
	return overlay_context.get_world_telemetry_context() if overlay_context != null else null


func _get_server_hitbox_overlay():
	return overlay_context.get_server_hitbox_overlay() if overlay_context != null else null


func request_set_score(target_scope: String, target_player_id: String, score: int) -> void:
	if command_context != null:
		command_context.request_set_score(target_scope, target_player_id, score)


func request_add_score(target_scope: String, target_player_id: String, amount: int) -> void:
	if command_context != null:
		command_context.request_add_score(target_scope, target_player_id, amount)


func request_set_lives(target_scope: String, target_player_id: String, lives: int) -> void:
	if command_context != null:
		command_context.request_set_lives(target_scope, target_player_id, lives)


func request_add_lives(target_scope: String, target_player_id: String, amount: int) -> void:
	if command_context != null:
		command_context.request_add_lives(target_scope, target_player_id, amount)


func request_clear_bullets() -> void:
	if command_context != null:
		command_context.request_clear_bullets()


func request_clear_asteroids() -> void:
	if command_context != null:
		command_context.request_clear_asteroids()


func _on_show_server_hitboxes_changed(enabled: bool) -> void:
	if overlay_context != null:
		overlay_context.set_server_hitboxes_enabled(enabled)


func configure_local_player_id(player_id: String) -> void:
	if state_context != null:
		state_context.set_local_player_id(player_id)


func request_set_game_target(target_player_id: String) -> void:
	if command_context != null:
		command_context.request_set_game_target(target_player_id)


func request_clear_game_target() -> void:
	if command_context != null:
		command_context.request_clear_game_target()


func request_respawn_player(target_scope: String = DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, target_player_id: String = "") -> void:
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		ClientLogger.game_warn("GameplayDevtoolsContext: respawn request ignored, target_player_id is empty")
		return
	if state_context == null or !state_context.has_gameplay_state():
		return
	if dev_connection_service == null || !dev_connection_service.is_configured():
		ClientLogger.game_warn("GameplayDevtoolsContext: respawn request ignored, dev_connection_service is unavailable")
		return
	dev_connection_service.send_respawn_player(target_scope, target_player_id)


func request_respawn_local_player() -> void:
	if state_context == null or state_context.get_local_player_id() == "":
		ClientLogger.game_warn("GameplayDevtoolsContext: local respawn request ignored, local_player_id is empty")
		return
	request_respawn_player(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, state_context.get_local_player_id())


func configure_placement_request_route(route: Callable) -> void:
	if placement_context != null:
		placement_context.configure_placement_request_route(route)


func request_placement_action(action_name: StringName, placement_context: Dictionary = {}) -> void:
	if self.placement_context != null:
		self.placement_context.request_placement_action(action_name, placement_context)


func handle_placement_result(result: Dictionary) -> void:
	if self.placement_context != null:
		self.placement_context.handle_placement_result(result)
