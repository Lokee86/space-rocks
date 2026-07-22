extends RefCounted
class_name AuthCredentialStore

const LegacyAuthTokenReaderScript := preload("res://scripts/auth/auth_token_store.gd")
const SERVICE_NAME := "ca.laughingskull.space-rocks"
const ACCOUNT_NAME := "session"
const WINDOWS_HELPER_NAME := "space-rocks-credential-helper.exe"
const MACOS_HELPER_NAME := "space-rocks-credential-helper"

var service_name := SERVICE_NAME
var account_name := ACCOUNT_NAME
var encrypted_blob_path := "user://auth_credential.bin"
var revocation_marker_path := "user://auth_credential_revoked"
var helper_path_override := ""
var request_handler: Callable
var legacy_token_reader
var _loaded_from_legacy := false


func load_token() -> String:
	_ensure_legacy_reader()
	_loaded_from_legacy = false

	if FileAccess.file_exists(revocation_marker_path):
		if bool(_request("clear").get("ok", false)):
			_remove_revocation_marker()
		legacy_token_reader.clear_token()
		return ""

	var result := _request("load")
	if !bool(result.get("ok", false)):
		legacy_token_reader.clear_token()
		return ""

	var secure_token := str(result.get("secret", ""))
	if !secure_token.is_empty():
		legacy_token_reader.clear_token()
		return secure_token

	var legacy_token: String = str(legacy_token_reader.load_token())
	if legacy_token.is_empty():
		return ""
	_loaded_from_legacy = true
	return legacy_token


func save_token(token: String) -> bool:
	if token.is_empty():
		return false
	var result := _request("save", token)
	if !bool(result.get("ok", false)):
		return false
	_ensure_legacy_reader()
	legacy_token_reader.clear_token()
	_remove_revocation_marker()
	_loaded_from_legacy = false
	return true


func clear_token() -> bool:
	_write_revocation_marker()
	var result := _request("clear")
	_ensure_legacy_reader()
	legacy_token_reader.clear_token()
	_loaded_from_legacy = false
	var cleared := bool(result.get("ok", false))
	if cleared:
		_remove_revocation_marker()
	return cleared


func requires_legacy_migration() -> bool:
	return _loaded_from_legacy


func _request(action: String, secret: String = "") -> Dictionary:
	var payload := {
		"action": action,
		"service": service_name,
		"account": account_name,
	}
	if OS.get_name() == "Windows":
		payload["blob_path"] = ProjectSettings.globalize_path(encrypted_blob_path)
	if !secret.is_empty():
		payload["secret"] = secret

	if request_handler.is_valid():
		var handled = request_handler.call(payload)
		return handled if typeof(handled) == TYPE_DICTIONARY else {"ok": false}

	var helper_path := _resolve_helper_path()
	if helper_path.is_empty() || !FileAccess.file_exists(helper_path):
		return {"ok": false, "error": "helper_unavailable"}

	var pipes := OS.execute_with_pipe(helper_path, PackedStringArray(), true)
	if pipes.is_empty():
		return {"ok": false, "error": "helper_launch_failed"}

	var stdio: FileAccess = pipes.get("stdio")
	var stderr: FileAccess = pipes.get("stderr")
	if stdio == null:
		if stderr != null:
			stderr.close()
		return {"ok": false, "error": "helper_pipe_failed"}

	stdio.store_line(JSON.stringify(payload))
	stdio.flush()
	var response_line := stdio.get_line()
	stdio.close()
	if stderr != null:
		stderr.close()
	if response_line.is_empty():
		return {"ok": false, "error": "helper_empty_response"}

	var parsed = JSON.parse_string(response_line)
	return parsed if typeof(parsed) == TYPE_DICTIONARY else {"ok": false, "error": "helper_invalid_response"}


func _resolve_helper_path() -> String:
	if !helper_path_override.is_empty():
		return ProjectSettings.globalize_path(helper_path_override)

	var platform := OS.get_name()
	var helper_name := ""
	if platform == "Windows":
		helper_name = WINDOWS_HELPER_NAME
	elif platform == "macOS":
		helper_name = MACOS_HELPER_NAME
	else:
		return ""

	if OS.has_feature("editor"):
		return ProjectSettings.globalize_path("res://native/credential-helper/bin/%s" % helper_name)

	var executable_directory := OS.get_executable_path().get_base_dir()
	if platform == "macOS":
		return executable_directory.path_join("../Helpers/%s" % helper_name).simplify_path()
	return executable_directory.path_join(helper_name)


func _write_revocation_marker() -> void:
	var file := FileAccess.open(revocation_marker_path, FileAccess.WRITE)
	if file == null:
		return
	file.store_string("revoked\n")
	file.close()


func _remove_revocation_marker() -> void:
	if FileAccess.file_exists(revocation_marker_path):
		DirAccess.remove_absolute(ProjectSettings.globalize_path(revocation_marker_path))


func _ensure_legacy_reader() -> void:
	if legacy_token_reader == null:
		legacy_token_reader = LegacyAuthTokenReaderScript.new()
