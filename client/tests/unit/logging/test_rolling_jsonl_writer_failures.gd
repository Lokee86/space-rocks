extends GutTest

const RollingJSONLWriterTestSupport := preload("res://tests/unit/logging/rolling_jsonl_writer_test_support.gd")
const FakeFilesystemWriter := RollingJSONLWriterTestSupport.FakeFilesystemWriter

func test_configuration_failure_records_once_and_resets_on_next_configure_attempt() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fail_next_make_dir = true

	assert_false(writer.configure("user://fake-writer", "client"))
	assert_eq(writer.failure_count, 1)
	assert_true(writer.last_failure_message.contains("failed to create active log directory"))
	assert_eq(writer.failure_warnings.size(), 1)

	writer.fail_next_make_dir = true
	assert_false(writer.configure("user://fake-writer", "client"))
	assert_eq(writer.failure_count, 1)
	assert_true(writer.last_failure_message.contains("failed to create active log directory"))
	assert_eq(writer.failure_warnings.size(), 2)

func test_recovery_failure_records_and_disables_file_output() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fail_rename = true
	writer.file_sizes["user://fake-writer/active/client.jsonl.open"] = 4

	assert_false(writer.configure("user://fake-writer", "client"))
	assert_eq(writer.failure_count, 1)
	assert_true(writer.last_failure_message.contains("failed to recover interrupted active log file"))
	assert_false(writer.enabled)
	assert_eq(writer.failure_warnings.size(), 1)

func test_rotation_reopen_failure_records_and_disables_file_output() -> void:
	var writer := FakeFilesystemWriter.new()
	var policy := {
		"segment_max_bytes": 1,
		"retention_max_age": 0,
		"retention_max_bytes": 0,
	}

	assert_true(writer.configure("user://fake-writer", "client", policy))
	writer.fail_open_after_first_call = true
	writer.write_line("ab")

	assert_eq(writer.failure_count, 1)
	assert_true(writer.last_failure_message.contains("failed to reopen active log file after rotation"))
	assert_false(writer.enabled)
	assert_eq(writer.failure_warnings.size(), 1)

func test_write_line_failure_disables_file_output_and_records_failure() -> void:
	var writer := FakeFilesystemWriter.new()
	assert_true(writer.configure("user://fake-writer", "client"))
	writer.handles[0].error = ERR_CANT_CREATE

	writer.write_line("boom")

	assert_eq(writer.handles[0].lines, ["boom"])
	assert_eq(writer.handles[0].flush_calls, 1)
	assert_eq(writer.handles[0].close_calls, 1)
	assert_eq(writer.failure_count, 1)
	assert_true(writer.last_failure_message.contains("failed to write active log line"))
	assert_false(writer.enabled)
	assert_eq(writer.current_path, "")
	assert_eq(writer.failure_warnings.size(), 1)

func test_retention_deletion_failures_are_nonfatal_and_warn_once() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 10000
	writer.fail_delete_paths = [
		"user://fake-writer/archive/client-1000-1000.jsonl",
		"user://fake-writer/archive/client-2000-2000.jsonl",
	]
	writer.file_sizes["user://fake-writer/archive/client-1000-1000.jsonl"] = 4
	writer.file_sizes["user://fake-writer/archive/client-2000-2000.jsonl"] = 5
	writer.file_modified_times["user://fake-writer/archive/client-1000-1000.jsonl"] = 1000
	writer.file_modified_times["user://fake-writer/archive/client-2000-2000.jsonl"] = 2000
	var policy := {
		"retention_max_age": 1,
		"retention_max_bytes": 0,
	}

	assert_true(writer.configure("user://fake-writer", "client", policy))
	assert_eq(writer.failure_count, 2)
	assert_true(writer.last_failure_message.contains("failed to delete archived log file"))
	assert_eq(writer.failure_warnings.size(), 1)
	assert_true(writer.enabled)
	assert_true(writer.file_sizes.has("user://fake-writer/archive/client-1000-1000.jsonl"))
	assert_true(writer.file_sizes.has("user://fake-writer/archive/client-2000-2000.jsonl"))
