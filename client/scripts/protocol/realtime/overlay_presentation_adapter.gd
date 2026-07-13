extends RefCounted
class_name OverlayPresentationAdapter

func apply_overlay_lane_state(hud_flow: GameplayHudFlow, overlay_lane_state) -> void:
	if hud_flow == null or overlay_lane_state == null:
		return
	hud_flow.apply_overlay_lane_state(overlay_lane_state)
