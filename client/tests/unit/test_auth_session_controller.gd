extends GutTest

const AuthSessionController := preload("res://scripts/auth/auth_session_controller.gd")
const AuthSession := preload("res://scripts/auth/auth_session.gd")
const AuthTokenStore := preload("res://scripts/auth/auth_token_store.gd")
const ApiRequestResult := preload("res://scripts/api/api_request_result.gd")

const TEST_TOKEN_PATH := "user://test_auth_session_controller_token.json"

var _auth_state_change_count := 0


class FakeAuthApiClient:
	extends RefCounted

	signal current_user_released
	signal discord_sign_in_released

	var current_user_result: ApiRequestResult
	var discord_sign_in_result: ApiRequestResult
	var logout_result: ApiRequestResult
	var wait_for_current_user := false
	var wait_for_discord_sign_in := false
	var logout_tokens: Array[String] = []

	func get_current_user(_token: String):
		if wait_for_current_user:
			await current_user_released
		return current_user_result

	func begin_discord_login_session():
		if wait_for_discord_sign_in:
			await discord_sign_in_released
		return discord_sign_in_result

	func release_current_user() -> void:
		current_user_released.emit()

	func release_discord_sign_in() -> void:
		discord_sign_in_released.emit()

	func logout(token: String):
		logout_tokens.append(token)
		return logout_result


func before_each() -> void:
	_auth_state_change_count = 0
	_cleanup_token_file()


func after_each() -> void:
	_cleanup_token_file()


func test_initialize_from_saved_token_with_no_token_emits_signed_out_state() -> void:
	var controller := _create_controller(FakeAuthApiClient.new())
	watch_signals(controller)

	controller.initialize_from_saved_token()

	assert_signal_emitted(controller, "auth_state_changed")
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


func test_logout_supersedes_awaiting_saved_token_validation() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.wait_for_current_user = true
	fake_client.current_user_result = ApiRequestResult.success(200, {
		"user": {"id": 42, "display_name": "Ada Lovelace"},
	})

	var controller := _create_controller(fake_client)
	controller.auth_token_store.save_token("old-token")
	watch_signals(controller)
	controller.auth_state_changed.connect(_count_auth_state_change)
	controller.initialize_from_saved_token()
	await get_tree().process_frame
	controller.logout()
	fake_client.release_current_user()
	await get_tree().process_frame

	assert_false(controller.get_session().is_signed_in())
	assert_eq(controller.auth_token_store.load_token(), "")
	assert_eq(_auth_state_change_count, 1)


func test_logout_supersedes_awaiting_discord_sign_in_start() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.wait_for_discord_sign_in = true
	fake_client.discord_sign_in_result = ApiRequestResult.success(200, {
		"login_session_id": "session-id",
		"poll_secret": "poll-secret",
		"login_url": "https://discord.com",
	})

	var controller := _create_controller(fake_client)
	watch_signals(controller)
	controller.auth_state_changed.connect(_count_auth_state_change)
	controller.request_discord_sign_in()
	await get_tree().process_frame
	controller.logout()
	fake_client.release_discord_sign_in()
	await get_tree().process_frame

	assert_false(controller.get_session().is_signed_in())
	assert_eq(controller.auth_token_store.load_token(), "")
	assert_eq(_auth_state_change_count, 1)
	assert_signal_not_emitted(controller, "auth_error")


func _create_controller(fake_client) -> AuthSessionController:
	var controller := AuthSessionController.new()
	controller.auth_session = AuthSession.new()
	controller.auth_token_store = AuthTokenStore.new()
	controller.auth_token_store.token_path = TEST_TOKEN_PATH
	controller.auth_api_client = fake_client
	add_child_autofree(controller)
	return controller


func _count_auth_state_change() -> void:
	_auth_state_change_count += 1


func _cleanup_token_file() -> void:
	if FileAccess.file_exists(TEST_TOKEN_PATH):
		DirAccess.remove_absolute(ProjectSettings.globalize_path(TEST_TOKEN_PATH))
