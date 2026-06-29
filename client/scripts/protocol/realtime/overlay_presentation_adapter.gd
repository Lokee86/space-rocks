extends RefCounted

func apply_overlay_lane_state(hud_flow, overlay_lane_state) -> void:
	if hud_flow == null or overlay_lane_state == null:
		return
	if hud_flow.has_method("apply_overlay_lane_state"):
		hud_flow.apply_overlay_lane_state(overlay_lane_state)
