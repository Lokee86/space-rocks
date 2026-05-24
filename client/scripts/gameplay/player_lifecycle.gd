extends RefCounted

const Packets = preload("res://scripts/networking/packets.gd")


static func from_state(data: Dictionary) -> Dictionary:
	var lifecycle_data = data.get(Packets.FIELD_PLAYER_LIFECYCLE, {})
	if !(lifecycle_data is Dictionary):
		return {}

	var lifecycle := {}
	for player_id in lifecycle_data.keys():
		lifecycle[str(player_id)] = str(lifecycle_data[player_id])
	return lifecycle
