extends RefCounted

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")


static func packet_type(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_TYPE, ""))


static func is_room_snapshot(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_ROOM_SNAPSHOT


static func is_loadout_options(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_LOADOUT_OPTIONS


static func is_room_state_changed(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_ROOM_STATE_CHANGED


static func is_room_error(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_ROOM_ERROR


static func is_authenticate_result(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_AUTHENTICATE_RESULT


static func is_world_full(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_WORLD_FULL


static func is_world_delta(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_WORLD_DELTA


static func is_ship_delta(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_SHIP_DELTA


static func is_player_locator(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_PLAYER_LOCATOR


static func is_ships_lifecycle(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_SHIPS_LIFECYCLE


static func is_asteroid_delta(packet: Dictionary) -> bool:
	return packet_type(packet) == "asteroid_delta"


static func is_bullet_delta(packet: Dictionary) -> bool:
	return packet_type(packet) == "bullet_delta"


static func is_asteroids_lifecycle(packet: Dictionary) -> bool:
	return packet_type(packet) == "asteroids_lifecycle"


static func is_bullets_lifecycle(packet: Dictionary) -> bool:
	return packet_type(packet) == "bullets_lifecycle"


static func is_overlay_full(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_OVERLAY_FULL


static func is_overlay_delta(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_OVERLAY_DELTA


static func is_session_full(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_SESSION_FULL


static func is_session_delta(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_SESSION_DELTA


static func is_event_batch(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_EVENT_BATCH


static func is_resync_request(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_RESYNC_REQUEST


static func is_resync_required(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_RESYNC_REQUIRED


static func is_webrtc_answer(packet: Dictionary) -> bool:
	return packet_type(packet) == "webrtc_answer"


static func is_webrtc_ice_candidate(packet: Dictionary) -> bool:
	return packet_type(packet) == "webrtc_ice_candidate"


static func is_webrtc_ready(packet: Dictionary) -> bool:
	return packet_type(packet) == "webrtc_ready"


static func is_webrtc_smoke(packet: Dictionary) -> bool:
	return packet_type(packet) == "webrtc_smoke"


static func is_webrtc_failed(packet: Dictionary) -> bool:
	return packet_type(packet) == "webrtc_failed"


static func is_player_pause_state(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_PLAYER_PAUSE_STATE
