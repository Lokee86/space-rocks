extends Node
class_name AuthSessionController

signal auth_state_changed
signal auth_error(message: String)

const AuthSessionScript := preload("res://scripts/auth/auth_session.gd")
const AuthTokenStoreScript := preload("res://scripts/auth/auth_token_store.gd")
const AuthApiClientScript := preload("res://scripts/auth/auth_api_client.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")
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
	var operation: Dictionary = {}
	if token.is_empty():
		_cancel_auth_operation()
	else:
		operation = _begin_auth_operation("logout")

	auth_token_store.clear_token()
	auth_session.clear()
	auth_state_changed.emit()

	if token.is_empty():
		return

	call_deferred("_logout_remote", token, operation["epoch"], operation["trace"])


func _run_discord_sign_in(operation_epoch: int, operation_trace: ClientOperationTrace) -> void:
	var begin_result = await auth_api_client.begin_discord_login_session()
	if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
		return

	if !begin_result.ok:
		_fail_auth_sign_in("Unable to start Discord sign-in.", operation_epoch, operation_trace.trace_id())
		return

	var login_session_id = begin_result.body.get("login_session_id", "")
	var poll_secret = begin_result.body.get("poll_secret", "")
	var login_url = begin_result.body.get("login_url", "")
	if str(login_session_id).is_empty() || str(poll_secret).is_empty() || str(login_url).is_empty():
		_fail_auth_sign_in("Unable to start Discord sign-in.", operation_epoch, operation_trace.trace_id())
		return

	OS.shell_open(str(login_url))
	await _poll_discord_login_session(str(login_session_id), str(poll_secret), operation_epoch, operation_trace)


func _poll_discord_login_session(
	login_session_id: String,
	poll_secret: String,
	operation_epoch: int,
	operation_trace: ClientOperationTrace
) -> void:
	var deadline := Time.get_unix_time_from_system() + DISCORD_POLL_TIMEOUT_SECONDS
	while _is_current_auth_operation(operation_epoch, operation_trace.trace_id()) && Time.get_unix_time_from_system() < deadline:
		var exchange_result = await auth_api_client.exchange_discord_login_session(login_session_id, poll_secret)
		if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
			return

		if exchange_result.status_code == 202:
			await get_tree().create_timer(DISCORD_POLL_INTERVAL_SECONDS).timeout
			if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
				return
			continue

		if exchange_result.ok:
			var token := str(exchange_result.body.get("token", ""))
			var user_payload = exchange_result.body.get("user", {})
			if token.is_empty() || typeof(user_payload) != TYPE_DICTIONARY:
				_fail_auth_sign_in("Discord sign-in failed.", operation_epoch, operation_trace.trace_id())
				return

			auth_token_store.save_token(token)
			auth_session.set_signed_in(token, user_payload)
			auth_state_changed.emit()
			_clear_auth_trace(operation_epoch, operation_trace.trace_id())
			return

		_fail_auth_sign_in("Discord sign-in failed.", operation_epoch, operation_trace.trace_id())
		return

	if _is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
		_fail_auth_sign_in("Discord sign-in timed out.", operation_epoch, operation_trace.trace_id())


func _fail_auth_sign_in(message: String, operation_epoch: int, trace_id: String = "") -> void:
	if !_is_current_auth_operation(operation_epoch, trace_id):
		return
	auth_token_store.clear_token()
	auth_session.clear()
	auth_error.emit(message)
	auth_state_changed.emit()
	_clear_auth_trace(operation_epoch, trace_id)


func _logout_remote(token: String, operation_epoch: int, operation_trace: ClientOperationTrace) -> void:
	await auth_api_client.logout(token)
	_clear_auth_trace(operation_epoch, operation_trace.trace_id())


func _ensure_auth_objects() -> void:
	if auth_session == null:
		auth_session = AuthSessionScript.new()
	if auth_token_store == null:
		auth_token_store = AuthTokenStoreScript.new()
	if auth_api_client == null:
		auth_api_client = AuthApiClientScript.new()


func _validate_saved_token(token: String, operation_epoch: int, operation_trace: ClientOperationTrace) -> void:
	var result = await auth_api_client.get_current_user(token)
	if !_is_current_auth_operation(operation_epoch, operation_trace.trace_id()):
		return
	if result.ok:
		var user_payload: Dictionary = result.body.get("user", {})
		if !user_payload.is_empty():
			auth_session.set_signed_in(token, user_payload)
		else:
			auth_token_store.clear_token()
			auth_session.clear()
	else:
		auth_token_store.clear_token()
		auth_session.clear()
	auth_state_changed.emit()
	_clear_auth_trace(operation_epoch, operation_trace.trace_id())


func active_auth_trace_id() -> String:
	if _active_auth_trace == null:
		return ""
	return _active_auth_trace.trace_id()


func _begin_auth_operation(operation_name: String) -> Dictionary:
	_auth_operation_epoch += 1
	_active_auth_trace = ClientOperationTrace.create(operation_name, _operation_trace_factory)
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