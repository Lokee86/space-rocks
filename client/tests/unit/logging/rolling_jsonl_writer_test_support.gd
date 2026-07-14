extends RefCounted

const RollingJSONLWriter := preload("res://scripts/logging/rolling_jsonl_writer.gd")

class FakeHandle extends RefCounted:
	var path := ""
	var writer = null
	var lines: Array[String] = []
	var flush_calls := 0
	var close_calls := 0
	var error := OK

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


class FakeFilesystemWriter extends RollingJSONLWriter:
	var fake_now := 123456
	var make_dir_paths: Array[String] = []
	var existing_paths: Array[String] = []
	var opened_paths: Array[String] = []
	var renamed_paths: Array[String] = []
	var deleted_paths: Array[String] = []
	var failure_warnings: Array[String] = []
	var fail_next_make_dir := false
	var fail_rename := false
	var fail_open_after_first_call := false
	var fail_delete_paths: Array[String] = []
	var file_sizes: Dictionary = {}
	var file_modified_times: Dictionary = {}
	var handles: Array[FakeHandle] = []

	func _current_time_unix_ms() -> int:
		return fake_now

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
		if fail_open_after_first_call and opened_paths.size() > 1:
			return null
		file_sizes[path] = int(file_sizes.get(path, 0))
		var handle := FakeHandle.new()
		handle.path = path
		handle.writer = self
		handles.append(handle)
		return handle

	func _emit_failure_warning(message: String) -> void:
		failure_warnings.append(message)
