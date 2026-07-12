extends Node
class_name AuthSessionController

signal auth_state_changed
signal auth_error(message: String)

const AuthSessionScript := preload("res://scripts/auth/auth_session.gd")
const AuthTokenStoreScript := preload("res://scripts/auth/auth_token_store.gd")
const AuthApiClientScript := preload("res://scripts/auth/auth_api_client.gd")
const DISCORD_POLL_INTERVAL_SECONDS := 1.0
const DISCORD_POLL_TIMEOUT_SECONDS := 120.0

var auth_session: AuthSession
var auth_token_store: AuthTokenStore
var auth_api_client
var _auth_operation_epoch := 0


func _ready() -> void:
	if auth_session == null:
		auth_session = AuthSessionScript.new()
	if auth_token_store == null:
		auth_token_store = AuthTokenStoreScript.new()
	if auth_api_client == null:
		auth_api_client = AuthApiClientScript.new()


func configure(auth_api_client_ref = null) -> void:
	if auth_api_client_ref != null:
		auth_api_client = auth_api_client_ref


func get_session() -> AuthSession:
	return auth_session


func initialize_from_saved_token() -> void:
	if auth_session == null:
		auth_session = AuthSessionScript.new()
	if auth_token_store == null:
		auth_token_store = AuthTokenStoreScript.new()
	if auth_api_client == null:
		auth_api_client = AuthApiClientScript.new()
	var operation_epoch := _advance_auth_operation_epoch()

	var token := auth_token_store.load_token()
	if token.is_empty():
		auth_session.clear()
		auth_state_changed.emit()
		return

	call_deferred("_validate_saved_token", token, operation_epoch)


func request_discord_sign_in() -> void:
	_ensure_auth_objects()
	var operation_epoch := _advance_auth_operation_epoch()
	call_deferred("_run_discord_sign_in", operation_epoch)


func logout() -> void:
	_ensure_auth_objects()
	_advance_auth_operation_epoch()
	var token := auth_token_store.load_token()
	auth_token_store.clear_token()
	auth_session.clear()
	auth_state_changed.emit()

	if token.is_empty():
		return

	call_deferred("_logout_remote", token)


func _run_discord_sign_in(operation_epoch: int) -> void:
	var begin_result = await auth_api_client.begin_discord_login_session()
	if !_is_current_auth_operation(operation_epoch):
		return

	if !begin_result.ok:
		_fail_auth_sign_in("Unable to start Discord sign-in.", operation_epoch)
		return

	var login_session_id = begin_result.body.get("login_session_id", "")
	var poll_secret = begin_result.body.get("poll_secret", "")
	var login_url = begin_result.body.get("login_url", "")
	if str(login_session_id).is_empty() || str(poll_secret).is_empty() || str(login_url).is_empty():
		_fail_auth_sign_in("Unable to start Discord sign-in.", operation_epoch)
		return

	OS.shell_open(str(login_url))
	await _poll_discord_login_session(str(login_session_id), str(poll_secret), operation_epoch)


func _poll_discord_login_session(login_session_id: String, poll_secret: String, operation_epoch: int) -> void:
	var deadline := Time.get_unix_time_from_system() + DISCORD_POLL_TIMEOUT_SECONDS
	while _is_current_auth_operation(operation_epoch) && Time.get_unix_time_from_system() < deadline:
		var exchange_result = await auth_api_client.exchange_discord_login_session(login_session_id, poll_secret)
		if !_is_current_auth_operation(operation_epoch):
			return

		if exchange_result.status_code == 202:
			await get_tree().create_timer(DISCORD_POLL_INTERVAL_SECONDS).timeout
			if !_is_current_auth_operation(operation_epoch):
				return
			continue

		if exchange_result.ok:
			var token := str(exchange_result.body.get("token", ""))
			var user_payload = exchange_result.body.get("user", {})
			if token.is_empty() || typeof(user_payload) != TYPE_DICTIONARY:
				_fail_auth_sign_in("Discord sign-in failed.", operation_epoch)
				return

			auth_token_store.save_token(token)
			auth_session.set_signed_in(token, user_payload)
			auth_state_changed.emit()
			return

		_fail_auth_sign_in("Discord sign-in failed.", operation_epoch)
		return

	if _is_current_auth_operation(operation_epoch):
		_fail_auth_sign_in("Discord sign-in timed out.", operation_epoch)


func _fail_auth_sign_in(message: String, operation_epoch: int) -> void:
	if !_is_current_auth_operation(operation_epoch):
		return
	auth_token_store.clear_token()
	auth_session.clear()
	auth_error.emit(message)
	auth_state_changed.emit()


func _logout_remote(token: String) -> void:
	await auth_api_client.logout(token)


func _ensure_auth_objects() -> void:
	if auth_session == null:
		auth_session = AuthSessionScript.new()
	if auth_token_store == null:
		auth_token_store = AuthTokenStoreScript.new()
	if auth_api_client == null:
		auth_api_client = AuthApiClientScript.new()


func _validate_saved_token(token: String, operation_epoch: int) -> void:
	var result = await auth_api_client.get_current_user(token)
	if !_is_current_auth_operation(operation_epoch):
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


func _advance_auth_operation_epoch() -> int:
	_auth_operation_epoch += 1
	return _auth_operation_epoch


func _is_current_auth_operation(operation_epoch: int) -> bool:
	return operation_epoch == _auth_operation_epoch
