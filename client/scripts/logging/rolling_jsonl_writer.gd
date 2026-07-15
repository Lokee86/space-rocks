extends RefCounted

const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const GzipArchiveCompressor := preload("res://scripts/logging/gzip_archive_compressor.gd")

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
	if _file_exists(candidate_path):
		if !_recover_interrupted_active_segment(candidate_path):
			return false

	var handle = _open_file(candidate_path, FileAccess.WRITE)
	if handle == null:
		_record_failure("failed to open active log file: %s" % candidate_path)
		return false

	enabled = true
	current_path = candidate_path
	_handle = handle
	_apply_retention()
	return true

func _recover_interrupted_active_segment(candidate_path: String) -> bool:
	var rotation_time_unix_ms := _current_time_unix_ms()
	var segment_start_unix_ms := _get_file_modified_time_unix_ms(candidate_path)
	if segment_start_unix_ms <= 0:
		segment_start_unix_ms = rotation_time_unix_ms

	var archive_path := archive_directory_path.path_join(_build_segment_archive_filename(configured_prefix, segment_start_unix_ms, rotation_time_unix_ms))
	if !_finalize_archive(candidate_path, archive_path):
		_record_failure("failed to recover interrupted active log file: %s" % candidate_path)
		return false
	return true

func _normalize_policy(policy: Dictionary) -> Dictionary:
	var normalized := {
		"segment_max_bytes": 16 * 1024 * 1024,
		"segment_max_age": ObservabilityContract.FILE_LOGGING_MAX_ACTIVE_SEGMENT_AGE_SECONDS,
		"retention_max_age": ObservabilityContract.RETENTION_DEFAULT_AGE_SECONDS_OPERATIONAL,
		"retention_max_bytes": 250 * 1024 * 1024,
		"compression_enabled": ObservabilityContract.FILE_LOGGING_COMPRESSION_ENABLED,
		"active_directory_name": "active",
		"archive_directory_name": "archive",
	}
	for key in normalized.keys():
		if policy.has(key):
			normalized[key] = policy[key]
	return normalized

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
	var archive_files := _archive_file_infos()
	var now_unix_ms := _current_time_unix_ms()
	var max_age_seconds := int(configuration["retention_max_age"])
	if max_age_seconds > 0:
		var cutoff_unix_ms := now_unix_ms - (max_age_seconds * 1000)
		for info in archive_files:
			if info.modified_time_unix_ms > 0 and info.modified_time_unix_ms < cutoff_unix_ms:
				if _delete_file(info.path) != OK:
					_record_failure("failed to delete archived log file: %s" % info.path)
		archive_files = _archive_file_infos()

	var max_bytes := int(configuration["retention_max_bytes"])
	if max_bytes <= 0:
		return

	archive_files.sort_custom(func(a: ArchiveFileInfo, b: ArchiveFileInfo) -> bool:
		if a.modified_time_unix_ms == b.modified_time_unix_ms:
			return a.path < b.path
		if a.modified_time_unix_ms <= 0:
			return false
		if b.modified_time_unix_ms <= 0:
			return true
		return a.modified_time_unix_ms < b.modified_time_unix_ms
	)

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
	if _rename_file(source_path, archive_path) != OK:
		return false
	if _compression_enabled():
		_compress_archive(archive_path)
	return true

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
	if _handle != null:
		_handle.flush()
		_handle.close()

	enabled = false
	current_path = ""
	_handle = null
	configured_prefix = ""
	segment_started_at_unix_ms = 0