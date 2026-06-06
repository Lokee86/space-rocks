extends RefCounted
class_name PickupSyncState

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")


static func pickup_id(state: Dictionary) -> String:
	return str(state.get(Packets.FIELD_ID, ""))


static func pickup_type(state: Dictionary) -> String:
	return str(state.get(Packets.FIELD_TYPE, ""))


static func server_position(state: Dictionary) -> Vector2:
	return Vector2(
		float(state.get(Packets.FIELD_X, 0.0)),
		float(state.get(Packets.FIELD_Y, 0.0))
	)


static func age_seconds(state: Dictionary) -> float:
	return float(state.get(Packets.FIELD_AGE_SECONDS, 0.0))


static func lifespan_seconds(state: Dictionary) -> float:
	return float(state.get(Packets.FIELD_LIFESPAN_SECONDS, 0.0))
