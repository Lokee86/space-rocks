extends GutTest

const RollingJSONLWriter := preload("res://scripts/logging/rolling_jsonl_writer.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const RollingJSONLWriterTestSupport := preload("res://tests/unit/logging/rolling_jsonl_writer_test_support.gd")
const FakeFilesystemWriter := RollingJSONLWriterTestSupport.FakeFilesystemWriter
func before_each() -> void:
	_cleanup_real_writer_test_root()
func after_each() -> void:
	_cleanup_real_writer_test_root()
func _cleanup_real_writer_test_root() -> void:
	_remove_directory_contents("user://writer-test")
	DirAccess.remove_absolute("user://writer-test")


func _remove_directory_contents(path: String) -> void:
	var dir := DirAccess.open(path)
	if dir == null:
		return
	dir.list_dir_begin()
	while true:
		var entry := dir.get_next()
		if entry == "":
			break
		if entry == "." or entry == "..":
			continue
		var entry_path := path.path_join(entry)
		if dir.current_is_dir():
			_remove_directory_contents(entry_path)
			DirAccess.remove_absolute(entry_path)
		else:
			DirAccess.remove_absolute(entry_path)
	dir.list_dir_end()

func test_policy_is_defensively_copied_and_layout_paths_are_prepared() -> void:
	var writer := RollingJSONLWriter.new()
	var policy := {
		"segment_max_bytes": 1024,
		"segment_max_age": 60,
		"retention_max_age": 3600,
		"retention_max_bytes": 8192,
		"active_directory_name": "current",
		"archive_directory_name": "history",
	}
	assert_true(writer.configure("user://writer-test", "client", policy))
	policy["segment_max_bytes"] = 1
	policy["active_directory_name"] = "changed"
	assert_eq(writer.configuration["segment_max_bytes"], 1024)
	assert_eq(writer.configuration["active_directory_name"], "current")
	assert_eq(writer.active_directory_path, "user://writer-test/current")
	assert_eq(writer.archive_directory_path, "user://writer-test/history")
	assert_eq(writer.current_path, "user://writer-test/current/client.jsonl.open")
	writer.close()

func test_configuration_uses_generated_client_defaults_when_policy_is_empty() -> void:
	var writer := RollingJSONLWriter.new()
	assert_true(writer.configure("user://writer-test", "client"))
	assert_eq(writer.configuration["segment_max_bytes"], 16 * 1024 * 1024)
	assert_eq(writer.configuration["segment_max_age"], ObservabilityContract.FILE_LOGGING_MAX_ACTIVE_SEGMENT_AGE_SECONDS)
	assert_eq(writer.configuration["retention_max_age"], ObservabilityContract.RETENTION_DEFAULT_AGE_SECONDS_OPERATIONAL)
	assert_eq(writer.configuration["retention_max_bytes"], 250 * 1024 * 1024)
	assert_eq(writer.current_path, "user://writer-test/active/client.jsonl.open")
	writer.close()





func test_configuration_uses_replaceable_clock_and_filesystem_boundaries() -> void:
	var writer := FakeFilesystemWriter.new()
	assert_true(writer.configure("user://fake-writer", "client"))

	assert_eq(writer.last_configured_at_unix_ms, 123456)
	assert_eq(writer.make_dir_paths, ["user://fake-writer/active", "user://fake-writer/archive"])
	assert_eq(writer.existing_paths, ["user://fake-writer/active/client.jsonl.open"])
	assert_eq(writer.opened_paths, ["user://fake-writer/active/client.jsonl.open"])
	assert_true(writer.enabled)
	assert_eq(writer.current_path, "user://fake-writer/active/client.jsonl.open")

	writer.write_line("{\"event\":\"test\"}")
	assert_eq(writer.handles.size(), 1)
	assert_eq(writer.handles[0].lines, ["{\"event\":\"test\"}"])
	writer.close()
	assert_eq(writer.handles[0].close_calls, 1)


func test_configuration_recovers_existing_active_file_into_archive_and_opens_fresh_active() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 9000
	var active_path := "user://fake-writer/active/client.jsonl.open"
	writer.file_sizes[active_path] = 17
	writer.file_modified_times[active_path] = 7000

	assert_true(writer.configure("user://fake-writer", "client"))

	assert_eq(writer.renamed_paths, ["user://fake-writer/active/client.jsonl.open -> user://fake-writer/archive/client-7000-9000.jsonl"])
	assert_eq(writer.opened_paths, ["user://fake-writer/active/client.jsonl.open"])
	assert_eq(writer.file_sizes["user://fake-writer/archive/client-7000-9000.jsonl"], 17)
	assert_eq(writer.file_sizes[writer.current_path], 0)
	assert_eq(writer.current_path, active_path)
	assert_true(writer.enabled)
	assert_eq(writer.last_configured_at_unix_ms, 9000)
	assert_eq(writer.segment_started_at_unix_ms, 9000)


func test_configuration_recovers_existing_active_file_uses_current_clock_when_modified_time_is_unavailable() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 5555
	var active_path := "user://fake-writer/active/client.jsonl.open"
	writer.file_sizes[active_path] = 9

	assert_true(writer.configure("user://fake-writer", "client"))

	assert_eq(writer.renamed_paths, ["user://fake-writer/active/client.jsonl.open -> user://fake-writer/archive/client-5555-5555.jsonl"])
	assert_eq(writer.file_sizes["user://fake-writer/archive/client-5555-5555.jsonl"], 9)
	assert_eq(writer.current_path, active_path)


func test_configuration_applies_retention_age_cleanup_after_open() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 10000
	var policy := {
		"retention_max_age": 2,
		"retention_max_bytes": 0,
	}
	writer.file_sizes["user://fake-writer/archive/client-1000-1000.jsonl"] = 4
	writer.file_sizes["user://fake-writer/archive/client-5000-5000.jsonl"] = 5
	writer.file_sizes["user://fake-writer/archive/client-9000-9000.jsonl"] = 6
	writer.file_modified_times["user://fake-writer/archive/client-1000-1000.jsonl"] = 1000
	writer.file_modified_times["user://fake-writer/archive/client-5000-5000.jsonl"] = 5000
	writer.file_modified_times["user://fake-writer/archive/client-9000-9000.jsonl"] = 9000

	assert_true(writer.configure("user://fake-writer", "client", policy))

	assert_eq(writer.deleted_paths, [
		"user://fake-writer/archive/client-1000-1000.jsonl",
		"user://fake-writer/archive/client-5000-5000.jsonl",
	])
	assert_true(writer.file_sizes.has("user://fake-writer/archive/client-9000-9000.jsonl"))
	assert_true(writer.file_sizes.has(writer.current_path))


func test_configuration_applies_retention_byte_cleanup_oldest_first() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 10000
	var policy := {
		"retention_max_age": 0,
		"retention_max_bytes": 15,
	}
	writer.file_sizes["user://fake-writer/archive/client-a-1000-1000.jsonl"] = 10
	writer.file_sizes["user://fake-writer/archive/client-b-1000-1000.jsonl"] = 10
	writer.file_sizes["user://fake-writer/archive/client-c-1000-1000.jsonl"] = 10
	writer.file_modified_times["user://fake-writer/archive/client-b-1000-1000.jsonl"] = 1000
	writer.file_modified_times["user://fake-writer/archive/client-a-1000-1000.jsonl"] = 1000
	writer.file_modified_times["user://fake-writer/archive/client-c-1000-1000.jsonl"] = 1000

	assert_true(writer.configure("user://fake-writer", "client", policy))

	assert_eq(writer.deleted_paths, [
		"user://fake-writer/archive/client-a-1000-1000.jsonl",
		"user://fake-writer/archive/client-b-1000-1000.jsonl",
	])
	assert_true(writer.file_sizes.has("user://fake-writer/archive/client-c-1000-1000.jsonl"))
	assert_true(writer.file_sizes.has(writer.current_path))


func test_write_line_rotates_when_segment_size_would_be_exceeded() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 1000
	var policy := {
		"segment_max_bytes": 12,
		"segment_max_age": 0,
	}

	assert_true(writer.configure("user://fake-writer", "client", policy))
	writer.file_sizes[writer.current_path] = 10
	writer.fake_now = 2000

	writer.write_line("ab")

	assert_eq(writer.renamed_paths, ["user://fake-writer/active/client.jsonl.open -> user://fake-writer/archive/client-1000-2000.jsonl"])
	assert_eq(writer.opened_paths, [
		"user://fake-writer/active/client.jsonl.open",
		"user://fake-writer/active/client.jsonl.open",
	])
	assert_eq(writer.handles.size(), 2)
	assert_eq(writer.handles[0].close_calls, 1)
	assert_eq(writer.handles[0].lines, [])
	assert_eq(writer.handles[1].lines, ["ab"])
	assert_eq(writer.current_path, "user://fake-writer/active/client.jsonl.open")
	assert_eq(writer.file_sizes["user://fake-writer/archive/client-1000-2000.jsonl"], 10)
	assert_eq(writer.file_sizes[writer.current_path], 3)


func test_write_line_rotates_when_segment_age_expires() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 1000
	var policy := {
		"segment_max_bytes": 0,
		"segment_max_age": 2,
	}

	assert_true(writer.configure("user://fake-writer", "client", policy))
	writer.fake_now = 3500

	writer.write_line("age")

	assert_eq(writer.renamed_paths, ["user://fake-writer/active/client.jsonl.open -> user://fake-writer/archive/client-1000-3500.jsonl"])
	assert_eq(writer.opened_paths, [
		"user://fake-writer/active/client.jsonl.open",
		"user://fake-writer/active/client.jsonl.open",
	])
	assert_eq(writer.handles.size(), 2)
	assert_eq(writer.handles[0].close_calls, 1)
	assert_eq(writer.handles[1].lines, ["age"])
	assert_eq(writer.file_sizes["user://fake-writer/archive/client-1000-3500.jsonl"], 0)
	assert_eq(writer.file_sizes[writer.current_path], 4)
