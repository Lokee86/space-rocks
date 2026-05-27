extends RefCounted
class_name GameplayPauseInputFlow

var menu_flow
var pending_open_menu_before_spawn := false


func configure(menu_flow_ref) -> void:
	menu_flow = menu_flow_ref


func reset() -> void:
	pending_open_menu_before_spawn = false


func process(has_received_state: bool) -> void:
	if menu_flow == null:
		return
	if !has_received_state && Input.is_action_just_pressed("OpenMenu"):
		pending_open_menu_before_spawn = true
	elif has_received_state:
		menu_flow.handle_open_menu_pressed(has_received_state)
	if pending_open_menu_before_spawn && has_received_state:
		pending_open_menu_before_spawn = false
		menu_flow.open_live_pause_from_request(true)
