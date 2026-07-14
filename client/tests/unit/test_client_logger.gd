extends GutTest

const ClientLogger := preload("res://scripts/logging/logger.gd")
const TEST_LOG_DIR := "user://logger_test_output"
const TEST_LOG_PREFIX := "client-test"

class FakeWriter extends RefCounted:
	var configure_calls: Array = []
	var written_lines: Array[String] = []
	var close_calls := 0
	var enabled := false
	var current_path := ""

	func configure(base_dir: String, prefix: String, policy: Dictionary = {}) -> bool:
		configure_calls.append([base_dir, prefix])
		enabled = true
		current_path = base_dir.path_join("active").path_join("%s.jsonl.open" % prefix)
		return true

	func write_line(line: String) -> void:
		written_lines.append(line)

	func close() -> void:
		close_calls += 1
		enabled = false
		current_path = ""


func before_each() -> void:
	ClientLogger.reset_for_tests()


func after_each() -> void:
	_cleanup_test_log_files()
	ClientLogger.reset_for_tests()


func test_level_name_maps_known_levels_and_unknown_values() -> void:
	assert_eq(ClientLogger.level_name(ClientLogger.LEVEL_DEBUG), "debug")
	assert_eq(ClientLogger.level_name(ClientLogger.LEVEL_INFO), "info")
	assert_eq(ClientLogger.level_name(ClientLogger.LEVEL_WARN), "warn")
	assert_eq(ClientLogger.level_name(ClientLogger.LEVEL_ERROR), "error")
	assert_eq(ClientLogger.level_name(ClientLogger.LEVEL_OFF), "unknown")
	assert_eq(ClientLogger.level_name(-1), "unknown")


func test_build_record_includes_expected_fields() -> void:
	var fields := {
		"packet_id": "abc123",
		"attempt": 2,
	}

	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_WARN,
		"connection_retry_scheduled",
		"Retrying socket connection",
		fields
	)

	assert_eq(typeof(record["timestamp_unix_ms"]), TYPE_INT)
	assert_eq(record["level"], "warn")
	assert_eq(record["category"], ClientLogger.CATEGORY_NETWORK)
	assert_eq(record["event"], "connection_retry_scheduled")
	assert_eq(record["message"], "Retrying socket connection")
	assert_eq(record["fields"], fields)


func test_build_record_duplicates_fields_dictionary() -> void:
	var fields := {
		"nested": {
			"count": 1,
		},
	}

	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_PACKETS,
		ClientLogger.LEVEL_DEBUG,
		"packet_decoded",
		"",
		fields
	)

	fields["new_key"] = "new value"
	fields["nested"]["count"] = 99

	assert_false(record["fields"].has("new_key"))
	assert_eq(record["fields"]["nested"]["count"], 1)


func test_format_json_line_round_trips_to_dictionary() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_INFO,
		"log_message",
		"realtime protocol state reset",
		{
			"raw_bytes": 42,
		}
	)

	var json_line := ClientLogger.format_json_line(record)
	var parsed = JSON.parse_string(json_line)

	assert_true(parsed is Dictionary)
	assert_eq(parsed["category"], ClientLogger.CATEGORY_NETWORK)
	assert_eq(parsed["level"], "info")
	assert_eq(parsed["event"], "log_message")
	assert_eq(parsed["message"], "realtime protocol state reset")
	assert_eq(parsed["fields"]["raw_bytes"], 42.0)


func test_format_console_line_for_log_message_keeps_old_style_output() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_INFO,
		"log_message",
		"realtime protocol state reset"
	)

	assert_eq(
		ClientLogger.format_console_line(record),
		"[network][info] realtime protocol state reset"
	)


func test_format_console_line_for_named_event_includes_event_bracket() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_WARN,
		"packet_decode_failed",
		"Packet decode failed"
	)

	assert_eq(
		ClientLogger.format_console_line(record),
		"[network][warn][packet_decode_failed] Packet decode failed"
	)


func test_format_console_line_sorts_fields_deterministically() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_WARN,
		"packet_decode_failed",
		"Packet decode failed",
		{
			"raw_bytes": 42,
			"error": "Invalid JSON",
		}
	)

	assert_eq(
		ClientLogger.format_console_line(record),
		"[network][warn][packet_decode_failed] Packet decode failed error=Invalid JSON raw_bytes=42"
	)


func test_should_log_respects_default_level() -> void:
	ClientLogger.default_level = ClientLogger.LEVEL_INFO
	ClientLogger.category_levels = {}

	assert_false(ClientLogger._should_log(ClientLogger.CATEGORY_NETWORK, ClientLogger.LEVEL_DEBUG))
	assert_true(ClientLogger._should_log(ClientLogger.CATEGORY_NETWORK, ClientLogger.LEVEL_INFO))


func test_should_log_respects_category_override() -> void:
	ClientLogger.default_level = ClientLogger.LEVEL_ERROR
	ClientLogger.category_levels = {
		ClientLogger.CATEGORY_NETWORK: ClientLogger.LEVEL_DEBUG,
	}

	assert_true(ClientLogger._should_log(ClientLogger.CATEGORY_NETWORK, ClientLogger.LEVEL_DEBUG))
	assert_false(ClientLogger._should_log(ClientLogger.CATEGORY_GAME, ClientLogger.LEVEL_WARN))


func test_format_console_line_for_built_log_message_record_keeps_old_compatible_output() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_INFO,
		"log_message",
		"realtime protocol state reset"
	)

	assert_eq(
		ClientLogger.format_console_line(record),
		"[network][info] realtime protocol state reset"
	)


func test_event_records_use_provided_event_name() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_WARN,
		"packet_decode_failed",
		"Packet decode failed"
	)

	assert_eq(record["event"], "packet_decode_failed")
	assert_eq(
		ClientLogger.format_console_line(record),
		"[network][warn][packet_decode_failed] Packet decode failed"
	)


func test_network_event_uses_category_network() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_INFO,
		"realtime_connected",
		"Realtime connected"
	)

	assert_eq(record["category"], ClientLogger.CATEGORY_NETWORK)


func test_packets_event_uses_category_packets() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_PACKETS,
		ClientLogger.LEVEL_INFO,
		"packet_decoded",
		"Packet decoded"
	)

	assert_eq(record["category"], ClientLogger.CATEGORY_PACKETS)


func test_disabled_category_blocks_event_path_via_should_log_and_helpers() -> void:
	ClientLogger.default_level = ClientLogger.LEVEL_INFO
	ClientLogger.category_levels = {
		ClientLogger.CATEGORY_NETWORK: ClientLogger.LEVEL_OFF,
	}

	assert_false(ClientLogger._should_log(ClientLogger.CATEGORY_NETWORK, ClientLogger.LEVEL_ERROR))

	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_ERROR,
		"packet_decode_failed",
		"Packet decode failed"
	)

	assert_eq(record["category"], ClientLogger.CATEGORY_NETWORK)
	assert_eq(record["event"], "packet_decode_failed")


func test_configure_file_output_creates_a_file_in_test_directory() -> void:
	var configured := ClientLogger.configure_file_output(TEST_LOG_DIR, TEST_LOG_PREFIX)
	var path := ClientLogger.current_file_output_path()

	assert_true(configured)
	assert_true(ClientLogger._file_writer.enabled)
	assert_ne(path, "")
	assert_true(path.begins_with(TEST_LOG_DIR.path_join("active")))
	assert_true(path.ends_with(".jsonl.open"))
	assert_true(FileAccess.file_exists(path))


func test_emitted_structured_log_writes_jsonl_line_to_file() -> void:
	assert_true(ClientLogger.configure_file_output(TEST_LOG_DIR, TEST_LOG_PREFIX))
	var path := ClientLogger.current_file_output_path()

	ClientLogger.event(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_WARN,
		"packet_decode_failed",
		"Packet decode failed",
		{
			"raw_bytes": 42,
		}
	)
	ClientLogger.close_file_output()

	var file = FileAccess.open(path, FileAccess.READ)
	assert_ne(file, null)
	var contents := file.get_as_text().strip_edges()
	var parsed = JSON.parse_string(contents)

	assert_true(parsed is Dictionary)
	assert_eq(parsed["category"], ClientLogger.CATEGORY_NETWORK)
	assert_eq(parsed["event"], "packet_decode_failed")
	assert_eq(parsed["message"], "Packet decode failed")
	assert_eq(parsed["fields"]["raw_bytes"], 42.0)


func test_close_file_output_resets_file_state() -> void:
	assert_true(ClientLogger.configure_file_output(TEST_LOG_DIR, TEST_LOG_PREFIX))
	assert_true(ClientLogger._file_writer.enabled)
	assert_ne(ClientLogger.current_file_output_path(), "")

	ClientLogger.close_file_output()

	assert_false(ClientLogger._file_writer.enabled)
	assert_eq(ClientLogger.current_file_output_path(), "")


func test_file_output_delegates_to_replaceable_writer() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)

	assert_true(ClientLogger.configure_file_output("user://fake-logs", "fake-client"))
	assert_eq(writer.configure_calls, [["user://fake-logs", "fake-client"]])
	assert_eq(ClientLogger.current_file_output_path(), "user://fake-logs/active/fake-client.jsonl.open")

	ClientLogger.event(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_INFO,
		"fake_writer_event",
		"delegated",
		{"attempt": 1}
	)

	assert_eq(writer.written_lines.size(), 1)
	var parsed = JSON.parse_string(writer.written_lines[0])
	assert_eq(parsed["event"], "fake_writer_event")
	assert_eq(parsed["message"], "delegated")

	ClientLogger.close_file_output()
	assert_eq(writer.close_calls, 1)
	assert_false(writer.enabled)
	assert_eq(writer.current_path, "")


func _cleanup_test_log_files() -> void:
	ClientLogger.close_file_output()
	for subdir in ["active", "archive"]:
		var dir_path := TEST_LOG_DIR.path_join(subdir)
		DirAccess.make_dir_recursive_absolute(dir_path)
		var dir := DirAccess.open(dir_path)
		if dir == null:
			continue

		dir.list_dir_begin()
		var file_name := dir.get_next()
		while file_name != "":
			if !dir.current_is_dir() and file_name.begins_with(TEST_LOG_PREFIX):
				dir.remove(file_name)
			file_name = dir.get_next()
		dir.list_dir_end()
		DirAccess.remove_absolute(dir_path)
	DirAccess.remove_absolute(TEST_LOG_DIR)
