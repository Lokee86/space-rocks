extends GutTest

const ApiConfig := preload("res://scripts/api/api_config.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")

var _environment := {}


func before_each() -> void:
	_environment.clear()
	for environment_name in [Constants.RAILS_API_BASE_URL_ENV, Constants.PLAYER_DATA_API_BASE_URL_ENV]:
		_environment[environment_name] = {
			"present": OS.has_environment(environment_name),
			"value": OS.get_environment(environment_name),
		}
		OS.unset_environment(environment_name)


func after_each() -> void:
	for environment_name in _environment:
		var original: Dictionary = _environment[environment_name]
		if original.present:
			OS.set_environment(environment_name, original.value)
		else:
			OS.unset_environment(environment_name)


func test_development_rails_api_base_url() -> void:
	assert_eq(ApiConfig.rails_api_base_url_for_runtime(false), Constants.DEVELOPMENT_RAILS_API_BASE_URL)


func test_multiplayer_rails_api_base_url() -> void:
	assert_eq(ApiConfig.rails_api_base_url_for_runtime(true), Constants.MULTIPLAYER_RAILS_API_BASE_URL)


func test_rails_api_accepts_full_url_environment_override() -> void:
	OS.set_environment(Constants.RAILS_API_BASE_URL_ENV, " https://rails.example.test/ ")

	assert_eq(ApiConfig.rails_api_base_url(), "https://rails.example.test")
	assert_eq(ApiConfig.auth_me_path(), "https://rails.example.test/api/auth/me")


func test_development_player_data_base_url() -> void:
	assert_eq(ApiConfig.player_data_base_url_for_runtime(false, false), Constants.DEVELOPMENT_PLAYER_DATA_API_BASE_URL)


func test_multiplayer_player_data_base_url() -> void:
	assert_eq(ApiConfig.player_data_base_url_for_runtime(false, true), Constants.MULTIPLAYER_PLAYER_DATA_API_BASE_URL)


func test_packaged_player_data_base_url_uses_ipv4_loopback() -> void:
	assert_eq(ApiConfig.player_data_base_url_for_runtime(true, false), Constants.LOCAL_PACKAGED_PLAYER_DATA_API_BASE_URL)


func test_player_data_accepts_full_url_environment_override() -> void:
	OS.set_environment(Constants.PLAYER_DATA_API_BASE_URL_ENV, " https://player-data.example.test/ ")

	assert_eq(ApiConfig.player_data_base_url(), "https://player-data.example.test")
	assert_eq(ApiConfig.player_data_profile_path(), "https://player-data.example.test/api/player-data/profile")
