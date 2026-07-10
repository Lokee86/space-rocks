extends RefCounted

const WorldPresentationAdapter = preload("res://scripts/protocol/realtime/world_presentation_adapter.gd")
const OverlayPresentationAdapter = preload("res://scripts/protocol/realtime/overlay_presentation_adapter.gd")
const SessionPresentationAdapter = preload("res://scripts/protocol/realtime/session_presentation_adapter.gd")
const EventPresentationAdapter = preload("res://scripts/protocol/realtime/event_presentation_adapter.gd")
const RealtimeQuantize = preload("res://scripts/protocol/realtime/realtime_quantize.gd")

var world_adapter := WorldPresentationAdapter.new()
var overlay_adapter := OverlayPresentationAdapter.new()
var session_adapter := SessionPresentationAdapter.new()
var event_adapter := EventPresentationAdapter.new()

func fanout_lane_states(presentation_state, world_sync_ref = null, gameplay_hud_flow_ref = null, event_flow_ref = null) -> void:
	if presentation_state == null:
		return

	var self_id := ""
	if presentation_state.overlay_lane_state != null and presentation_state.overlay_lane_state.self_id != null:
		self_id = str(presentation_state.overlay_lane_state.self_id)

	world_adapter.apply_world_lane_state(world_sync_ref, presentation_state.world_lane_state, self_id)
	overlay_adapter.apply_overlay_lane_state(gameplay_hud_flow_ref, RealtimeQuantize.decode_overlay_state(presentation_state.overlay_lane_state))
	session_adapter.apply_session_lane_state(gameplay_hud_flow_ref, RealtimeQuantize.decode_session_state(presentation_state.session_lane_state), self_id)

	var event_flow = null
	if event_flow_ref != null and event_flow_ref.has_method("apply_server_events"):
		event_flow = event_flow_ref
	elif gameplay_hud_flow_ref != null and gameplay_hud_flow_ref.has_method("apply_server_events"):
		event_flow = gameplay_hud_flow_ref
	event_adapter.apply_event_batch_output(event_flow, presentation_state.event_batch_applier, self_id)
