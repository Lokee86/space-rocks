extends GutTest

const AuthSessionController := preload("res://scripts/auth/auth_session_controller.gd")
const AuthSession := preload("res://scripts/auth/auth_session.gd")
const AuthCredentialStore := preload("res://scripts/auth/auth_credential_store.gd")
const ApiRequestResult := preload("res://scripts/api/api_request_result.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const TEST_TOKEN_PATH := "user://test_auth_session_controller_token.json"

var _auth_state_change_count := 0
class InMemoryAuthCredentialStore:
	extends "res://scripts/auth/auth_credential_store.gd"

	var stored_token := ""
	var migration_required := false
	var save_succeeds := true

	func load_token() -> String:
		return stored_token

	func save_token(token: String) -> bool:
		if !save_succeeds:
			return false
		stored_token = token
		migration_required = false
		return true

	func clear_token() -> bool:
		stored_token = ""
		migration_required = false
		return true

	func requires_legacy_migration() -> bool:
		return migration_required


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
	var current_user_trace_ids: Array[String] = []

	func get_current_user(_token: String, trace_id: String = ""):
		current_user_trace_ids.append(trace_id)
		if wait_for_current_user:
			await current_user_released
		return current_user_result

	func begin_discord_login_session(_trace_id: String = ""):
		if wait_for_discord_sign_in:
			await discord_sign_in_released
		return discord_sign_in_result

	func release_current_user() -> void:
		current_user_released.emit()

	func release_discord_sign_in() -> void:
		discord_sign_in_released.emit()

	func logout(token: String, _trace_id: String = ""):
		logout_tokens.append(token)
		return logout_result

class FakeWriter extends RefCounted:
	var written_lines: Array[String] = []

	func write_line(line: String) -> void:
		written_lines.append(line)

	func close() -> void:
		pass

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
	controller.auth_credential_store.save_token("bearer-token")

	controller.initialize_from_saved_token()
	await get_tree().process_frame
	await get_tree().process_frame

	var session := controller.get_session()
	assert_true(session.is_signed_in())
	assert_eq(session.token, "bearer-token")
	assert_eq(session.user_id, 42)
	assert_eq(session.display_name, "Ada Lovelace")
	assert_eq(session.email, "ada@example.com")


func test_failed_legacy_migration_signs_out_and_revokes_token() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.current_user_result = ApiRequestResult.success(200, {
		"user": {"id": 42, "display_name": "Ada Lovelace"},
	})
	fake_client.logout_result = ApiRequestResult.success(200, {})

	var controller := _create_controller(fake_client)
	var store := InMemoryAuthCredentialStore.new()
	store.stored_token = "legacy-token"
	store.migration_required = true
	store.save_succeeds = false
	controller.auth_credential_store = store
	watch_signals(controller)

	controller.initialize_from_saved_token()
	await get_tree().process_frame
	await get_tree().process_frame
	await get_tree().process_frame

	assert_false(controller.get_session().is_signed_in())
	assert_eq(store.stored_token, "")
	assert_signal_emitted_with_parameters(controller, "auth_error", ["Secure credential migration failed. Sign in again."])
	assert_eq(fake_client.logout_tokens, ["legacy-token"])


func test_initialize_from_saved_token_with_invalid_token_clears_saved_token() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.current_user_result = ApiRequestResult.failure(401, "invalid")

	var controller := _create_controller(fake_client)
	controller.auth_credential_store.save_token("bearer-token")

	controller.initialize_from_saved_token()
	await get_tree().process_frame
	await get_tree().process_frame

	assert_false(controller.get_session().is_signed_in())
	assert_eq(controller.get_session().token, "")
	assert_eq(controller.auth_credential_store.load_token(), "")


func test_logout_clears_auth_session_and_token_store() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.logout_result = ApiRequestResult.success(200, {})

	var controller := _create_controller(fake_client)
	controller.auth_credential_store.save_token("bearer-token")
	controller.auth_session.set_signed_in("bearer-token", {
		"id": 42,
		"display_name": "Ada Lovelace",
	})

	controller.logout()
	await get_tree().process_frame

	assert_false(controller.get_session().is_signed_in())
	assert_eq(controller.get_session().token, "")
	assert_eq(controller.auth_credential_store.load_token(), "")
	assert_eq(fake_client.logout_tokens, ["bearer-token"])


func test_logout_supersedes_awaiting_saved_token_validation() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.wait_for_current_user = true
	fake_client.current_user_result = ApiRequestResult.success(200, {
		"user": {"id": 42, "display_name": "Ada Lovelace"},
	})

	var controller := _create_controller(fake_client)
	controller.auth_credential_store.save_token("old-token")
	watch_signals(controller)
	controller.auth_state_changed.connect(_count_auth_state_change)
	controller.initialize_from_saved_token()
	await get_tree().process_frame
	controller.logout()
	fake_client.release_current_user()
	await get_tree().process_frame

	assert_false(controller.get_session().is_signed_in())
	assert_eq(controller.auth_credential_store.load_token(), "")
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
	assert_eq(controller.auth_credential_store.load_token(), "")
	assert_eq(_auth_state_change_count, 1)
	assert_signal_not_emitted(controller, "auth_error")


func _create_controller(fake_client, operation_trace_factory: Callable = Callable()) -> AuthSessionController:
	var controller := AuthSessionController.new()
	controller.auth_session = AuthSession.new()
	controller.auth_credential_store = InMemoryAuthCredentialStore.new()
	controller.auth_api_client = fake_client
	controller.configure(fake_client, operation_trace_factory)
	add_child_autofree(controller)
	return controller


func _count_auth_state_change() -> void:
	_auth_state_change_count += 1


func _cleanup_token_file() -> void:
	if FileAccess.file_exists(TEST_TOKEN_PATH):
		DirAccess.remove_absolute(ProjectSettings.globalize_path(TEST_TOKEN_PATH))

func test_saved_token_validation_owns_trace_only_when_remote_validation_occurs() -> void:
	var controller := _create_controller(FakeAuthApiClient.new())
	controller.initialize_from_saved_token()
	assert_eq(controller.active_auth_trace_id(), "")

	var fake_client := FakeAuthApiClient.new()
	fake_client.wait_for_current_user = true
	fake_client.current_user_result = ApiRequestResult.success(200, {
		"user": {
			"id": 42,
			"display_name": "Ada Lovelace",
		}
	})
	var controller_with_token := _create_controller(
		fake_client,
		func(operation_name: String):
			return ClientOperationTrace.new(
				operation_name,
				func() -> String: return "00000000-0000-4000-8000-000000000021"
			)
	)
	var token_store := InMemoryAuthCredentialStore.new()
	token_store.stored_token = "bearer-token"
	controller_with_token.auth_credential_store = token_store
	controller_with_token.initialize_from_saved_token()

	assert_eq(controller_with_token.active_auth_trace_id(), "00000000-0000-4000-8000-000000000021")
	await get_tree().process_frame
	fake_client.release_current_user()
	await get_tree().process_frame
	await get_tree().process_frame
	assert_eq(controller_with_token.active_auth_trace_id(), "")


func test_replacing_discord_sign_in_replaces_active_trace() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.wait_for_discord_sign_in = true
	var state := {"index": 0}
	var ids := [
		"00000000-0000-4000-8000-000000000031",
		"00000000-0000-4000-8000-000000000032",
	]
	var controller := _create_controller(fake_client, func(operation_name: String):
		var trace_id: String = ids[state["index"]]
		state["index"] += 1
		return ClientOperationTrace.new(operation_name, func() -> String: return trace_id)
	)

	controller.request_discord_sign_in()
	var first_trace_id := controller.active_auth_trace_id()
	controller.request_discord_sign_in()

	assert_eq(first_trace_id, ids[0])
	assert_eq(controller.active_auth_trace_id(), ids[1])

func _auth_records(writer: FakeWriter) -> Array:
	var records: Array = []
	for line in writer.written_lines:
		records.append(JSON.parse_string(line))
	return records


func _fixed_auth_trace(operation_name: String) -> ClientOperationTrace:
	return ClientOperationTrace.new(
		operation_name,
		func() -> String: return "00000000-0000-4000-8000-000000000061"
	)


func test_saved_token_auth_start_and_success_share_one_trace() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var fake_client := FakeAuthApiClient.new()
	fake_client.current_user_result = ApiRequestResult.success(200, {
		"user": {"id": 42, "display_name": "Ada Lovelace"},
	})
	var controller := _create_controller(fake_client, func(operation_name: String): return _fixed_auth_trace(operation_name))
	var token_store := InMemoryAuthCredentialStore.new()
	token_store.stored_token = "bearer-token"
	controller.auth_credential_store = token_store

	controller.initialize_from_saved_token()
	await get_tree().process_frame
	await get_tree().process_frame

	var records := _auth_records(writer)
	assert_eq(records[0]["event"], ObservabilityContract.EVENT_AUTH_FLOW_STARTED)
	assert_eq(records[1]["event"], ObservabilityContract.EVENT_AUTH_SUCCEEDED)
	assert_eq(records[0]["trace_id"], records[1]["trace_id"])
	assert_eq(records[0]["fields"]["auth_operation"], "saved_token_validation")
	assert_eq(fake_client.current_user_trace_ids, ["00000000-0000-4000-8000-000000000061"])


func test_saved_token_provider_failure_is_distinct_from_invalid_token_failure() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var fake_client := FakeAuthApiClient.new()
	fake_client.current_user_result = ApiRequestResult.failure(503, "unavailable")
	var controller := _create_controller(fake_client, func(operation_name: String): return _fixed_auth_trace(operation_name))
	var token_store := InMemoryAuthCredentialStore.new()
	token_store.stored_token = "bearer-token"
	controller.auth_credential_store = token_store

	controller.initialize_from_saved_token()
	await get_tree().process_frame
	await get_tree().process_frame
	assert_push_error_count(1)

	var records := _auth_records(writer)
	assert_eq(records[1]["event"], ObservabilityContract.EVENT_AUTH_PROVIDER_UNAVAILABLE)
	assert_eq(records[1]["fields"]["failure_mode"], "http_5xx")
	assert_eq(token_store.stored_token, "bearer-token")

	ClientLogger.reset_for_tests()
	writer = FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var invalid_client := FakeAuthApiClient.new()
	invalid_client.current_user_result = ApiRequestResult.failure(401, "invalid")
	var invalid_controller := _create_controller(invalid_client, func(operation_name: String): return _fixed_auth_trace(operation_name))
	var invalid_store := InMemoryAuthCredentialStore.new()
	invalid_store.stored_token = "bearer-token"
	invalid_controller.auth_credential_store = invalid_store

	invalid_controller.initialize_from_saved_token()
	await get_tree().process_frame
	await get_tree().process_frame

	records = _auth_records(writer)
	assert_eq(records[1]["event"], ObservabilityContract.EVENT_AUTH_FAILED)
	assert_eq(records[1]["fields"]["failure_mode"], "invalid_or_expired_token")
	assert_eq(invalid_store.stored_token, "")


func test_saved_token_network_failure_preserves_saved_credential() -> void:
	var fake_client := FakeAuthApiClient.new()
	fake_client.current_user_result = ApiRequestResult.failure(0, "network_failure_7")
	var controller := _create_controller(fake_client)
	var token_store := InMemoryAuthCredentialStore.new()
	token_store.stored_token = "bearer-token"
	controller.auth_credential_store = token_store

	controller.initialize_from_saved_token()
	await get_tree().process_frame
	await get_tree().process_frame
	assert_push_error_count(1)

	assert_false(controller.get_session().is_signed_in())
	assert_eq(token_store.stored_token, "bearer-token")


func test_cancelled_saved_token_validation_emits_no_stale_terminal_event() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var fake_client := FakeAuthApiClient.new()
	fake_client.wait_for_current_user = true
	fake_client.current_user_result = ApiRequestResult.success(200, {
		"user": {"id": 42, "display_name": "Ada Lovelace"},
	})
	var controller := _create_controller(fake_client, func(operation_name: String): return _fixed_auth_trace(operation_name))
	var token_store := InMemoryAuthCredentialStore.new()
	token_store.stored_token = "bearer-token"
	controller.auth_credential_store = token_store

	controller.initialize_from_saved_token()
	await get_tree().process_frame
	controller.logout()
	fake_client.release_current_user()
	await get_tree().process_frame

	var records := _auth_records(writer)
	assert_eq(records.size(), 1)
	assert_eq(records[0]["event"], ObservabilityContract.EVENT_AUTH_FLOW_STARTED)

func test_saved_token_scene_tree_failure_is_auth_failed() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var fake_client := FakeAuthApiClient.new()
	fake_client.current_user_result = ApiRequestResult.failure(0, "scene_tree_unavailable")
	var controller := _create_controller(fake_client, func(operation_name: String): return _fixed_auth_trace(operation_name))
	var token_store := InMemoryAuthCredentialStore.new()
	token_store.stored_token = "bearer-token"
	controller.auth_credential_store = token_store

	controller.initialize_from_saved_token()
	await get_tree().process_frame
	await get_tree().process_frame

	var records := _auth_records(writer)
	assert_eq(records[1]["event"], ObservabilityContract.EVENT_AUTH_FAILED)
	assert_eq(records[1]["fields"]["failure_mode"], "scene_tree_unavailable")
