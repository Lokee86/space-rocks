extends RefCounted
class_name PlayerDataProfileApiClient

const ApiConfigScript := preload("res://scripts/api/api_config.gd")
const ApiHttpClientScript := preload("res://scripts/api/api_http_client.gd")

var api_http_client: ApiHttpClient


func _init(api_http_client_value = null) -> void:
	api_http_client = api_http_client_value if api_http_client_value != null else ApiHttpClientScript.new()


func load_profile(play_mode: String, identity_kind: String, local_profile_id := "", token := ""):
	var body := {
		"play_mode": play_mode,
		"identity_kind": identity_kind,
		"local_profile_id": local_profile_id,
	}
	return await api_http_client.post_json(ApiConfigScript.player_data_profile_path(), body, token)
