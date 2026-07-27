extends RefCounted

const RollingJSONLWriter := preload("res://scripts/logging/rolling_jsonl_writer.gd")

class FakeHandle extends RefCounted:
	var path := ""
	var writer = null
	var lines: Array[String] = []
	var flush_calls := 0
	var close_calls := 0
	var error := OK
	var seek_end_calls := 0

	func store_line(line: String) -> void:
		lines.append(line)
		if writer != null and path != "":
			writer.file_sizes[path] = int(writer.file_sizes.get(path, 0)) + line.to_utf8_buffer().size() + 1

	func flush() -> void:
		flush_calls += 1

	func close() -> void:
		close_calls += 1

	func get_error() -> Error:
		return error

	func seek_end() -> void:
		seek_end_calls += 1


class FakeFilesystemWriter extends RollingJSONLWriter:
	var fake_now := 123456
	var make_dir_paths: Array[String] = []
	var existing_paths: Array[String] = []
	var opened_paths: Array[String] = []
	var renamed_paths: Array[String] = []
	var compressed_paths: Array[String] = []
	var deleted_paths: Array[String] = []
	var failure_warnings: Array[String] = []
	var fail_next_make_dir := false
	var fail_rename := false
	var fail_compression := false
	var fail_open_after_first_call := false
	var fail_all_open := false
	var fake_process_id := 4242
	var running_process_ids: Array[int] = []
	var fail_delete_paths: Array[String] = []
	var file_sizes: Dictionary = {}
	var file_modified_times: Dictionary = {}
	var handles: Array[FakeHandle] = []
	var clean_marker_exists := false
	var clean_marker_write_calls := 0
	var clean_marker_remove_calls := 0
	var fail_clean_marker_write := false
	var fail_clean_marker_remove := false

	func _should_run_startup_maintenance_async() -> bool:
		return false

	func _current_time_unix_ms() -> int:
		return fake_now

	func _process_id() -> int:
		return fake_process_id

	func _is_process_running(process_id: int) -> bool:
		return running_process_ids.has(process_id)

	func _make_dir_recursive(path: String) -> Error:
		make_dir_paths.append(path)
		if fail_next_make_dir:
			fail_next_make_dir = false
			return ERR_CANT_CREATE
		return OK

	func _file_exists(path: String) -> bool:
		existing_paths.append(path)
		return file_sizes.has(path)

	func _get_file_size(path: String) -> int:
		return int(file_sizes.get(path, 0))

	func _get_file_modified_time_unix_ms(path: String) -> int:
		return int(file_modified_times.get(path, 0))

	func _list_active_files() -> Array[String]:
		var active_files: Array[String] = []
		for path in file_sizes.keys():
			if path.begins_with("user://fake-writer/active/") and path.ends_with(".jsonl.open"):
				active_files.append(path)
		active_files.sort()
		return active_files

	func _list_archive_files() -> Array[String]:
		var archive_files: Array[String] = []
		for path in file_sizes.keys():
			if path.begins_with("user://fake-writer/archive/"):
				archive_files.append(path)
		archive_files.sort()
		return archive_files

	func _delete_file(path: String) -> Error:
		deleted_paths.append(path)
		if fail_delete_paths.has(path):
			return ERR_CANT_OPEN
		file_sizes.erase(path)
		file_modified_times.erase(path)
		return OK

	func _rename_file(from_path: String, to_path: String) -> Error:
		renamed_paths.append("%s -> %s" % [from_path, to_path])
		if fail_rename:
			return ERR_CANT_CREATE
		if !file_sizes.has(from_path):
			return ERR_DOES_NOT_EXIST
		file_sizes[to_path] = int(file_sizes.get(from_path, 0))
		file_sizes.erase(from_path)
		if file_modified_times.has(from_path):
			file_modified_times[to_path] = int(file_modified_times[from_path])
			file_modified_times.erase(from_path)
		return OK

	func _open_file(path: String, _mode: int):
		opened_paths.append(path)
		if fail_all_open or (fail_open_after_first_call and opened_paths.size() > 1):
			return null
		file_sizes[path] = int(file_sizes.get(path, 0))
		var handle := FakeHandle.new()
		handle.path = path
		handle.writer = self
		handles.append(handle)
		return handle

	func _seek_file_to_end(handle) -> void:
		handle.seek_end()

	func _clean_shutdown_marker_exists(_path: String) -> bool:
		return clean_marker_exists

	func _write_clean_shutdown_marker(_path: String) -> Error:
		clean_marker_write_calls += 1
		if fail_clean_marker_write:
			return ERR_CANT_CREATE
		clean_marker_exists = true
		return OK

	func _remove_clean_shutdown_marker(_path: String) -> Error:
		clean_marker_remove_calls += 1
		if fail_clean_marker_remove:
			return ERR_CANT_OPEN
		clean_marker_exists = false
		return OK

	func _compress_archive(archive_path: String) -> bool:
		compressed_paths.append(archive_path)
		if fail_compression:
			_record_failure("failed to compress archived log file: %s" % archive_path)
			return false
		var compressed_path := "%s.gz" % archive_path
		file_sizes[compressed_path] = int(file_sizes.get(archive_path, 0))
		file_sizes.erase(archive_path)
		if file_modified_times.has(archive_path):
			file_modified_times[compressed_path] = int(file_modified_times[archive_path])
			file_modified_times.erase(archive_path)
		return true

	func _emit_failure_warning(message: String) -> void:
		failure_warnings.append(message)
