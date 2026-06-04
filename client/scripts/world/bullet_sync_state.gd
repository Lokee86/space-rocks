extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")


static func server_position(state: Dictionary) -> Vector2:
	return Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])

