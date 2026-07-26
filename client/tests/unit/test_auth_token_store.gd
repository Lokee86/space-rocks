extends GutTest

const AuthCredentialStore := preload("res://scripts/auth/auth_credential_store.gd")
const LegacyAuthTokenReader := preload("res://scripts/auth/auth_token_store.gd")

const TEST_LEGACY_PATH := "user://test_auth_token.json"
const TEST_BLOB_PATH := "user://test_auth_credential.bin"
const TEST_MARKER_PATH := "user://test_auth_credential_revoked"

var _secure_token := ""
var _request_failure := false


func before_each() -> void:
	_secure_token = ""
	_request_failure = false
	_cleanup_files()


func after_each() -> void:
	_cleanup_files()


func test_test_process_uses_isolated_credential_identity_and_paths() -> void:
	var store := AuthCredentialStore.new()

	var test_scope := str(OS.get_process_id())
	assert_eq(store.service_name, "ca.laughingskull.space-rocks.test.%s" % test_scope)
	assert_eq(store.encrypted_blob_path, "user://test_auth_credential_%s.bin" % test_scope)
	assert_eq(store.revocation_marker_path, "user://test_auth_credential_revoked_%s" % test_scope)
	assert_eq(store.legacy_token_path, "user://test_auth_token_%s.json" % test_scope)


func test_secure_store_round_trip_uses_helper_protocol() -> void:
	var store = _create_store()

	assert_true(store.save_token("bearer-token"))
	assert_eq(store.load_token(), "bearer-token")
	assert_true(store.clear_token())
	assert_eq(store.load_token(), "")


func test_legacy_token_is_loaded_only_for_migration() -> void:
	_write_legacy_token("legacy-token")
	var store = _create_store()

	assert_eq(store.load_token(), "legacy-token")
	assert_true(store.requires_legacy_migration())
	assert_false(FileAccess.file_exists(TEST_LEGACY_PATH))

	assert_true(store.save_token("legacy-token"))
	assert_false(store.requires_legacy_migration())
	assert_false(FileAccess.file_exists(TEST_LEGACY_PATH))
	assert_eq(store.load_token(), "legacy-token")


func test_helper_failure_deletes_plaintext_legacy_token_and_signs_out() -> void:
	_write_legacy_token("legacy-token")
	_request_failure = true
	var store = _create_store()

	assert_eq(store.load_token(), "")
	assert_false(FileAccess.file_exists(TEST_LEGACY_PATH))


func test_failed_clear_blocks_credential_from_reappearing() -> void:
	var store = _create_store()
	_secure_token = "secure-token"
	_request_failure = true

	assert_false(store.clear_token())
	assert_true(FileAccess.file_exists(TEST_MARKER_PATH))
	assert_eq(_secure_token, "secure-token")

	_request_failure = false
	assert_eq(store.load_token(), "")
	assert_eq(_secure_token, "")
	assert_false(FileAccess.file_exists(TEST_MARKER_PATH))


func test_clear_removes_secure_and_legacy_credentials() -> void:
	_write_legacy_token("legacy-token")
	var store = _create_store()
	assert_true(store.save_token("secure-token"))
	_write_legacy_token("stale-token")

	assert_true(store.clear_token())
	assert_eq(_secure_token, "")
	assert_false(FileAccess.file_exists(TEST_LEGACY_PATH))


func _create_store():
	var store := AuthCredentialStore.new()
	store.encrypted_blob_path = TEST_BLOB_PATH
	store.revocation_marker_path = TEST_MARKER_PATH
	var legacy_reader := LegacyAuthTokenReader.new()
	legacy_reader.token_path = TEST_LEGACY_PATH
	store.legacy_token_reader = legacy_reader
	store.request_handler = _handle_request
	return store


func _handle_request(request: Dictionary) -> Dictionary:
	if _request_failure:
		return {"ok": false, "error": "test_failure"}
	match str(request.get("action", "")):
		"load":
			return {"ok": true, "secret": _secure_token}
		"save":
			_secure_token = str(request.get("secret", ""))
			return {"ok": true}
		"clear":
			_secure_token = ""
			return {"ok": true}
	return {"ok": false}


func _write_legacy_token(token: String) -> void:
	var file := FileAccess.open(TEST_LEGACY_PATH, FileAccess.WRITE)
	assert_not_null(file)
	file.store_string(JSON.stringify({"token": token}))
	file.close()


func _cleanup_files() -> void:
	for path in [TEST_LEGACY_PATH, TEST_BLOB_PATH, TEST_MARKER_PATH]:
		if FileAccess.file_exists(path):
			DirAccess.remove_absolute(ProjectSettings.globalize_path(path))
