extends Node
class_name AuthSessionController

signal auth_state_changed
signal auth_error(message: String)

const AuthSessionScript := preload("res://scripts/auth/auth_session.gd")
const AuthCredentialStoreScript := preload("res://scripts/auth/auth_credential_store.gd")
const AuthApiClientScript := preload("res://scripts/auth/auth_api_client.gd")

const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const DISCORD_POLL_INTERVAL_SECONDS := 1.0
const DISCORD_POLL_TIMEOUT_SECONDS := 120.0
const SAVED_TOKEN_RETRY_DELAYS_SECONDS := [1.0, 2.0, 4.0, 8.0, 15.0, 30.0]
const DISCORD_BEGIN_RETRY_DELAYS_SECONDS := [1.0, 2.0, 4.0, 8.0]
const CREDENTIAL_SAVE_RETRY_DELAYS_SECONDS := [0.25, 0.5, 1.0, 2.0]

var auth_session: AuthSession
var auth_credential_store
var auth_api_client
var _auth_operation_epoch := 0
var _operation_trace_factory: Callable
var _active_auth_trace: ClientOperationTrace
var saved_token_retry_delays_seconds: Array = SAVED_TOKEN_RETRY_DELAYS_SECONDS.duplicate()
var discord_begin_retry_delays_seconds: Array = DISCORD_BEGIN_RETRY_DELAYS_SECONDS.duplicate()
var credential_save_retry_delays_seconds: Array = CREDENTIAL_SAVE_RETRY_DELAYS_SECONDS.duplicate()
var discord_poll_interval_seconds := DISCORD_POLL_INTERVAL_SECONDS
var discord_poll_timeout_seconds := DISCORD_POLL_TIMEOUT_SECONDS


func _ready() -> void:
	if auth_session == null:
		auth_session = AuthSessionScript.new()
	if auth_credential_store == null:
		auth_credential_store = AuthCredentialStoreScript.new()
	if auth_api_client == null:
		auth_api_client = AuthApiClientScript.new()


func configure(auth_api_client_ref = null, operation_trace_factory: Callable = Callable()) -> void:
	if auth_api_client_ref != null:
		auth_api_client = auth_api_client_ref
	_operation_trace_factory = operation_trace_factory


func get_session() -> AuthSession:
	return auth_session


func initialize_from_saved_token() -> void:
	if auth_session == null:
		auth_session = AuthSessionScript.new()
	if auth_credential_store == null:
		auth_credential_store = AuthCredentialStoreScript.new()
	if auth_api_client == null:
		auth_api_client = AuthApiClientScript.new()

	var token: String = str(auth_credential_store.load_token())
	if token.is_empty():
		_cancel_auth_operation()
		auth_session.clear()
		auth_state_changed.emit()
		return

	var operation := _begin_auth_operation("saved_token_validation")
	call_deferred("_validate_saved_token", token, operation["epoch"], operation["trace"])


func request_discord_sign_in() -> void:
	_ensure_auth_objects()
	var operation := _begin_auth_operation("discord_sign_in")
	call_deferred("_run_discord_sign_in", operation["epoch"], operation["trace"])


func logout() -> void:
	_ensure_auth_objects()
	var token: String = str(auth_credential_store.load_token())
	_cancel_auth_operation()

	auth_credential_store.clear_token()
	auth_session.clear()
	auth_state_changed.emit()

	if token.is_empty():
		return

	call_deferred("_logout_remote", token)


func _run_discord_sign_in(operation_epoch: int, operation_trace: ClientOperationTrace) -> void:
	var begin_result = await auth_api_client.begin_discord_login_session(operation_trace.trace_id())
	if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
		return

	var retry_index := 0
	while _is_provider_unavailable(begin_result) && retry_index < discord_begin_retry_delays_seconds.size():
		var scene_tree := get_tree()
		if scene_tree == null:
			break
		var retry_delay := float(discord_begin_retry_delays_seconds[retry_index])
		retry_index += 1
		await scene_tree.create_timer(retry_delay).timeout
		if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
			return
		begin_result = await auth_api_client.begin_discord_login_session(operation_trace.trace_id())
		if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
			return

	if begin_result == null || !begin_result.ok:
		_fail_auth_sign_in(
			"Unable to start Discord sign-in.",
			operation_epoch,
			operation_trace.trace_id(),
			_failure_mode_for_result(begin_result),
			_is_provider_unavailable(begin_result)
		)
		return

	var login_session_id = begin_result.body.get("login_session_id", "")
	var poll_secret = begin_result.body.get("poll_secret", "")
	var login_url = begin_result.body.get("login_url", "")
	if str(login_session_id).is_empty() || str(poll_secret).is_empty() || str(login_url).is_empty():
		_fail_auth_sign_in(
			"Unable to start Discord sign-in.",
			operation_epoch,
			operation_trace.trace_id(),
			"malformed_response"
		)
		return

	var shell_open_error := OS.shell_open(str(login_url))
	if shell_open_error != OK:
		_fail_auth_sign_in(
			"Unable to start Discord sign-in.",
			operation_epoch,
			operation_trace.trace_id(),
			"login_launch_failed",
			false
		)
		return
	await _poll_discord_login_session(str(login_session_id), str(poll_secret), operation_epoch, operation_trace)


func _poll_discord_login_session(
	login_session_id: String,
	poll_secret: String,
	operation_epoch: int,
	operation_trace: ClientOperationTrace
) -> void:
	var deadline := Time.get_unix_time_from_system() + discord_poll_timeout_seconds
	while _is_current_auth_operation(operation_epoch, operation_trace.trace_id()) && Time.get_unix_time_from_system() < deadline:
		var exchange_result = await auth_api_client.exchange_discord_login_session(login_session_id, poll_secret, operation_trace.trace_id())
		if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
			return

		if exchange_result != null && exchange_result.status_code == 202:
			await get_tree().create_timer(discord_poll_interval_seconds).timeout
			if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
				return
			continue

		if _is_provider_unavailable(exchange_result):
			await get_tree().create_timer(discord_poll_interval_seconds).timeout
			if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
				return
			continue

		if exchange_result != null && exchange_result.ok:
			var token := str(exchange_result.body.get("token", ""))
			var user_payload = exchange_result.body.get("user", {})
			if token.is_empty() || typeof(user_payload) != TYPE_DICTIONARY || user_payload.is_empty():
				_fail_auth_sign_in(
					"Discord sign-in failed.",
					operation_epoch,
					operation_trace.trace_id(),
					"malformed_response"
				)
				return

			var token_saved := await _save_token_with_retries(token, operation_epoch, operation_trace.trace_id())
			if !token_saved:
				if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
					return
				call_deferred("_logout_remote", token)
				_fail_auth_sign_in(
					"Secure credential storage is unavailable.",
					operation_epoch,
					operation_trace.trace_id(),
					"credential_storage_unavailable"
				)
				return

			auth_session.set_signed_in(token, user_payload)
			_emit_auth_terminal(
				ObservabilityContract.EVENT_AUTH_SUCCEEDED,
				operation_trace.trace_id(),
				"discord"
			)
			auth_state_changed.emit()
			_clear_auth_trace(operation_epoch, operation_trace.trace_id())
			return

		_fail_auth_sign_in(
			"Discord sign-in failed.",
			operation_epoch,
			operation_trace.trace_id(),
			_failure_mode_for_result(exchange_result),
			_is_provider_unavailable(exchange_result)
		)
		return

	if _is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
		_fail_auth_sign_in(
			"Discord sign-in timed out.",
			operation_epoch,
			operation_trace.trace_id(),
			"timeout",
		)


func _fail_auth_sign_in(
	message: String,
	operation_epoch: int,
	trace_id: String = "",
	failure_mode: String = "authentication_failed",
	provider_unavailable: bool = false
) -> void:
	if !_is_current_auth_operation(operation_epoch, trace_id):
		return
	auth_credential_store.clear_token()
	auth_session.clear()
	_emit_auth_terminal(
		ObservabilityContract.EVENT_AUTH_PROVIDER_UNAVAILABLE if provider_unavailable else ObservabilityContract.EVENT_AUTH_FAILED,
		trace_id,
		"discord",
		failure_mode
	)
	auth_error.emit(message)
	auth_state_changed.emit()
	_clear_auth_trace(operation_epoch, trace_id)


func _logout_remote(token: String) -> void:
	await auth_api_client.logout(token)


func _ensure_auth_objects() -> void:
	if auth_session == null:
		auth_session = AuthSessionScript.new()
	if auth_credential_store == null:
		auth_credential_store = AuthCredentialStoreScript.new()
	if auth_api_client == null:
		auth_api_client = AuthApiClientScript.new()


func _validate_saved_token(token: String, operation_epoch: int, operation_trace: ClientOperationTrace) -> void:
	var result = await auth_api_client.get_current_user(token, operation_trace.trace_id())
	if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
		return

	var retry_index := 0
	while _is_provider_unavailable(result) && retry_index < saved_token_retry_delays_seconds.size():
		var scene_tree := get_tree()
		if scene_tree == null:
			break
		var retry_delay := float(saved_token_retry_delays_seconds[retry_index])
		retry_index += 1
		await scene_tree.create_timer(retry_delay).timeout
		if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
			return
		result = await auth_api_client.get_current_user(token, operation_trace.trace_id())
		if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
			return

	if result != null && result.ok:
		var user_payload: Dictionary = result.body.get("user", {})
		if !user_payload.is_empty():
			var migration_failed := false
			if auth_credential_store.requires_legacy_migration():
				migration_failed = !await _save_token_with_retries(token, operation_epoch, operation_trace.trace_id())
			if migration_failed:
				if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
					return
				auth_credential_store.clear_token()
				auth_session.clear()
				call_deferred("_logout_remote", token)
				_emit_auth_terminal(
					ObservabilityContract.EVENT_AUTH_FAILED,
					operation_trace.trace_id(),
					"saved_token",
					"credential_storage_unavailable"
				)
				auth_error.emit("Secure credential migration failed. Sign in again.")
			else:
				auth_session.set_signed_in(token, user_payload)
				_emit_auth_terminal(
					ObservabilityContract.EVENT_AUTH_SUCCEEDED,
					operation_trace.trace_id(),
					"saved_token"
				)
		else:
			auth_session.clear()
			_emit_auth_terminal(
				ObservabilityContract.EVENT_AUTH_FAILED,
				operation_trace.trace_id(),
				"saved_token",
				"malformed_response"
			)
	else:
		var provider_unavailable := _is_provider_unavailable(result)
		if _should_clear_saved_credential(result):
			auth_credential_store.clear_token()
		auth_session.clear()
		_emit_auth_terminal(
			ObservabilityContract.EVENT_AUTH_PROVIDER_UNAVAILABLE if provider_unavailable else ObservabilityContract.EVENT_AUTH_FAILED,
			operation_trace.trace_id(),
			"saved_token",
			_failure_mode_for_result(result)
		)
	auth_state_changed.emit()
	_clear_auth_trace(operation_epoch, operation_trace.trace_id())


func _save_token_with_retries(token: String, operation_epoch: int, trace_id: String) -> bool:
	if auth_credential_store.save_token(token):
		return true

	for retry_delay_value in credential_save_retry_delays_seconds:
		var scene_tree := get_tree()
		if scene_tree == null:
			return false
		await scene_tree.create_timer(float(retry_delay_value)).timeout
		if !_is_current_auth_operation(operation_epoch, trace_id):
			return false
		if auth_credential_store.save_token(token):
			return true
	return false


func active_auth_trace_id() -> String:
	if _active_auth_trace == null:
		return ""
	return _active_auth_trace.trace_id()


func _begin_auth_operation(operation_name: String) -> Dictionary:
	_auth_operation_epoch += 1
	_active_auth_trace = ClientOperationTrace.create(operation_name, _operation_trace_factory)
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_AUTH_FLOW_STARTED,
		"",
		{"trace_id": _active_auth_trace.trace_id()},
		{"auth_operation": operation_name}
	)
	return {"epoch": _auth_operation_epoch, "trace": _active_auth_trace}


func _cancel_auth_operation() -> void:
	_auth_operation_epoch += 1
	_active_auth_trace = null


func _clear_auth_trace(operation_epoch: int, trace_id: String = "") -> void:
	if !_is_current_auth_operation(operation_epoch, trace_id):
		return
	_active_auth_trace = null


func _is_current_auth_operation(operation_epoch: int, trace_id: String = "") -> bool:
	if operation_epoch != _auth_operation_epoch:
		return false
	if trace_id.is_empty():
		return true
	return _active_auth_trace != null && _active_auth_trace.trace_id() == trace_id


func _emit_auth_terminal(
	event_name: String,
	trace_id: String,
	provider: String,
	failure_mode: String = ""
) -> void:
	if trace_id.is_empty():
		return
	var fields := {"provider": provider}
	if !failure_mode.is_empty():
		fields["failure_mode"] = failure_mode
	ClientLogger.emit_canonical(event_name, "", {"trace_id": trace_id}, fields)


func _is_provider_unavailable(result) -> bool:
	if result == null:
		return true
	var error_message := str(result.error_message)
	if error_message.begins_with("network_failure_") || error_message == "request_failed":
		return true
	var status_code := int(result.status_code)
	return status_code == 408 || status_code == 429 || status_code >= 500


func _should_clear_saved_credential(result) -> bool:
	if result == null:
		return false
	var status_code := int(result.status_code)
	return status_code == 401 || status_code == 403


func _failure_mode_for_result(result) -> String:
	if result == null:
		return "request_failed"
	var error_message := str(result.error_message)
	if error_message.begins_with("network_failure_"):
		return "network_failure"
	if error_message == "scene_tree_unavailable":
		return "scene_tree_unavailable"
	if error_message == "request_failed":
		return "request_setup_failed"
	var status_code := int(result.status_code)
	if status_code >= 500:
		return "http_5xx"
	if status_code == 401:
		return "invalid_or_expired_token"
	if status_code == 403:
		return "authentication_denied"
	return "authentication_failed"
