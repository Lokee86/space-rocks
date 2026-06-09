extends GutTest

const AuthTokenStore := preload("res://scripts/auth/auth_token_store.gd")


func before_each() -> void:
	_cleanup_token_file()


func after_each() -> void:
	_cleanup_token_file()


func test_save_and_load_token_round_trip() -> void:
	var store := AuthTokenStore.new()
	store.token_path = "user://test_auth_token.json"

	store.save_token("bearer-token")

	assert_eq(store.load_token(), "bearer-token")


func test_clear_token_removes_stored_token() -> void:
	var store := AuthTokenStore.new()
	store.token_path = "user://test_auth_token.json"
	store.save_token("bearer-token")

	store.clear_token()

	assert_eq(store.load_token(), "")


func _cleanup_token_file() -> void:
	var path := "user://test_auth_token.json"
	if FileAccess.file_exists(path):
		DirAccess.remove_absolute(ProjectSettings.globalize_path(path))
