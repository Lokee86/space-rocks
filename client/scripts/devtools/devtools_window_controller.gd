extends RefCounted
class_name DevtoolsWindowController

signal toggle_invincible_requested(target_scope: String, target_player_id: String)
signal toggle_infinite_lives_requested(target_scope: String, target_player_id: String)
signal toggle_freeze_world_requested(freeze_target: String)
signal toggle_freeze_player_requested(target_scope: String, target_player_id: String)
signal set_score_requested(target_scope: String, target_player_id: String, score: int)
signal add_score_requested(target_scope: String, target_player_id: String, amount: int)
signal set_lives_requested(target_scope: String, target_player_id: String, lives: int)
signal add_lives_requested(target_scope: String, target_player_id: String, amount: int)
signal clear_bullets_requested
signal clear_asteroids_requested
signal game_target_set_requested(target_player_id: String)
signal game_target_clear_requested
signal placement_action_requested(action_name: StringName, placement_context: Dictionary)
signal respawn_player_requested(target_scope: String, target_player_id: String)
signal show_server_hitboxes_changed(enabled: bool)
signal telemetry_sources_changed(local_source: String, target_source: String)

const DevtoolsWindowScene := preload("res://scenes/devtools/devtools_window.tscn")
const ClientLogger = preload("res://scripts/logging/logger.gd")

var window: Window
var parent: Node
var latest_debug_status := {}
var latest_kill_player_rows: Array = []
var latest_target_rows: Array = []
var latest_invincible_rows: Array = []
var latest_infinite_lives_rows: Array = []
var latest_player_frozen_rows: Array = []
var latest_game_target_rows: Array = []
var latest_game_target_player_id := ""
var latest_local_player_state := {}
var latest_target_kind := ""
var latest_target_id := ""
var latest_target_state := {}
var latest_local_telemetry_source := "players"
var latest_target_telemetry_source := "players"
var connection_service
var self_player_id := ""
var game_target_kind := ""
var game_target_id := ""
var show_server_hitboxes := false


func ensure_window() -> Window:
	if window != null && is_instance_valid(window):
		return window

	window = DevtoolsWindowScene.instantiate()
	parent = Engine.get_main_loop().root
	parent.add_child(window)
	_connect_window_signals()
	window.set_debug_status(latest_debug_status)
	window.refresh_kill_player_targets(latest_kill_player_rows)
	if window.has_method("refresh_respawn_player_targets"):
		window.refresh_respawn_player_targets(latest_target_rows)
	if window.has_method("refresh_invincible_targets"):
		window.refresh_invincible_targets(latest_invincible_rows)
	if window.has_method("refresh_infinite_lives_targets"):
		window.refresh_infinite_lives_targets(latest_infinite_lives_rows)
	if window.has_method("refresh_player_frozen_targets"):
		window.refresh_player_frozen_targets(latest_player_frozen_rows)
	if window.has_method("refresh_game_target_options"):
		window.refresh_game_target_options(
			latest_game_target_rows,
			latest_game_target_player_id,
			game_target_kind,
			game_target_id
		)
	if window.has_method("refresh_local_player_state"):
		window.refresh_local_player_state(latest_local_player_state)
	if window.has_method("refresh_target_state"):
		window.refresh_target_state(latest_target_kind, latest_target_id, latest_target_state)
	if window.has_method("set_show_server_hitboxes"):
		window.set_show_server_hitboxes(show_server_hitboxes)
	if window.has_method("set_telemetry_sources"):
		window.set_telemetry_sources(latest_local_telemetry_source, latest_target_telemetry_source)
	else:
		if window.has_method("set_local_telemetry_source"):
			window.set_local_telemetry_source(latest_local_telemetry_source)
		if window.has_method("set_target_telemetry_source"):
			window.set_target_telemetry_source(latest_target_telemetry_source)
	if window.has_method("refresh_local_player_state"):
		window.refresh_local_player_state(latest_local_player_state)
	if window.has_method("refresh_target_state"):
		window.refresh_target_state(latest_target_kind, latest_target_id, latest_target_state)
	return window


func show_window() -> void:
	ensure_window().show_window()


func hide_window() -> void:
	if window != null && is_instance_valid(window):
		window.hide_window()


func toggle_window() -> void:
	var devtools_window := ensure_window()
	if devtools_window.visible:
		devtools_window.hide_window()
	else:
		devtools_window.show_window()


func apply_debug_status(status: Dictionary) -> void:
	latest_debug_status = status
	if window != null && is_instance_valid(window):
		window.set_debug_status(latest_debug_status)


func refresh_kill_player_targets(target_rows: Array) -> void:
	latest_kill_player_rows = target_rows
	var devtools_window := ensure_window()
	devtools_window.refresh_kill_player_targets(latest_kill_player_rows)


func refresh_debug_player_targets(
	kill_player_rows: Array,
	target_rows: Array,
	invincible_rows: Array,
	infinite_lives_rows: Array,
	player_frozen_rows: Array
) -> void:
	latest_kill_player_rows = kill_player_rows
	latest_target_rows = target_rows
	latest_invincible_rows = invincible_rows
	latest_infinite_lives_rows = infinite_lives_rows
	latest_player_frozen_rows = player_frozen_rows

	if window == null || !is_instance_valid(window):
		return

	window.refresh_kill_player_targets(latest_kill_player_rows)
	if window.has_method("refresh_respawn_player_targets"):
		window.refresh_respawn_player_targets(latest_target_rows)
	if window.has_method("refresh_invincible_targets"):
		window.refresh_invincible_targets(latest_invincible_rows)
	if window.has_method("refresh_infinite_lives_targets"):
		window.refresh_infinite_lives_targets(latest_infinite_lives_rows)
	if window.has_method("refresh_player_frozen_targets"):
		window.refresh_player_frozen_targets(latest_player_frozen_rows)


func refresh_counter_player_targets(rows: Array) -> void:
	var devtools_window := ensure_window()
	if devtools_window.has_method("refresh_counter_player_targets"):
		devtools_window.refresh_counter_player_targets(rows)


func refresh_game_target_options(rows: Array, current_target_kind: String, current_target_id: String) -> void:
	latest_game_target_rows = rows
	game_target_kind = current_target_kind
	game_target_id = current_target_id
	if game_target_kind == DevtoolsTargetResolver.TARGET_KIND_PLAYER:
		latest_game_target_player_id = game_target_id
	else:
		latest_game_target_player_id = ""

	if window == null || !is_instance_valid(window):
		return
	if window.has_method("refresh_game_target_options"):
		window.refresh_game_target_options(
			latest_game_target_rows,
			latest_game_target_player_id,
			game_target_kind,
			game_target_id
		)


func refresh_local_player_state(state: Dictionary) -> void:
	latest_local_player_state = state
	if window == null or !is_instance_valid(window):
		return
	if window.has_method("refresh_local_player_state"):
		window.refresh_local_player_state(state)


func refresh_target_state(target_kind: String, target_id: String, state: Dictionary) -> void:
	latest_target_kind = target_kind
	latest_target_id = target_id
	latest_target_state = state
	if window == null or !is_instance_valid(window):
		return
	if window.has_method("refresh_target_state"):
		window.refresh_target_state(target_kind, target_id, state)


func local_telemetry_source() -> String:
	return latest_local_telemetry_source


func target_telemetry_source() -> String:
	return latest_target_telemetry_source


func refresh_spawn_player_slots(max_players: int) -> void:
	var devtools_window := ensure_window()
	if devtools_window.has_method("refresh_spawn_player_slots"):
		devtools_window.refresh_spawn_player_slots(max_players)


func configure_kill_player_routing(
	connection_service_ref,
	self_id: String,
	target_kind: String,
	target_id: String
) -> void:
	connection_service = connection_service_ref
	self_player_id = self_id
	game_target_kind = target_kind
	game_target_id = target_id
	if game_target_kind == "player":
		latest_game_target_player_id = game_target_id
	else:
		latest_game_target_player_id = ""


func request_placement_action(action_name: StringName, placement_context: Dictionary = {}) -> void:
	placement_action_requested.emit(action_name, placement_context)


func _connect_window_signals() -> void:
	if !window.toggle_invincible_requested.is_connected(_on_toggle_invincible_requested):
		window.toggle_invincible_requested.connect(_on_toggle_invincible_requested)
	if !window.toggle_infinite_lives_requested.is_connected(_on_toggle_infinite_lives_requested):
		window.toggle_infinite_lives_requested.connect(_on_toggle_infinite_lives_requested)
	if !window.toggle_freeze_world_requested.is_connected(_on_toggle_freeze_world_requested):
		window.toggle_freeze_world_requested.connect(_on_toggle_freeze_world_requested)
	if !window.toggle_freeze_player_requested.is_connected(_on_toggle_freeze_player_requested):
		window.toggle_freeze_player_requested.connect(_on_toggle_freeze_player_requested)
	if !window.spawn_asteroid_placement_requested.is_connected(_on_spawn_asteroid_placement_requested):
		window.spawn_asteroid_placement_requested.connect(_on_spawn_asteroid_placement_requested)
	if !window.spawn_player_placement_requested.is_connected(_on_spawn_player_placement_requested):
		window.spawn_player_placement_requested.connect(_on_spawn_player_placement_requested)
	if !window.spawn_bullet_placement_requested.is_connected(_on_spawn_bullet_placement_requested):
		window.spawn_bullet_placement_requested.connect(_on_spawn_bullet_placement_requested)
	if !window.respawn_player_placement_requested.is_connected(_on_respawn_player_placement_requested):
		window.respawn_player_placement_requested.connect(_on_respawn_player_placement_requested)
	if !window.kill_player_requested.is_connected(_on_kill_player_requested):
		window.kill_player_requested.connect(_on_kill_player_requested)
	if window.has_signal("set_score_requested") and !window.set_score_requested.is_connected(_on_set_score_requested):
		window.set_score_requested.connect(_on_set_score_requested)
	if window.has_signal("add_score_requested") and !window.add_score_requested.is_connected(_on_add_score_requested):
		window.add_score_requested.connect(_on_add_score_requested)
	if window.has_signal("set_lives_requested") and !window.set_lives_requested.is_connected(_on_set_lives_requested):
		window.set_lives_requested.connect(_on_set_lives_requested)
	if window.has_signal("add_lives_requested") and !window.add_lives_requested.is_connected(_on_add_lives_requested):
		window.add_lives_requested.connect(_on_add_lives_requested)
	if window.has_signal("clear_bullets_requested") and !window.clear_bullets_requested.is_connected(_on_clear_bullets_requested):
		window.clear_bullets_requested.connect(_on_clear_bullets_requested)
	if window.has_signal("clear_asteroids_requested") and !window.clear_asteroids_requested.is_connected(_on_clear_asteroids_requested):
		window.clear_asteroids_requested.connect(_on_clear_asteroids_requested)
	if window.has_signal("game_target_set_requested") and !window.game_target_set_requested.is_connected(_on_game_target_set_requested):
		window.game_target_set_requested.connect(_on_game_target_set_requested)
	if window.has_signal("game_target_clear_requested") and !window.game_target_clear_requested.is_connected(_on_game_target_clear_requested):
		window.game_target_clear_requested.connect(_on_game_target_clear_requested)
	if window.has_signal("show_server_hitboxes_changed") and !window.show_server_hitboxes_changed.is_connected(_on_show_server_hitboxes_changed):
		window.show_server_hitboxes_changed.connect(_on_show_server_hitboxes_changed)
	if window.has_signal("telemetry_sources_changed") and !window.telemetry_sources_changed.is_connected(_on_telemetry_sources_changed):
		window.telemetry_sources_changed.connect(_on_telemetry_sources_changed)


func _on_telemetry_sources_changed(local_source: String, target_source: String) -> void:
	if local_source == latest_local_telemetry_source and target_source == latest_target_telemetry_source:
		return
	latest_local_telemetry_source = local_source
	latest_target_telemetry_source = target_source
	telemetry_sources_changed.emit(local_source, target_source)


func _on_toggle_invincible_requested(target_player_id: String) -> void:
	var target_context := _effective_target_context(target_player_id)
	if str(target_context.get("target_scope", "")) == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and str(target_context.get("target_player_id", "")) == "":
		return
	toggle_invincible_requested.emit(
		str(target_context.get("target_scope", "")),
		str(target_context.get("target_player_id", ""))
	)


func _on_toggle_infinite_lives_requested(target_player_id: String) -> void:
	var target_context := _effective_target_context(target_player_id)
	if str(target_context.get("target_scope", "")) == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and str(target_context.get("target_player_id", "")) == "":
		return
	toggle_infinite_lives_requested.emit(
		str(target_context.get("target_scope", "")),
		str(target_context.get("target_player_id", ""))
	)


func _on_toggle_freeze_world_requested(freeze_target: String) -> void:
	toggle_freeze_world_requested.emit(freeze_target)


func _on_toggle_freeze_player_requested(target_player_id: String) -> void:
	var target_context := _effective_target_context(target_player_id)
	if str(target_context.get("target_scope", "")) == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and str(target_context.get("target_player_id", "")) == "":
		return
	toggle_freeze_player_requested.emit(
		str(target_context.get("target_scope", "")),
		str(target_context.get("target_player_id", ""))
	)


func _on_set_score_requested(target_player_id: String, score: int) -> void:
	var target_context := _effective_target_context(target_player_id)
	if str(target_context.get("target_scope", "")) == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and str(target_context.get("target_player_id", "")) == "":
		return
	set_score_requested.emit(
		str(target_context.get("target_scope", "")),
		str(target_context.get("target_player_id", "")),
		score
	)


func _on_add_score_requested(target_player_id: String, amount: int) -> void:
	var target_context := _effective_target_context(target_player_id)
	if str(target_context.get("target_scope", "")) == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and str(target_context.get("target_player_id", "")) == "":
		return
	add_score_requested.emit(
		str(target_context.get("target_scope", "")),
		str(target_context.get("target_player_id", "")),
		amount
	)


func _on_set_lives_requested(target_player_id: String, lives: int) -> void:
	var target_context := _effective_target_context(target_player_id)
	if str(target_context.get("target_scope", "")) == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and str(target_context.get("target_player_id", "")) == "":
		return
	set_lives_requested.emit(
		str(target_context.get("target_scope", "")),
		str(target_context.get("target_player_id", "")),
		lives
	)


func _on_add_lives_requested(target_player_id: String, amount: int) -> void:
	var target_context := _effective_target_context(target_player_id)
	if str(target_context.get("target_scope", "")) == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and str(target_context.get("target_player_id", "")) == "":
		return
	add_lives_requested.emit(
		str(target_context.get("target_scope", "")),
		str(target_context.get("target_player_id", "")),
		amount
	)


func _on_clear_bullets_requested() -> void:
	clear_bullets_requested.emit()


func _on_clear_asteroids_requested() -> void:
	clear_asteroids_requested.emit()


func _on_game_target_set_requested(target_player_id: String) -> void:
	game_target_set_requested.emit(target_player_id)


func _on_game_target_clear_requested() -> void:
	game_target_clear_requested.emit()


func _on_show_server_hitboxes_changed(enabled: bool) -> void:
	show_server_hitboxes = enabled
	show_server_hitboxes_changed.emit(enabled)


func _on_kill_player_requested(selected_player_id: String) -> void:
	if connection_service == null:
		return

	var target_context := _effective_target_context(selected_player_id)
	var target_scope := str(target_context.get("target_scope", ""))
	var target_player_id := str(target_context.get("target_player_id", ""))
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS:
		connection_service.send_debug_kill_player_request(target_scope, "")
		return

	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		return

	if target_player_id == self_player_id:
		connection_service.send_debug_kill_player_request(target_scope, "")
	else:
		connection_service.send_debug_kill_target_player_request(target_player_id, target_scope)


func _on_spawn_asteroid_placement_requested() -> void:
	request_placement_action(&"spawn_asteroid")


func _on_spawn_player_placement_requested(target_player_id: String) -> void:
	var placement_context := {}
	if target_player_id != "":
		placement_context["target_player_id"] = target_player_id
	request_placement_action(&"spawn_player", placement_context)


func _on_spawn_bullet_placement_requested() -> void:
	request_placement_action(&"spawn_bullet")


func _on_respawn_player_placement_requested(target_player_id: String) -> void:
	ClientLogger.game_info("Devtools respawn placement received target_player_id='%s'" % target_player_id)
	var target_context := _effective_target_context(target_player_id)
	var target_scope := str(target_context.get("target_scope", ""))
	var effective_target_player_id := str(target_context.get("target_player_id", ""))
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS:
		ClientLogger.game_info("Devtools respawn all-players request starting")
		respawn_player_requested.emit(target_scope, "")
		return

	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and effective_target_player_id == "":
		ClientLogger.game_warn("Devtools respawn placement blocked: effective target_player_id is empty")
		return
	ClientLogger.game_info("Devtools respawn direct request starting")
	respawn_player_requested.emit(target_scope, effective_target_player_id)


func _effective_target(selected_tool_target: String) -> String:
	return DevtoolsTargetResolver.resolve_player_target(
		selected_tool_target,
		game_target_kind,
		game_target_id,
		self_player_id
	)


func _effective_target_context(selected_tool_target: String) -> Dictionary:
	return DevtoolsTargetResolver.resolve_player_target_scope(
		selected_tool_target,
		game_target_kind,
		game_target_id,
		self_player_id
	)
