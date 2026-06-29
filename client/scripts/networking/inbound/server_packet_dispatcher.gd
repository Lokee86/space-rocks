extends Node

const ServerPacketRouter := preload("res://scripts/networking/inbound/server_packet_router.gd")

signal room_snapshot_received(packet: Dictionary)
signal authenticate_result_received(packet: Dictionary)
signal room_state_changed(packet: Dictionary)
signal room_error_received(packet: Dictionary)
signal world_full_received(packet: Dictionary)
signal world_delta_received(packet: Dictionary)
signal overlay_full_received(packet: Dictionary)
signal overlay_delta_received(packet: Dictionary)
signal session_full_received(packet: Dictionary)
signal session_delta_received(packet: Dictionary)
signal event_batch_received(packet: Dictionary)
signal resync_request_received(packet: Dictionary)
signal resync_required_received(packet: Dictionary)
signal debug_shape_catalog_received(packet: Dictionary)
signal debug_status_received(packet: Dictionary)
signal player_pause_state_received(packet: Dictionary)
signal telemetry_pong_received(packet: Dictionary)
signal unknown_packet_received(packet: Dictionary)


func dispatch(packet: Dictionary) -> void:
	if ServerPacketRouter.is_room_snapshot(packet):
		room_snapshot_received.emit(packet)
	elif ServerPacketRouter.is_authenticate_result(packet):
		authenticate_result_received.emit(packet)
	elif ServerPacketRouter.is_room_state_changed(packet):
		room_state_changed.emit(packet)
	elif ServerPacketRouter.is_room_error(packet):
		room_error_received.emit(packet)
	elif ServerPacketRouter.is_world_full(packet):
		world_full_received.emit(packet)
	elif ServerPacketRouter.is_world_delta(packet):
		world_delta_received.emit(packet)
	elif ServerPacketRouter.is_overlay_full(packet):
		overlay_full_received.emit(packet)
	elif ServerPacketRouter.is_overlay_delta(packet):
		overlay_delta_received.emit(packet)
	elif ServerPacketRouter.is_session_full(packet):
		session_full_received.emit(packet)
	elif ServerPacketRouter.is_session_delta(packet):
		session_delta_received.emit(packet)
	elif ServerPacketRouter.is_event_batch(packet):
		event_batch_received.emit(packet)
	elif ServerPacketRouter.is_resync_request(packet):
		resync_request_received.emit(packet)
	elif ServerPacketRouter.is_resync_required(packet):
		resync_required_received.emit(packet)
	elif ServerPacketRouter.is_debug_shape_catalog(packet):
		debug_shape_catalog_received.emit(packet)
	elif ServerPacketRouter.is_debug_status(packet):
		debug_status_received.emit(packet)
	elif ServerPacketRouter.is_player_pause_state(packet):
		player_pause_state_received.emit(packet)
	elif ServerPacketRouter.is_telemetry_pong(packet):
		telemetry_pong_received.emit(packet)
	else:
		unknown_packet_received.emit(packet)
