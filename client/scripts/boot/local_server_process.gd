extends Node
class_name LocalServerProcess

const WINDOWS_SERVER_NAME := "space-rocks-server.exe"
const MACOS_SERVER_NAME := "space-rocks-server"

var process_id := -1
var server_path_override := ""
var required_override := -1
var platform_name_override := ""
var executable_path_override := ""
var file_exists_handler: Callable
var create_process_handler: Callable
var process_running_handler: Callable
var kill_process_handler: Callable


func is_required() -> bool:
	if required_override >= 0:
		return required_override == 1
	return OS.has_feature("local_packaged_alpha")


func start() -> Error:
	if !is_required():
		return OK
	if is_running():
		return OK

	var server_path := resolve_server_path()
	if server_path.is_empty() || !_file_exists(server_path):
		return ERR_FILE_NOT_FOUND

	var started_process_id := _create_process(server_path)
	if started_process_id <= 0:
		return ERR_CANT_FORK
	process_id = started_process_id
	return OK


func stop() -> Error:
	if process_id <= 0:
		return OK
	if !_is_process_running(process_id):
		process_id = -1
		return OK

	var result := _kill_process(process_id)
	if result == OK:
		process_id = -1
	return result


func is_running() -> bool:
	return process_id > 0 && _is_process_running(process_id)


func resolve_server_path() -> String:
	if !server_path_override.is_empty():
		return ProjectSettings.globalize_path(server_path_override)

	var platform := platform_name_override if !platform_name_override.is_empty() else OS.get_name()
	var executable_path := executable_path_override if !executable_path_override.is_empty() else OS.get_executable_path()
	var executable_directory := executable_path.get_base_dir()
	if platform == "Windows":
		return executable_directory.path_join(WINDOWS_SERVER_NAME)
	if platform == "macOS":
		return executable_directory.path_join("../Helpers/%s" % MACOS_SERVER_NAME).simplify_path()
	return ""


func _file_exists(path: String) -> bool:
	if file_exists_handler.is_valid():
		return bool(file_exists_handler.call(path))
	return FileAccess.file_exists(path)


func _create_process(path: String) -> int:
	if create_process_handler.is_valid():
		return int(create_process_handler.call(path))
	return OS.create_process(path, PackedStringArray(), false)


func _is_process_running(pid: int) -> bool:
	if process_running_handler.is_valid():
		return bool(process_running_handler.call(pid))
	return OS.is_process_running(pid)


func _kill_process(pid: int) -> Error:
	if kill_process_handler.is_valid():
		return int(kill_process_handler.call(pid)) as Error
	return OS.kill(pid)
