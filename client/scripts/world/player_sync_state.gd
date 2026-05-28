extends RefCounted

const Packets = preload("res://scripts/networking/packets/packets.gd")


static func server_position(state: Dictionary) -> Vector2:
	return Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])


static func server_rotation(state: Dictionary) -> float:
	return state[Packets.FIELD_ROTATION]
