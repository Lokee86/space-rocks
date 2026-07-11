extends RefCounted

const FIELD_PLAYER_LIFECYCLE := "player_lifecycle"
const STATUS_ACTIVE := "active"
const STATUS_PENDING_RESPAWN := "pending_respawn"
const STATUS_ELIMINATED := "eliminated"


static func from_state(data: Dictionary) -> Dictionary:
	var lifecycle_data = data.get(FIELD_PLAYER_LIFECYCLE, {})
	if !(lifecycle_data is Dictionary):
		return {}

	var lifecycle := {}
	for player_id in lifecycle_data.keys():
		lifecycle[str(player_id)] = str(lifecycle_data[player_id])
	return lifecycle


static func is_player_active(lifecycle: Dictionary, player_id: String) -> bool:
	return status_for(lifecycle, player_id) == STATUS_ACTIVE


static func status_for(lifecycle: Dictionary, player_id: String) -> String:
	if player_id.is_empty():
		return ""

	var lifecycle_value = lifecycle.get(str(player_id), "")
	if lifecycle_value is Dictionary:
		var lifecycle_record: Dictionary = lifecycle_value
		if lifecycle_record.has("state"):
			return str(lifecycle_record.get("state", ""))
		if lifecycle_record.has("status"):
			return str(lifecycle_record.get("status", ""))
		return ""

	if lifecycle_value == null:
		return ""
	return str(lifecycle_value)
