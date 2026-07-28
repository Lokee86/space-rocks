extends RefCounted
class_name LegacyAuthTokenReader

var token_path := "user://auth_token.json"


func _init() -> void:
	if _is_test_process():
		token_path = "user://test_auth_token_%s.json" % str(OS.get_process_id())


func load_token() -> String:
	if !FileAccess.file_exists(token_path):
		return ""

	var file := FileAccess.open(token_path, FileAccess.READ)
	if file == null:
		return ""

	var text := file.get_as_text()
	file.close()
	clear_token()

	var parsed = JSON.parse_string(text)
	if typeof(parsed) != TYPE_DICTIONARY:
		return ""

	return str(parsed.get("token", ""))


func clear_token() -> void:
	if FileAccess.file_exists(token_path):
		DirAccess.remove_absolute(ProjectSettings.globalize_path(token_path))


func _is_test_process() -> bool:
	for argument in OS.get_cmdline_args():
		if str(argument).contains("gut_cmdln.gd"):
			return true
	return false
