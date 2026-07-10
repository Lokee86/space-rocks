extends RefCounted
class_name RealtimePacketPipeline

const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const CompactLanePacket := preload("res://scripts/protocol/realtime/compact_lane_packet.gd")
const RealtimeRouter := preload("res://scripts/protocol/realtime/realtime_router.gd")

signal gameplay_packet_applied(packet)

var _router: RealtimeRouter

func _init() -> void:
	_router = RealtimeRouter.new()

func get_router() -> RealtimeRouter:
	return _router

func get_readiness():
	return _router.get_gameplay_readiness()

func apply_packet(packet: Dictionary) -> void:
	if _router == null:
		return
	var expanded_packet: Dictionary = packet
	if not expanded_packet.has("type"):
		if expanded_packet.has("t"):
			expanded_packet = CompactLanePacket.expand_packet(packet)
		else:
			return
	var packet_type = expanded_packet.get("type")
	match packet_type:
		LaneMetadata.PACKET_FAMILY_WORLD[0], LaneMetadata.PACKET_FAMILY_WORLD[1], "asteroid_delta", "bullet_delta", "asteroids_lifecycle", "bullets_lifecycle", LaneMetadata.PACKET_FAMILY_OVERLAY[0], LaneMetadata.PACKET_FAMILY_OVERLAY[1], LaneMetadata.PACKET_FAMILY_SESSION[0], LaneMetadata.PACKET_FAMILY_SESSION[1], LaneMetadata.PACKET_FAMILY_EVENT[0], LaneMetadata.PACKET_FAMILY_CONTROL[0], LaneMetadata.PACKET_FAMILY_CONTROL[1]:
			_router.route_lane_packet(expanded_packet)
			gameplay_packet_applied.emit(expanded_packet)
		_:
			return

func reset() -> void:
	_router = RealtimeRouter.new()
