extends GutTest

const ApiConfig := preload("res://scripts/api/api_config.gd")


func test_development_rails_api_base_url() -> void:
	assert_eq(ApiConfig.rails_api_base_url_for_runtime(false), "http://localhost:3000")


func test_multiplayer_rails_api_base_url() -> void:
	assert_eq(ApiConfig.rails_api_base_url_for_runtime(true), "https://api.laughingskull.ca")


func test_development_player_data_base_url() -> void:
	assert_eq(ApiConfig.player_data_base_url_for_runtime(false, false), "http://localhost:8080")


func test_multiplayer_player_data_base_url() -> void:
	assert_eq(ApiConfig.player_data_base_url_for_runtime(false, true), "https://game.laughingskull.ca")


func test_packaged_player_data_base_url_uses_ipv4_loopback() -> void:
	assert_eq(ApiConfig.player_data_base_url_for_runtime(true, false), "http://127.0.0.1:8080")
