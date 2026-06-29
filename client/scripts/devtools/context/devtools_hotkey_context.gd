extends RefCounted
class_name DevtoolsHotkeyContext

var state_context
var overlay_context
var hotkey_flow
var window_toggle_route: Callable


func configure(state_context_ref, overlay_context_ref, hotkey_flow_ref, window_toggle_route_ref: Callable) -> void:
	state_context = state_context_ref
	overlay_context = overlay_context_ref
	hotkey_flow = hotkey_flow_ref
	window_toggle_route = window_toggle_route_ref


func process(required_lane_baselines_synced: bool) -> void:
	if Input.is_action_just_pressed("DevToggle0"):
		if !window_toggle_route.is_null():
			window_toggle_route.call()
	if Input.is_action_just_pressed("DevToggle8"):
		if Input.is_key_pressed(KEY_SHIFT):
			if state_context != null and state_context.get_player_dev_label_mode() == "network":
				state_context.set_player_dev_label_mode("")
			else:
				if state_context != null:
					state_context.set_player_dev_label_mode("network")
		else:
			if state_context != null and state_context.get_player_dev_label_mode() == "basic":
				state_context.set_player_dev_label_mode("")
			else:
				if state_context != null:
					state_context.set_player_dev_label_mode("basic")
		if overlay_context != null and state_context != null:
			overlay_context.set_player_dev_label_mode(state_context.get_player_dev_label_mode())
	if Input.is_action_just_pressed("DevToggle9") and overlay_context != null:
		overlay_context.toggle_world_telemetry_overlay()
	if hotkey_flow != null:
		hotkey_flow.process(required_lane_baselines_synced)
