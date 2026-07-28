extends GutTest

const ConfigScript := preload("res://scripts/devtools/runtime_scenarios/runtime_scenario_config.gd")
const RosterScript := preload("res://scripts/devtools/runtime_scenarios/runtime_scenario_roster.gd")
const StatusWriterScript := preload("res://scripts/devtools/runtime_scenarios/runtime_scenario_status_writer.gd")
const PhaseRunnerScript := preload("res://scripts/devtools/runtime_scenarios/runtime_scenario_phase_runner.gd")
const RoundsScript := preload("res://scripts/devtools/runtime_scenarios/runtime_scenario_rounds.gd")

const SCENARIO_PATH := "user://runtime_scenario_support_test.json"
const STATUS_PATH := "user://runtime_scenario_support_status.json"


class FakeRoomSessionController:
	var snapshot: Dictionary = {}

	func lobby_state_snapshot() -> Dictionary:
		return snapshot.duplicate(true)


class FakeDevConnectionService:
	var spawn_results: Array[Dictionary] = []

	func send_spawn_from_placement_result(result: Dictionary) -> void:
		spawn_results.append(result.duplicate(true))


class FakeStatusWriter:
	var states: Array[String] = []

	func write(state: String, _fields: Dictionary = {}) -> bool:
		states.append(state)
		return true


func after_each() -> void:
	for path in [SCENARIO_PATH, STATUS_PATH]:
		if FileAccess.file_exists(path):
			DirAccess.remove_absolute(ProjectSettings.globalize_path(path))


func test_config_loads_valid_client_arguments() -> void:
	_write_json(SCENARIO_PATH, {
		"id": "scenario-test",
		"seed": 7,
		"phases": [{"name": "warmup", "duration_seconds": 1}],
	})
	var config: Dictionary = ConfigScript.from_arguments([
		"--runtime-scenario=%s" % ProjectSettings.globalize_path(SCENARIO_PATH),
		"--runtime-scenario-role=coordinator",
		"--runtime-scenario-client-id=client-1",
		"--runtime-scenario-status=%s" % ProjectSettings.globalize_path(STATUS_PATH),
	])

	assert_true(bool(config.get("enabled", false)))
	assert_true(bool(config.get("valid", false)))
	assert_eq(config.get("id"), "scenario-test")
	assert_eq(config.get("role"), "coordinator")
	assert_eq(config.get("client_id"), "client-1")


func test_config_rejects_unknown_role() -> void:
	_write_json(SCENARIO_PATH, {
		"id": "scenario-test",
		"seed": 7,
		"phases": [{"name": "warmup", "duration_seconds": 1}],
	})
	var config: Dictionary = ConfigScript.from_arguments([
		"--runtime-scenario=%s" % ProjectSettings.globalize_path(SCENARIO_PATH),
		"--runtime-scenario-role=observer",
		"--runtime-scenario-client-id=client-1",
		"--runtime-scenario-status=%s" % ProjectSettings.globalize_path(STATUS_PATH),
	])

	assert_false(bool(config.get("valid", true)))
	assert_string_contains(str(config.get("error", "")), "coordinator or participant")


func test_status_writer_persists_terminal_state() -> void:
	var writer = StatusWriterScript.new()
	writer.configure(STATUS_PATH, "participant-1", "participant")
	assert_true(writer.write("completed", {"measurement_report": "report.json"}))
	var payload: Dictionary = JSON.parse_string(FileAccess.get_file_as_string(STATUS_PATH))

	assert_eq(payload.get("state"), "completed")
	assert_eq(payload.get("client_id"), "participant-1")
	assert_eq(payload.get("measurement_report"), "report.json")


func test_phase_runner_prepares_deterministic_asteroid_pressure() -> void:
	var service := FakeDevConnectionService.new()
	var writer := FakeStatusWriter.new()
	var runner = PhaseRunnerScript.new()
	runner.role = "coordinator"
	runner.dev_connection_service = service
	runner.status_writer = writer

	var result: Dictionary = await runner.prepare({
		"asteroid_spawns": 3,
		"settle_seconds": 0.0,
	})

	assert_true(bool(result.get("ok", false)))
	assert_eq(service.spawn_results.size(), 3)
	assert_eq(service.spawn_results[0].get("action_name"), &"spawn_asteroid")
	assert_true(service.spawn_results[0].get("server_position") is Vector2)
	assert_true(service.spawn_results[0].get("direction") is Vector2)
	assert_eq(writer.states, ["setup_started", "setup_completed"])


func test_repeated_rounds_expand_with_unique_names() -> void:
	var rounds: Array = RoundsScript.expand([
		{
			"name": "soak-cycle",
			"repeat": 3,
			"phases": [{"name": "pressure", "duration_seconds": 1}],
		}
	])

	assert_eq(rounds.size(), 3)
	assert_eq(rounds[0].get("name"), "soak-cycle-001")
	assert_eq(rounds[2].get("name"), "soak-cycle-003")
	assert_false(rounds[0].has("repeat"))


func test_roster_counts_humans_bots_and_other_player() -> void:
	var controller := FakeRoomSessionController.new()
	controller.snapshot = {
		"local_player_id": "Player-1",
		"can_start_game": true,
		"members": [
			{"player_id": "Player-1", "is_bot": false, "ready": true},
			{"player_id": "Player-2", "is_bot": false, "ready": true},
			{"player_id": "Player-3", "is_bot": true},
		],
	}
	var roster = RosterScript.new()
	roster.configure(controller, 2)

	assert_true(roster.humans_joined())
	assert_true(roster.lobby_can_start())
	assert_true(roster.local_member_ready())
	assert_eq(roster.human_count(), 2)
	assert_eq(roster.bot_count(), 1)
	assert_eq(roster.other_human_player_id(), "Player-2")


func _write_json(path: String, payload: Dictionary) -> void:
	var file := FileAccess.open(path, FileAccess.WRITE)
	file.store_string(JSON.stringify(payload))
	file.close()
