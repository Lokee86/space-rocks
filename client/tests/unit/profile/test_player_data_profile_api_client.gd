extends GutTest

const PlayerDataProfileApiClient := preload("res://scripts/profile/player_data_profile_api_client.gd")
const ApiConfig := preload("res://scripts/api/api_config.gd")
const ApiHttpClient := preload("res://scripts/api/api_http_client.gd")
const ApiRequestResult := preload("res://scripts/api/api_request_result.gd")


class FakeApiHttpClient:
	extends ApiHttpClient

	var last_url := ""
	var last_body := {}
	var last_token := ""
	var result: ApiRequestResult = ApiRequestResult.success(200, {
		"profile": {
			"callsign": "Guest",
			"activity_status": "OFFLINE",
			"identity_kind": "guest",
			"stats": {
				"total_score": 0,
				"high_score": 0,
				"ship_deaths": 0,
				"games_played": 0,
				"wins": 0,
			},
		}
	})

	func post_json(url: String, body: Dictionary = {}, bearer_token: String = "") -> ApiRequestResult:
		last_url = url
		last_body = body
		last_token = bearer_token
		return result


func test_load_profile_calls_profile_endpoint_with_body_and_bearer_token() -> void:
	var fake_client := FakeApiHttpClient.new()
	var client := PlayerDataProfileApiClient.new(fake_client)

	var profile_result = await client.load_profile("multiplayer", "authenticated_account", "local-profile-123", "bearer-token")

	assert_eq(fake_client.last_url, ApiConfig.player_data_profile_path())
	assert_eq(fake_client.last_body.get("play_mode"), "multiplayer")
	assert_eq(fake_client.last_body.get("identity_kind"), "authenticated_account")
	assert_eq(fake_client.last_body.get("local_profile_id"), "local-profile-123")
	assert_eq(fake_client.last_token, "bearer-token")
	assert_eq(profile_result, fake_client.result)


func test_load_profile_guest_omits_bearer_token() -> void:
	var fake_client := FakeApiHttpClient.new()
	var client := PlayerDataProfileApiClient.new(fake_client)

	await client.load_profile("single_player", "guest")

	assert_eq(fake_client.last_url, ApiConfig.player_data_profile_path())
	assert_eq(fake_client.last_body.get("play_mode"), "single_player")
	assert_eq(fake_client.last_body.get("identity_kind"), "guest")
	assert_eq(fake_client.last_body.get("local_profile_id"), "")
	assert_eq(fake_client.last_token, "")
