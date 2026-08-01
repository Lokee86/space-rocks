extends Node
class_name LocalAlphaProfileSmoke

const ApiConfig := preload("res://scripts/api/api_config.gd")
const HEALTH_ATTEMPTS := 50
const CONDITION_ATTEMPTS := 150
const CLEANUP_ATTEMPTS := 20
const RETRY_SECONDS := 0.1


func wait_for_server() -> bool:
	for _attempt in HEALTH_ATTEMPTS:
		var response := await _request(HTTPClient.METHOD_GET, "/health")
		if int(response.get("code", 0)) == 200 \
				&& str(response.get("text", "")).strip_edges() == "OK":
			return true
		await get_tree().create_timer(RETRY_SECONDS).timeout
	return false


func create_selected_profile(display_name: String) -> Dictionary:
	var list_response := await _request(HTTPClient.METHOD_GET, "/api/player-data/local-profiles")
	if int(list_response.get("code", 0)) != 200:
		return _failure("could not list local profiles")
	for profile in _profiles_from(list_response):
		if str(profile.get("display_name", "")) == display_name:
			await _request(
				HTTPClient.METHOD_DELETE,
				"/api/player-data/local-profiles/%s" % str(profile.get("local_profile_id", ""))
			)

	var create_response := await _request(
		HTTPClient.METHOD_POST,
		"/api/player-data/local-profiles",
		{"display_name": display_name, "seed_from_guest_stats": false}
	)
	if int(create_response.get("code", 0)) != 201:
		return _failure("could not create local profile")
	var created_profile: Dictionary = create_response.get("json", {}).get("profile", {})
	var profile_id := str(created_profile.get("local_profile_id", ""))
	if profile_id.is_empty():
		return _failure("created local profile had no id")

	var default_response := await _request(
		HTTPClient.METHOD_PUT,
		"/api/player-data/local-profiles/default",
		{"identity_kind": "local_profile", "local_profile_id": profile_id}
	)
	if int(default_response.get("code", 0)) != 200:
		return _failure("could not set default local profile")
	return {"ok": true, "profile_id": profile_id}


func wait_for_persisted_stats(profile_id: String, score: int) -> bool:
	for _attempt in CONDITION_ATTEMPTS:
		var stats := await _load_profile_stats(profile_id)
		if _stats_match(stats, score):
			return true
		await get_tree().create_timer(RETRY_SECONDS).timeout
	return false


func verify_persistent_profile_and_stats(display_name: String, score: int) -> String:
	var list_response := await _request(HTTPClient.METHOD_GET, "/api/player-data/local-profiles")
	if int(list_response.get("code", 0)) != 200:
		return "could not list persisted local profiles"

	var profile_id := ""
	for profile in _profiles_from(list_response):
		if str(profile.get("display_name", "")) == display_name:
			profile_id = str(profile.get("local_profile_id", ""))
			break
	if profile_id.is_empty():
		return "seeded local profile did not persist across restart"

	var default_response := await _request(HTTPClient.METHOD_GET, "/api/player-data/local-profiles/default")
	var default_profile: Dictionary = default_response.get("json", {}).get("default_profile", {})
	if int(default_response.get("code", 0)) != 200 \
			|| str(default_profile.get("local_profile_id", "")) != profile_id:
		return "default local profile did not persist across restart"
	if !_stats_match(await _load_profile_stats(profile_id), score):
		return "completed match stats did not persist across restart"

	var reset_response := await _request(
		HTTPClient.METHOD_PUT,
		"/api/player-data/local-profiles/default",
		{"identity_kind": "guest", "local_profile_id": ""}
	)
	if int(reset_response.get("code", 0)) != 200:
		return "could not reset default profile after verification"
	return await _delete_profile_with_confirmation(profile_id)


func _delete_profile_with_confirmation(profile_id: String) -> String:
	var last_response := {}
	for _attempt in CLEANUP_ATTEMPTS:
		last_response = await _request(
			HTTPClient.METHOD_DELETE,
			"/api/player-data/local-profiles/%s" % profile_id
		)
		if int(last_response.get("code", 0)) == 204:
			return ""

		var list_response := await _request(HTTPClient.METHOD_GET, "/api/player-data/local-profiles")
		if int(list_response.get("code", 0)) == 200:
			var profile_still_exists := false
			for profile in _profiles_from(list_response):
				if str(profile.get("local_profile_id", "")) == profile_id:
					profile_still_exists = true
					break
			if !profile_still_exists:
				return ""
		await get_tree().create_timer(RETRY_SECONDS).timeout

	return "could not clean up release-gate local profile (HTTP %d: %s)" % [
		int(last_response.get("code", 0)),
		str(last_response.get("text", "")).strip_edges(),
	]


func _load_profile_stats(profile_id: String) -> Dictionary:
	var profile_response := await _request(
		HTTPClient.METHOD_POST,
		"/api/player-data/profile",
		{
			"play_mode": "single_player",
			"identity_kind": "local_profile",
			"local_profile_id": profile_id,
		}
	)
	if int(profile_response.get("code", 0)) != 200:
		return {}
	var profile = profile_response.get("json", {}).get("profile", {})
	if typeof(profile) != TYPE_DICTIONARY:
		return {}
	var stats = profile.get("stats", {})
	return stats if typeof(stats) == TYPE_DICTIONARY else {}


func _stats_match(stats: Dictionary, score: int) -> bool:
	return int(stats.get("games_played", 0)) >= 1 \
			&& int(stats.get("total_score", 0)) >= score \
			&& int(stats.get("high_score", 0)) >= score \
			&& int(stats.get("ship_deaths", 0)) >= 1


func _profiles_from(response: Dictionary) -> Array:
	var parsed = response.get("json", {})
	if typeof(parsed) != TYPE_DICTIONARY:
		return []
	var profiles = parsed.get("profiles", [])
	return profiles if typeof(profiles) == TYPE_ARRAY else []


func _request(method: int, path: String, payload: Dictionary = {}) -> Dictionary:
	var request := HTTPRequest.new()
	request.timeout = 2.0
	add_child(request)
	var headers := PackedStringArray()
	var body := ""
	if !payload.is_empty():
		headers.append("Content-Type: application/json")
		body = JSON.stringify(payload)
	var start_error := request.request(_base_url() + path, headers, method, body)
	if start_error != OK:
		request.queue_free()
		return {"transport": start_error, "code": 0, "text": "", "json": {}}
	var completed = await request.request_completed
	request.queue_free()
	var response_text := (completed[3] as PackedByteArray).get_string_from_utf8()
	return {
		"transport": int(completed[0]),
		"code": int(completed[1]),
		"text": response_text,
		"json": _parse_json_dictionary(response_text),
	}


func _parse_json_dictionary(text: String) -> Dictionary:
	if text.is_empty():
		return {}
	var parser := JSON.new()
	if parser.parse(text) != OK || typeof(parser.data) != TYPE_DICTIONARY:
		return {}
	return parser.data


func _base_url() -> String:
	return ApiConfig.player_data_base_url()


func _failure(message: String) -> Dictionary:
	return {"ok": false, "error": message}
