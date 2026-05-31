class_name DevtoolsWindowController
extends RefCounted

signal toggle_invincible_requested(target_player_id: String)
signal toggle_infinite_lives_requested(target_player_id: String)
signal toggle_freeze_world_requested(freeze_target: String)
signal toggle_freeze_player_requested(target_player_id: String)
signal set_score_requested(target_player_id: String, score: int)
signal add_score_requested(target_player_id: String, amount: int)
signal set_lives_requested(target_player_id: String, lives: int)
signal add_lives_requested(target_player_id: String, amount: int)
signal clear_bullets_requested
signal clear_asteroids_requested
signal placement_action_requested(action_name: StringName, placement_context: Dictionary)
signal respawn_player_requested(target_player_id: String)

const DevtoolsWindowScene := preload("res://scenes/devtools/devtools_window.tscn")
const ClientLogger = preload("res://scripts/logging/logger.gd")

var window: Window
var parent: Node
var latest_debug_status := {}
var latest_target_rows: Array = []
var latest_invincible_rows: Array = []
var latest_infinite_lives_rows: Array = []
var latest_player_frozen_rows: Array = []
var connection_service
var self_player_id := ""


func ensure_window() -> Window:
	if window != null && is_instance_valid(window):
		return window

	window = DevtoolsWindowScene.instantiate()
	parent = Engine.get_main_loop().root
	parent.add_child(window)
	_connect_window_signals()
	window.set_debug_status(latest_debug_status)
	window.refresh_kill_player_targets(latest_target_rows)
	if window.has_method("refresh_respawn_player_targets"):
		window.refresh_respawn_player_targets(latest_target_rows)
	if window.has_method("refresh_invincible_targets"):
		window.refresh_invincible_targets(latest_invincible_rows)
	if window.has_method("refresh_infinite_lives_targets"):
		window.refresh_infinite_lives_targets(latest_infinite_lives_rows)
	if window.has_method("refresh_player_frozen_targets"):
		window.refresh_player_frozen_targets(latest_player_frozen_rows)
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
	var devtools_window := ensure_window()
	devtools_window.refresh_kill_player_targets(target_rows)
	if devtools_window.has_method("refresh_respawn_player_targets"):
		devtools_window.refresh_respawn_player_targets(target_rows)


func refresh_debug_player_targets(
	target_rows: Array,
	invincible_rows: Array,
	infinite_lives_rows: Array,
	player_frozen_rows: Array
) -> void:
	latest_target_rows = target_rows
	latest_invincible_rows = invincible_rows
	latest_infinite_lives_rows = infinite_lives_rows
	latest_player_frozen_rows = player_frozen_rows

	if window == null || !is_instance_valid(window):
		return

	window.refresh_kill_player_targets(latest_target_rows)
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


func refresh_spawn_player_slots(max_players: int) -> void:
	var devtools_window := ensure_window()
	if devtools_window.has_method("refresh_spawn_player_slots"):
		devtools_window.refresh_spawn_player_slots(max_players)


func configure_kill_player_routing(connection_service_ref, self_id: String) -> void:
	connection_service = connection_service_ref
	self_player_id = self_id


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


func _on_toggle_invincible_requested(target_player_id: String) -> void:
	toggle_invincible_requested.emit(target_player_id)


func _on_toggle_infinite_lives_requested(target_player_id: String) -> void:
	toggle_infinite_lives_requested.emit(target_player_id)


func _on_toggle_freeze_world_requested(freeze_target: String) -> void:
	toggle_freeze_world_requested.emit(freeze_target)


func _on_toggle_freeze_player_requested(target_player_id: String) -> void:
	toggle_freeze_player_requested.emit(target_player_id)


func _on_set_score_requested(target_player_id: String, score: int) -> void:
	set_score_requested.emit(target_player_id, score)


func _on_add_score_requested(target_player_id: String, amount: int) -> void:
	add_score_requested.emit(target_player_id, amount)


func _on_set_lives_requested(target_player_id: String, lives: int) -> void:
	set_lives_requested.emit(target_player_id, lives)


func _on_add_lives_requested(target_player_id: String, amount: int) -> void:
	add_lives_requested.emit(target_player_id, amount)


func _on_clear_bullets_requested() -> void:
	clear_bullets_requested.emit()


func _on_clear_asteroids_requested() -> void:
	clear_asteroids_requested.emit()


func _on_kill_player_requested(selected_player_id: String) -> void:
	if connection_service == null || selected_player_id == "":
		return

	if selected_player_id == self_player_id:
		connection_service.send_debug_kill_player_request()
	else:
		connection_service.send_debug_kill_target_player_request(selected_player_id)


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
	if target_player_id == "":
		ClientLogger.game_warn("Devtools respawn placement blocked: target_player_id is empty")
		return
	ClientLogger.game_info("Devtools respawn direct request starting")
	respawn_player_requested.emit(target_player_id)
