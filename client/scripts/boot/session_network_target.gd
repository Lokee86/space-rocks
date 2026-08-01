extends RefCounted
class_name SessionNetworkTarget

const Constants := preload("res://scripts/generated/constants/constants.gd")
const LOCAL_PACKAGED_ALPHA_FEATURE := "local_packaged_alpha"


static func websocket_url_for_mode(mode: String) -> String:
	var environment_name := ""
	match mode:
		Constants.SESSION_MODE_SINGLE_PLAYER:
			environment_name = Constants.SINGLE_PLAYER_WS_URL_ENV
		Constants.SESSION_MODE_MULTIPLAYER:
			environment_name = Constants.MULTIPLAYER_WS_URL_ENV
		_:
			return ""

	var override_url := OS.get_environment(environment_name).strip_edges()
	if !override_url.is_empty():
		return override_url
	return websocket_url_for_runtime(mode, OS.has_feature(LOCAL_PACKAGED_ALPHA_FEATURE))


static func websocket_url_for_runtime(mode: String, local_packaged_alpha: bool) -> String:
	match mode:
		Constants.SESSION_MODE_SINGLE_PLAYER:
			if local_packaged_alpha:
				return Constants.LOCAL_PACKAGED_SINGLE_PLAYER_WS_URL
			return Constants.SINGLE_PLAYER_WS_URL
		Constants.SESSION_MODE_MULTIPLAYER:
			return Constants.MULTIPLAYER_WS_URL
		_:
			return ""
