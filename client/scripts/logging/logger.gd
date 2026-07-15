extends RefCounted

const RollingJSONLWriter := preload("res://scripts/logging/rolling_jsonl_writer.gd")

const LEVEL_DEBUG := 10
const LEVEL_INFO := 20
const LEVEL_WARN := 30
const LEVEL_ERROR := 40
const LEVEL_OFF := 999

const CATEGORY_DEFAULT := "default"
const CATEGORY_SHELL := "shell"
const CATEGORY_LOBBY := "lobby"
const CATEGORY_NETWORK := "network"
const CATEGORY_GAME := "game"
const CATEGORY_WORLD_SYNC := "world_sync"
const CATEGORY_HUD := "hud"
const CATEGORY_INPUT := "input"
const CATEGORY_PACKETS := "packets"

static var default_level := LEVEL_INFO
static var category_levels := {}
static var _file_writer = RollingJSONLWriter.new()


static func set_default_level(level: int) -> void:
	default_level = level


static func set_category_level(category: String, level: int) -> void:
	category_levels[category] = level


static func set_all_categories_level(level: int) -> void:
	default_level = level
	for category in category_levels.keys():
		category_levels[category] = level


static func enable_debug() -> void:
	set_default_level(LEVEL_DEBUG)


static func disable() -> void:
	set_default_level(LEVEL_OFF)


static func reset_for_tests() -> void:
	_reset_file_writer_for_tests()
	default_level = LEVEL_INFO
	category_levels = {}


static func _set_file_writer_for_tests(writer: RefCounted) -> void:
	close_file_output()
	_file_writer = writer


static func _reset_file_writer_for_tests() -> void:
	close_file_output()
	_file_writer = RollingJSONLWriter.new()


static func debug(category: String, message: String) -> void:
	_log(category, LEVEL_DEBUG, message)


static func info(category: String, message: String) -> void:
	_log(category, LEVEL_INFO, message)


static func warn(category: String, message: String) -> void:
	_log(category, LEVEL_WARN, message)


static func error(category: String, message: String) -> void:
	_log(category, LEVEL_ERROR, message)


static func event(
	category: String,
	level: int,
	event_name: String,
	message: String = "",
	fields: Dictionary = {}
) -> void:
	if !_should_log(category, level):
		return

	var record := build_record(category, level, event_name, message, fields)
	_output_record(record)


static func network_event(
	level: int,
	event_name: String,
	message: String = "",
	fields: Dictionary = {}
) -> void:
	event(CATEGORY_NETWORK, level, event_name, message, fields)


static func packets_event(
	level: int,
	event_name: String,
	message: String = "",
	fields: Dictionary = {}
) -> void:
	event(CATEGORY_PACKETS, level, event_name, message, fields)


static func level_name(level: int) -> String:
	match level:
		LEVEL_DEBUG:
			return "debug"
		LEVEL_INFO:
			return "info"
		LEVEL_WARN:
			return "warn"
		LEVEL_ERROR:
			return "error"
		_:
			return "unknown"


static func build_record(
	category: String,
	level: int,
	event_name: String,
	message: String = "",
	fields: Dictionary = {}
) -> Dictionary:
	return {
		"timestamp_unix_ms": int(Time.get_unix_time_from_system() * 1000.0),
		"level": level_name(level),
		"category": category,
		"event": event_name,
		"message": message,
		"fields": fields.duplicate(true),
	}


static func format_console_line(record: Dictionary) -> String:
	var line := "[%s][%s]" % [record.get("category", ""), record.get("level", "unknown")]
	var event_name := str(record.get("event", ""))
	if event_name != "" && event_name != "log_message":
		line += "[%s]" % event_name

	var message := str(record.get("message", ""))
	if message != "":
		line += " %s" % message

	var field_parts: Array[String] = []
	var fields = record.get("fields", {})
	if fields is Dictionary:
		var keys: Array = fields.keys()
		keys.sort()
		for key in keys:
			field_parts.append("%s=%s" % [str(key), _format_field_value(fields[key])])

	if !field_parts.is_empty():
		line += " %s" % " ".join(field_parts)

	return line


static func format_json_line(record: Dictionary) -> String:
	return JSON.stringify(record)


static func configure_file_output(base_dir: String = "user://logs", prefix: String = "client", policy: Dictionary = {}) -> bool:
	if policy.is_empty():
		return _file_writer.configure(base_dir, prefix)
	return _file_writer.configure(base_dir, prefix, policy)


static func close_file_output() -> void:
	_file_writer.close()


static func current_file_output_path() -> String:
	return _file_writer.current_path


static func file_output_status() -> Dictionary:
	return {
		"enabled": _file_writer.enabled,
		"current_path": _file_writer.current_path,
		"failure_count": _file_writer.failure_count,
		"last_failure_message": _file_writer.last_failure_message,
	}


static func shell_debug(message: String) -> void:
	debug(CATEGORY_SHELL, message)


static func shell_info(message: String) -> void:
	info(CATEGORY_SHELL, message)


static func shell_warn(message: String) -> void:
	warn(CATEGORY_SHELL, message)


static func shell_error(message: String) -> void:
	error(CATEGORY_SHELL, message)


static func lobby_debug(message: String) -> void:
	debug(CATEGORY_LOBBY, message)


static func lobby_info(message: String) -> void:
	info(CATEGORY_LOBBY, message)


static func lobby_warn(message: String) -> void:
	warn(CATEGORY_LOBBY, message)


static func lobby_error(message: String) -> void:
	error(CATEGORY_LOBBY, message)


static func network_debug(message: String) -> void:
	debug(CATEGORY_NETWORK, message)


static func network_info(message: String) -> void:
	info(CATEGORY_NETWORK, message)


static func network_warn(message: String) -> void:
	warn(CATEGORY_NETWORK, message)


static func network_error(message: String) -> void:
	error(CATEGORY_NETWORK, message)


static func game_debug(message: String) -> void:
	debug(CATEGORY_GAME, message)


static func game_info(message: String) -> void:
	info(CATEGORY_GAME, message)


static func game_warn(message: String) -> void:
	warn(CATEGORY_GAME, message)


static func game_error(message: String) -> void:
	error(CATEGORY_GAME, message)


static func world_sync_debug(message: String) -> void:
	debug(CATEGORY_WORLD_SYNC, message)


static func world_sync_info(message: String) -> void:
	info(CATEGORY_WORLD_SYNC, message)


static func world_sync_warn(message: String) -> void:
	warn(CATEGORY_WORLD_SYNC, message)


static func world_sync_error(message: String) -> void:
	error(CATEGORY_WORLD_SYNC, message)


static func hud_debug(message: String) -> void:
	debug(CATEGORY_HUD, message)


static func hud_info(message: String) -> void:
	info(CATEGORY_HUD, message)


static func hud_warn(message: String) -> void:
	warn(CATEGORY_HUD, message)


static func hud_error(message: String) -> void:
	error(CATEGORY_HUD, message)


static func input_debug(message: String) -> void:
	debug(CATEGORY_INPUT, message)


static func input_info(message: String) -> void:
	info(CATEGORY_INPUT, message)


static func input_warn(message: String) -> void:
	warn(CATEGORY_INPUT, message)


static func input_error(message: String) -> void:
	error(CATEGORY_INPUT, message)


static func packets_debug(message: String) -> void:
	debug(CATEGORY_PACKETS, message)


static func packets_info(message: String) -> void:
	info(CATEGORY_PACKETS, message)


static func packets_warn(message: String) -> void:
	warn(CATEGORY_PACKETS, message)


static func packets_error(message: String) -> void:
	error(CATEGORY_PACKETS, message)


static func _log(category: String, level: int, message: String) -> void:
	if !_should_log(category, level):
		return

	var record := build_record(category, level, "log_message", message)
	_output_record(record)


static func _output_record(record: Dictionary) -> void:
	var line := format_console_line(record)
	var level := str(record.get("level", "unknown"))

	match level:
		"warn":
			push_warning(line)
		"error":
			push_error(line)
		_:
			print(line)

	_file_writer.write_line(format_json_line(record))


static func _should_log(category: String, level: int) -> bool:
	var active_level := default_level
	if category_levels.has(category):
		active_level = category_levels[category]

	return level >= active_level && active_level != LEVEL_OFF


static func _format_field_value(value) -> String:
	match typeof(value):
		TYPE_DICTIONARY, TYPE_ARRAY:
			return JSON.stringify(value)
		_:
			return str(value)
