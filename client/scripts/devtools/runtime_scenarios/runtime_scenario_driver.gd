extends Node
class_name RuntimeScenarioDriver

const StatusWriterScript := preload("res://scripts/devtools/runtime_scenarios/runtime_scenario_status_writer.gd")
const RosterScript := preload("res://scripts/devtools/runtime_scenarios/runtime_scenario_roster.gd")
const PhaseRunnerScript := preload("res://scripts/devtools/runtime_scenarios/runtime_scenario_phase_runner.gd")
const GameplayDebugFlowScript := preload("res://scripts/devtools/gameplay_debug_flow.gd")
const DevConnectionServiceScript := preload("res://scripts/devtools/dev_connection_service.gd")
const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")

var scenario: Dictionary = {}
var main_menu_session_controller
var room_session_controller
var gameplay_session_controller
var connection_service
var status_writer
var debug_flow
var dev_connection_service
var roster
var phase_runner
var role := ""
var client_id := ""
var _connection_closed := false
var _failure_message := ""
var _terminal_status_written := false


func configure(
	scenario_value: Dictionary,
	main_menu_session_controller_ref,
	room_session_controller_ref,
	gameplay_session_controller_ref,
	connection_service_ref
) -> void:
	scenario = scenario_value.duplicate(true)
	main_menu_session_controller = main_menu_session_controller_ref
	room_session_controller = room_session_controller_ref
	gameplay_session_controller = gameplay_session_controller_ref
	connection_service = connection_service_ref
	role = str(scenario.get("role", ""))
	client_id = str(scenario.get("client_id", ""))
	status_writer = StatusWriterScript.new()
	status_writer.configure(str(scenario.get("status_path", "")), client_id, role)
	debug_flow = GameplayDebugFlowScript.new()
	debug_flow.configure(connection_service)
	dev_connection_service = DevConnectionServiceScript.new()
	dev_connection_service.configure(connection_service)
	roster = RosterScript.new()
	roster.configure(room_session_controller, _expected_human_count())
	phase_runner = PhaseRunnerScript.new()
	phase_runner.configure(
		role,
		debug_flow,
		dev_connection_service,
		roster,
		room_session_controller,
		gameplay_session_controller,
		status_writer
	)
	var status_path: String = ProjectSettings.globalize_path(str(scenario.get("status_path", "")))
	var report_directory: String = status_path.get_base_dir().path_join("measurements").path_join(client_id)
	gameplay_session_controller.configure_measurement_report_directory(report_directory)
	if connection_service != null and connection_service.has_signal("closed"):
		connection_service.closed.connect(_on_connection_closed)


func start() -> void:
	call_deferred("_run")


func _run() -> void:
	if !bool(scenario.get("valid", false)):
		_fail(str(scenario.get("error", "runtime scenario configuration is invalid")))
		return
	_status("starting", {"scenario_id": _scenario_id()})
	if role == "coordinator":
		main_menu_session_controller.request_create_room(_room_config())
	else:
		var requested_room_code := str(scenario.get("room_code", ""))
		if requested_room_code.is_empty():
			_fail("participant room code is required")
			return
		main_menu_session_controller.request_join_room(requested_room_code)
	if !await _wait_until(Callable(self, "_is_in_lobby"), "enter lobby", _setup_timeout()):
		return

	var lobby: Dictionary = room_session_controller.lobby_state_snapshot()
	_status("room_ready", {
		"room_code": str(lobby.get("room_code", "")),
		"local_player_id": str(lobby.get("local_player_id", "")),
	})
	if role == "coordinator":
		if !await _prepare_coordinator_lobby():
			return
	room_session_controller.request_ready(true)
	if role == "coordinator":
		if !await _wait_until(Callable(roster, "lobby_can_start"), "all members ready", _setup_timeout()):
			return
		room_session_controller.request_start_game()
	if !await _wait_until(Callable(self, "_is_in_game"), "match start", _setup_timeout()):
		return
	if !await _wait_until(Callable(self, "_is_tooling_ready"), "tooling readiness", _setup_timeout()):
		return
	if role == "coordinator":
		debug_flow.set_lives(DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, "", 99)
		await get_tree().create_timer(0.5).timeout

	var setup_value = scenario.get("setup", {})
	var setup: Dictionary = setup_value if setup_value is Dictionary else {}
	var setup_result: Dictionary = await phase_runner.prepare(setup)
	if !bool(setup_result.get("ok", false)):
		_fail(str(setup_result.get("error", "runtime scenario setup failed")))
		return

	var measurement_request: String = gameplay_session_controller.start_measurement(
		_scenario_id(),
		{
			"scenario_id": _scenario_id(),
			"client_id": client_id,
			"role": role,
			"scenario": scenario.duplicate(true),
		}
	)
	if measurement_request.is_empty():
		_fail("measurement start request was rejected")
		return
	if !await _wait_until(Callable(self, "_measurement_is_recording"), "measurement start", _setup_timeout()):
		return

	for raw_phase in scenario.get("phases", []):
		if !(raw_phase is Dictionary):
			continue
		var phase_result: Dictionary = await phase_runner.run(raw_phase)
		if !bool(phase_result.get("ok", false)):
			_fail(str(phase_result.get("error", "runtime scenario phase failed")))
			return
		if _connection_closed:
			_fail("connection closed during phase %s" % str(raw_phase.get("name", "phase")))
			return

	phase_runner.release_actions()
	_status("cleanup_started", {"scenario_id": _scenario_id()})
	if role == "coordinator":
		debug_flow.clear_bullets()
	await get_tree().create_timer(5.0).timeout
	if gameplay_session_controller.stop_measurement().is_empty():
		_fail("measurement stop request was rejected")
		return
	if !await _wait_until(Callable(self, "_measurement_is_stopped"), "measurement stop", _setup_timeout()):
		return
	if !await _wait_until(Callable(self, "_measurement_export_finished"), "measurement export", _setup_timeout()):
		return
	var export_result: Dictionary = gameplay_session_controller.get_latest_measurement_export_result()
	if !bool(export_result.get("success", false)):
		_fail("measurement export failed: %s" % str(export_result.get("error", "unknown error")))
		return
	_terminal_status_written = true
	_status("completed", {
		"scenario_id": _scenario_id(),
		"room_code": str(room_session_controller.lobby_state_snapshot().get("room_code", "")),
		"measurement_report": ProjectSettings.globalize_path(str(export_result.get("path", ""))),
	})
	get_tree().quit(0)


func _prepare_coordinator_lobby() -> bool:
	if !await _wait_until(Callable(roster, "humans_joined"), "real clients to join", _setup_timeout()):
		return false
	var requested_bots := int(scenario.get("bots", 0))
	while roster.bot_count() < requested_bots:
		var required_count: int = roster.bot_count() + 1
		room_session_controller.request_add_bot()
		if !await _wait_until(
			Callable(roster, "has_bot_count").bind(required_count),
			"bot admission",
			_setup_timeout()
		):
			return false
	return true


func _wait_until(predicate: Callable, description: String, timeout_seconds: float) -> bool:
	var deadline := Time.get_ticks_msec() + int(maxf(timeout_seconds, 0.1) * 1000.0)
	while Time.get_ticks_msec() < deadline:
		if _connection_closed:
			_fail("connection closed while waiting to %s" % description)
			return false
		if bool(predicate.call()):
			return true
		await get_tree().process_frame
	_fail("timed out waiting to %s" % description)
	return false


func _is_in_lobby() -> bool:
	var lobby: Dictionary = room_session_controller.lobby_state_snapshot()
	return str(lobby.get("room_code", "")) != "" \
		and str(lobby.get("room_state", "")) == Constants.ROOM_STATE_LOBBY


func _is_in_game() -> bool:
	return room_session_controller.current_room_state() == Constants.ROOM_STATE_IN_GAME \
		and gameplay_session_controller.is_gameplay_active()


func _is_tooling_ready() -> bool:
	return connection_service != null and connection_service.is_tooling_ready()


func _measurement_is_recording() -> bool:
	return bool(gameplay_session_controller.get_measurement_state().get("recording", false))


func _measurement_is_stopped() -> bool:
	return !bool(gameplay_session_controller.get_measurement_state().get("recording", false))


func _measurement_export_finished() -> bool:
	return !gameplay_session_controller.get_latest_measurement_export_result().is_empty()


func _expected_human_count() -> int:
	var clients = scenario.get("clients", {})
	if !(clients is Dictionary):
		return 1
	return maxi(int(clients.get("visible", 1)) + int(clients.get("headless", 0)), 1)


func _room_config() -> Dictionary:
	var room = scenario.get("room", {})
	return room.duplicate(true) if room is Dictionary else {}


func _setup_timeout() -> float:
	return maxf(float(scenario.get("setup_timeout_seconds", 30.0)), 1.0)


func _scenario_id() -> String:
	return str(scenario.get("id", "runtime_scenario"))


func _status(state: String, fields: Dictionary = {}) -> void:
	status_writer.write(state, fields)


func _fail(message: String) -> void:
	if !_failure_message.is_empty():
		return
	_failure_message = message
	_terminal_status_written = true
	if phase_runner != null:
		phase_runner.release_actions()
	_status("failed", {"scenario_id": _scenario_id(), "error": message})
	get_tree().quit(1)


func _on_connection_closed() -> void:
	_connection_closed = true


func _exit_tree() -> void:
	if status_writer != null and !_terminal_status_written:
		status_writer.write("failed", {
			"scenario_id": _scenario_id(),
			"error": "client exited before the runtime scenario completed",
		})
