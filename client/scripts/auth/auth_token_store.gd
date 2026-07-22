extends RefCounted
class_name LegacyAuthTokenReader

var token_path := "user://auth_token.json"


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
