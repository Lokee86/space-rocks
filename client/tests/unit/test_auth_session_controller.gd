extends GutTest

const AuthSessionController := preload("res://scripts/auth/auth_session_controller.gd")
const AuthSession := preload("res://scripts/auth/auth_session.gd")
const AuthTokenStore := preload("res://scripts/auth/auth_token_store.gd")
const ApiRequestResult := preload("res://scripts/api/api_request_result.gd")

const TEST_TOKEN_PATH := "user://test_auth_session_controller_token.json"


class FakeAuthApiClient:
	extends RefCounted

	var current_user_result: ApiRequestResult
	var logout_result: ApiRequestResult
	var logout_tokens: Array[String] = []

	func get_current_user(_token: String):
		return current_user_result

	func logout(token: String):
		logout_tokens.append(token)
		return logout_result


func before_each() -> void:
	_cleanup_token_file()


func after_each() -> void:
	_cleanup_token_file()


func test_initialize_from_saved_token_with_no_token_emits_signed_out_state() -> void:
	var controller := _create_controller(FakeAuthApiClient.new())
	var auth_state_changed := false

	controller.auth_state_changed.connect(func() -> void:
		auth_state_changed = true
	)

	controller.initialize_from_saved_token()

	assert_true(auth_state_changed)
	assert_false(controller.get_session().is_signed_in())
	assert_eq(controller.get_session().display_name, "")


func test_initialize_from_saved_token_with_valid_token_populates_auth_session() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.current_user_result = ApiRequestResult.success(200, {
		"user": {
			"id": 42,
			"display_name": "Ada Lovelace",
			"email": "ada@example.com",
		}
	})

	var controller := _create_controller(fake_client)
	controller.auth_token_store.save_token("bearer-token")

	controller.initialize_from_saved_token()
	await get_tree().process_frame
	await get_tree().process_frame

	var session := controller.get_session()
	assert_true(session.is_signed_in())
	assert_eq(session.token, "bearer-token")
	assert_eq(session.user_id, 42)
	assert_eq(session.display_name, "Ada Lovelace")
	assert_eq(session.email, "ada@example.com")


func test_initialize_from_saved_token_with_invalid_token_clears_saved_token() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.current_user_result = ApiRequestResult.failure(401, "invalid")

	var controller := _create_controller(fake_client)
	controller.auth_token_store.save_token("bearer-token")

	controller.initialize_from_saved_token()
	await get_tree().process_frame
	await get_tree().process_frame

	assert_false(controller.get_session().is_signed_in())
	assert_eq(controller.get_session().token, "")
	assert_eq(controller.auth_token_store.load_token(), "")


func test_logout_clears_auth_session_and_token_store() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.logout_result = ApiRequestResult.success(200, {})

	var controller := _create_controller(fake_client)
	controller.auth_token_store.save_token("bearer-token")
	controller.auth_session.set_signed_in("bearer-token", {
		"id": 42,
		"display_name": "Ada Lovelace",
	})

	controller.logout()
	await get_tree().process_frame

	assert_false(controller.get_session().is_signed_in())
	assert_eq(controller.get_session().token, "")
	assert_eq(controller.auth_token_store.load_token(), "")
	assert_eq(fake_client.logout_tokens, ["bearer-token"])


func _create_controller(fake_client) -> AuthSessionController:
	var controller := AuthSessionController.new()
	controller.auth_session = AuthSession.new()
	controller.auth_token_store = AuthTokenStore.new()
	controller.auth_token_store.token_path = TEST_TOKEN_PATH
	controller.auth_api_client = fake_client
	add_child_autofree(controller)
	return controller


func _cleanup_token_file() -> void:
	if FileAccess.file_exists(TEST_TOKEN_PATH):
		DirAccess.remove_absolute(ProjectSettings.globalize_path(TEST_TOKEN_PATH))
