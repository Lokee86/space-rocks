extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")


static func input_packet(forward, back, right, left, primary_fire, secondary_fire) -> Dictionary:
	return Packets.input_packet(forward, back, right, left, primary_fire, secondary_fire)


static func respawn_packet() -> Dictionary:
	return Packets.respawn_packet()


static func pause_request_packet() -> Dictionary:
	return Packets.pause_request_packet()


static func set_target_player_request_packet(target_kind, target_id) -> Dictionary:
	return Packets.set_target_player_request_packet(target_kind, target_id)


static func select_target_at_position_request_packet(x, y, target_kind, target_id) -> Dictionary:
	return Packets.select_target_at_position_request_packet(x, y, target_kind, target_id)


static func clear_target_request_packet() -> Dictionary:
	return Packets.clear_target_request_packet()
