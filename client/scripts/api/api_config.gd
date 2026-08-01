class_name ApiConfig
extends RefCounted

const Constants := preload("res://scripts/generated/constants/constants.gd")
const MULTIPLAYER_ALPHA_FEATURE := "multiplayer_alpha"
const LOCAL_PACKAGED_ALPHA_FEATURE := "local_packaged_alpha"


static func rails_api_base_url() -> String:
	var override_url := _environment_base_url(Constants.RAILS_API_BASE_URL_ENV)
	if !override_url.is_empty():
		return override_url
	return rails_api_base_url_for_runtime(OS.has_feature(MULTIPLAYER_ALPHA_FEATURE))


static func rails_api_base_url_for_runtime(multiplayer_alpha: bool) -> String:
	if multiplayer_alpha:
		return Constants.MULTIPLAYER_RAILS_API_BASE_URL
	return Constants.DEVELOPMENT_RAILS_API_BASE_URL


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
	var override_url := _environment_base_url(Constants.PLAYER_DATA_API_BASE_URL_ENV)
	if !override_url.is_empty():
		return override_url
	return player_data_base_url_for_runtime(
		OS.has_feature(LOCAL_PACKAGED_ALPHA_FEATURE),
		OS.has_feature(MULTIPLAYER_ALPHA_FEATURE)
	)


static func player_data_base_url_for_runtime(local_packaged_alpha: bool, multiplayer_alpha: bool = false) -> String:
	if local_packaged_alpha:
		return Constants.LOCAL_PACKAGED_PLAYER_DATA_API_BASE_URL
	if multiplayer_alpha:
		return Constants.MULTIPLAYER_PLAYER_DATA_API_BASE_URL
	return Constants.DEVELOPMENT_PLAYER_DATA_API_BASE_URL


static func player_data_profile_path() -> String:
	return "%s/api/player-data/profile" % player_data_base_url()


static func player_data_local_profiles_path() -> String:
	return "%s/api/player-data/local-profiles" % player_data_base_url()


static func player_data_local_profile_path(local_profile_id: String) -> String:
	return "%s/api/player-data/local-profiles/%s" % [player_data_base_url(), local_profile_id.uri_encode()]


static func player_data_local_profiles_default_path() -> String:
	return "%s/api/player-data/local-profiles/default" % player_data_base_url()


static func _environment_base_url(environment_name: String) -> String:
	return OS.get_environment(environment_name).strip_edges().trim_suffix("/")
