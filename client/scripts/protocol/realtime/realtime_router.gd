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
const CompactLanePacket = preload("res://scripts/protocol/realtime/compact_lane_packet.gd")
const LifecycleLaneGate = preload("res://scripts/protocol/realtime/lifecycle_lane_gate.gd")

var world_lane_state := WorldLaneState.new()
var overlay_lane_state := OverlayLaneState.new()
var session_lane_state := SessionLaneState.new()
var event_batch_applier := EventBatchApplier.new()
var baseline_tracker := BaselineTracker.new()
var gameplay_readiness := GameplayReadiness.new()
var resync_state := ResyncState.new()
var lifecycle_lane_gate := LifecycleLaneGate.new()

var _world_applier := WorldLaneApplier.new()
var _overlay_applier := OverlayLaneApplier.new()
var _session_applier := SessionLaneApplier.new()

func _init() -> void:
	baseline_tracker.bind_readiness(gameplay_readiness)


func route_lane_packet(packet: Dictionary) -> Dictionary:
	var expanded_packet: Dictionary = packet
	if not expanded_packet.has("type"):
		if expanded_packet.has("t"):
			expanded_packet = CompactLanePacket.expand_packet(packet)
		else:
			return {}
	var packet_type = expanded_packet.get("type")
	match packet_type:
		"world_full":
			_route_world_full(expanded_packet)
		"world_delta":
			_world_applier.apply_world_delta(world_lane_state, baseline_tracker, LaneMetadata.LANE_WORLD, expanded_packet)
		"asteroid_delta":
			_world_applier.apply_asteroid_delta(world_lane_state, LaneMetadata.LANE_ASTEROIDS, expanded_packet)
		"bullet_delta":
			_world_applier.apply_bullet_delta(world_lane_state, LaneMetadata.LANE_BULLETS, expanded_packet)
		"asteroids_lifecycle":
			_route_asteroids_lifecycle(expanded_packet)
		"bullets_lifecycle":
			_route_bullets_lifecycle(expanded_packet)
		"overlay_full":
			_overlay_applier.apply_overlay_full(overlay_lane_state, baseline_tracker, LaneMetadata.LANE_OVERLAY, expanded_packet)
		"overlay_delta":
			_overlay_applier.apply_overlay_delta(overlay_lane_state, baseline_tracker, LaneMetadata.LANE_OVERLAY, expanded_packet)
		"session_full":
			_session_applier.apply_session_full(session_lane_state, baseline_tracker, LaneMetadata.LANE_SESSION, expanded_packet)
		"session_delta":
			_session_applier.apply_session_delta(session_lane_state, baseline_tracker, LaneMetadata.LANE_SESSION, expanded_packet)
		"event_batch":
			event_batch_applier.apply_event_batch(expanded_packet, self)
		"resync_request", "resync_required":
			_route_resync(expanded_packet)
	return {}



func get_gameplay_readiness():
	return gameplay_readiness


func is_presentable() -> bool:
	return gameplay_readiness != null and gameplay_readiness.is_gameplay_ready()


func _route_world_full(packet: Dictionary) -> void:
	_world_applier.apply_world_full(world_lane_state, baseline_tracker, LaneMetadata.LANE_WORLD, packet)
	var active_baseline_id := baseline_tracker.get_active_baseline_id(LaneMetadata.LANE_WORLD)
	if not baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD) or active_baseline_id != packet.get("baseline_id"):
		return
	for entry in lifecycle_lane_gate.take_pending_for_baseline(active_baseline_id):
		var lane = entry.get("lane")
		var lifecycle_packet: Dictionary = entry.get("packet", {})
		var sequence = entry.get("sequence")
		if lane == LaneMetadata.LANE_ASTEROIDS_LIFECYCLE:
			if _world_applier.apply_asteroids_lifecycle(world_lane_state, lifecycle_packet):
				lifecycle_lane_gate.mark_applied(lane, sequence)
		elif lane == LaneMetadata.LANE_BULLETS_LIFECYCLE:
			if _world_applier.apply_bullets_lifecycle(world_lane_state, lifecycle_packet):
				lifecycle_lane_gate.mark_applied(lane, sequence)
	lifecycle_lane_gate.discard_obsolete_baselines(active_baseline_id)

func _route_asteroids_lifecycle(packet: Dictionary) -> void:
	var decision := lifecycle_lane_gate.submit(
		LaneMetadata.LANE_ASTEROIDS_LIFECYCLE,
		packet,
		baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD),
		baseline_tracker.get_active_baseline_id(LaneMetadata.LANE_WORLD)
	)
	if decision.status != LifecycleLaneGate.DECISION_APPLY:
		return
	if _world_applier.apply_asteroids_lifecycle(world_lane_state, packet):
		lifecycle_lane_gate.mark_applied(LaneMetadata.LANE_ASTEROIDS_LIFECYCLE, decision.sequence)

func _route_bullets_lifecycle(packet: Dictionary) -> void:
	var decision := lifecycle_lane_gate.submit(
		LaneMetadata.LANE_BULLETS_LIFECYCLE,
		packet,
		baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD),
		baseline_tracker.get_active_baseline_id(LaneMetadata.LANE_WORLD)
	)
	if decision.status != LifecycleLaneGate.DECISION_APPLY:
		return
	if _world_applier.apply_bullets_lifecycle(world_lane_state, packet):
		lifecycle_lane_gate.mark_applied(LaneMetadata.LANE_BULLETS_LIFECYCLE, decision.sequence)

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


