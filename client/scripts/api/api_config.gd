class_name ApiConfig
extends RefCounted

const RAILS_API_BASE_URL := "http://localhost:3000"
const DATA_HANDLER_API_BASE_URL := "http://localhost:8080"


static func auth_me_path() -> String:
	return "%s/api/auth/me" % RAILS_API_BASE_URL


static func auth_logout_path() -> String:
	return "%s/api/auth/logout" % RAILS_API_BASE_URL


static func discord_login_sessions_path() -> String:
	return "%s/api/auth/discord/login_sessions" % RAILS_API_BASE_URL


static func discord_login_session_exchange_path(login_session_id: String) -> String:
	return "%s/api/auth/discord/login_sessions/%s/exchange" % [RAILS_API_BASE_URL, login_session_id]


static func player_stats_path() -> String:
	return "%s/api/player/stats" % RAILS_API_BASE_URL


static func player_data_profile_path() -> String:
	return "%s/api/player-data/profile" % DATA_HANDLER_API_BASE_URL


static func player_data_local_profiles_path() -> String:
	return "%s/api/player-data/local-profiles" % DATA_HANDLER_API_BASE_URL


static func player_data_local_profile_path(local_profile_id: String) -> String:
	return "%s/api/player-data/local-profiles/%s" % [DATA_HANDLER_API_BASE_URL, local_profile_id.uri_encode()]


static func player_data_local_profiles_default_path() -> String:
	return "%s/api/player-data/local-profiles/default" % DATA_HANDLER_API_BASE_URL
