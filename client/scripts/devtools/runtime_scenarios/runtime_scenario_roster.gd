extends RefCounted
class_name RuntimeScenarioRoster

var room_session_controller
var expected_human_count := 1


func configure(room_session_controller_ref, expected_humans: int) -> void:
	room_session_controller = room_session_controller_ref
	expected_human_count = maxi(expected_humans, 1)


func humans_joined() -> bool:
	return human_count() >= expected_human_count


func lobby_can_start() -> bool:
	return bool(_snapshot().get("can_start_game", false))


func has_bot_count(required_count: int) -> bool:
	return bot_count() >= required_count


func human_count() -> int:
	var count := 0
	for member in _members():
		if member is Dictionary and !bool(member.get("is_bot", false)):
			count += 1
	return count


func bot_count() -> int:
	var count := 0
	for member in _members():
		if member is Dictionary and bool(member.get("is_bot", false)):
			count += 1
	return count


func other_human_player_id() -> String:
	var snapshot := _snapshot()
	var local_player_id := str(snapshot.get("local_player_id", ""))
	for member in snapshot.get("members", []):
		if !(member is Dictionary) or bool(member.get("is_bot", false)):
			continue
		var player_id := str(member.get("player_id", ""))
		if !player_id.is_empty() and player_id != local_player_id:
			return player_id
	return ""


func _members() -> Array:
	var members = _snapshot().get("members", [])
	return members if members is Array else []


func _snapshot() -> Dictionary:
	if room_session_controller == null:
		return {}
	return room_session_controller.lobby_state_snapshot()
