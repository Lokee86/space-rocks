extends RefCounted
class_name LocalAlphaCredentialSmoke

const AuthCredentialStoreScript := preload("res://scripts/auth/auth_credential_store.gd")
const LegacyAuthTokenReaderScript := preload("res://scripts/auth/auth_token_store.gd")

const SMOKE_TOKEN := "space-rocks-release-gate-token"


func run() -> bool:
	var store = AuthCredentialStoreScript.new()
	store.service_name = "ca.laughingskull.space-rocks.release-gate-smoke"
	store.account_name = "session"
	store.encrypted_blob_path = "user://release_gate_credential.bin"
	store.revocation_marker_path = "user://release_gate_credential_revoked"
	var legacy_reader = LegacyAuthTokenReaderScript.new()
	legacy_reader.token_path = "user://release_gate_credential_legacy.json"
	store.legacy_token_reader = legacy_reader

	store.clear_token()
	if !store.save_token(SMOKE_TOKEN):
		return false
	if store.load_token() != SMOKE_TOKEN:
		return false
	if !store.clear_token():
		return false
	return store.load_token().is_empty()
