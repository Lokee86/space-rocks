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
	if world_sync != null:
		if world_sync.has_method("clear_view_reference_player"):
			world_sync.clear_view_reference_player()
		if world_sync.has_method("clear_view_target_player"):
			world_sync.clear_view_target_player()


func process() -> void:
	pass


func request_open_spectate_menu() -> void:
	if !is_spectating || menu_flow == null:
		return
	menu_flow.show_spectating_menu()


func request_cycle_target() -> void:
	if !is_spectating || spectate_menu_state == null || world_sync == null:
		return

	var target_id: String = spectate_menu_state.cycle_next_target()
	if !target_id.is_empty():
		if world_sync.has_method("set_view_reference_player"):
			world_sync.set_view_reference_player(target_id)
		if world_sync.has_method("set_view_target_player"):
			world_sync.set_view_target_player(target_id)
		world_sync.focus_camera_on_player(target_id)


func begin_spectating() -> void:
	if spectate_menu_state == null || world_sync == null:
		return

	var target_id: String = spectate_menu_state.begin_spectating()
	if !target_id.is_empty():
		if world_sync.has_method("set_view_reference_player"):
			world_sync.set_view_reference_player(target_id)
		if world_sync.has_method("set_view_target_player"):
			world_sync.set_view_target_player(target_id)
		var focused: bool = bool(world_sync.focus_camera_on_player(target_id))
		if focused:
			is_spectating = true
