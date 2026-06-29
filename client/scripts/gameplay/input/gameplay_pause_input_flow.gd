extends RefCounted
class_name GameplayPauseInputFlow

var menu_flow
var pending_open_menu_before_spawn := false


func configure(menu_flow_ref) -> void:
	menu_flow = menu_flow_ref


func reset() -> void:
	pending_open_menu_before_spawn = false


func process(required_lane_baselines_synced: bool) -> bool:
	if menu_flow == null:
		return false
	if !required_lane_baselines_synced && Input.is_action_just_pressed("OpenMenu"):
		pending_open_menu_before_spawn = true
		return true
	elif required_lane_baselines_synced:
		if menu_flow.handle_open_menu_pressed(required_lane_baselines_synced):
			return true
	if pending_open_menu_before_spawn && required_lane_baselines_synced:
		pending_open_menu_before_spawn = false
		menu_flow.open_live_pause_from_request(true)
		return true
	return false

