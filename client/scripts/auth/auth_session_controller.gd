extends Node
class_name AuthSessionController

signal auth_state_changed
signal auth_error(message: String)

const AuthSessionScript := preload("res://scripts/auth/auth_session.gd")
const AuthTokenStoreScript := preload("res://scripts/auth/auth_token_store.gd")
const AuthApiClientScript := preload("res://scripts/auth/auth_api_client.gd")

const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const DISCORD_POLL_INTERVAL_SECONDS := 1.0
const DISCORD_POLL_TIMEOUT_SECONDS := 120.0

var auth_session: AuthSession
var auth_token_store: AuthTokenStore
var auth_api_client
var _auth_operation_epoch := 0
var _operation_trace_factory: Callable
var _active_auth_trace: ClientOperationTrace


func _ready() -> void:
	if auth_session == null:
		auth_session = AuthSessionScript.new()
	if auth_token_store == null:
		auth_token_store = AuthTokenStoreScript.new()
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
	if auth_token_store == null:
		auth_token_store = AuthTokenStoreScript.new()
	if auth_api_client == null:
		auth_api_client = AuthApiClientScript.new()

	var token := auth_token_store.load_token()
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
	var token := auth_token_store.load_token()
	_cancel_auth_operation()

	auth_token_store.clear_token()
	auth_session.clear()
	auth_state_changed.emit()

	if token.is_empty():
		return

	call_deferred("_logout_remote", token)


func _run_discord_sign_in(operation_epoch: int, operation_trace: ClientOperationTrace) -> void:
	var begin_result = await auth_api_client.begin_discord_login_session(operation_trace.trace_id())
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
	var deadline := Time.get_unix_time_from_system() + DISCORD_POLL_TIMEOUT_SECONDS
	while _is_current_auth_operation(operation_epoch, operation_trace.trace_id()) && Time.get_unix_time_from_system() < deadline:
		var exchange_result = await auth_api_client.exchange_discord_login_session(login_session_id, poll_secret, operation_trace.trace_id())
		if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
			return

		if exchange_result != null && exchange_result.status_code == 202:
			await get_tree().create_timer(DISCORD_POLL_INTERVAL_SECONDS).timeout
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

			auth_token_store.save_token(token)
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
	auth_token_store.clear_token()
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
	if auth_token_store == null:
		auth_token_store = AuthTokenStoreScript.new()
	if auth_api_client == null:
		auth_api_client = AuthApiClientScript.new()


func _validate_saved_token(token: String, operation_epoch: int, operation_trace: ClientOperationTrace) -> void:
	var result = await auth_api_client.get_current_user(token, operation_trace.trace_id())
	if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
		return
	if result != null && result.ok:
		var user_payload: Dictionary = result.body.get("user", {})
		if !user_payload.is_empty():
			auth_session.set_signed_in(token, user_payload)
			_emit_auth_terminal(
				ObservabilityContract.EVENT_AUTH_SUCCEEDED,
				operation_trace.trace_id(),
				"saved_token"
			)
		else:
			auth_token_store.clear_token()
			auth_session.clear()
			_emit_auth_terminal(
				ObservabilityContract.EVENT_AUTH_FAILED,
				operation_trace.trace_id(),
				"saved_token",
				"malformed_response"
			)
	else:
		auth_token_store.clear_token()
		auth_session.clear()
		_emit_auth_terminal(
			ObservabilityContract.EVENT_AUTH_PROVIDER_UNAVAILABLE if _is_provider_unavailable(result) else ObservabilityContract.EVENT_AUTH_FAILED,
			operation_trace.trace_id(),
			"saved_token",
			_failure_mode_for_result(result)
		)
	auth_state_changed.emit()
	_clear_auth_trace(operation_epoch, operation_trace.trace_id())


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
		return false
	var error_message := str(result.error_message)
	if error_message.begins_with("network_failure_"):
		return true
	return int(result.status_code) >= 500


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
