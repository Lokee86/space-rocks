extends RefCounted
class_name GameplayAliveRestoreFlow

var world_sync
var respawn_flow
var hud_flow
var match_end_flow
var player


func configure(world_sync_ref, respawn_flow_ref, hud_flow_ref, match_end_flow_ref, player_ref) -> void:
	world_sync = world_sync_ref
	respawn_flow = respawn_flow_ref
	hud_flow = hud_flow_ref
	match_end_flow = match_end_flow_ref
	player = player_ref


func reset() -> void:
	pass


func apply_state(state: Dictionary) -> void:
	if hud_flow == null || respawn_flow == null:
		return

	var has_stale_dead_presentation: bool = false
	if match_end_flow != null && match_end_flow.has_method("has_stale_dead_presentation"):
		has_stale_dead_presentation = match_end_flow.has_stale_dead_presentation()

	if !respawn_flow.should_restore_alive_hud(state, player, has_stale_dead_presentation):
		return

	if world_sync != null:
		world_sync.clear_view_target_player()
	hud_flow.set_alive()
	if match_end_flow != null && match_end_flow.has_method("handle_alive_restored"):
		match_end_flow.handle_alive_restored()
	respawn_flow.clear_awaiting_confirmation()

