extends GutTest

const SessionNetworkTarget := preload("res://scripts/boot/session_network_target.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")

var _original_port := ""
var _had_original_port := false


func before_each() -> void:
	_had_original_port = OS.has_environment(SessionNetworkTarget.LOCAL_SERVER_PORT_ENV)
	_original_port = OS.get_environment(SessionNetworkTarget.LOCAL_SERVER_PORT_ENV)
	OS.set_environment(SessionNetworkTarget.LOCAL_SERVER_PORT_ENV, "")


func after_each() -> void:
	if _had_original_port:
		OS.set_environment(SessionNetworkTarget.LOCAL_SERVER_PORT_ENV, _original_port)
	else:
		OS.unset_environment(SessionNetworkTarget.LOCAL_SERVER_PORT_ENV)


func test_websocket_url_for_single_player_mode_returns_single_player_url() -> void:
	var url := SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_SINGLE_PLAYER)

	assert_eq(url, Constants.SINGLE_PLAYER_WS_URL)


func test_single_player_mode_accepts_isolated_local_server_port() -> void:
	OS.set_environment(SessionNetworkTarget.LOCAL_SERVER_PORT_ENV, "43127")

	var url := SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_SINGLE_PLAYER)

	assert_eq(url, "ws://127.0.0.1:43127/ws")


func test_invalid_local_server_port_uses_default_url() -> void:
	OS.set_environment(SessionNetworkTarget.LOCAL_SERVER_PORT_ENV, "65536")

	var url := SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_SINGLE_PLAYER)

	assert_eq(url, Constants.SINGLE_PLAYER_WS_URL)


func test_websocket_url_for_multiplayer_mode_returns_multiplayer_url() -> void:
	var url := SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_MULTIPLAYER)

	assert_eq(url, Constants.MULTIPLAYER_WS_URL)


func test_websocket_url_for_unknown_mode_returns_empty_string() -> void:
	var url := SessionNetworkTarget.websocket_url_for_mode("unknown")

	assert_eq(url, "")
