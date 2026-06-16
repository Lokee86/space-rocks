extends RefCounted
class_name AuthApiClient

const ApiConfigScript := preload("res://scripts/api/api_config.gd")
const ApiHttpClientScript := preload("res://scripts/api/api_http_client.gd")

var api_http_client: ApiHttpClient


func _init(api_http_client_value: ApiHttpClient = null) -> void:
	api_http_client = api_http_client_value if api_http_client_value != null else ApiHttpClientScript.new()


func get_current_user(token: String):
	return await api_http_client.get_json(ApiConfigScript.auth_me_path(), token)


func logout(token: String):
	return await api_http_client.delete_json(ApiConfigScript.auth_logout_path(), {}, token)


func begin_discord_login_session():
	return await api_http_client.post_json(ApiConfigScript.discord_login_sessions_path(), {})


func exchange_discord_login_session(login_session_id: String, poll_secret: String):
	return await api_http_client.post_json(
		ApiConfigScript.discord_login_session_exchange_path(login_session_id),
		{ "poll_secret": poll_secret }
	)
