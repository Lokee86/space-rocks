extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const STATUS_ACTIVE := "active"


static func from_state(data: Dictionary) -> Dictionary:
	var lifecycle_data = data.get(Packets.FIELD_PLAYER_LIFECYCLE, {})
	if !(lifecycle_data is Dictionary):
		return {}

	var lifecycle := {}
	for player_id in lifecycle_data.keys():
		lifecycle[str(player_id)] = str(lifecycle_data[player_id])
	return lifecycle


static func is_player_active(lifecycle: Dictionary, player_id: String) -> bool:
	if player_id.is_empty():
		return false

	var lifecycle_value = lifecycle.get(str(player_id), "")
	if lifecycle_value is Dictionary:
		var lifecycle_record: Dictionary = lifecycle_value
		if lifecycle_record.has("state"):
			return str(lifecycle_record.get("state", "")) == STATUS_ACTIVE
		if lifecycle_record.has("status"):
			return str(lifecycle_record.get("status", "")) == STATUS_ACTIVE
		return false

	return str(lifecycle_value) == STATUS_ACTIVE


