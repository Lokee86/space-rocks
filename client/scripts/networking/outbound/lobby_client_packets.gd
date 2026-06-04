extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")


static func create_room_request_packet() -> Dictionary:
	return Packets.create_room_request_packet()


static func join_room_request_packet(room_code) -> Dictionary:
	return Packets.join_room_request_packet(room_code)


static func leave_room_request_packet() -> Dictionary:
	return Packets.leave_room_request_packet()


static func set_ready_request_packet(ready) -> Dictionary:
	return Packets.set_ready_request_packet(ready)


static func start_game_request_packet() -> Dictionary:
	return Packets.start_game_request_packet()


static func start_single_player_request_packet() -> Dictionary:
	return Packets.start_single_player_request_packet()


static func return_to_lobby_request_packet() -> Dictionary:
	return Packets.return_to_lobby_request_packet()

