extends GutTest

const LocalServerProcess := preload("res://scripts/boot/local_server_process.gd")

var _created_paths: Array[String] = []
var _running := false
var _killed_ids: Array[int] = []


func before_each() -> void:
	_created_paths.clear()
	_killed_ids.clear()
	_running = false


func test_non_packaged_build_is_noop() -> void:
	var process = _make_process()
	process.required_override = 0
	assert_eq(process.start(), OK)
	assert_eq(_created_paths, [])


func test_packaged_build_requires_bundled_server() -> void:
	var process = _make_process()
	process.file_exists_handler = func(_path: String) -> bool: return false
	assert_eq(process.start(), ERR_FILE_NOT_FOUND)


func test_start_is_idempotent_and_stop_kills_owned_process() -> void:
	var process = _make_process()
	assert_eq(process.start(), OK)
	assert_eq(process.process_id, 4242)
	assert_eq(_created_paths.size(), 1)
	_running = true
	assert_eq(process.start(), OK)
	assert_eq(_created_paths.size(), 1)
	assert_eq(process.stop(), OK)
	assert_eq(_killed_ids, [4242])
	assert_eq(process.process_id, -1)


func test_windows_and_macos_paths_match_package_layout() -> void:
	var process = _make_process()
	process.executable_path_override = "C:/Games/SpaceRocks/SpaceRocks.exe"
	process.platform_name_override = "Windows"
	assert_eq(process.resolve_server_path(), "C:/Games/SpaceRocks/space-rocks-server.exe")

	process.executable_path_override = "/Applications/Space Rocks.app/Contents/MacOS/Space Rocks"
	process.platform_name_override = "macOS"
	assert_eq(
		process.resolve_server_path(),
		"/Applications/Space Rocks.app/Contents/Helpers/space-rocks-server"
	)


func _make_process():
	var process := LocalServerProcess.new()
	add_child_autofree(process)
	process.required_override = 1
	process.platform_name_override = "Windows"
	process.executable_path_override = "C:/Games/SpaceRocks/SpaceRocks.exe"
	process.file_exists_handler = func(_path: String) -> bool: return true
	process.create_process_handler = func(path: String) -> int:
		_created_paths.append(path)
		_running = true
		return 4242
	process.process_running_handler = func(_pid: int) -> bool: return _running
	process.kill_process_handler = func(pid: int) -> int:
		_killed_ids.append(pid)
		_running = false
		return OK
	return process
