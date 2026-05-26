extends RefCounted
class_name GameplaySpectateFlow

var menu_flow
var spectate_menu_state
var world_sync
var is_spectating := false


func configure(menu_flow_ref, spectate_menu_state_ref, world_sync_ref) -> void:
	menu_flow = menu_flow_ref
	spectate_menu_state = spectate_menu_state_ref
	world_sync = world_sync_ref


func reset() -> void:
	is_spectating = false


func process() -> void:
	if !is_spectating:
		return

	if Input.is_action_just_pressed("OpenMenu") && menu_flow != null:
		menu_flow.show_spectating_menu()

	if (
		Input.is_action_just_pressed("SwitchCamera")
		&& spectate_menu_state != null
		&& world_sync != null
	):
		var target_id: String = spectate_menu_state.cycle_next_target()
		if !target_id.is_empty():
			world_sync.focus_camera_on_player(target_id)


func begin_spectating() -> void:
	if spectate_menu_state == null || world_sync == null:
		return

	var target_id: String = spectate_menu_state.begin_spectating()
	if !target_id.is_empty() && world_sync.focus_camera_on_player(target_id):
		is_spectating = true
