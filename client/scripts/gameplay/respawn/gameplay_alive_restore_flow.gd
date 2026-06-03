extends RefCounted
class_name GameplayAliveRestoreFlow

var world_sync
var respawn_flow
var hud_flow
var menu_flow
var player


func configure(world_sync_ref, respawn_flow_ref, hud_flow_ref, menu_flow_ref, player_ref) -> void:
	world_sync = world_sync_ref
	respawn_flow = respawn_flow_ref
	hud_flow = hud_flow_ref
	menu_flow = menu_flow_ref
	player = player_ref


func reset() -> void:
	pass


func apply_state(state: Dictionary) -> void:
	if hud_flow == null || respawn_flow == null:
		return

	var has_stale_dead_presentation: bool = false
	has_stale_dead_presentation = bool(hud_flow.is_dead) || bool(hud_flow.is_game_over)
	if menu_flow != null:
		has_stale_dead_presentation = has_stale_dead_presentation || bool(menu_flow.is_game_over)

	if !respawn_flow.should_restore_alive_hud(state, player, has_stale_dead_presentation):
		return

	if world_sync != null:
		world_sync.clear_view_target_player()
	hud_flow.set_alive()
	if menu_flow != null:
		menu_flow.set_alive()
	respawn_flow.clear_awaiting_confirmation()
