extends RefCounted

func apply_session_lane_state(hud_flow, session_lane_state, self_id = "") -> void:
	if hud_flow == null or session_lane_state == null:
		return
	if hud_flow.has_method("apply_session_lane_state"):
		hud_flow.apply_session_lane_state(session_lane_state, self_id)
