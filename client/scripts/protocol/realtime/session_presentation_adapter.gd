extends RefCounted
class_name SessionPresentationAdapter

func apply_session_lane_state(hud_flow: GameplayHudFlow, session_lane_state, self_id := "") -> void:
	if hud_flow == null or session_lane_state == null:
		return
	hud_flow.apply_session_lane_state(session_lane_state, self_id)
