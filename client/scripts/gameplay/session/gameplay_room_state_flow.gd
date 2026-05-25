extends RefCounted

const Packets = preload("res://scripts/networking/packets/packets.gd")
const RoomState = preload("res://scripts/session/room_state.gd")


static func room_state_from_packet(data: Dictionary, fallback_room_state: String) -> String:
	return str(data.get(Packets.FIELD_ROOM_STATE, fallback_room_state)).strip_edges()


static func should_stop_spectating_for_room_state(room_state: String) -> bool:
	return RoomState.is_game_over(room_state)
