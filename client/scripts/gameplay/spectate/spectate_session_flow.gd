extends RefCounted
class_name SpectateSessionFlow

const SpectateMenuState := preload("res://scripts/gameplay/spectate/spectate_menu_state.gd")

var gameplay_menu_flow
var gameplay_shell_flow
var spectate_menu_state


func configure(gameplay_menu_flow_ref, gameplay_shell_flow_ref, spectate_menu_state_ref = null) -> void:
	gameplay_menu_flow = gameplay_menu_flow_ref
	gameplay_shell_flow = gameplay_shell_flow_ref
	if spectate_menu_state_ref != null:
		spectate_menu_state = spectate_menu_state_ref
	else:
		spectate_menu_state = SpectateMenuState.new()
	if gameplay_menu_flow != null && gameplay_menu_flow.has_method("configure_spectate_menu_state"):
		gameplay_menu_flow.configure_spectate_menu_state(spectate_menu_state)


func apply_gameplay_state(state: Dictionary) -> void:
	if state == null:
		return
	if spectate_menu_state == null:
		return
	spectate_menu_state.apply_gameplay_state(state)


func reset() -> void:
	if spectate_menu_state == null:
		return
	spectate_menu_state.reset()

