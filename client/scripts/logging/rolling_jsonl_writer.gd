extends RefCounted

const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const GzipArchiveCompressor := preload("res://scripts/logging/gzip_archive_compressor.gd")
const LogArchiveMaintenance := preload("res://scripts/logging/log_archive_maintenance.gd")

var enabled := false
var current_path := ""
var configuration: Dictionary = {}
var active_directory_path := ""
var archive_directory_path := ""
var last_configured_at_unix_ms := 0
var configured_prefix := ""
var segment_started_at_unix_ms := 0
var failure_count := 0
var last_failure_message := ""
var _failure_warning_emitted := false
var _handle = null
var _archive_compressor := GzipArchiveCompressor.new()
var _archive_mutex := Mutex.new()
var _startup_maintenance
var startup_maintenance_completed := false

const DEFAULT_RETENTION_MAX_FILES := 256

class ArchiveFileInfo extends RefCounted:
	var path := ""
	var modified_time_unix_ms := 0
	var size_bytes := 0

func configure(base_dir: String, prefix: String, policy: Dictionary = {}) -> bool:
	close()
	_reset_failure_state()
	configuration = _normalize_policy(policy)
	configured_prefix = prefix
	active_directory_path = base_dir.path_join(configuration["active_directory_name"])
	archive_directory_path = base_dir.path_join(configuration["archive_directory_name"])
	last_configured_at_unix_ms = _current_time_unix_ms()
	segment_started_at_unix_ms = last_configured_at_unix_ms

	var make_error := _make_dir_recursive(active_directory_path)
	if make_error != OK:
		_record_failure("failed to create active log directory: %s" % active_directory_path)
		return false

	make_error = _make_dir_recursive(archive_directory_path)
	if make_error != OK:
		_record_failure("failed to create archive log directory: %s" % archive_directory_path)
		return false

	var candidate_path := _build_active_file_path()
	var clean_marker_path := _build_clean_shutdown_marker_path()
	var active_exists := _file_exists(candidate_path)
	var resume_clean_segment := active_exists and _clean_shutdown_marker_exists(clean_marker_path)
	var startup_archives: Array[String] = []
	if resume_clean_segment:
		if _remove_clean_shutdown_marker(clean_marker_path) != OK:
			_record_failure("failed to remove clean shutdown marker: %s" % clean_marker_path)
			return false
		segment_started_at_unix_ms = _get_file_modified_time_unix_ms(candidate_path)
		if segment_started_at_unix_ms <= 0:
			segment_started_at_unix_ms = last_configured_at_unix_ms
	elif active_exists:
		var recovered_archive_path := _recover_interrupted_active_segment(candidate_path)
		if recovered_archive_path.is_empty():
			return false
		startup_archives.append(recovered_archive_path)
	elif _clean_shutdown_marker_exists(clean_marker_path):
		_remove_clean_shutdown_marker(clean_marker_path)

	var handle = _open_file(candidate_path, FileAccess.READ_WRITE if resume_clean_segment else FileAccess.WRITE)
	if handle == null:
		_record_failure("failed to open active log file: %s" % candidate_path)
		return false
	if resume_clean_segment:
		_seek_file_to_end(handle)

	enabled = true
	current_path = candidate_path
	_handle = handle
	_start_startup_maintenance(startup_archives)
	return true

func _recover_interrupted_active_segment(candidate_path: String) -> String:
	var rotation_time_unix_ms := _current_time_unix_ms()
	var segment_start_unix_ms := _get_file_modified_time_unix_ms(candidate_path)
	if segment_start_unix_ms <= 0:
		segment_start_unix_ms = rotation_time_unix_ms

	var archive_path := archive_directory_path.path_join(_build_segment_archive_filename(configured_prefix, segment_start_unix_ms, rotation_time_unix_ms))
	_archive_mutex.lock()
	var rename_error := _rename_file(candidate_path, archive_path)
	_archive_mutex.unlock()
	if rename_error != OK:
		_record_failure("failed to recover interrupted active log file: %s" % candidate_path)
		return ""
	return archive_path

func _normalize_policy(policy: Dictionary) -> Dictionary:
	var normalized := {
		"segment_max_bytes": 16 * 1024 * 1024,
		"segment_max_age": ObservabilityContract.FILE_LOGGING_MAX_ACTIVE_SEGMENT_AGE_SECONDS,
		"retention_max_age": ObservabilityContract.RETENTION_DEFAULT_AGE_SECONDS_OPERATIONAL,
		"retention_max_bytes": 250 * 1024 * 1024,
		"retention_max_files": DEFAULT_RETENTION_MAX_FILES,
		"compression_enabled": ObservabilityContract.FILE_LOGGING_COMPRESSION_ENABLED,
		"startup_maintenance_async": true,
		"active_directory_name": "active",
		"archive_directory_name": "archive",
	}
	for key in normalized.keys():
		if policy.has(key):
			normalized[key] = policy[key]
	return normalized

func _start_startup_maintenance(startup_archives: Array[String]) -> void:
	_startup_maintenance = null
	startup_maintenance_completed = false

	if !_should_run_startup_maintenance_async():
		_archive_mutex.lock()
		for archive_path in startup_archives:
			if _compression_enabled():
				_compress_archive(archive_path)
		_apply_retention_unlocked()
		_archive_mutex.unlock()
		startup_maintenance_completed = true
		return

	var maintenance = _create_startup_maintenance()
	var start_error: Error = maintenance.start(
		startup_archives,
		archive_directory_path,
		configured_prefix,
		configuration,
		_archive_mutex
	)
	if start_error != OK:
		_record_failure("failed to start background log archive maintenance")
		return
	_startup_maintenance = maintenance


func _create_startup_maintenance():
	return LogArchiveMaintenance.new()


func _should_run_startup_maintenance_async() -> bool:
	return bool(configuration.get("startup_maintenance_async", true))


func poll_startup_maintenance() -> void:
	if _startup_maintenance == null:
		return
	_record_startup_maintenance_failures(_startup_maintenance.poll_failures())
	var status: Dictionary = _startup_maintenance.status()
	if bool(status.get("completed", false)):
		startup_maintenance_completed = true


func wait_for_startup_maintenance() -> void:
	if _startup_maintenance == null:
		return
	_record_startup_maintenance_failures(_startup_maintenance.wait_for_completion())
	startup_maintenance_completed = true


func startup_maintenance_status() -> Dictionary:
	poll_startup_maintenance()
	if _startup_maintenance != null:
		return _startup_maintenance.status()
	return {
		"running": false,
		"completed": startup_maintenance_completed,
	}


func _record_startup_maintenance_failures(failures: Array[String]) -> void:
	for failure in failures:
		_record_failure(str(failure))


func _current_time_unix_ms() -> int:
	return int(Time.get_unix_time_from_system() * 1000.0)

func _make_dir_recursive(path: String) -> Error:
	return DirAccess.make_dir_recursive_absolute(path)

func _file_exists(path: String) -> bool:
	return FileAccess.file_exists(path)

func _get_file_size(path: String) -> int:
	if _handle != null and path == current_path:
		return _handle.get_length()

	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		return 0
	var size := file.get_length()
	file.close()
	return size

func _get_file_modified_time_unix_ms(path: String) -> int:
	var modified_time_unix_s := FileAccess.get_modified_time(path)
	if modified_time_unix_s <= 0:
		return 0
	return int(modified_time_unix_s * 1000.0)

func _list_archive_files() -> Array[String]:
	var archive_files: Array[String] = []
	var dir := DirAccess.open(archive_directory_path)
	if dir == null:
		return archive_files

	dir.list_dir_begin()
	while true:
		var entry := dir.get_next()
		if entry == "":
			break
		if dir.current_is_dir():
			continue
		archive_files.append(archive_directory_path.path_join(entry))
	dir.list_dir_end()
	return archive_files

func _delete_file(path: String) -> Error:
	return DirAccess.remove_absolute(path)

func _rename_file(from_path: String, to_path: String) -> Error:
	return DirAccess.rename_absolute(from_path, to_path)

func _open_file(path: String, mode: int):
	return FileAccess.open(path, mode)

func _seek_file_to_end(handle) -> void:
	handle.seek_end()

func _clean_shutdown_marker_exists(path: String) -> bool:
	return FileAccess.file_exists(path)

func _write_clean_shutdown_marker(path: String) -> Error:
	var marker = FileAccess.open(path, FileAccess.WRITE)
	if marker == null:
		return FileAccess.get_open_error()
	marker.store_string("clean")
	var marker_error := marker.get_error()
	marker.close()
	return marker_error

func _remove_clean_shutdown_marker(path: String) -> Error:
	if !FileAccess.file_exists(path):
		return OK
	return DirAccess.remove_absolute(path)


func _get_handle_error(handle) -> Error:
	if handle != null and handle.has_method("get_error"):
		return handle.get_error()
	return OK

func _emit_failure_warning(message: String) -> void:
	push_warning(message)

func _reset_failure_state() -> void:
	failure_count = 0
	last_failure_message = ""
	_failure_warning_emitted = false

func _record_failure(message: String) -> void:
	failure_count += 1
	last_failure_message = message
	if !_failure_warning_emitted:
		_failure_warning_emitted = true
		_emit_failure_warning(message)

func _disable_active_file_output(message: String) -> void:
	_record_failure(message)
	if _handle != null:
		_handle.close()
	enabled = false
	current_path = ""
	_handle = null
	segment_started_at_unix_ms = 0
static func _build_segment_archive_filename(prefix: String, segment_start_unix_ms: int, rotation_unix_ms: int) -> String:
	return "%s-%d-%d.jsonl" % [prefix, segment_start_unix_ms, rotation_unix_ms]

func _build_active_file_path() -> String:
	return active_directory_path.path_join("%s.jsonl.open" % configured_prefix)

func _build_clean_shutdown_marker_path() -> String:
	return active_directory_path.path_join("%s.jsonl.clean" % configured_prefix)

func _build_archive_file_path(rotation_unix_ms: int) -> String:
	return archive_directory_path.path_join(_build_segment_archive_filename(configured_prefix, segment_started_at_unix_ms, rotation_unix_ms))

func _segment_max_bytes() -> int:
	return int(configuration["segment_max_bytes"])

func _compression_enabled() -> bool:
	return bool(configuration["compression_enabled"])

func _segment_max_age_ms() -> int:
	return int(configuration["segment_max_age"]) * 1000

func _line_size_bytes(line: String) -> int:
	return line.to_utf8_buffer().size() + 1

func _segment_age_expired(now_unix_ms: int) -> bool:
	var max_age_ms := _segment_max_age_ms()
	if max_age_ms <= 0:
		return false
	return now_unix_ms - segment_started_at_unix_ms >= max_age_ms

func _segment_size_would_exceed(line: String) -> bool:
	var max_bytes := _segment_max_bytes()
	if max_bytes <= 0:
		return false
	return _get_file_size(current_path) + _line_size_bytes(line) > max_bytes

func _archive_file_infos() -> Array[ArchiveFileInfo]:
	var files: Array[ArchiveFileInfo] = []
	for path in _list_archive_files():
		var filename := path.get_file()
		if !filename.begins_with("%s-" % configured_prefix):
			continue
		if !filename.ends_with(".jsonl") and !filename.ends_with(".jsonl.gz"):
			continue
		var info := ArchiveFileInfo.new()
		info.path = path
		info.modified_time_unix_ms = _get_file_modified_time_unix_ms(path)
		info.size_bytes = _get_file_size(path)
		files.append(info)
	return files

func _apply_retention() -> void:
	_archive_mutex.lock()
	_apply_retention_unlocked()
	_archive_mutex.unlock()


func _apply_retention_unlocked() -> void:
	var archive_files := _archive_file_infos()
	var now_unix_ms := _current_time_unix_ms()
	var max_age_seconds := int(configuration["retention_max_age"])
	if max_age_seconds > 0:
		var cutoff_unix_ms := now_unix_ms - (max_age_seconds * 1000)
		var age_retained: Array[ArchiveFileInfo] = []
		for info in archive_files:
			if info.modified_time_unix_ms > 0 and info.modified_time_unix_ms < cutoff_unix_ms:
				if _delete_file(info.path) != OK:
					_record_failure("failed to delete archived log file: %s" % info.path)
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

	var max_files := int(configuration["retention_max_files"])
	if max_files > 0 and archive_files.size() > max_files:
		var count_retained: Array[ArchiveFileInfo] = []
		var files_to_remove := archive_files.size() - max_files
		for index in archive_files.size():
			var info := archive_files[index]
			if index < files_to_remove:
				if _delete_file(info.path) != OK:
					_record_failure("failed to delete archived log file: %s" % info.path)
					count_retained.append(info)
			else:
				count_retained.append(info)
		archive_files = count_retained

	var max_bytes := int(configuration["retention_max_bytes"])
	if max_bytes <= 0:
		return

	var total_bytes := 0
	for info in archive_files:
		total_bytes += info.size_bytes

	for info in archive_files:
		if total_bytes <= max_bytes:
			break
		if _delete_file(info.path) != OK:
			_record_failure("failed to delete archived log file: %s" % info.path)
		else:
			total_bytes -= info.size_bytes

func _finalize_archive(source_path: String, archive_path: String) -> bool:
	_archive_mutex.lock()
	var finalized := true
	if _rename_file(source_path, archive_path) != OK:
		finalized = false
	elif _compression_enabled():
		_compress_archive(archive_path)
	_archive_mutex.unlock()
	return finalized

func _compress_archive(archive_path: String) -> bool:
	if _archive_compressor.compress(archive_path):
		return true
	_record_failure(_archive_compressor.last_failure_message)
	return false

func _rotate_active_segment() -> bool:
	var rotation_time_unix_ms := _current_time_unix_ms()
	var archive_path := _build_archive_file_path(rotation_time_unix_ms)
	var previous_path := current_path

	if _handle != null:
		_handle.flush()
		_handle.close()
		_handle = null

	if !_finalize_archive(previous_path, archive_path):
		_record_failure("failed to archive active log segment: %s" % previous_path)
		enabled = false
		current_path = ""
		return false

	var new_active_path := _build_active_file_path()
	var handle = _open_file(new_active_path, FileAccess.WRITE)
	if handle == null:
		_record_failure("failed to reopen active log file after rotation: %s" % new_active_path)
		enabled = false
		current_path = ""
		return false

	current_path = new_active_path
	_handle = handle
	segment_started_at_unix_ms = rotation_time_unix_ms
	_apply_retention()
	return true

func write_line(line: String) -> void:
	poll_startup_maintenance()
	if !enabled || _handle == null:
		return

	var now_unix_ms := _current_time_unix_ms()
	if _segment_age_expired(now_unix_ms) or _segment_size_would_exceed(line):
		if !_rotate_active_segment():
			return

	_handle.store_line(line)
	var store_error := _get_handle_error(_handle)
	_handle.flush()
	var flush_error := _get_handle_error(_handle)
	if store_error != OK or flush_error != OK:
		_disable_active_file_output("failed to write active log line: %s" % current_path)


func close() -> void:
	var clean_marker_path := _build_clean_shutdown_marker_path() if !configured_prefix.is_empty() else ""
	if _handle != null:
		_handle.flush()
		_handle.close()
		if !clean_marker_path.is_empty() and _write_clean_shutdown_marker(clean_marker_path) != OK:
			_record_failure("failed to write clean shutdown marker: %s" % clean_marker_path)

	wait_for_startup_maintenance()
	enabled = false
	current_path = ""
	_handle = null
	configured_prefix = ""
	segment_started_at_unix_ms = 0