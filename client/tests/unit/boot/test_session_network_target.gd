extends GutTest

const SessionNetworkTarget := preload("res://scripts/boot/session_network_target.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")

var _environment := {}


func before_each() -> void:
	_environment.clear()
	for environment_name in [Constants.SINGLE_PLAYER_WS_URL_ENV, Constants.MULTIPLAYER_WS_URL_ENV]:
		_environment[environment_name] = {
			"present": OS.has_environment(environment_name),
			"value": OS.get_environment(environment_name),
		}
		OS.unset_environment(environment_name)


func after_each() -> void:
	for environment_name in _environment:
		var original: Dictionary = _environment[environment_name]
		if original.present:
			OS.set_environment(environment_name, original.value)
		else:
			OS.unset_environment(environment_name)


func test_websocket_url_for_single_player_mode_returns_development_url() -> void:
	var url := SessionNetworkTarget.websocket_url_for_runtime(Constants.SESSION_MODE_SINGLE_PLAYER, false)

	assert_eq(url, Constants.SINGLE_PLAYER_WS_URL)


func test_packaged_single_player_mode_returns_ipv4_loopback_url() -> void:
	var url := SessionNetworkTarget.websocket_url_for_runtime(Constants.SESSION_MODE_SINGLE_PLAYER, true)

	assert_eq(url, Constants.LOCAL_PACKAGED_SINGLE_PLAYER_WS_URL)


func test_single_player_mode_accepts_full_url_environment_override() -> void:
	OS.set_environment(Constants.SINGLE_PLAYER_WS_URL_ENV, " wss://single.example.test/socket ")

	var url := SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_SINGLE_PLAYER)

	assert_eq(url, "wss://single.example.test/socket")


func test_websocket_url_for_multiplayer_mode_returns_shared_url() -> void:
	var url := SessionNetworkTarget.websocket_url_for_runtime(Constants.SESSION_MODE_MULTIPLAYER, false)

	assert_eq(url, Constants.MULTIPLAYER_WS_URL)


func test_multiplayer_mode_accepts_full_url_environment_override() -> void:
	OS.set_environment(Constants.MULTIPLAYER_WS_URL_ENV, " wss://multiplayer.example.test/ws ")

	var url := SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_MULTIPLAYER)

	assert_eq(url, "wss://multiplayer.example.test/ws")


func test_websocket_url_for_unknown_mode_returns_empty_string() -> void:
	var url := SessionNetworkTarget.websocket_url_for_mode("unknown")

	assert_eq(url, "")
