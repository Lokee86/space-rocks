extends RefCounted
class_name RuntimeScenarioPhaseRunner

const DevtoolsTargetResolverScript := preload("res://scripts/devtools/devtools_target_resolver.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")

const ASTEROID_SPAWN_BATCH_SIZE := 24
const ASTEROID_SPAWN_MARGIN := 500.0

const KNOWN_ACTIONS := [
	&"move_forward",
	&"move_backward",
	&"turn_left",
	&"turn_right",
	&"PrimaryFire",
	&"SecondaryFire",
]

var role := ""
var debug_flow
var dev_connection_service
var roster
var room_session_controller
var gameplay_session_controller
var status_writer


func configure(
	role_value: String,
	debug_flow_ref,
	dev_connection_service_ref,
	roster_ref,
	room_session_controller_ref,
	gameplay_session_controller_ref,
	status_writer_ref
) -> void:
	role = role_value
	debug_flow = debug_flow_ref
	dev_connection_service = dev_connection_service_ref
	roster = roster_ref
	room_session_controller = room_session_controller_ref
	gameplay_session_controller = gameplay_session_controller_ref
	status_writer = status_writer_ref


func prepare(setup: Dictionary) -> Dictionary:
	var asteroid_spawns := maxi(int(setup.get("asteroid_spawns", 0)), 0)
	var settle_seconds := maxf(float(setup.get("settle_seconds", 0.0)), 0.0)
	_status("setup_started", {
		"asteroid_spawns": asteroid_spawns,
		"settle_seconds": settle_seconds,
	})
	if role == "coordinator":
		await _spawn_asteroids(asteroid_spawns)
	if settle_seconds > 0.0:
		await Engine.get_main_loop().create_timer(settle_seconds).timeout
	_status("setup_completed", {"asteroid_spawns": asteroid_spawns})
	return {"ok": true}


func run(phase: Dictionary) -> Dictionary:
	var phase_name := str(phase.get("name", "phase"))
	_status("phase_started", {"phase": phase_name})
	_apply_actions(_actions_for_phase(phase))
	if role == "coordinator":
		_start_bullet_streams(int(phase.get("bullet_streams", 0)))
		if str(phase.get("kill_target_role", "")) == "participant":
			var target_player_id: String = roster.other_human_player_id()
			if target_player_id.is_empty():
				return _failure("phase %s could not resolve participant target" % phase_name)
			debug_flow.set_lives(DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER, target_player_id, 1)
			await Engine.get_main_loop().create_timer(0.25).timeout
			debug_flow.kill_player(DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER, target_player_id)
			_status("phase_action", {"phase": phase_name, "killed_player_id": target_player_id})

	var duration := maxf(float(phase.get("duration_seconds", 0.0)), 0.0)
	var spectate_delay := 0.0
	if role == "participant" and bool(phase.get("spectate_owner", false)):
		spectate_delay = clampf(float(phase.get("spectate_delay_seconds", 1.0)), 0.0, duration)
		if spectate_delay > 0.0:
			await Engine.get_main_loop().create_timer(spectate_delay).timeout
		var owner_id := str(room_session_controller.lobby_state_snapshot().get("owner_id", ""))
		if !owner_id.is_empty():
			gameplay_session_controller.request_spectate_target(owner_id)
	if duration > spectate_delay:
		await Engine.get_main_loop().create_timer(duration - spectate_delay).timeout
	_status("phase_completed", {"phase": phase_name})
	return {"ok": true}


func release_actions() -> void:
	for action in KNOWN_ACTIONS:
		Input.action_release(action)


func _spawn_asteroids(count: int) -> void:
	var spawn_count := maxi(count, 0)
	if spawn_count == 0:
		return
	var columns := maxi(int(ceil(sqrt(float(spawn_count)))), 1)
	var rows := maxi(int(ceil(float(spawn_count) / float(columns))), 1)
	var usable_width := Constants.WORLD_WIDTH - ASTEROID_SPAWN_MARGIN * 2.0
	var usable_height := Constants.WORLD_HEIGHT - ASTEROID_SPAWN_MARGIN * 2.0
	for index in range(spawn_count):
		var column := index % columns
		var row := index / columns
		var position := Vector2(
			ASTEROID_SPAWN_MARGIN + usable_width * (float(column) + 0.5) / float(columns),
			ASTEROID_SPAWN_MARGIN + usable_height * (float(row) + 0.5) / float(rows),
		)
		var angle := TAU * float(index) / float(spawn_count)
		dev_connection_service.send_spawn_from_placement_result({
			"action_name": &"spawn_asteroid",
			"server_position": position,
			"direction": Vector2(cos(angle), sin(angle)),
			"has_direction": true,
		})
		if (index + 1) % ASTEROID_SPAWN_BATCH_SIZE == 0:
			await Engine.get_main_loop().process_frame


func _start_bullet_streams(count: int) -> void:
	for index in range(maxi(count, 0)):
		var angle := TAU * float(index) / float(maxi(count, 1))
		dev_connection_service.send_begin_continuous_bullet_stream_from_placement_result({
			"server_position": Vector2(1200.0 + 12.0 * index, 900.0 + 7.0 * index),
			"direction": Vector2(cos(angle), sin(angle)),
			"has_direction": true,
		})


func _actions_for_phase(phase: Dictionary) -> Array:
	var by_role = phase.get("actions_by_role", {})
	if by_role is Dictionary and by_role.has(role) and by_role[role] is Array:
		return by_role[role]
	var actions = phase.get("actions", [])
	return actions if actions is Array else []


func _apply_actions(actions: Array) -> void:
	release_actions()
	for action_value in actions:
		var action := StringName(str(action_value))
		if InputMap.has_action(action):
			Input.action_press(action)


func _status(state: String, fields: Dictionary = {}) -> void:
	status_writer.write(state, fields)


func _failure(message: String) -> Dictionary:
	return {"ok": false, "error": message}
