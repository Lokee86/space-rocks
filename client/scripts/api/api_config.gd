class_name ApiConfig
extends RefCounted

const DEVELOPMENT_RAILS_API_BASE_URL := "http://localhost:3000"
const MULTIPLAYER_RAILS_API_BASE_URL := "https://api.laughingskull.ca"
const DEVELOPMENT_DATA_HANDLER_API_BASE_URL := "http://localhost:8080"
const MULTIPLAYER_DATA_HANDLER_API_BASE_URL := "https://game.laughingskull.ca"
const MULTIPLAYER_ALPHA_FEATURE := "multiplayer_alpha"
const LOCAL_PACKAGED_ALPHA_FEATURE := "local_packaged_alpha"
const LOCAL_SERVER_PORT_ENV := "SPACE_ROCKS_LOCAL_SERVER_PORT"
const DEFAULT_LOCAL_SERVER_PORT := 8080


static func rails_api_base_url() -> String:
	if OS.has_feature(MULTIPLAYER_ALPHA_FEATURE):
		return MULTIPLAYER_RAILS_API_BASE_URL
	return DEVELOPMENT_RAILS_API_BASE_URL


static func auth_me_path() -> String:
	return "%s/api/auth/me" % rails_api_base_url()


static func auth_logout_path() -> String:
	return "%s/api/auth/logout" % rails_api_base_url()


static func discord_login_sessions_path() -> String:
	return "%s/api/auth/discord/login_sessions" % rails_api_base_url()


static func discord_login_session_exchange_path(login_session_id: String) -> String:
	return "%s/api/auth/discord/login_sessions/%s/exchange" % [rails_api_base_url(), login_session_id]


static func player_stats_path() -> String:
	return "%s/api/player/stats" % rails_api_base_url()


static func player_data_base_url() -> String:
	return player_data_base_url_for_runtime(OS.has_feature(LOCAL_PACKAGED_ALPHA_FEATURE))


static func player_data_base_url_for_runtime(local_packaged_alpha: bool) -> String:
	var port := OS.get_environment(LOCAL_SERVER_PORT_ENV).strip_edges()
	if port.is_valid_int():
		var parsed_port := port.to_int()
		if parsed_port >= 1 && parsed_port <= 65535:
			return "http://127.0.0.1:%d" % parsed_port
	if local_packaged_alpha:
		return "http://127.0.0.1:%d" % DEFAULT_LOCAL_SERVER_PORT
	if OS.has_feature(MULTIPLAYER_ALPHA_FEATURE):
		return MULTIPLAYER_DATA_HANDLER_API_BASE_URL
	return DEVELOPMENT_DATA_HANDLER_API_BASE_URL


static func player_data_profile_path() -> String:
	return "%s/api/player-data/profile" % player_data_base_url()


static func player_data_local_profiles_path() -> String:
	return "%s/api/player-data/local-profiles" % player_data_base_url()


static func player_data_local_profile_path(local_profile_id: String) -> String:
	return "%s/api/player-data/local-profiles/%s" % [player_data_base_url(), local_profile_id.uri_encode()]


static func player_data_local_profiles_default_path() -> String:
	return "%s/api/player-data/local-profiles/default" % player_data_base_url()
