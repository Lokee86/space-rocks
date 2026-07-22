extends SceneTree

const AuthCredentialStoreScript := preload("res://scripts/auth/auth_credential_store.gd")
const LegacyAuthTokenReaderScript := preload("res://scripts/auth/auth_token_store.gd")
const SMOKE_TOKEN := "space-rocks-credential-smoke"


func _init() -> void:
	var store = AuthCredentialStoreScript.new()
	store.service_name = "ca.laughingskull.space-rocks.smoke"
	store.account_name = "credential-helper-smoke"
	store.encrypted_blob_path = "user://credential_helper_smoke.bin"
	store.revocation_marker_path = "user://credential_helper_smoke_revoked"
	var legacy_reader = LegacyAuthTokenReaderScript.new()
	legacy_reader.token_path = "user://credential_helper_smoke_legacy.json"
	store.legacy_token_reader = legacy_reader

	store.clear_token()
	if !store.save_token(SMOKE_TOKEN):
		_fail("secure save failed")
		return
	if store.load_token() != SMOKE_TOKEN:
		_fail("secure load did not return the saved credential")
		return
	if !store.clear_token():
		_fail("secure clear failed")
		return
	if !store.load_token().is_empty():
		_fail("credential remained available after clear")
		return

	print("credential helper smoke passed on %s" % OS.get_name())
	quit(0)


func _fail(message: String) -> void:
	push_error("credential helper smoke failed: %s" % message)
	quit(1)
