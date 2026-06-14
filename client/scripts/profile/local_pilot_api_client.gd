extends RefCounted
class_name LocalPilotApiClient

const ApiConfig := preload("res://scripts/api/api_config.gd")
const ApiHttpClient := preload("res://scripts/api/api_http_client.gd")

var api_http_client: ApiHttpClient


func _init(api_http_client_value = null) -> void:
	api_http_client = api_http_client_value if api_http_client_value != null else ApiHttpClient.new()


func list_profiles(token := ""):
	return await api_http_client.get_json(ApiConfig.player_data_local_profiles_path(), token)


func create_profile(display_name: String, seed_from_guest_stats: bool, token := ""):
	var body := {
		"display_name": display_name,
		"seed_from_guest_stats": seed_from_guest_stats,
	}
	return await api_http_client.post_json(ApiConfig.player_data_local_profiles_path(), body, token)
