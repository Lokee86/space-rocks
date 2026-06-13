extends GutTest

const ApiConfig := preload("res://scripts/api/api_config.gd")


func test_auth_me_path_returns_rails_api_url() -> void:
	assert_eq(ApiConfig.auth_me_path(), "http://localhost:3000/api/auth/me")


func test_player_stats_path_returns_rails_api_player_stats_url() -> void:
	assert_eq(ApiConfig.player_stats_path(), "http://localhost:3000/api/player/stats")


func test_player_data_profile_path_returns_data_handler_url() -> void:
	assert_eq(ApiConfig.player_data_profile_path(), "http://localhost:8080/api/player-data/profile")
