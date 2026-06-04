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

	return str(lifecycle.get(str(player_id), "")) == STATUS_ACTIVE

