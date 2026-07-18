extends RefCounted
class_name ClientMeasurementReportWriter

const REPORT_VERSION := 1
const DEFAULT_BASE_DIRECTORY := "user://measurements"
const TEMPORARY_SUFFIX := ".tmp"

var base_directory := DEFAULT_BASE_DIRECTORY
var _timestamp_provider: Callable


func _init(base_directory_ref: String = DEFAULT_BASE_DIRECTORY, timestamp_provider: Callable = Callable()) -> void:
	base_directory = base_directory_ref
	_timestamp_provider = timestamp_provider


func write(report: Dictionary, run_id: String = "") -> Dictionary:
	var final_path := build_path(run_id)
	var make_error := DirAccess.make_dir_recursive_absolute(base_directory)
	if make_error != OK or !DirAccess.dir_exists_absolute(base_directory):
		return _failure(final_path, "failed to create measurement directory: %s" % base_directory)

	var temporary_path := "%s%s" % [final_path, TEMPORARY_SUFFIX]
	var encoded_report := JSON.stringify(report, "\t")
	var file := FileAccess.open(temporary_path, FileAccess.WRITE)
	if file == null:
		return _failure(final_path, "failed to open temporary measurement report: %s" % temporary_path)

	file.store_string(encoded_report)
	file.flush()
	var write_error := file.get_error()
	file.close()
	if write_error != OK:
		_cleanup_temporary_file(temporary_path)
		return _failure(final_path, "failed to write measurement report: %s" % temporary_path)

	var replace_error := _replace_file(temporary_path, final_path)
	if replace_error != OK:
		_cleanup_temporary_file(temporary_path)
		return _failure(final_path, "failed to finalize measurement report: %s" % final_path)
	return {
		"success": true,
		"path": final_path,
		"error": "",
	}


func build_path(run_id: String = "") -> String:
	var timestamp := _sanitize_component(_current_timestamp(), "unknown-time")
	var sanitized_run_id := _sanitize_component(run_id, "run")
	var filename := "measurement-v%d-%s-%s.json" % [REPORT_VERSION, timestamp, sanitized_run_id]
	return base_directory.path_join(filename)


func _current_timestamp() -> String:
	if _timestamp_provider.is_valid():
		return str(_timestamp_provider.call())
	return Time.get_datetime_string_from_system(true, false)


func _sanitize_component(value: String, fallback: String) -> String:
	var regex := RegEx.new()
	regex.compile("[^A-Za-z0-9_-]")
	var sanitized := regex.sub(value, "_", true)
	return sanitized if !sanitized.is_empty() else fallback


func _replace_file(temporary_path: String, final_path: String) -> Error:
	var rename_error := DirAccess.rename_absolute(temporary_path, final_path)
	if rename_error == OK:
		return OK
	if !FileAccess.file_exists(final_path):
		return rename_error
	if DirAccess.remove_absolute(final_path) != OK:
		return rename_error
	return DirAccess.rename_absolute(temporary_path, final_path)


func _cleanup_temporary_file(temporary_path: String) -> void:
	if FileAccess.file_exists(temporary_path):
		DirAccess.remove_absolute(temporary_path)


func _failure(final_path: String, message: String) -> Dictionary:
	return {
		"success": false,
		"path": final_path,
		"error": message,
	}
