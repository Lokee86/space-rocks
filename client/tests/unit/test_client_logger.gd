extends GutTest

const ClientLogger := preload("res://scripts/logging/logger.gd")
const Contract := preload("res://scripts/generated/observability/contract_generated.gd")
const TEST_LOG_DIR := "user://logger_test_output"
const TEST_LOG_PREFIX := "client-test"

class FakeWriter extends RefCounted:
	var configure_calls: Array = []
	var written_lines: Array[String] = []
	var close_calls := 0
	var enabled := false
	var current_path := ""
	var failure_count := 0
	var last_failure_message := ""

	func configure(base_dir: String, prefix: String, _policy: Dictionary = {}) -> bool:
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
	assert_eq(ClientLogger.level_name(ClientLogger.LEVEL_CRITICAL), "critical")
	assert_eq(ClientLogger.level_name(ClientLogger.LEVEL_OFF), "unknown")


func test_build_record_uses_canonical_bridge_envelope() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_WARN,
		"connection_retry_scheduled",
		"Retrying socket connection",
		{"attempt": 2}
	)

	assert_eq(record["level"], "warn")
	assert_eq(record["category"], ClientLogger.CATEGORY_NETWORK)
	assert_eq(record["event"], Contract.EVENT_LOG_MESSAGE)
	assert_eq(record["message"], "Retrying socket connection")
	assert_eq(record["fields"]["attempt"], 2)
	assert_eq(record["fields"]["legacy_event"], "connection_retry_scheduled")
	assert_eq(record["service"], "client")
	assert_eq(record["schema_version"], Contract.SCHEMA_VERSION)
	assert_true(_valid_uuid(record["event_id"]))
	assert_true(_valid_uuid(record["service_instance_id"]))
	assert_false(record.has("timestamp_unix_ms"))


func test_canonical_record_contains_only_contract_top_level_fields() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_PACKETS,
		ClientLogger.LEVEL_INFO,
		"packet_decoded",
		"decoded",
		{"raw_bytes": 42}
	)
	for key in record:
		assert_true(Contract.ALLOWED_TOP_LEVEL_FIELDS.has(key), "unexpected canonical field: %s" % key)
	for key in Contract.REQUIRED_FIELDS:
		assert_true(record.has(key), "missing required canonical field: %s" % key)


func test_legacy_bridge_normalizes_container_fields_to_scalars() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_INFO,
		"webrtc_data_channel_ready",
		"ready",
		{"channels": ["reliable", "unreliable"]}
	)
	assert_eq(record["fields"]["channels"], '["reliable","unreliable"]')


func test_format_json_line_round_trips_to_dictionary() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_INFO,
		"log_message",
		"realtime protocol state reset",
		{"raw_bytes": 42}
	)
	var parsed = JSON.parse_string(ClientLogger.format_json_line(record))
	assert_true(parsed is Dictionary)
	assert_eq(parsed["event"], "log_message")
	assert_eq(parsed["message"], "realtime protocol state reset")
	assert_eq(parsed["fields"]["raw_bytes"], 42.0)


func test_format_console_line_keeps_legacy_style_output() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_INFO,
		"log_message",
		"realtime protocol state reset"
	)
	assert_eq(ClientLogger.format_console_line(record), "[network][info] realtime protocol state reset")


func test_format_console_line_sorts_fields_deterministically() -> void:
	var record := ClientLogger.build_record(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_WARN,
		"packet_decode_failed",
		"Packet decode failed",
		{"raw_bytes": 42, "error": "Invalid JSON"}
	)
	assert_eq(
		ClientLogger.format_console_line(record),
		"[network][warn] Packet decode failed error=Invalid JSON legacy_event=packet_decode_failed raw_bytes=42"
	)


func test_should_log_respects_default_and_category_override() -> void:
	ClientLogger.default_level = ClientLogger.LEVEL_INFO
	assert_false(ClientLogger._should_log(ClientLogger.CATEGORY_NETWORK, ClientLogger.LEVEL_DEBUG))
	assert_true(ClientLogger._should_log(ClientLogger.CATEGORY_NETWORK, ClientLogger.LEVEL_INFO))
	ClientLogger.category_levels = {ClientLogger.CATEGORY_NETWORK: ClientLogger.LEVEL_OFF}
	assert_false(ClientLogger._should_log(ClientLogger.CATEGORY_NETWORK, ClientLogger.LEVEL_ERROR))


func test_emit_canonical_uses_generated_metadata_and_scalar_fields() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	assert_true(ClientLogger.configure_file_output("user://fake-logs", "fake-client"))

	var result := ClientLogger.emit_canonical(
		Contract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Presentation contract failed",
		{},
		{
			"subsystem": "world_sync",
			"entity_kind": "pickup",
			"failure_mode": "wrong_scene_root",
			"expected_type": "PickupPresentation",
			"actual_type": "Control",
		}
	)

	assert_push_error_count(1)
	assert_true(result["accepted"])
	assert_eq(result["record"]["event"], Contract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION)
	assert_eq(result["record"]["level"], "error")
	assert_eq(result["record"]["category"], "client_presentation")
	assert_eq(result["record"]["retention_tier"], "diagnostic_report")
	assert_false(result["record"].has("trace_id"))
	assert_eq(result["record"]["fields"]["actual_type"], "Control")
	assert_eq(JSON.parse_string(writer.written_lines[0])["event"], Contract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION)

	var state_result := ClientLogger.emit_canonical(
		Contract.EVENT_CLIENT_PRESENTATION_STATE_INVALID,
		"Presentation state is invalid",
		{},
		{"subsystem": "world_sync", "failure_mode": "missing_state_field", "field_name": "scale"}
	)
	assert_true(state_result["accepted"])
	assert_eq(state_result["record"]["level"], "warn")
	assert_eq(state_result["record"]["category"], "client_presentation")
	assert_eq(state_result["record"]["retention_tier"], "operational")

	var unsafe_node := Node2D.new()
	var rejected := ClientLogger.emit_canonical(
		Contract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"unsafe",
		{},
		{"node": unsafe_node}
	)
	assert_false(rejected["accepted"])
	assert_eq(rejected["rejection_code"], Contract.REJECTION_INVALID_FIELD_TYPE)
	unsafe_node.free()


func test_emitted_log_writes_one_canonical_jsonl_line_to_replaceable_writer() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	assert_true(ClientLogger.configure_file_output("user://fake-logs", "fake-client"))
	ClientLogger.event(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_WARN,
		"packet_decode_failed",
		"Packet decode failed",
		{"raw_bytes": 42}
	)
	assert_eq(writer.written_lines.size(), 1)
	var parsed = JSON.parse_string(writer.written_lines[0])
	assert_eq(parsed["event"], Contract.EVENT_LOG_MESSAGE)
	assert_eq(parsed["service"], "client")
	assert_eq(parsed["fields"]["legacy_event"], "packet_decode_failed")
	assert_eq(parsed["fields"]["raw_bytes"], 42.0)
	assert_false(parsed.has("timestamp_unix_ms"))


func test_configure_and_close_file_output_preserve_status() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	assert_true(ClientLogger.configure_file_output("user://fake-logs", "fake-client"))
	assert_eq(writer.configure_calls, [["user://fake-logs", "fake-client"]])
	assert_eq(ClientLogger.current_file_output_path(), "user://fake-logs/active/fake-client.jsonl.open")
	assert_true(ClientLogger.file_output_status().has("emitter"))
	ClientLogger.close_file_output()
	assert_eq(writer.close_calls, 1)
	assert_false(writer.enabled)


func test_configure_file_output_creates_active_file() -> void:
	assert_true(ClientLogger.configure_file_output(TEST_LOG_DIR, TEST_LOG_PREFIX))
	var path := ClientLogger.current_file_output_path()
	assert_true(ClientLogger._file_writer.enabled)
	assert_ne(path, "")
	assert_true(path.begins_with(TEST_LOG_DIR.path_join("active")))
	assert_true(path.ends_with(".jsonl.open"))
	assert_true(FileAccess.file_exists(path))


func test_real_file_output_writes_canonical_jsonl_line() -> void:
	assert_true(ClientLogger.configure_file_output(TEST_LOG_DIR, TEST_LOG_PREFIX))
	var path := ClientLogger.current_file_output_path()
	ClientLogger.event(
		ClientLogger.CATEGORY_NETWORK,
		ClientLogger.LEVEL_WARN,
		"packet_decode_failed",
		"Packet decode failed",
		{"raw_bytes": 42}
	)
	ClientLogger.close_file_output()
	assert_false(FileAccess.file_exists(path))
	var archive_files := _matching_archive_files(TEST_LOG_PREFIX)
	assert_eq(archive_files.size(), 1)
	var file = FileAccess.open_compressed(
		archive_files[0],
		FileAccess.READ,
		FileAccess.COMPRESSION_GZIP
	)
	assert_ne(file, null)
	var parsed = JSON.parse_string(file.get_as_text().strip_edges())
	assert_true(parsed is Dictionary)
	assert_eq(parsed["event"], Contract.EVENT_LOG_MESSAGE)
	assert_eq(parsed["service"], "client")
	assert_eq(parsed["fields"]["legacy_event"], "packet_decode_failed")
	assert_eq(parsed["fields"]["raw_bytes"], 42.0)


func test_close_file_output_resets_real_writer_state() -> void:
	assert_true(ClientLogger.configure_file_output(TEST_LOG_DIR, TEST_LOG_PREFIX))
	assert_true(ClientLogger._file_writer.enabled)
	ClientLogger.close_file_output()
	assert_false(ClientLogger._file_writer.enabled)
	assert_eq(ClientLogger.current_file_output_path(), "")


func test_emit_canonical_uses_generated_event_definition() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var result := ClientLogger.emit_canonical(
		Contract.EVENT_CLIENT_STARTED,
		"",
		{},
		{"subsystem": "shell"}
	)

	assert_true(result["accepted"])
	var record: Dictionary = result["record"]
	assert_eq(record["event"], Contract.EVENT_CLIENT_STARTED)
	assert_eq(record["level"], "info")
	assert_eq(record["category"], "client_startup")
	assert_eq(record["retention_tier"], "operational")
	assert_eq(record["service"], "client")
	assert_eq(record["fields"]["subsystem"], "shell")
	assert_eq(writer.written_lines.size(), 1)


func test_emit_canonical_rejection_is_nonfatal() -> void:
	var result := ClientLogger.emit_canonical("event_that_does_not_exist")
	assert_false(result["accepted"])
	assert_eq(result["rejection_code"], Contract.REJECTION_UNKNOWN_EVENT)


func test_emit_canonical_applies_generated_level_thresholds() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var context := {"trace_id": "00000000-0000-4000-8000-000000000001"}
	var result := ClientLogger.emit_canonical(
		Contract.EVENT_CONNECTION_ATTEMPT_STARTED,
		"",
		context
	)

	assert_false(result["accepted"])
	assert_true(result.get("suppressed", false))
	assert_eq(writer.written_lines.size(), 0)

	ClientLogger.enable_debug()
	result = ClientLogger.emit_canonical(
		Contract.EVENT_CONNECTION_ATTEMPT_STARTED,
		"",
		context
	)
	assert_true(result["accepted"])
	assert_eq(result["record"]["event"], Contract.EVENT_CONNECTION_ATTEMPT_STARTED)
	assert_eq(writer.written_lines.size(), 1)


func test_emit_canonical_respects_level_off_and_category_overrides() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	ClientLogger.disable()
	var result := ClientLogger.emit_canonical(Contract.EVENT_CLIENT_STARTED)
	assert_true(result.get("suppressed", false))

	ClientLogger.set_default_level(ClientLogger.LEVEL_INFO)
	var category := str(Contract.EVENT_DEFINITIONS[Contract.EVENT_CLIENT_STARTED]["category"])
	ClientLogger.set_category_level(category, ClientLogger.LEVEL_OFF)
	result = ClientLogger.emit_canonical(Contract.EVENT_CLIENT_STARTED)
	assert_true(result.get("suppressed", false))
	assert_eq(writer.written_lines.size(), 0)


func _matching_archive_files(prefix: String) -> Array[String]:
	var files: Array[String] = []
	var archive_path := TEST_LOG_DIR.path_join("archive")
	var dir := DirAccess.open(archive_path)
	if dir == null:
		return files
	dir.list_dir_begin()
	while true:
		var entry := dir.get_next()
		if entry == "":
			break
		if !dir.current_is_dir() and entry.begins_with(prefix) and entry.ends_with(".jsonl.gz"):
			files.append(archive_path.path_join(entry))
	dir.list_dir_end()
	files.sort()
	return files


func _valid_uuid(value: String) -> bool:
	var regex := RegEx.new()
	regex.compile(Contract.UUID_REGEX)
	return regex.search(value) != null


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
