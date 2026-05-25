extends RefCounted

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

static var default_level := LEVEL_DEBUG
static var category_levels := {}


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


static func debug(category: String, message: String) -> void:
	_log(category, LEVEL_DEBUG, "debug", message)


static func info(category: String, message: String) -> void:
	_log(category, LEVEL_INFO, "info", message)


static func warn(category: String, message: String) -> void:
	_log(category, LEVEL_WARN, "warn", message)


static func error(category: String, message: String) -> void:
	_log(category, LEVEL_ERROR, "error", message)


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


static func _log(category: String, level: int, level_name: String, message: String) -> void:
	if !_should_log(category, level):
		return

	var line := "[%s][%s] %s" % [category, level_name, message]

	match level:
		LEVEL_WARN:
			push_warning(line)
		LEVEL_ERROR:
			push_error(line)
		_:
			print(line)


static func _should_log(category: String, level: int) -> bool:
	var active_level := default_level
	if category_levels.has(category):
		active_level = category_levels[category]

	return level >= active_level && active_level != LEVEL_OFF