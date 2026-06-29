extends RefCounted

const LaneMetadata = preload("res://scripts/protocol/realtime/lane_metadata.gd")
const WorldLaneState = preload("res://scripts/protocol/realtime/world_lane_state.gd")
const WorldLaneApplier = preload("res://scripts/protocol/realtime/world_lane_applier.gd")
const OverlayLaneState = preload("res://scripts/protocol/realtime/overlay_lane_state.gd")
const OverlayLaneApplier = preload("res://scripts/protocol/realtime/overlay_lane_applier.gd")
const SessionLaneState = preload("res://scripts/protocol/realtime/session_lane_state.gd")
const SessionLaneApplier = preload("res://scripts/protocol/realtime/session_lane_applier.gd")
const EventBatchApplier = preload("res://scripts/protocol/realtime/event_batch_applier.gd")
const BaselineTracker = preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const GameplayReadiness = preload("res://scripts/protocol/realtime/gameplay_readiness.gd")
const ResyncState = preload("res://scripts/protocol/realtime/resync_state.gd")
const PresentationAdapter = preload("res://scripts/protocol/realtime/presentation_adapter.gd")

var world_lane_state := WorldLaneState.new()
var overlay_lane_state := OverlayLaneState.new()
var session_lane_state := SessionLaneState.new()
var event_batch_applier := EventBatchApplier.new()
var baseline_tracker := BaselineTracker.new()
var gameplay_readiness := GameplayReadiness.new()
var resync_state := ResyncState.new()
var presentation_adapter := PresentationAdapter.new()

var _world_applier := WorldLaneApplier.new()
var _overlay_applier := OverlayLaneApplier.new()
var _session_applier := SessionLaneApplier.new()

func _init() -> void:
	baseline_tracker.bind_readiness(gameplay_readiness)
	presentation_adapter.bind_gameplay_readiness(gameplay_readiness)

func route_packet(packet: Dictionary) -> Dictionary:
	var packet_type = packet.get("type")
	match packet_type:
		LaneMetadata.PACKET_FAMILY_WORLD[0]:
			_world_applier.apply_world_full(world_lane_state, baseline_tracker, LaneMetadata.LANE_WORLD, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_WORLD[1]:
			_world_applier.apply_world_delta(world_lane_state, baseline_tracker, LaneMetadata.LANE_WORLD, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_OVERLAY[0]:
			_overlay_applier.apply_overlay_full(overlay_lane_state, baseline_tracker, LaneMetadata.LANE_OVERLAY, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_OVERLAY[1]:
			_overlay_applier.apply_overlay_delta(overlay_lane_state, baseline_tracker, LaneMetadata.LANE_OVERLAY, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_SESSION[0]:
			_session_applier.apply_session_full(session_lane_state, baseline_tracker, LaneMetadata.LANE_SESSION, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_SESSION[1]:
			_session_applier.apply_session_delta(session_lane_state, baseline_tracker, LaneMetadata.LANE_SESSION, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_EVENT[0]:
			event_batch_applier.apply_event_batch(packet, self)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_CONTROL[0], LaneMetadata.PACKET_FAMILY_CONTROL[1]:
			_route_resync(packet)
	return {}

func route_packet_for_protocol_mode(packet: Dictionary, protocol_mode: String) -> Dictionary:
	if protocol_mode == "lane_protocol":
		return route_lane_packet(packet)
	return route_packet(packet)

func route_lane_packet(packet: Dictionary) -> Dictionary:
	var packet_type = packet.get("type")
	match packet_type:
		LaneMetadata.PACKET_FAMILY_WORLD[0]:
			_world_applier.apply_world_full(world_lane_state, baseline_tracker, LaneMetadata.LANE_WORLD, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_WORLD[1]:
			_world_applier.apply_world_delta(world_lane_state, baseline_tracker, LaneMetadata.LANE_WORLD, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_OVERLAY[0]:
			_overlay_applier.apply_overlay_full(overlay_lane_state, baseline_tracker, LaneMetadata.LANE_OVERLAY, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_OVERLAY[1]:
			_overlay_applier.apply_overlay_delta(overlay_lane_state, baseline_tracker, LaneMetadata.LANE_OVERLAY, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_SESSION[0]:
			_session_applier.apply_session_full(session_lane_state, baseline_tracker, LaneMetadata.LANE_SESSION, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_SESSION[1]:
			_session_applier.apply_session_delta(session_lane_state, baseline_tracker, LaneMetadata.LANE_SESSION, packet)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_EVENT[0]:
			event_batch_applier.apply_event_batch(packet, self)
			presentation_adapter.fanout_lane_states(self)
		LaneMetadata.PACKET_FAMILY_CONTROL[0], LaneMetadata.PACKET_FAMILY_CONTROL[1]:
			_route_resync(packet)
	return {}

func handle_presentation_event(event_type, payload, event_packet) -> void:
	pass

func apply_presentation_event(event_type, payload, event_packet) -> void:
	pass

func _route_resync(packet: Dictionary) -> void:
	var packet_type = packet.get("type")
	if packet_type == "resync_request":
		resync_state.mark_missing_baseline(_lane_from_packet(packet))
	elif packet_type == "resync_required":
		resync_state.mark_wrong_baseline(_lane_from_packet(packet))

func _lane_from_packet(packet: Dictionary) -> String:
	var lane = packet.get("lane")
	if lane != null:
		return lane
	return LaneMetadata.LANE_WORLD