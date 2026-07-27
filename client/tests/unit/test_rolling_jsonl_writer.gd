extends GutTest

const RollingJSONLWriter := preload("res://scripts/logging/rolling_jsonl_writer.gd")
const LogArchiveMaintenance := preload("res://scripts/logging/log_archive_maintenance.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const RollingJSONLWriterTestSupport := preload("res://tests/unit/logging/rolling_jsonl_writer_test_support.gd")
const FakeFilesystemWriter := RollingJSONLWriterTestSupport.FakeFilesystemWriter

class BlockingLogArchiveMaintenance extends LogArchiveMaintenance:
	var maintenance_started := Semaphore.new()
	var release_maintenance := Semaphore.new()

	func _perform_maintenance(
		_archive_paths_to_compress: Array[String],
		_archive_directory_path: String,
		_configured_prefix: String,
		_configuration: Dictionary
	) -> Array[String]:
		maintenance_started.post()
		release_maintenance.wait()
		return []

class BlockingMaintenanceWriter extends RollingJSONLWriter:
	var maintenance := BlockingLogArchiveMaintenance.new()

	func _create_startup_maintenance():
		return maintenance

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
		"retention_max_files": 12,
		"compression_enabled": false,
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
	assert_true(writer.current_path.begins_with("user://writer-test/current/client-"))
	assert_true(writer.current_path.ends_with(".jsonl.open"))
	writer.close()

func test_configuration_uses_generated_client_defaults_when_policy_is_empty() -> void:
	var writer := RollingJSONLWriter.new()
	assert_true(writer.configure("user://writer-test", "client"))
	assert_eq(writer.configuration["segment_max_bytes"], 16 * 1024 * 1024)
	assert_eq(writer.configuration["segment_max_age"], ObservabilityContract.FILE_LOGGING_MAX_ACTIVE_SEGMENT_AGE_SECONDS)
	assert_eq(writer.configuration["retention_max_age"], ObservabilityContract.RETENTION_DEFAULT_AGE_SECONDS_OPERATIONAL)
	assert_eq(writer.configuration["retention_max_bytes"], 250 * 1024 * 1024)
	assert_eq(writer.configuration["retention_max_files"], RollingJSONLWriter.DEFAULT_RETENTION_MAX_FILES)
	assert_eq(writer.configuration["compression_enabled"], ObservabilityContract.FILE_LOGGING_COMPRESSION_ENABLED)
	assert_true(writer.configuration["startup_maintenance_async"])
	assert_true(writer.current_path.begins_with("user://writer-test/active/client-"))
	assert_true(writer.current_path.ends_with(".jsonl.open"))
	writer.close()





func test_configuration_uses_replaceable_clock_and_filesystem_boundaries() -> void:
	var writer := FakeFilesystemWriter.new()
	assert_true(writer.configure("user://fake-writer", "client"))

	assert_eq(writer.last_configured_at_unix_ms, 123456)
	assert_eq(writer.make_dir_paths, ["user://fake-writer/active", "user://fake-writer/archive"])
	assert_eq(writer.existing_paths, ["user://fake-writer/active/client-4242.jsonl.open"])
	assert_eq(writer.opened_paths, ["user://fake-writer/active/client-4242.jsonl.open"])
	assert_true(writer.enabled)
	assert_eq(writer.current_path, "user://fake-writer/active/client-4242.jsonl.open")

	writer.write_line("{\"event\":\"test\"}")
	assert_eq(writer.handles.size(), 1)
	assert_eq(writer.handles[0].lines, ["{\"event\":\"test\"}"])
	writer.close()
	assert_eq(writer.handles[0].close_calls, 1)


func test_clean_close_archives_process_segment_and_reopens_fresh() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 7000
	assert_true(writer.configure("user://fake-writer", "client", {"compression_enabled": false}))
	writer.file_sizes[writer.current_path] = 17

	writer.fake_now = 8000
	writer.close()
	var archive_path := "user://fake-writer/archive/client-4242-7000-8000.jsonl"
	assert_true(writer.file_sizes.has(archive_path))
	assert_false(writer.file_sizes.has("user://fake-writer/active/client-4242.jsonl.open"))
	assert_eq(writer.clean_marker_write_calls, 0)

	writer.fake_now = 9000
	assert_true(writer.configure("user://fake-writer", "client", {"compression_enabled": false}))
	assert_eq(writer.current_path, "user://fake-writer/active/client-4242.jsonl.open")
	assert_eq(writer.file_sizes[writer.current_path], 0)
	assert_eq(writer.segment_started_at_unix_ms, 9000)


func test_startup_maintenance_does_not_block_fresh_active_logging() -> void:
	var active_directory := "user://writer-test/active"
	var archive_directory := "user://writer-test/archive"
	assert_eq(DirAccess.make_dir_recursive_absolute(active_directory), OK)
	assert_eq(DirAccess.make_dir_recursive_absolute(archive_directory), OK)
	var stale_path := active_directory.path_join("client-999999.jsonl.open")
	var previous_active := FileAccess.open(stale_path, FileAccess.WRITE)
	assert_ne(previous_active, null)
	if previous_active != null:
		previous_active.store_line("previous-session")
		previous_active.close()

	var writer := BlockingMaintenanceWriter.new()
	assert_true(writer.configure("user://writer-test", "client", {
		"compression_enabled": false,
		"retention_max_age": 0,
		"retention_max_bytes": 0,
		"startup_maintenance_async": true,
	}))
	writer.maintenance.maintenance_started.wait()

	assert_true(writer.enabled)
	assert_ne(writer.current_path, stale_path)
	assert_true(writer.startup_maintenance_status()["running"])
	writer.write_line("startup-event")

	var active_reader := FileAccess.open(writer.current_path, FileAccess.READ)
	assert_ne(active_reader, null)
	if active_reader != null:
		assert_eq(active_reader.get_as_text(), "startup-event\n")
		active_reader.close()

	writer.maintenance.release_maintenance.post()
	writer.wait_for_startup_maintenance()
	var maintenance_status := writer.startup_maintenance_status()
	assert_false(maintenance_status["running"])
	assert_true(maintenance_status["completed"])
	writer.close()


func test_async_startup_recovery_compresses_previous_segment_after_active_open() -> void:
	var active_directory := "user://writer-test/active"
	var archive_directory := "user://writer-test/archive"
	assert_eq(DirAccess.make_dir_recursive_absolute(active_directory), OK)
	assert_eq(DirAccess.make_dir_recursive_absolute(archive_directory), OK)
	var stale_path := active_directory.path_join("client-999999.jsonl.open")
	var previous_active := FileAccess.open(stale_path, FileAccess.WRITE)
	assert_ne(previous_active, null)
	if previous_active != null:
		previous_active.store_line("previous-session")
		previous_active.close()

	var writer := RollingJSONLWriter.new()
	assert_true(writer.configure("user://writer-test", "client", {
		"compression_enabled": true,
		"retention_max_age": 0,
		"retention_max_bytes": 0,
		"startup_maintenance_async": true,
	}))
	writer.write_line("startup-event")
	writer.wait_for_startup_maintenance()

	var active_reader := FileAccess.open(writer.current_path, FileAccess.READ)
	assert_ne(active_reader, null)
	if active_reader != null:
		assert_eq(active_reader.get_as_text(), "startup-event\n")
		active_reader.close()

	var archive_files := _archive_files(archive_directory)
	assert_eq(archive_files.size(), 1)
	assert_true(archive_files[0].ends_with(".jsonl.gz"))
	if archive_files.size() == 1:
		var archive := FileAccess.open_compressed(
			archive_files[0],
			FileAccess.READ,
			FileAccess.COMPRESSION_GZIP
		)
		assert_ne(archive, null)
		if archive != null:
			assert_eq(archive.get_as_text(), "previous-session\n")
			archive.close()
	writer.close()


func test_configuration_recovers_clean_legacy_active_segment() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 9000
	writer.clean_marker_exists = true
	var legacy_path := "user://fake-writer/active/client.jsonl.open"
	writer.file_sizes[legacy_path] = 17
	writer.file_modified_times[legacy_path] = 7000

	assert_true(writer.configure("user://fake-writer", "client", {"compression_enabled": false}))

	assert_eq(writer.renamed_paths, ["user://fake-writer/active/client.jsonl.open -> user://fake-writer/archive/client-7000-9000.jsonl"])
	assert_eq(writer.clean_marker_remove_calls, 1)
	assert_false(writer.clean_marker_exists)
	assert_true(writer.file_sizes.has("user://fake-writer/archive/client-7000-9000.jsonl"))
	assert_eq(writer.current_path, "user://fake-writer/active/client-4242.jsonl.open")


func test_configuration_leaves_unmarked_legacy_active_segment_untouched() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 9000
	var legacy_path := "user://fake-writer/active/client.jsonl.open"
	writer.file_sizes[legacy_path] = 17
	writer.file_modified_times[legacy_path] = 7000

	assert_true(writer.configure("user://fake-writer", "client", {"compression_enabled": false}))

	assert_true(writer.renamed_paths.is_empty())
	assert_true(writer.file_sizes.has(legacy_path))
	assert_eq(writer.current_path, "user://fake-writer/active/client-4242.jsonl.open")


func test_configuration_recovers_dead_process_active_file_into_archive_and_opens_fresh_active() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 9000
	var stale_path := "user://fake-writer/active/client-9001.jsonl.open"
	writer.file_sizes[stale_path] = 17
	writer.file_modified_times[stale_path] = 7000

	assert_true(writer.configure("user://fake-writer", "client", {"compression_enabled": false}))

	assert_eq(writer.renamed_paths, ["user://fake-writer/active/client-9001.jsonl.open -> user://fake-writer/archive/client-9001-7000-9000.jsonl"])
	assert_eq(writer.opened_paths, ["user://fake-writer/active/client-4242.jsonl.open"])
	assert_eq(writer.file_sizes["user://fake-writer/archive/client-9001-7000-9000.jsonl"], 17)
	assert_eq(writer.file_sizes[writer.current_path], 0)
	assert_eq(writer.current_path, "user://fake-writer/active/client-4242.jsonl.open")
	assert_true(writer.enabled)
	assert_eq(writer.last_configured_at_unix_ms, 9000)
	assert_eq(writer.segment_started_at_unix_ms, 9000)


func test_configuration_leaves_running_process_active_segment_untouched() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 9000
	writer.running_process_ids = [9001]
	var active_path := "user://fake-writer/active/client-9001.jsonl.open"
	writer.file_sizes[active_path] = 17
	writer.file_modified_times[active_path] = 7000

	assert_true(writer.configure("user://fake-writer", "client", {"compression_enabled": false}))

	assert_true(writer.renamed_paths.is_empty())
	assert_true(writer.file_sizes.has(active_path))
	assert_eq(writer.current_path, "user://fake-writer/active/client-4242.jsonl.open")
	assert_true(writer.enabled)


func test_recovery_fallback_is_archived_on_clean_close() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 9000
	writer.fail_rename = true
	writer.file_sizes["user://fake-writer/active/client-4242.jsonl.open"] = 17
	writer.file_modified_times["user://fake-writer/active/client-4242.jsonl.open"] = 7000

	assert_true(writer.configure("user://fake-writer", "client", {"compression_enabled": false}))
	assert_eq(writer.current_path, "user://fake-writer/active/client-4242-1.jsonl.open")
	writer.write_line("fallback-session")

	writer.fail_rename = false
	writer.fake_now = 12000
	writer.close()

	var fallback_archive := "user://fake-writer/archive/client-4242-1-9000-12000.jsonl"
	assert_true(writer.file_sizes.has(fallback_archive))
	assert_false(writer.file_sizes.has("user://fake-writer/active/client-4242-1.jsonl.open"))
	assert_eq(writer.clean_marker_write_calls, 0)


func test_configuration_recovers_dead_process_active_file_uses_current_clock_when_modified_time_is_unavailable() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 5555
	var stale_path := "user://fake-writer/active/client-9001.jsonl.open"
	writer.file_sizes[stale_path] = 9

	assert_true(writer.configure("user://fake-writer", "client", {"compression_enabled": false}))

	assert_eq(writer.renamed_paths, ["user://fake-writer/active/client-9001.jsonl.open -> user://fake-writer/archive/client-9001-5555-5555.jsonl"])
	assert_eq(writer.file_sizes["user://fake-writer/archive/client-9001-5555-5555.jsonl"], 9)
	assert_eq(writer.current_path, "user://fake-writer/active/client-4242.jsonl.open")


func test_configuration_applies_retention_age_cleanup_after_open() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 10000
	var policy := {
		"retention_max_age": 2,
		"retention_max_bytes": 0,
		"compression_enabled": false,
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
		"compression_enabled": false,
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


func test_configuration_applies_retention_file_count_cleanup_oldest_first() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 10000
	var policy := {
		"retention_max_age": 0,
		"retention_max_bytes": 0,
		"retention_max_files": 2,
		"compression_enabled": false,
	}
	var oldest := "user://fake-writer/archive/client-a-1000-1000.jsonl"
	var middle := "user://fake-writer/archive/client-b-2000-2000.jsonl"
	var newest := "user://fake-writer/archive/client-c-3000-3000.jsonl"
	writer.file_sizes[oldest] = 1
	writer.file_sizes[middle] = 1
	writer.file_sizes[newest] = 1
	writer.file_modified_times[oldest] = 1000
	writer.file_modified_times[middle] = 2000
	writer.file_modified_times[newest] = 3000

	assert_true(writer.configure("user://fake-writer", "client", policy))

	assert_eq(writer.deleted_paths, [oldest])
	assert_false(writer.file_sizes.has(oldest))
	assert_true(writer.file_sizes.has(middle))
	assert_true(writer.file_sizes.has(newest))


func test_write_line_rotates_when_segment_size_would_be_exceeded() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 1000
	var policy := {
		"segment_max_bytes": 12,
		"segment_max_age": 0,
		"compression_enabled": false,
	}

	assert_true(writer.configure("user://fake-writer", "client", policy))
	writer.file_sizes[writer.current_path] = 10
	writer.fake_now = 2000

	writer.write_line("ab")

	assert_eq(writer.renamed_paths, ["user://fake-writer/active/client-4242.jsonl.open -> user://fake-writer/archive/client-4242-1000-2000.jsonl"])
	assert_eq(writer.opened_paths, [
		"user://fake-writer/active/client-4242.jsonl.open",
		"user://fake-writer/active/client-4242.jsonl.open",
	])
	assert_eq(writer.handles.size(), 2)
	assert_eq(writer.handles[0].close_calls, 1)
	assert_eq(writer.handles[0].lines, [])
	assert_eq(writer.handles[1].lines, ["ab"])
	assert_eq(writer.current_path, "user://fake-writer/active/client-4242.jsonl.open")
	assert_eq(writer.file_sizes["user://fake-writer/archive/client-4242-1000-2000.jsonl"], 10)
	assert_eq(writer.file_sizes[writer.current_path], 3)


func test_write_line_rotates_when_segment_age_expires() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 1000
	var policy := {
		"segment_max_bytes": 0,
		"segment_max_age": 2,
		"compression_enabled": false,
	}

	assert_true(writer.configure("user://fake-writer", "client", policy))
	writer.fake_now = 3500

	writer.write_line("age")

	assert_eq(writer.renamed_paths, ["user://fake-writer/active/client-4242.jsonl.open -> user://fake-writer/archive/client-4242-1000-3500.jsonl"])
	assert_eq(writer.opened_paths, [
		"user://fake-writer/active/client-4242.jsonl.open",
		"user://fake-writer/active/client-4242.jsonl.open",
	])
	assert_eq(writer.handles.size(), 2)
	assert_eq(writer.handles[0].close_calls, 1)
	assert_eq(writer.handles[1].lines, ["age"])
	assert_eq(writer.file_sizes["user://fake-writer/archive/client-4242-1000-3500.jsonl"], 0)
	assert_eq(writer.file_sizes[writer.current_path], 4)
func test_rotation_compresses_completed_segment_when_enabled() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 1000
	assert_true(writer.configure("user://fake-writer", "client", {
		"segment_max_bytes": 1,
		"segment_max_age": 0,
		"compression_enabled": true,
		"retention_max_age": 0,
		"retention_max_bytes": 0,
	}))
	writer.fake_now = 2000
	writer.write_line("ab")
	var archive_path := "user://fake-writer/archive/client-4242-1000-2000.jsonl"
	assert_eq(writer.compressed_paths, [archive_path])
	assert_false(writer.file_sizes.has(archive_path))
	assert_true(writer.file_sizes.has("%s.gz" % archive_path))
	assert_true(writer.enabled)


func test_rotation_keeps_completed_segment_uncompressed_when_disabled() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 1000
	assert_true(writer.configure("user://fake-writer", "client", {
		"segment_max_bytes": 1,
		"segment_max_age": 0,
		"compression_enabled": false,
		"retention_max_age": 0,
		"retention_max_bytes": 0,
	}))
	writer.fake_now = 2000
	writer.write_line("ab")
	var archive_path := "user://fake-writer/archive/client-4242-1000-2000.jsonl"
	assert_true(writer.compressed_paths.is_empty())
	assert_true(writer.file_sizes.has(archive_path))
	assert_false(writer.file_sizes.has("%s.gz" % archive_path))


func test_interrupted_active_recovery_compresses_completed_segment() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 9000
	var active_path := "user://fake-writer/active/client-9001.jsonl.open"
	writer.file_sizes[active_path] = 17
	writer.file_modified_times[active_path] = 7000
	assert_true(writer.configure("user://fake-writer", "client", {"compression_enabled": true}))
	var archive_path := "user://fake-writer/archive/client-9001-7000-9000.jsonl"
	assert_eq(writer.compressed_paths, [archive_path])
	assert_false(writer.file_sizes.has(archive_path))
	assert_true(writer.file_sizes.has("%s.gz" % archive_path))
	assert_true(writer.enabled)


func test_retention_counts_compressed_archive_sizes() -> void:
	var writer := FakeFilesystemWriter.new()
	writer.fake_now = 10000
	var oldest := "user://fake-writer/archive/client-1000-1000.jsonl.gz"
	var newest := "user://fake-writer/archive/client-2000-2000.jsonl.gz"
	writer.file_sizes[oldest] = 8
	writer.file_sizes[newest] = 7
	writer.file_modified_times[oldest] = 1000
	writer.file_modified_times[newest] = 2000
	assert_true(writer.configure("user://fake-writer", "client", {
		"compression_enabled": false,
		"retention_max_age": 0,
		"retention_max_bytes": 10,
	}))
	assert_eq(writer.deleted_paths, [oldest])
	assert_false(writer.file_sizes.has(oldest))
	assert_true(writer.file_sizes.has(newest))


func test_rotation_creates_valid_gzip_archive() -> void:
	var writer := RollingJSONLWriter.new()
	assert_true(writer.configure("user://writer-test", "client", {
		"segment_max_bytes": 8,
		"segment_max_age": 0,
		"compression_enabled": true,
		"retention_max_age": 0,
		"retention_max_bytes": 0,
	}))
	writer.write_line("first")
	writer.write_line("second")
	var archive_files := _archive_files("user://writer-test/archive")
	assert_eq(archive_files.size(), 1)
	assert_true(archive_files[0].ends_with(".jsonl.gz"))
	assert_false(FileAccess.file_exists(archive_files[0].trim_suffix(".gz")))
	var archive = FileAccess.open_compressed(archive_files[0], FileAccess.READ, FileAccess.COMPRESSION_GZIP)
	assert_ne(archive, null)
	if archive != null:
		assert_eq(archive.get_as_text(), "first\n")
		archive.close()
	writer.close()


func _archive_files(path: String) -> Array[String]:
	var files: Array[String] = []
	var dir := DirAccess.open(path)
	if dir == null:
		return files
	dir.list_dir_begin()
	while true:
		var entry := dir.get_next()
		if entry == "":
			break
		if !dir.current_is_dir():
			files.append(path.path_join(entry))
	dir.list_dir_end()
	files.sort()
	return files