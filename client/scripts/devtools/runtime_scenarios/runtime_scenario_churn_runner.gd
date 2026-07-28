extends RefCounted
class_name RuntimeScenarioChurnRunner

const Constants := preload("res://scripts/generated/constants/constants.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")

var scenario: Dictionary = {}
var role := ""
var client_id := ""
var room_session_controller
var gameplay_session_controller
var connection_service
var debug_flow
var roster
var phase_runner
var status_writer
var connection_closed: Callable


func configure(
	scenario_value: Dictionary,
	role_value: String,
	client_id_value: String,
	room_session_controller_ref,
	gameplay_session_controller_ref,
	connection_service_ref,
	debug_flow_ref,
	roster_ref,
	phase_runner_ref,
	status_writer_ref,
	connection_closed_ref: Callable
) -> void:
	scenario = scenario_value.duplicate(true)
	role = role_value
	client_id = client_id_value
	room_session_controller = room_session_controller_ref
	gameplay_session_controller = gameplay_session_controller_ref
	connection_service = connection_service_ref
	debug_flow = debug_flow_ref
	roster = roster_ref
	phase_runner = phase_runner_ref
	status_writer = status_writer_ref
	connection_closed = connection_closed_ref


func run(rounds: Array) -> Dictionary:
	var match_ids: Array[String] = []
	var measurement_reports: Array[String] = []
	for round_index in range(rounds.size()):
		var round_value = rounds[round_index]
		if !(round_value is Dictionary):
			return _failure("round %d is not an object" % (round_index + 1))
		var round: Dictionary = round_value
		var round_number := round_index + 1
		var start_result: Dictionary = await _start_round(round, round_number)
		if !bool(start_result.get("ok", false)):
			return start_result
		var match_id := str(start_result.get("match_id", ""))
		if match_id.is_empty():
			return _failure("round %d did not publish a match id" % round_number)
		if match_ids.has(match_id):
			return _failure("round %d reused match id %s" % [round_number, match_id])

		var setup_value = round.get("setup", {})
		var setup: Dictionary = setup_value if setup_value is Dictionary else {}
		var setup_result: Dictionary = await phase_runner.prepare(setup)
		if !bool(setup_result.get("ok", false)):
			return setup_result

		var measurement_result: Dictionary = await _start_measurement(round, round_number, match_id)
		if !bool(measurement_result.get("ok", false)):
			return measurement_result
		for phase_value in round.get("phases", []):
			if !(phase_value is Dictionary):
				continue
			var phase_result: Dictionary = await phase_runner.run(phase_value)
			if !bool(phase_result.get("ok", false)):
				return phase_result
			if _connection_is_closed():
				return _failure("connection closed during round %d" % round_number)

		phase_runner.release_actions()
		var end_result: Dictionary = await _end_round(round_number, match_id)
		if !bool(end_result.get("ok", false)):
			return end_result
		var export_result: Dictionary = await _stop_measurement(round_number)
		if !bool(export_result.get("ok", false)):
			return export_result
		match_ids.append(match_id)
		measurement_reports.append(str(export_result.get("path", "")))
		_status("round_completed", {
			"round": round_number,
			"match_id": match_id,
			"measurement_report": str(export_result.get("path", "")),
		})

		if round_index < rounds.size() - 1:
			var lobby_result: Dictionary = await _return_to_lobby(round_number)
			if !bool(lobby_result.get("ok", false)):
				return lobby_result

	return {
		"ok": true,
		"match_ids": match_ids,
		"measurement_reports": measurement_reports,
		"rounds_completed": rounds.size(),
	}


func _start_round(round: Dictionary, round_number: int) -> Dictionary:
	_status("round_starting", {"round": round_number, "name": str(round.get("name", ""))})
	if role == "coordinator":
		if !await _wait_until(Callable(roster, "humans_joined"), "round %d humans" % round_number):
			return _failure("round %d did not regain all human players" % round_number)
	room_session_controller.request_ready(true)
	if role == "coordinator":
		if !await _wait_until(Callable(roster, "lobby_can_start"), "round %d readiness" % round_number):
			return _failure("round %d did not become startable" % round_number)
		room_session_controller.request_start_game()
	if !await _wait_until(Callable(self, "_is_in_game"), "round %d match start" % round_number):
		return _failure("round %d did not enter gameplay" % round_number)
	if !await _wait_until(Callable(self, "_is_tooling_ready"), "round %d tooling" % round_number):
		return _failure("round %d tooling did not become ready" % round_number)
	if role == "coordinator":
		debug_flow.set_lives(
			DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS,
			"",
			maxi(int(round.get("lives", 2)), 1)
		)
		await Engine.get_main_loop().create_timer(0.25).timeout
	var match_id: String = str(room_session_controller.current_match_id())
	_status("round_started", {"round": round_number, "match_id": match_id})
	return {"ok": true, "match_id": match_id}


func _end_round(round_number: int, match_id: String) -> Dictionary:
	_status("round_ending", {"round": round_number, "match_id": match_id})
	if role == "coordinator":
		debug_flow.clear_bullets()
		debug_flow.set_lives(DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, "", 0)
		await Engine.get_main_loop().create_timer(0.25).timeout
		debug_flow.kill_player(DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, "")
	if !await _wait_until(Callable(self, "_is_game_over"), "round %d game over" % round_number):
		return _failure("round %d did not reach game over" % round_number)
	var result: Dictionary = room_session_controller.current_match_result()
	if str(result.get(Packets.FIELD_MATCH_ID, "")) != match_id:
		return _failure("round %d result did not match active match" % round_number)
	var players = result.get(Packets.FIELD_PLAYERS, [])
	if !(players is Array) or players.size() != _expected_participant_count():
		return _failure("round %d result did not contain all participants" % round_number)
	return {"ok": true}


func _return_to_lobby(round_number: int) -> Dictionary:
	_status("lobby_return_started", {"round": round_number})
	if role == "coordinator":
		connection_service.send_return_to_lobby_request()
	if !await _wait_until(Callable(self, "_is_in_lobby"), "round %d lobby return" % round_number):
		return _failure("round %d did not return to lobby" % round_number)
	var lobby: Dictionary = room_session_controller.lobby_state_snapshot()
	if bool(lobby.get("all_members_ready", true)):
		return _failure("round %d did not clear lobby readiness" % round_number)
	_status("lobby_return_completed", {"round": round_number})
	return {"ok": true}


func _start_measurement(round: Dictionary, round_number: int, match_id: String) -> Dictionary:
	var request_id: String = gameplay_session_controller.start_measurement(
		"%s-round-%d" % [str(scenario.get("id", "match_churn")), round_number],
		{
			"scenario_id": str(scenario.get("id", "match_churn")),
			"client_id": client_id,
			"role": role,
			"round": round_number,
			"round_name": str(round.get("name", "")),
			"match_id": match_id,
		}
	)
	if request_id.is_empty():
		return _failure("round %d measurement start was rejected" % round_number)
	if !await _wait_until(Callable(self, "_measurement_is_recording"), "round %d measurement start" % round_number):
		return _failure("round %d measurement did not start" % round_number)
	return {"ok": true}


func _stop_measurement(round_number: int) -> Dictionary:
	if gameplay_session_controller.stop_measurement().is_empty():
		return _failure("round %d measurement stop was rejected" % round_number)
	if !await _wait_until(Callable(self, "_measurement_is_stopped"), "round %d measurement stop" % round_number):
		return _failure("round %d measurement did not stop" % round_number)
	if !await _wait_until(Callable(self, "_measurement_export_finished"), "round %d measurement export" % round_number):
		return _failure("round %d measurement export did not finish" % round_number)
	var export_result: Dictionary = gameplay_session_controller.get_latest_measurement_export_result()
	if !bool(export_result.get("success", false)):
		return _failure("round %d measurement export failed: %s" % [
			round_number,
			str(export_result.get("error", "unknown error")),
		])
	return {"ok": true, "path": ProjectSettings.globalize_path(str(export_result.get("path", "")))}


func _wait_until(predicate: Callable, description: String) -> bool:
	var timeout_seconds := maxf(float(scenario.get("setup_timeout_seconds", 30.0)), 1.0)
	var deadline := Time.get_ticks_msec() + int(timeout_seconds * 1000.0)
	while Time.get_ticks_msec() < deadline:
		if _connection_is_closed():
			return false
		if bool(predicate.call()):
			return true
		await Engine.get_main_loop().process_frame
	_status("wait_failed", {"description": description})
	return false


func _connection_is_closed() -> bool:
	return connection_closed.is_valid() and bool(connection_closed.call())


func _is_in_game() -> bool:
	return room_session_controller.current_room_state() == Constants.ROOM_STATE_IN_GAME \
		and gameplay_session_controller.is_gameplay_active()


func _is_game_over() -> bool:
	return room_session_controller.current_room_state() == Constants.ROOM_STATE_GAME_OVER


func _is_in_lobby() -> bool:
	return room_session_controller.current_room_state() == Constants.ROOM_STATE_LOBBY \
		and !gameplay_session_controller.is_gameplay_active()


func _is_tooling_ready() -> bool:
	return connection_service != null and connection_service.is_tooling_ready()


func _measurement_is_recording() -> bool:
	return bool(gameplay_session_controller.get_measurement_state().get("recording", false))


func _measurement_is_stopped() -> bool:
	return !bool(gameplay_session_controller.get_measurement_state().get("recording", false))


func _measurement_export_finished() -> bool:
	return !gameplay_session_controller.get_latest_measurement_export_result().is_empty()


func _expected_participant_count() -> int:
	var clients = scenario.get("clients", {})
	var human_count := 1
	if clients is Dictionary:
		human_count = maxi(int(clients.get("visible", 1)) + int(clients.get("headless", 0)), 1)
	return human_count + maxi(int(scenario.get("bots", 0)), 0)


func _status(state: String, fields: Dictionary = {}) -> void:
	status_writer.write(state, fields)


func _failure(message: String) -> Dictionary:
	return {"ok": false, "error": message}
