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

	var world_value: Dictionary = state.get("world", {}) if state.has("world") and state["world"] is Dictionary else {}
	var session_value: Dictionary = state.get("session", {}) if state.has("session") and state["session"] is Dictionary else {}
	var world_ships: Dictionary = world_value.get("ships", {}) if world_value.has("ships") and world_value["ships"] is Dictionary else {}
	var player_lifecycle: Dictionary = session_value.get("player_lifecycle", {}) if session_value.has("player_lifecycle") and session_value["player_lifecycle"] is Dictionary else {}
	var self_id: String = str(state.get("self_id", ""))

	var has_stale_dead_presentation: bool = false
	if match_end_flow != null && match_end_flow.has_method("has_stale_dead_presentation"):
		has_stale_dead_presentation = match_end_flow.has_stale_dead_presentation()

	if !respawn_flow.should_restore_alive_hud(world_ships, player_lifecycle, self_id, player, has_stale_dead_presentation):
		return

	if world_sync != null:
		world_sync.clear_view_target_player()
	hud_flow.set_alive()
	if match_end_flow != null && match_end_flow.has_method("handle_alive_restored"):
		match_end_flow.handle_alive_restored()
	respawn_flow.clear_awaiting_confirmation()


func apply_lane_state(world_lane_state, session_lane_state, self_id: String) -> void:
	if hud_flow == null || respawn_flow == null || world_lane_state == null || session_lane_state == null || self_id == "":
		return
	if hud_flow.hidden_for_match_over or hud_flow.is_game_over:
		return

	var world_ships: Dictionary = {}
	if world_lane_state.ships is Dictionary:
		world_ships = world_lane_state.ships

	var player_lifecycle: Dictionary = {}
	if session_lane_state.player_lifecycle is Dictionary:
		player_lifecycle = session_lane_state.player_lifecycle

	var has_stale_dead_presentation := false
	if hud_flow.has_method("has_dead_presentation"):
		has_stale_dead_presentation = hud_flow.has_dead_presentation()
	elif hud_flow.has_method("_has_dead_presentation"):
		has_stale_dead_presentation = hud_flow._has_dead_presentation()

	if !respawn_flow.should_restore_alive_hud(world_ships, player_lifecycle, self_id, player, has_stale_dead_presentation):
		return

	if world_sync != null:
		world_sync.clear_view_target_player()
	hud_flow.clear_dead_presentation()
	if match_end_flow != null && match_end_flow.has_method("handle_alive_restored"):
		match_end_flow.handle_alive_restored()
	respawn_flow.clear_awaiting_confirmation()
