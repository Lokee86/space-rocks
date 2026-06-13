extends GutTest

const ProfileStatsProvider := preload("res://scripts/profile/profile_stats_provider.gd")
const PlayerDataProfileApiClient := preload("res://scripts/profile/player_data_profile_api_client.gd")
const ApiRequestResult := preload("res://scripts/api/api_request_result.gd")


class FakePlayerDataProfileApiClient:
	extends RefCounted

	var call_count := 0
	var last_play_mode := ""
	var last_identity_kind := ""
	var last_local_profile_id := ""
	var last_token := ""
	var result: ApiRequestResult = ApiRequestResult.success(200, {
		"profile": {
			"callsign": "Guest",
			"activity_status": "OFFLINE",
			"identity_kind": "guest",
			"stats": {
				"total_score": 100,
				"high_score": 75,
				"ship_deaths": 3,
				"games_played": 4,
				"wins": 2,
			},
		}
	})

	func load_profile(play_mode: String, identity_kind: String, local_profile_id := "", token := ""):
		call_count += 1
		last_play_mode = play_mode
		last_identity_kind = identity_kind
		last_local_profile_id = local_profile_id
		last_token = token
		return result


class FakeSession:
	extends RefCounted

	var token := ""
	var signed_in := false

	func is_signed_in() -> bool:
		return signed_in


class FakeAuthSessionController:
	extends RefCounted

	var session

	func get_session():
		return session


func test_guest_loads_profile_stats_via_profile_api() -> void:
	var provider := ProfileStatsProvider.new()
	var api_client := FakePlayerDataProfileApiClient.new()
	provider.configure(null, api_client)

	var stats: Dictionary = await provider.load_stats({
		"play_mode": "single_player",
		"identity_kind": "guest",
	})

	assert_eq(api_client.call_count, 1)
	assert_eq(api_client.last_play_mode, "single_player")
	assert_eq(api_client.last_identity_kind, "guest")
	assert_eq(api_client.last_local_profile_id, "")
	assert_eq(api_client.last_token, "")
	assert_eq(stats, {
		"total_score": 100,
		"high_score": 75,
		"ship_deaths": 3,
		"games_played": 4,
		"wins": 2,
	})


func test_local_profile_loads_profile_stats_with_local_profile_id() -> void:
	var provider := ProfileStatsProvider.new()
	var api_client := FakePlayerDataProfileApiClient.new()
	provider.configure(null, api_client)

	var stats: Dictionary = await provider.load_stats({
		"play_mode": "single_player",
		"identity_kind": "local_profile",
		"local_profile_id": "local-profile-123",
	})

	assert_eq(api_client.call_count, 1)
	assert_eq(api_client.last_play_mode, "single_player")
	assert_eq(api_client.last_identity_kind, "local_profile")
	assert_eq(api_client.last_local_profile_id, "local-profile-123")
	assert_eq(api_client.last_token, "")
	assert_eq(stats, {
		"total_score": 100,
		"high_score": 75,
		"ship_deaths": 3,
		"games_played": 4,
		"wins": 2,
	})


func test_authenticated_loads_profile_stats_with_token() -> void:
	var provider := ProfileStatsProvider.new()
	var api_client := FakePlayerDataProfileApiClient.new()
	var controller := FakeAuthSessionController.new()
	controller.session = _create_session(true, "bearer-token")
	provider.configure(controller, api_client)

	var stats: Dictionary = await provider.load_stats({
		"play_mode": "multiplayer",
		"identity_kind": "authenticated_account",
	})

	assert_eq(api_client.call_count, 1)
	assert_eq(api_client.last_play_mode, "multiplayer")
	assert_eq(api_client.last_identity_kind, "authenticated_account")
	assert_eq(api_client.last_local_profile_id, "")
	assert_eq(api_client.last_token, "bearer-token")
	assert_eq(stats, {
		"total_score": 100,
		"high_score": 75,
		"ship_deaths": 3,
		"games_played": 4,
		"wins": 2,
	})


func test_authenticated_stats_maps_profile_response_to_profile_stats() -> void:
	var provider := ProfileStatsProvider.new()
	var api_client := FakePlayerDataProfileApiClient.new()
	api_client.result = ApiRequestResult.success(200, {
		"profile": {
			"callsign": "Ada",
			"activity_status": "ACTIVE",
			"identity_kind": "authenticated_account",
			"stats": {
				"total_score": 12,
				"high_score": 9,
				"ship_deaths": 3,
				"games_played": 4,
				"wins": 2,
			},
		},
	})
	var controller := FakeAuthSessionController.new()
	controller.session = _create_session(true, "token-123")
	provider.configure(controller, api_client)

	var stats: Dictionary = await provider.load_stats({
		"play_mode": "multiplayer",
		"identity_kind": "authenticated_account",
	})

	assert_eq(api_client.last_play_mode, "multiplayer")
	assert_eq(api_client.last_identity_kind, "authenticated_account")
	assert_eq(api_client.last_token, "token-123")
	assert_eq(stats, {
		"total_score": 12,
		"high_score": 9,
		"ship_deaths": 3,
		"games_played": 4,
		"wins": 2,
	})


func test_authenticated_stats_ignores_extra_sensitive_fields() -> void:
	var provider := ProfileStatsProvider.new()
	var api_client := FakePlayerDataProfileApiClient.new()
	api_client.result = ApiRequestResult.success(200, {
		"profile": {
			"callsign": "Ada",
			"activity_status": "ACTIVE",
			"identity_kind": "authenticated_account",
			"stats": {
				"total_score": 12,
				"high_score": 9,
				"ship_deaths": 3,
				"games_played": 4,
				"wins": 2,
				"user_id": 99,
				"email": "ada@example.com",
				"token": "secret-token",
				"created_at": "2026-06-12T00:00:00Z",
			},
		},
	})
	var controller := FakeAuthSessionController.new()
	controller.session = _create_session(true, "token-123")
	provider.configure(controller, api_client)

	var stats: Dictionary = await provider.load_stats({
		"play_mode": "multiplayer",
		"identity_kind": "authenticated_account",
	})

	assert_false(stats.has("user_id"))
	assert_false(stats.has("email"))
	assert_false(stats.has("token"))
	assert_false(stats.has("created_at"))
	assert_eq(stats, {
		"total_score": 12,
		"high_score": 9,
		"ship_deaths": 3,
		"games_played": 4,
		"wins": 2,
	})


func test_authenticated_stats_missing_stats_body_returns_zero() -> void:
	var provider := ProfileStatsProvider.new()
	var api_client := FakePlayerDataProfileApiClient.new()
	api_client.result = ApiRequestResult.success(200, {
		"profile": {
			"callsign": "Ada",
		},
	})
	var controller := FakeAuthSessionController.new()
	controller.session = _create_session(true, "token-123")
	provider.configure(controller, api_client)

	var stats: Dictionary = await provider.load_stats({
		"play_mode": "multiplayer",
		"identity_kind": "authenticated_account",
	})

	assert_eq(stats, {
		"total_score": 0,
		"high_score": 0,
		"ship_deaths": 0,
		"games_played": 0,
		"wins": 0,
	})


func test_authenticated_without_token_returns_zero_stats() -> void:
	var provider := ProfileStatsProvider.new()
	var api_client := FakePlayerDataProfileApiClient.new()
	var controller := FakeAuthSessionController.new()
	controller.session = _create_session(true, "")
	provider.configure(controller, api_client)

	var stats: Dictionary = await provider.load_stats({
		"play_mode": "multiplayer",
		"identity_kind": "authenticated_account",
	})

	assert_eq(api_client.call_count, 0)
	assert_eq(stats, {
		"total_score": 0,
		"high_score": 0,
		"ship_deaths": 0,
		"games_played": 0,
		"wins": 0,
	})


func test_failed_api_returns_cached_stats_when_available() -> void:
	var provider := ProfileStatsProvider.new()
	var api_client := FakePlayerDataProfileApiClient.new()
	var controller := FakeAuthSessionController.new()
	controller.session = _create_session(true, "bearer-token")
	provider.configure(controller, api_client)

	var cached_stats: Dictionary = await provider.load_stats({
		"play_mode": "multiplayer",
		"identity_kind": "authenticated_account",
	})
	assert_eq(cached_stats["total_score"], 100)

	api_client.result = ApiRequestResult.failure(500, "server_error")
	var stats: Dictionary = await provider.load_stats({
		"play_mode": "multiplayer",
		"identity_kind": "authenticated_account",
	})

	assert_eq(api_client.call_count, 2)
	assert_eq(stats, {
		"total_score": 100,
		"high_score": 75,
		"ship_deaths": 3,
		"games_played": 4,
		"wins": 2,
	})


func test_failed_api_without_cache_returns_zero_stats() -> void:
	var provider := ProfileStatsProvider.new()
	var api_client := FakePlayerDataProfileApiClient.new()
	api_client.result = ApiRequestResult.failure(500, "server_error")
	var controller := FakeAuthSessionController.new()
	controller.session = _create_session(true, "bearer-token")
	provider.configure(controller, api_client)

	var stats: Dictionary = await provider.load_stats({
		"identity_kind": "authenticated_account",
	})

	assert_eq(stats, {
		"total_score": 0,
		"high_score": 0,
		"ship_deaths": 0,
		"games_played": 0,
		"wins": 0,
	})


func _create_session(is_signed_in_value: bool, token_value: String) -> FakeSession:
	var session := FakeSession.new()
	session.signed_in = is_signed_in_value
	session.token = token_value
	return session
