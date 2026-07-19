extends RefCounted
class_name PresentationAdapter





const RealtimeQuantize = preload("res://scripts/protocol/realtime/realtime_quantize.gd")

var world_adapter := WorldPresentationAdapter.new()
var overlay_adapter := OverlayPresentationAdapter.new()
var session_adapter := SessionPresentationAdapter.new()
var event_adapter := EventPresentationAdapter.new()

func fanout_lane_states(presentation_state: RealtimePresentationState, world_sync_ref: WorldSync = null, gameplay_hud_flow_ref: GameplayHudFlow = null, event_flow_ref: GameplayEventLifecycleFlow = null, local_lifecycle_flow_ref: GameplayLocalLifecycleFlow = null) -> void:
	if presentation_state == null:
		return

	var self_id := ""
	if presentation_state.overlay_lane_state != null and presentation_state.overlay_lane_state.self_id != null:
		self_id = str(presentation_state.overlay_lane_state.self_id)

	world_adapter.apply_world_lane_state(world_sync_ref, presentation_state.world_lane_state, self_id)
	overlay_adapter.apply_overlay_lane_state(gameplay_hud_flow_ref, RealtimeQuantize.decode_overlay_state(presentation_state.overlay_lane_state))
	var decoded_session_state = RealtimeQuantize.decode_session_state(presentation_state.session_lane_state)
	session_adapter.apply_session_lane_state(gameplay_hud_flow_ref, decoded_session_state, self_id)
	if local_lifecycle_flow_ref != null:
		local_lifecycle_flow_ref.apply_lane_state(
			presentation_state.world_lane_state,
			decoded_session_state,
			self_id
		)

	event_adapter.apply_event_batch_output(event_flow_ref, presentation_state.event_batch_applier, self_id)
