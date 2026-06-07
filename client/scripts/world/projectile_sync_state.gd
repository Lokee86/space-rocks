extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")


static func server_position(state: Dictionary) -> Vector2:
	return Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])


static func projectile_type(state: Dictionary) -> String:
	if not state.has(Packets.FIELD_PROJECTILE_TYPE):
		return "bullet"

	var value := str(state[Packets.FIELD_PROJECTILE_TYPE])
	if value == "":
		return "bullet"

	return value

