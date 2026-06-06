extends RefCounted

const DevtoolsPlayerTargetModel = preload("res://scripts/devtools/devtools_player_target_model.gd")

var window_controller
var target_model := DevtoolsPlayerTargetModel.new()
var latest_max_players := 0


func configure(window_controller_ref) -> void:
	window_controller = window_controller_ref
	if window_controller != null and window_controller.has_signal("telemetry_sources_changed"):
		if !window_controller.telemetry_sources_changed.is_connected(_refresh_telemetry):
			window_controller.telemetry_sources_changed.connect(_refresh_telemetry)


func reset() -> void:
	target_model.reset()
	latest_max_players = 0


func local_player_id() -> String:
	return str(target_model.self_id) if target_model != null else ""


func game_target_kind() -> String:
	return str(target_model.game_target_kind) if target_model != null else ""


func game_target_id() -> String:
	return str(target_model.game_target_id) if target_model != null else ""


func game_target_player_id() -> String:
	return str(target_model.game_target_player_id) if target_model != null else ""


func refresh_gameplay_state(state: Dictionary) -> void:
	target_model.apply_gameplay_state(state)
	_refresh_telemetry()
	_refresh_debug_player_targets()
	if window_controller == null:
		return

	if window_controller.has_method("refresh_game_target_options"):
		window_controller.refresh_game_target_options(
			target_model.target_rows(),
			target_model.game_target_kind,
			target_model.game_target_id
		)
	window_controller.refresh_counter_player_targets(target_model.active_player_target_rows())
	if window_controller.has_method("refresh_spawn_player_slots"):
		window_controller.refresh_spawn_player_slots(latest_max_players)


func apply_debug_status_packet(state: Dictionary) -> void:
	if target_model != null:
		target_model.apply_debug_statuses(state.get("debug_statuses", {}))
	if window_controller != null and window_controller.has_method("apply_debug_status"):
		window_controller.apply_debug_status(state.get("debug_status", {}))
	_refresh_debug_player_targets()


func _refresh_telemetry() -> void:
	if window_controller == null:
		return

	if window_controller.has_method("refresh_local_player_state"):
		window_controller.refresh_local_player_state(
			target_model.local_player_state_for_source(_local_telemetry_source())
		)
	if window_controller.has_method("refresh_target_state"):
		window_controller.refresh_target_state(
			target_model.game_target_kind,
			target_model.game_target_id,
			target_model.target_state_for_source(_target_telemetry_source())
		)


func _refresh_debug_player_targets() -> void:
	if window_controller == null:
		return

	window_controller.refresh_debug_player_targets(
		target_model.kill_player_target_rows(),
		target_model.respawn_player_target_rows(),
		target_model.invincible_target_rows(),
		target_model.infinite_lives_target_rows(),
		target_model.player_frozen_target_rows()
	)


func _local_telemetry_source() -> String:
	if window_controller == null or !window_controller.has_method("local_telemetry_source"):
		return "players"
	return window_controller.local_telemetry_source()


func _target_telemetry_source() -> String:
	if window_controller == null or !window_controller.has_method("target_telemetry_source"):
		return "players"
	return window_controller.target_telemetry_source()


func refresh_spawn_player_slots(max_players: int) -> void:
	latest_max_players = max_players
	if window_controller == null:
		return
	if window_controller.has_method("refresh_spawn_player_slots"):
		window_controller.refresh_spawn_player_slots(latest_max_players)
