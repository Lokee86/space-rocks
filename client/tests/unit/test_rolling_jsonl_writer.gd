extends GutTest

const RollingJSONLWriter := preload("res://scripts/logging/rolling_jsonl_writer.gd")


class FakeHandle extends RefCounted:
	var lines: Array[String] = []
	var flush_calls := 0
	var close_calls := 0

	func store_line(line: String) -> void:
		lines.append(line)

	func flush() -> void:
		flush_calls += 1

	func close() -> void:
		close_calls += 1


class FakeFilesystemWriter extends RollingJSONLWriter:
	var fake_now := 123456
	var make_dir_paths: Array[String] = []
	var existing_paths: Array[String] = []
	var opened_paths: Array[String] = []
	var fake_handle := FakeHandle.new()

	func _current_time_unix_ms() -> int:
		return fake_now

	func _make_dir_recursive(path: String) -> Error:
		make_dir_paths.append(path)
		return OK

	func _file_exists(path: String) -> bool:
		existing_paths.append(path)
		return false

	func _open_file(path: String, _mode: int):
		opened_paths.append(path)
		return fake_handle


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
	assert_true(writer.current_path.begins_with("user://writer-test/client-"))
	assert_true(writer.current_path.ends_with(".jsonl"))


func test_configuration_uses_replaceable_clock_and_filesystem_boundaries() -> void:
	var writer := FakeFilesystemWriter.new()
	assert_true(writer.configure("user://fake-writer", "client"))

	assert_eq(writer.last_configured_at_unix_ms, 123456)
	assert_eq(writer.make_dir_paths, ["user://fake-writer"])
	assert_eq(writer.existing_paths, ["user://fake-writer/client-000001.jsonl"])
	assert_eq(writer.opened_paths, ["user://fake-writer/client-000001.jsonl"])
	assert_true(writer.enabled)

	writer.write_line("{\"event\":\"test\"}")
	assert_eq(writer.fake_handle.lines, ["{\"event\":\"test\"}"])
	writer.close()
	assert_eq(writer.fake_handle.close_calls, 1)
