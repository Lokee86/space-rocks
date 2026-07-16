extends GutTest

const AppEntry := preload("res://scripts/shell/app_entry.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

class FakeWriter extends RefCounted:
	var enabled := true
	var current_path := "user://fake-logs/active/client.jsonl.open"
	var failure_count := 0
	var last_failure_message := ""
	var close_calls := 0
	var written_lines: Array[String] = []

	func write_line(line: String) -> void:
		written_lines.append(line)
		if written_lines.size() == 1:
			enabled = false
			current_path = ""
			failure_count = 1
			last_failure_message = "simulated file failure"

	func close() -> void:
		close_calls += 1
		enabled = false
		current_path = ""


func before_each() -> void:
	ClientLogger.reset_for_tests()


func after_each() -> void:
	ClientLogger.reset_for_tests()


func test_status_exposes_writer_runtime_state() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var status := ClientLogger.file_output_status()
	assert_eq(status.size(), 5)
	assert_true(status["enabled"])
	assert_eq(status["current_path"], "user://fake-logs/active/client.jsonl.open")
	assert_eq(status["failure_count"], 0)
	assert_eq(status["last_failure_message"], "")
	assert_eq(status["emitter"], {
		"accepted_count": 0,
		"rejected_count": 0,
		"redacted_count": 0,
		"write_failure_count": 0,
		"last_rejection_code": "",
		"last_write_error": "",
	})
	ClientLogger.shell_info("first line disables file output")
	status = ClientLogger.file_output_status()
	assert_false(status["enabled"])
	assert_eq(status["current_path"], "")
	assert_eq(status["failure_count"], 1)
	assert_eq(status["last_failure_message"], "simulated file failure")


func test_console_logging_path_continues_after_file_output_degrades() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	ClientLogger.shell_info("first")
	ClientLogger.shell_info("second")
	assert_eq(writer.written_lines.size(), 2)
	assert_eq(ClientLogger.file_output_status()["failure_count"], 1)


func test_application_exit_closes_static_writer_and_repeated_close_is_safe() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var app_entry = AppEntry.new()
	app_entry._exit_tree()
	app_entry._exit_tree()
	assert_eq(writer.close_calls, 2)
	assert_false(writer.enabled)
	assert_eq(writer.current_path, "")
	app_entry.free()

func test_app_entry_uses_canonical_startup_lifecycle_and_degradation_events() -> void:
	var source := FileAccess.get_file_as_string("res://scripts/shell/app_entry.gd")
	assert_true(source.contains("ClientLogger.emit_canonical(ObservabilityContract.EVENT_CLIENT_STARTING)"))
	assert_true(source.contains("ClientLogger.emit_canonical(ObservabilityContract.EVENT_CLIENT_STARTED)"))
	var starting_position := source.find("ObservabilityContract.EVENT_CLIENT_STARTING")
	var started_position := source.find("ObservabilityContract.EVENT_CLIENT_STARTED")

	assert_gte(starting_position, 0)
	assert_gte(started_position, 0)
	assert_lt(starting_position, started_position)
	assert_true(source.contains("ObservabilityContract.EVENT_OBSERVABILITY_UNAVAILABLE"))
	assert_true(source.contains('"subsystem": "file_logging"'))
	assert_true(source.contains('"failure_mode": "configure_failed"'))
	assert_false(source.contains("App entry booted"))
	assert_false(source.contains("Client structured log file output unavailable"))