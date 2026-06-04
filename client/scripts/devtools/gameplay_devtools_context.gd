extends RefCounted
class_name GameplayDevtoolsContext

const DevConnectionService := preload("res://scripts/devtools/dev_connection_service.gd")
const DevtoolsDisplayRefreshFlow := preload("res://scripts/devtools/devtools_display_refresh_flow.gd")
const PlayerDevLabelsContext := preload("res://scripts/devtools/player_labels/player_dev_labels_context.gd")
const WorldTelemetryContext := preload("res://scripts/devtools/telemetry/world_telemetry_context.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

var debug_flow
var devtools_window_controller
var display_refresh_flow
var dev_connection_service
var player_dev_labels_context
var world_telemetry_context
var connection_service
var hotkey_flow
var server_hitbox_overlay
var has_received_gameplay_state := false
var placement_request_route: Callable
var remote_player_nodes_provider: Callable
var player_dev_label_mode := ""
var local_player_id := ""
var game_target_kind := ""
var game_target_id := ""
var game_target_player_id := ""


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
	player_dev_labels_context = PlayerDevLabelsContext.new()
	if !remote_player_nodes_provider.is_null() and remote_player_nodes_provider.is_valid():
		player_dev_labels_context.configure(remote_player_nodes_provider)
	world_telemetry_context = WorldTelemetryContext.new()
	world_telemetry_context.configure(connection_service_ref)
	_connect_window_controller_signals()


func configure_remote_player_nodes_provider(provider: Callable) -> void:
	remote_player_nodes_provider = provider
	if player_dev_labels_context != null:
		player_dev_labels_context.configure(remote_player_nodes_provider)


func reset() -> void:
	if debug_flow != null:
		debug_flow.reset()
	if display_refresh_flow != null:
		display_refresh_flow.reset()
	if server_hitbox_overlay != null && is_instance_valid(server_hitbox_overlay) and server_hitbox_overlay.has_method("set_hitbox_entries"):
		server_hitbox_overlay.set_hitbox_entries([])
	if player_dev_labels_context != null && player_dev_labels_context.has_method("clear_labels"):
		player_dev_labels_context.clear_labels()
	if world_telemetry_context != null:
		world_telemetry_context.reset()
	game_target_kind = ""
	game_target_id = ""
	game_target_player_id = ""


func process(has_received_state: bool) -> void:
	has_received_gameplay_state = has_received_state
	if Input.is_action_just_pressed("DevToggle0"):
		toggle_devtools_window()
	if Input.is_action_just_pressed("DevToggle8"):
		if Input.is_key_pressed(KEY_SHIFT):
			if player_dev_label_mode == "network":
				player_dev_label_mode = ""
			else:
				player_dev_label_mode = "network"
		else:
			if player_dev_label_mode == "basic":
				player_dev_label_mode = ""
			else:
				player_dev_label_mode = "basic"
		if player_dev_labels_context != null && player_dev_labels_context.has_method("set_mode"):
			player_dev_labels_context.set_mode(player_dev_label_mode)
	if Input.is_action_just_pressed("DevToggle9") and world_telemetry_context != null:
		world_telemetry_context.toggle_overlay()
	if hotkey_flow != null:
		hotkey_flow.process(has_received_state)
	if debug_flow != null:
		debug_flow.process(has_received_state)
	if player_dev_labels_context != null and world_telemetry_context != null:
		if world_telemetry_context.has_method("telemetry_snapshot") and player_dev_labels_context.has_method("apply_network_metrics"):
			player_dev_labels_context.apply_network_metrics(world_telemetry_context.telemetry_snapshot())
	if player_dev_labels_context != null && player_dev_labels_context.has_method("sync_remote_labels"):
		player_dev_labels_context.sync_remote_labels()
	if world_telemetry_context != null:
		world_telemetry_context.process(has_received_state, 0.0)


func toggle_devtools_window() -> void:
	if devtools_window_controller != null:
		devtools_window_controller.toggle_window()


func apply_debug_status(status: Dictionary) -> void:
	if devtools_window_controller != null:
		devtools_window_controller.apply_debug_status(status)


func apply_gameplay_state(state: Dictionary) -> void:
	apply_debug_status(state.get("debug_status", {}))
	if display_refresh_flow != null:
		display_refresh_flow.refresh_gameplay_state(state)
		local_player_id = display_refresh_flow.local_player_id()
		game_target_kind = display_refresh_flow.game_target_kind()
		game_target_id = display_refresh_flow.game_target_id()
		if game_target_kind == "player":
			game_target_player_id = game_target_id
		else:
			game_target_player_id = ""
	if devtools_window_controller != null:
		devtools_window_controller.configure_kill_player_routing(
			connection_service,
			local_player_id,
			game_target_kind,
			game_target_id
		)
	if player_dev_labels_context != null && player_dev_labels_context.has_method("apply_gameplay_state"):
		player_dev_labels_context.apply_gameplay_state(state)
	if world_telemetry_context != null:
		world_telemetry_context.apply_gameplay_state(state)


func refresh_spawn_player_slots(max_players: int) -> void:
	if display_refresh_flow == null:
		return
	display_refresh_flow.refresh_spawn_player_slots(max_players)


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
	if devtools_window_controller.has_signal("set_score_requested"):
		var set_score_callable := Callable(self, "request_set_score")
		if !devtools_window_controller.is_connected("set_score_requested", set_score_callable):
			devtools_window_controller.connect("set_score_requested", set_score_callable)
	if devtools_window_controller.has_signal("add_score_requested"):
		var add_score_callable := Callable(self, "request_add_score")
		if !devtools_window_controller.is_connected("add_score_requested", add_score_callable):
			devtools_window_controller.connect("add_score_requested", add_score_callable)
	if devtools_window_controller.has_signal("set_lives_requested"):
		var set_lives_callable := Callable(self, "request_set_lives")
		if !devtools_window_controller.is_connected("set_lives_requested", set_lives_callable):
			devtools_window_controller.connect("set_lives_requested", set_lives_callable)
	if devtools_window_controller.has_signal("add_lives_requested"):
		var add_lives_callable := Callable(self, "request_add_lives")
		if !devtools_window_controller.is_connected("add_lives_requested", add_lives_callable):
			devtools_window_controller.connect("add_lives_requested", add_lives_callable)
	if devtools_window_controller.has_signal("clear_bullets_requested"):
		var clear_bullets_callable := Callable(self, "request_clear_bullets")
		if !devtools_window_controller.is_connected("clear_bullets_requested", clear_bullets_callable):
			devtools_window_controller.connect("clear_bullets_requested", clear_bullets_callable)
	if devtools_window_controller.has_signal("clear_asteroids_requested"):
		var clear_asteroids_callable := Callable(self, "request_clear_asteroids")
		if !devtools_window_controller.is_connected("clear_asteroids_requested", clear_asteroids_callable):
			devtools_window_controller.connect("clear_asteroids_requested", clear_asteroids_callable)
	if devtools_window_controller.has_signal("game_target_set_requested"):
		var game_target_set_callable := Callable(self, "request_set_game_target")
		if !devtools_window_controller.is_connected("game_target_set_requested", game_target_set_callable):
			devtools_window_controller.connect("game_target_set_requested", game_target_set_callable)
	if devtools_window_controller.has_signal("game_target_clear_requested"):
		var game_target_clear_callable := Callable(self, "request_clear_game_target")
		if !devtools_window_controller.is_connected("game_target_clear_requested", game_target_clear_callable):
			devtools_window_controller.connect("game_target_clear_requested", game_target_clear_callable)
	if devtools_window_controller.has_signal("show_server_hitboxes_changed"):
		var show_server_hitboxes_callable := Callable(self, "_on_show_server_hitboxes_changed")
		if !devtools_window_controller.is_connected("show_server_hitboxes_changed", show_server_hitboxes_callable):
			devtools_window_controller.connect("show_server_hitboxes_changed", show_server_hitboxes_callable)


func configure_server_hitbox_overlay(overlay_ref) -> void:
	server_hitbox_overlay = overlay_ref
	if server_hitbox_overlay != null && server_hitbox_overlay.has_method("set_hitbox_entries"):
		server_hitbox_overlay.set_hitbox_entries([])


func request_toggle_invincible(target_scope: String = "", target_player_id: String = "") -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_invincible(target_scope, target_player_id)


func request_toggle_infinite_lives(target_scope: String = "", target_player_id: String = "") -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_infinite_lives(target_scope, target_player_id)


func request_toggle_freeze_world(freeze_target: String = "") -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_freeze_world(freeze_target)


func request_toggle_freeze_player(target_scope: String = "", target_player_id: String = "") -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_freeze_player(target_scope, target_player_id)


func request_set_score(target_scope: String, target_player_id: String, score: int) -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		return
	debug_flow.set_score(target_scope, target_player_id, score)


func request_add_score(target_scope: String, target_player_id: String, amount: int) -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		return
	debug_flow.add_score(target_scope, target_player_id, amount)


func request_set_lives(target_scope: String, target_player_id: String, lives: int) -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		return
	debug_flow.set_lives(target_scope, target_player_id, lives)


func request_add_lives(target_scope: String, target_player_id: String, amount: int) -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		return
	debug_flow.add_lives(target_scope, target_player_id, amount)


func request_clear_bullets() -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.clear_bullets()


func request_clear_asteroids() -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.clear_asteroids()


func _on_show_server_hitboxes_changed(enabled: bool) -> void:
	if server_hitbox_overlay == null || !is_instance_valid(server_hitbox_overlay):
		return
	if server_hitbox_overlay.has_method("set_enabled"):
		server_hitbox_overlay.set_enabled(enabled)


func configure_local_player_id(player_id: String) -> void:
	local_player_id = player_id


func request_set_game_target(target_player_id: String) -> void:
	if !has_received_gameplay_state:
		return
	if connection_service == null:
		return
	connection_service.send_packet(Packets.set_target_player_request_packet(target_player_id))


func request_clear_game_target() -> void:
	request_set_game_target("")


func request_respawn_player(target_scope: String = DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, target_player_id: String = "") -> void:
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		ClientLogger.game_warn("GameplayDevtoolsContext: respawn request ignored, target_player_id is empty")
		return
	if !has_received_gameplay_state:
		return
	if dev_connection_service == null || !dev_connection_service.is_configured():
		ClientLogger.game_warn("GameplayDevtoolsContext: respawn request ignored, dev_connection_service is unavailable")
		return
	dev_connection_service.send_respawn_player(target_scope, target_player_id)


func request_respawn_local_player() -> void:
	if local_player_id == "":
		ClientLogger.game_warn("GameplayDevtoolsContext: local respawn request ignored, local_player_id is empty")
		return
	request_respawn_player(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, local_player_id)


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

