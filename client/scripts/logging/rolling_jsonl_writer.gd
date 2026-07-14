extends RefCounted

var enabled := false
var current_path := ""
var configuration: Dictionary = {}
var active_directory_path := ""
var archive_directory_path := ""
var last_configured_at_unix_ms := 0
var _handle = null


func configure(base_dir: String, prefix: String, policy: Dictionary = {}) -> bool:
	close()
	configuration = _normalize_policy(policy)
	active_directory_path = base_dir.path_join(configuration["active_directory_name"])
	archive_directory_path = base_dir.path_join(configuration["archive_directory_name"])
	last_configured_at_unix_ms = _current_time_unix_ms()

	var make_error := _make_dir_recursive(base_dir)
	if make_error != OK:
		return false

	for index in range(1, 1000000):
		var candidate_name := build_numbered_filename(prefix, index)
		var candidate_path := base_dir.path_join(candidate_name)
		if _file_exists(candidate_path):
			continue

		var handle = _open_file(candidate_path, FileAccess.WRITE)
		if handle == null:
			return false

		enabled = true
		current_path = candidate_path
		_handle = handle
		return true

	return false


func _normalize_policy(policy: Dictionary) -> Dictionary:
	var normalized := {
		"segment_max_bytes": 0,
		"segment_max_age": 0,
		"retention_max_age": 0,
		"retention_max_bytes": 0,
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


func _open_file(path: String, mode: int):
	return FileAccess.open(path, mode)


static func build_numbered_filename(prefix: String, index: int) -> String:
	return "%s-%06d.jsonl" % [prefix, index]


func write_line(line: String) -> void:
	if !enabled || _handle == null:
		return

	_handle.store_line(line)
	_handle.flush()


func close() -> void:
	if _handle != null:
		_handle.flush()
		_handle.close()

	enabled = false
	current_path = ""
	_handle = null