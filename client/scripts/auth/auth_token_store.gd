extends RefCounted
class_name AuthTokenStore

var token_path := "user://auth_token.json"


func load_token() -> String:
	if !FileAccess.file_exists(token_path):
		return ""

	var file := FileAccess.open(token_path, FileAccess.READ)
	if file == null:
		return ""

	var text := file.get_as_text()
	file.close()

	var parser := JSON.new()
	if parser.parse(text) != OK:
		return ""

	if typeof(parser.data) != TYPE_DICTIONARY:
		return ""

	return str(parser.data.get("token", ""))


func save_token(token: String) -> void:
	var file := FileAccess.open(token_path, FileAccess.WRITE)
	if file == null:
		return

	file.store_string(JSON.stringify({ "token": token }))
	file.close()


func clear_token() -> void:
	if FileAccess.file_exists(token_path):
		DirAccess.remove_absolute(ProjectSettings.globalize_path(token_path))
