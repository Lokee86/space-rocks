extends RefCounted

const GameplayReadiness = preload("res://scripts/protocol/realtime/gameplay_readiness.gd")
const WorldPresentationAdapter = preload("res://scripts/protocol/realtime/world_presentation_adapter.gd")
const OverlayPresentationAdapter = preload("res://scripts/protocol/realtime/overlay_presentation_adapter.gd")
const SessionPresentationAdapter = preload("res://scripts/protocol/realtime/session_presentation_adapter.gd")
const EventPresentationAdapter = preload("res://scripts/protocol/realtime/event_presentation_adapter.gd")

var gameplay_readiness := GameplayReadiness.new()
var world_adapter := WorldPresentationAdapter.new()
var overlay_adapter := OverlayPresentationAdapter.new()
var session_adapter := SessionPresentationAdapter.new()
var event_adapter := EventPresentationAdapter.new()
var _presented_once := false

func is_presentable() -> bool:
	return gameplay_readiness.is_gameplay_ready()

func can_fanout() -> bool:
	return is_presentable()

func bind_gameplay_readiness(readiness) -> void:
	if readiness == null:
		return
	gameplay_readiness = readiness

func fanout_lane_states(router, world_sync_ref = null, gameplay_hud_flow_ref = null, event_flow_ref = null) -> void:
	if router == null:
		return
	if not is_presentable():
		return

	var self_id := ""
	if router.overlay_lane_state != null and router.overlay_lane_state.self_id != null:
		self_id = str(router.overlay_lane_state.self_id)

	world_adapter.apply_world_lane_state(world_sync_ref, router.world_lane_state, self_id)
	overlay_adapter.apply_overlay_lane_state(gameplay_hud_flow_ref, router.overlay_lane_state)
	session_adapter.apply_session_lane_state(gameplay_hud_flow_ref, router.session_lane_state, self_id)

	var event_flow = null
	if event_flow_ref != null and event_flow_ref.has_method("apply_server_events"):
		event_flow = event_flow_ref
	elif gameplay_hud_flow_ref != null and gameplay_hud_flow_ref.has_method("apply_server_events"):
		event_flow = gameplay_hud_flow_ref
	event_adapter.apply_event_batch_output(event_flow, router.event_batch_applier, self_id)

func has_fanned_out() -> bool:
	return _presented_once

func mark_fanned_out() -> void:
	_presented_once = true
