extends Node

const ServerPacketRouter := preload("res://scripts/networking/inbound/server_packet_router.gd")

signal room_snapshot_received(packet: Dictionary)
signal room_state_changed(packet: Dictionary)
signal room_error_received(packet: Dictionary)
signal gameplay_state_received(packet: Dictionary)
signal debug_status_received(packet: Dictionary)
signal player_pause_state_received(packet: Dictionary)
signal telemetry_pong_received(packet: Dictionary)
signal unknown_packet_received(packet: Dictionary)


func dispatch(packet: Dictionary) -> void:
	if ServerPacketRouter.is_room_snapshot(packet):
		room_snapshot_received.emit(packet)
	elif ServerPacketRouter.is_room_state_changed(packet):
		room_state_changed.emit(packet)
	elif ServerPacketRouter.is_room_error(packet):
		room_error_received.emit(packet)
	elif ServerPacketRouter.is_gameplay_state(packet):
		gameplay_state_received.emit(packet)
	elif ServerPacketRouter.is_debug_status(packet):
		debug_status_received.emit(packet)
	elif ServerPacketRouter.is_player_pause_state(packet):
		player_pause_state_received.emit(packet)
	elif ServerPacketRouter.is_telemetry_pong(packet):
		telemetry_pong_received.emit(packet)
	else:
		unknown_packet_received.emit(packet)
