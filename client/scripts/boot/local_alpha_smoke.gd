extends Node
class_name LocalAlphaSmoke

const CredentialSmokeScript := preload("res://scripts/boot/local_alpha_credential_smoke.gd")
const ProfileSmokeScript := preload("res://scripts/boot/local_alpha_profile_smoke.gd")
const MatchSmokeScript := preload("res://scripts/boot/local_alpha_match_smoke.gd")

const SMOKE_DISPLAY_NAME := "ReleaseGateSmoke"
const SMOKE_SCORE := 321


func run(phase: String) -> int:
	_clear_smoke_result()
	if !CredentialSmokeScript.new().run():
		return _fail(10, "packaged credential round trip failed")

	var profile_smoke = ProfileSmokeScript.new()
	add_child(profile_smoke)
	if !await profile_smoke.wait_for_server():
		return _fail(11, "bundled server did not become healthy")
	if phase == "seed":
		return await _run_seed(profile_smoke)
	if phase == "verify":
		return await _run_verify(profile_smoke)
	return _fail(12, "unsupported phase %s" % phase)


func _run_seed(profile_smoke) -> int:
	var profile_result: Dictionary = await profile_smoke.create_selected_profile(SMOKE_DISPLAY_NAME)
	if !bool(profile_result.get("ok", false)):
		return _fail(20, str(profile_result.get("error", "could not create local profile")))
	var profile_id := str(profile_result.get("profile_id", ""))

	var match_smoke = MatchSmokeScript.new()
	add_child(match_smoke)
	var match_error: String = await match_smoke.run(profile_id, SMOKE_SCORE)
	if !match_error.is_empty():
		return _fail(21, match_error)
	if !await profile_smoke.wait_for_persisted_stats(profile_id, SMOKE_SCORE):
		return _fail(22, "completed match did not persist local profile stats")

	print("local alpha smoke seed passed with completed single-player match")
	return 0


func _run_verify(profile_smoke) -> int:
	var verify_error: String = await profile_smoke.verify_persistent_profile_and_stats(
		SMOKE_DISPLAY_NAME,
		SMOKE_SCORE
	)
	if !verify_error.is_empty():
		return _fail(30, verify_error)
	print("local alpha smoke verify passed with persisted match stats")
	return 0


func _clear_smoke_result() -> void:
	var result_path := "user://release_gate_smoke_result.json"
	if FileAccess.file_exists(result_path):
		DirAccess.remove_absolute(ProjectSettings.globalize_path(result_path))


func _fail(code: int, message: String) -> int:
	var file := FileAccess.open("user://release_gate_smoke_result.json", FileAccess.WRITE)
	if file != null:
		file.store_string(JSON.stringify({"code": code, "message": message}) + "\n")
		file.close()
	push_error("local alpha smoke: %s" % message)
	return code
