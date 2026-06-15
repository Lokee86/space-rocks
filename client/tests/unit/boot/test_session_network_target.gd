extends GutTest

const SessionNetworkTarget := preload("res://scripts/boot/session_network_target.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")


func test_websocket_url_for_single_player_mode_returns_single_player_url() -> void:
	var url := SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_SINGLE_PLAYER)

	assert_eq(url, Constants.SINGLE_PLAYER_WS_URL)


func test_websocket_url_for_multiplayer_mode_returns_multiplayer_url() -> void:
	var url := SessionNetworkTarget.websocket_url_for_mode(Constants.SESSION_MODE_MULTIPLAYER)

	assert_eq(url, Constants.MULTIPLAYER_WS_URL)


func test_websocket_url_for_unknown_mode_returns_empty_string() -> void:
	var url := SessionNetworkTarget.websocket_url_for_mode("unknown")

	assert_eq(url, "")
