extends RefCounted
class_name SessionNetworkTarget

const Constants := preload("res://scripts/generated/constants/constants.gd")
const LOCAL_PACKAGED_ALPHA_FEATURE := "local_packaged_alpha"
const MULTIPLAYER_ALPHA_FEATURE := "multiplayer_alpha"
const HOSTED_MULTIPLAYER_WS_URL := "wss://game.laughingskull.ca/ws"
const LOCAL_SERVER_PORT_ENV := "SPACE_ROCKS_LOCAL_SERVER_PORT"
const DEFAULT_LOCAL_SERVER_PORT := 8080


static func websocket_url_for_mode(mode: String) -> String:
	return websocket_url_for_runtime(
		mode,
		OS.has_feature(LOCAL_PACKAGED_ALPHA_FEATURE),
		OS.has_feature(MULTIPLAYER_ALPHA_FEATURE)
	)


static func websocket_url_for_runtime(
	mode: String,
	local_packaged_alpha: bool,
	multiplayer_alpha: bool = false
) -> String:
	match mode:
		Constants.SESSION_MODE_SINGLE_PLAYER:
			return _single_player_websocket_url(local_packaged_alpha)
		Constants.SESSION_MODE_MULTIPLAYER:
			if multiplayer_alpha:
				return HOSTED_MULTIPLAYER_WS_URL
			return Constants.MULTIPLAYER_WS_URL
		_:
			return ""


static func _single_player_websocket_url(local_packaged_alpha: bool) -> String:
	var port := OS.get_environment(LOCAL_SERVER_PORT_ENV).strip_edges()
	if port.is_valid_int():
		var parsed_port := port.to_int()
		if parsed_port >= 1 && parsed_port <= 65535:
			return "ws://127.0.0.1:%d/ws" % parsed_port
	if local_packaged_alpha:
		return "ws://127.0.0.1:%d/ws" % DEFAULT_LOCAL_SERVER_PORT
	return Constants.SINGLE_PLAYER_WS_URL
