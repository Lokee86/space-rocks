class_name ApiConfig
extends RefCounted

const API_BASE_URL := "http://localhost:3000"


static func auth_me_path() -> String:
	return "%s/auth/me" % API_BASE_URL


static func auth_logout_path() -> String:
	return "%s/auth/logout" % API_BASE_URL


static func discord_login_sessions_path() -> String:
	return "%s/auth/discord/login_sessions" % API_BASE_URL


static func discord_login_session_exchange_path(login_session_id: String) -> String:
	return "%s/auth/discord/login_sessions/%s/exchange" % [API_BASE_URL, login_session_id]
