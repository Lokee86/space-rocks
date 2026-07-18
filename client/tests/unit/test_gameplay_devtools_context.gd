extends GutTest

const GameplayDevtoolsContext := preload("res://scripts/devtools/gameplay_devtools_context.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const PresentationEventCapture := preload("res://tests/unit/logging/presentation_event_capture.gd")


class FakeConnectionService extends RefCounted:
	func send_packet(_packet: Dictionary, _trace_id: String = "") -> void:
		pass


class FakeMeasurementCoordinator extends RefCounted:
	signal state_changed(state: Dictionary)
	signal error_received(error: Dictionary)

	var state := {
		"active_run_id": "",
		"recording": false,
		"pending_request_ids": {},
	}
	var start_calls := 0
	var stop_calls := 0
	var last_scenario_label := ""

	func get_state() -> Dictionary:
		return state.duplicate(true)

	func start(scenario_label: String = "") -> String:
		start_calls += 1
		last_scenario_label = scenario_label
		state["pending_request_ids"] = {"start": "start-request"}
		return "start-request"

	func stop() -> String:
		stop_calls += 1
		return "stop-request"

	func reset() -> String:
		return "reset-request"

	func set_state(next_state: Dictionary) -> void:
		state = next_state.duplicate(true)
		state_changed.emit(state)


func before_each() -> void:
	ClientLogger.reset_for_tests()


func after_each() -> void:
	ClientLogger.reset_for_tests()


func test_emits_devtools_enabled_after_successful_configuration() -> void:
	var writer := PresentationEventCapture.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var context := GameplayDevtoolsContext.new()

	context.configure(FakeConnectionService.new())

	assert_eq(writer.written_lines.size(), 1)
	var record: Dictionary = JSON.parse_string(writer.written_lines[0])
	assert_eq(record["event"], "devtools_enabled")


func test_does_not_emit_devtools_enabled_when_configuration_is_incomplete() -> void:
	var writer := PresentationEventCapture.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var context := GameplayDevtoolsContext.new()

	context.configure(null)

	assert_eq(writer.written_lines.size(), 0)


func test_emits_devtools_enabled_exactly_once_across_repeated_configuration() -> void:
	var writer := PresentationEventCapture.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var context := GameplayDevtoolsContext.new()

	context.configure(FakeConnectionService.new())
	context.configure(FakeConnectionService.new())

	assert_eq(writer.written_lines.size(), 1)
	var record: Dictionary = JSON.parse_string(writer.written_lines[0])
	assert_eq(record["event"], "devtools_enabled")


func test_measurement_actions_route_through_context_and_hotkey_uses_active_state() -> void:
	var context := GameplayDevtoolsContext.new()
	context.configure(null)
	var coordinator := FakeMeasurementCoordinator.new()
	context.configure_measurement(coordinator)

	context.devtools_window_controller.measurement_start_requested.emit("ui-scenario")
	assert_eq(coordinator.start_calls, 1)
	assert_eq(coordinator.last_scenario_label, "ui-scenario")
	assert_eq(context.devtools_window_controller.latest_measurement_state["pending_request_ids"]["start"], "start-request")

	coordinator.set_state({"active_run_id": "", "recording": false, "pending_request_ids": {"start": "request-start"}})
	context.hotkey_context.measurement_toggle_route.call()
	assert_eq(coordinator.stop_calls, 0)

	coordinator.set_state({"active_run_id": "run-1", "recording": true, "pending_request_ids": {}})
	context.hotkey_context.measurement_toggle_route.call()
	assert_eq(coordinator.stop_calls, 1)

	coordinator.set_state({"active_run_id": "", "recording": false, "pending_request_ids": {}})
	context.hotkey_context.measurement_toggle_route.call()
	assert_eq(coordinator.start_calls, 2)
