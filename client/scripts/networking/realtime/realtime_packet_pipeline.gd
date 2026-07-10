extends RefCounted
class_name RealtimePacketPipeline

const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const CompactLanePacket := preload("res://scripts/protocol/realtime/compact_lane_packet.gd")
const RealtimeRouter := preload("res://scripts/protocol/realtime/realtime_router.gd")
const RealtimePresentationState := preload("res://scripts/networking/realtime/realtime_presentation_state.gd")

signal gameplay_packet_applied(packet)

var _router: RealtimeRouter
var _presentation_state: RealtimePresentationState

func _init() -> void:
	_router = RealtimeRouter.new()
	_presentation_state = RealtimePresentationState.new()
	_presentation_state.update_from_router(_router)

func get_router() -> RealtimeRouter:
	return _router

func is_gameplay_ready() -> bool:
	if _router == null:
		return false
	if _router.has_method("is_presentable"):
		return _router.is_presentable()
	if _router.has_method("get_gameplay_readiness"):
		var gameplay_readiness = _router.get_gameplay_readiness()
		if gameplay_readiness != null and gameplay_readiness.has_method("is_gameplay_ready"):
			return gameplay_readiness.is_gameplay_ready()
	return false

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
			_presentation_state.update_from_router(_router)
			gameplay_packet_applied.emit(expanded_packet)
		_:
			return

func reset() -> void:
	_router = RealtimeRouter.new()
	_presentation_state.update_from_router(_router)

func get_presentation_state() -> RealtimePresentationState:
	return _presentation_state
