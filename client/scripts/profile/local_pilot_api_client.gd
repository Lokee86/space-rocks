extends RefCounted
class_name LocalPilotApiClient

const ApiConfig := preload("res://scripts/api/api_config.gd")
const ApiHttpClient := preload("res://scripts/api/api_http_client.gd")

var api_http_client: ApiHttpClient


func _init(api_http_client_value = null) -> void:
	api_http_client = api_http_client_value if api_http_client_value != null else ApiHttpClient.new()


func list_profiles(token := ""):
	return await api_http_client.get_json(ApiConfig.player_data_local_profiles_path(), token)


func get_default_profile(token := ""):
	return await api_http_client.get_json(ApiConfig.player_data_local_profiles_default_path(), token)


func create_profile(display_name: String, seed_from_guest_stats: bool, token := ""):
	var body := {
		"display_name": display_name,
		"seed_from_guest_stats": seed_from_guest_stats,
	}
	return await api_http_client.post_json(ApiConfig.player_data_local_profiles_path(), body, token)


func delete_profile(local_profile_id: String, token := ""):
	return await api_http_client.delete_json(ApiConfig.player_data_local_profile_path(local_profile_id), {}, token)


func update_profile_display_name(local_profile_id: String, display_name: String, token := ""):
	var body := {
		"display_name": display_name,
	}
	return await api_http_client.put_json(ApiConfig.player_data_local_profile_path(local_profile_id), body, token)


func set_default_profile(identity_kind: String, local_profile_id := "", token := ""):
	var body := {
		"identity_kind": identity_kind,
		"local_profile_id": local_profile_id,
	}
	return await api_http_client.put_json(ApiConfig.player_data_local_profiles_default_path(), body, token)
