extends RefCounted

const Packets = preload("res://scripts/networking/packets.gd")

const ROOM_ERROR_MESSAGES := {
	"room_not_found": "Room was not found.",
	"room_closed": "Room is closed.",
	"room_in_game": "Room is already in game.",
	"room_full": "Room is full.",
	"already_in_room": "Already in a room.",
	"not_in_room": "Not in a room.",
	"invalid_room_code": "Room code is invalid.",
	"not_ready": "Not all players are ready.",
	"invalid_room_state": "Room is not available.",
}
const FALLBACK_MESSAGE := "Room request failed."


static func message_for_packet(data: Dictionary) -> String:
	var message := str(data.get(Packets.FIELD_MESSAGE, "")).strip_edges()
	if message != "":
		return message

	var error_code := str(data.get(Packets.FIELD_ERROR_CODE, "")).strip_edges()
	return str(ROOM_ERROR_MESSAGES.get(error_code, FALLBACK_MESSAGE))
