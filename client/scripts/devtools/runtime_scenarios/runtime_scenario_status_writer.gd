extends RefCounted
class_name RuntimeScenarioStatusWriter

var path := ""
var client_id := ""
var role := ""


func configure(path_value: String, client_id_value: String, role_value: String) -> void:
	path = path_value
	client_id = client_id_value
	role = role_value


func write(state: String, fields: Dictionary = {}) -> bool:
	if path.is_empty():
		return false
	var absolute_path := ProjectSettings.globalize_path(path)
	var base_directory := absolute_path.get_base_dir()
	if DirAccess.make_dir_recursive_absolute(base_directory) != OK:
		return false
	var payload := {
		"state": state,
		"client_id": client_id,
		"role": role,
		"timestamp_msec": Time.get_ticks_msec(),
	}
	for key in fields.keys():
		payload[key] = fields[key]
	var file := FileAccess.open(absolute_path, FileAccess.WRITE)
	if file == null:
		return false
	file.store_string(JSON.stringify(payload, "\t"))
	file.flush()
	var succeeded := file.get_error() == OK
	file.close()
	return succeeded
