extends Node

const ServerPacketRouter := preload("res://scripts/networking/inbound/server_packet_router.gd")

signal room_snapshot_received(packet: Dictionary)
signal authenticate_result_received(packet: Dictionary)
signal room_state_changed(packet: Dictionary)
signal room_error_received(packet: Dictionary)
signal world_full_received(packet: Dictionary)
signal world_delta_received(packet: Dictionary)
signal ship_delta_received(packet: Dictionary)
signal player_locator_received(packet: Dictionary)
signal ships_lifecycle_received(packet: Dictionary)
signal asteroid_delta_received(packet: Dictionary)
signal bullet_delta_received(packet: Dictionary)
signal asteroids_lifecycle_received(packet: Dictionary)
signal bullets_lifecycle_received(packet: Dictionary)
signal overlay_full_received(packet: Dictionary)
signal overlay_delta_received(packet: Dictionary)
signal session_full_received(packet: Dictionary)
signal session_delta_received(packet: Dictionary)
signal event_batch_received(packet: Dictionary)
signal resync_request_received(packet: Dictionary)
signal resync_required_received(packet: Dictionary)
signal player_pause_state_received(packet: Dictionary)
signal webrtc_answer_received(packet: Dictionary)
signal webrtc_ice_candidate_received(packet: Dictionary)
signal webrtc_ready_received(packet: Dictionary)
signal webrtc_smoke_received(packet: Dictionary)
signal webrtc_failed_received(packet: Dictionary)
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
	elif ServerPacketRouter.is_ship_delta(packet):
		ship_delta_received.emit(packet)
	elif ServerPacketRouter.is_player_locator(packet):
		player_locator_received.emit(packet)
	elif ServerPacketRouter.is_ships_lifecycle(packet):
		ships_lifecycle_received.emit(packet)
	elif ServerPacketRouter.is_asteroid_delta(packet):
		asteroid_delta_received.emit(packet)
	elif ServerPacketRouter.is_bullet_delta(packet):
		bullet_delta_received.emit(packet)
	elif ServerPacketRouter.is_asteroids_lifecycle(packet):
		asteroids_lifecycle_received.emit(packet)
	elif ServerPacketRouter.is_bullets_lifecycle(packet):
		bullets_lifecycle_received.emit(packet)
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
	elif ServerPacketRouter.is_webrtc_answer(packet):
		webrtc_answer_received.emit(packet)
	elif ServerPacketRouter.is_webrtc_ice_candidate(packet):
		webrtc_ice_candidate_received.emit(packet)
	elif ServerPacketRouter.is_webrtc_ready(packet):
		webrtc_ready_received.emit(packet)
	elif ServerPacketRouter.is_webrtc_smoke(packet):
		webrtc_smoke_received.emit(packet)
	elif ServerPacketRouter.is_webrtc_failed(packet):
		webrtc_failed_received.emit(packet)
	elif ServerPacketRouter.is_player_pause_state(packet):
		player_pause_state_received.emit(packet)
	else:
		unknown_packet_received.emit(packet)


func _as_world_delta_packet(packet: Dictionary) -> Dictionary:
	var world_delta_packet: Dictionary = packet.duplicate(true)
	world_delta_packet["type"] = "world_delta"
	return world_delta_packet
