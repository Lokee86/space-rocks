extends RefCounted

const GzipArchiveCompressor := preload("res://scripts/logging/gzip_archive_compressor.gd")

class ArchiveFileInfo extends RefCounted:
	var path := ""
	var modified_time_unix_ms := 0
	var size_bytes := 0

var _thread: Thread
var _archive_mutex: Mutex
var _state_mutex := Mutex.new()
var _failures: Array[String] = []
var _running := false
var _completed := false


func start(
	archive_paths_to_compress: Array[String],
	archive_directory_path: String,
	configured_prefix: String,
	configuration: Dictionary,
	archive_mutex: Mutex
) -> Error:
	_archive_mutex = archive_mutex
	_state_mutex.lock()
	_failures.clear()
	_running = true
	_completed = false
	_state_mutex.unlock()

	_thread = Thread.new()
	var start_error := _thread.start(
		Callable(self, "_run_worker").bind(
			archive_paths_to_compress.duplicate(),
			archive_directory_path,
			configured_prefix,
			configuration.duplicate(true)
		)
	)
	if start_error != OK:
		_state_mutex.lock()
		_running = false
		_state_mutex.unlock()
		_thread = null
	return start_error


func poll_failures() -> Array[String]:
	_join_if_finished()
	return _take_failures()


func wait_for_completion() -> Array[String]:
	if _thread != null:
		if _thread.is_started():
			_thread.wait_to_finish()
		_thread = null
	return _take_failures()


func status() -> Dictionary:
	_join_if_finished()
	_state_mutex.lock()
	var current_status := {
		"running": _running,
		"completed": _completed,
	}
	_state_mutex.unlock()
	return current_status


func _join_if_finished() -> void:
	if _thread == null:
		return
	if _thread.is_started() and !_thread.is_alive():
		_thread.wait_to_finish()
		_thread = null


func _take_failures() -> Array[String]:
	_state_mutex.lock()
	var failures := _failures.duplicate()
	_failures.clear()
	_state_mutex.unlock()
	return failures


func _run_worker(
	archive_paths_to_compress: Array[String],
	archive_directory_path: String,
	configured_prefix: String,
	configuration: Dictionary
) -> void:
	if _archive_mutex != null:
		_archive_mutex.lock()
	var failures := _perform_maintenance(
		archive_paths_to_compress,
		archive_directory_path,
		configured_prefix,
		configuration
	)
	if _archive_mutex != null:
		_archive_mutex.unlock()

	_state_mutex.lock()
	_failures.append_array(failures)
	_running = false
	_completed = true
	_state_mutex.unlock()


func _perform_maintenance(
	archive_paths_to_compress: Array[String],
	archive_directory_path: String,
	configured_prefix: String,
	configuration: Dictionary
) -> Array[String]:
	var failures: Array[String] = []
	if bool(configuration.get("compression_enabled", false)):
		for archive_path in archive_paths_to_compress:
			var compressor := GzipArchiveCompressor.new()
			if !compressor.compress(archive_path):
				failures.append(compressor.last_failure_message)

	_apply_retention(archive_directory_path, configured_prefix, configuration, failures)
	return failures


func _apply_retention(
	archive_directory_path: String,
	configured_prefix: String,
	configuration: Dictionary,
	failures: Array[String]
) -> void:
	var archive_files := _archive_file_infos(archive_directory_path, configured_prefix)
	var now_unix_ms := int(Time.get_unix_time_from_system() * 1000.0)
	var max_age_seconds := int(configuration.get("retention_max_age", 0))
	if max_age_seconds > 0:
		var cutoff_unix_ms := now_unix_ms - (max_age_seconds * 1000)
		var age_retained: Array[ArchiveFileInfo] = []
		for info in archive_files:
			if info.modified_time_unix_ms > 0 and info.modified_time_unix_ms < cutoff_unix_ms:
				if DirAccess.remove_absolute(info.path) != OK:
					failures.append("failed to delete archived log file: %s" % info.path)
					age_retained.append(info)
			else:
				age_retained.append(info)
		archive_files = age_retained

	archive_files.sort_custom(func(a: ArchiveFileInfo, b: ArchiveFileInfo) -> bool:
		if a.modified_time_unix_ms == b.modified_time_unix_ms:
			return a.path < b.path
		if a.modified_time_unix_ms <= 0:
			return false
		if b.modified_time_unix_ms <= 0:
			return true
		return a.modified_time_unix_ms < b.modified_time_unix_ms
	)

	var max_files := int(configuration.get("retention_max_files", 0))
	if max_files > 0 and archive_files.size() > max_files:
		var count_retained: Array[ArchiveFileInfo] = []
		var files_to_remove := archive_files.size() - max_files
		for index in archive_files.size():
			var info := archive_files[index]
			if index < files_to_remove:
				if DirAccess.remove_absolute(info.path) != OK:
					failures.append("failed to delete archived log file: %s" % info.path)
					count_retained.append(info)
			else:
				count_retained.append(info)
		archive_files = count_retained

	var max_bytes := int(configuration.get("retention_max_bytes", 0))
	if max_bytes <= 0:
		return

	var total_bytes := 0
	for info in archive_files:
		total_bytes += info.size_bytes

	for info in archive_files:
		if total_bytes <= max_bytes:
			break
		if DirAccess.remove_absolute(info.path) != OK:
			failures.append("failed to delete archived log file: %s" % info.path)
		else:
			total_bytes -= info.size_bytes


func _archive_file_infos(
	archive_directory_path: String,
	configured_prefix: String
) -> Array[ArchiveFileInfo]:
	var files: Array[ArchiveFileInfo] = []
	var dir := DirAccess.open(archive_directory_path)
	if dir == null:
		return files

	dir.list_dir_begin()
	while true:
		var entry := dir.get_next()
		if entry == "":
			break
		if dir.current_is_dir():
			continue
		if !entry.begins_with("%s-" % configured_prefix):
			continue
		if !entry.ends_with(".jsonl") and !entry.ends_with(".jsonl.gz"):
			continue

		var path := archive_directory_path.path_join(entry)
		var info := ArchiveFileInfo.new()
		info.path = path
		var modified_time_unix_s := FileAccess.get_modified_time(path)
		if modified_time_unix_s > 0:
			info.modified_time_unix_ms = int(modified_time_unix_s * 1000.0)
		var file := FileAccess.open(path, FileAccess.READ)
		if file != null:
			info.size_bytes = file.get_length()
			file.close()
		files.append(info)
	dir.list_dir_end()
	return files
