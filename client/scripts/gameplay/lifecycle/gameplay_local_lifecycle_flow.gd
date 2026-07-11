extends RefCounted
class_name GameplayLocalLifecycleFlow

const PlayerLifecycle = preload("res://scripts/gameplay/lifecycle/player_lifecycle.gd")

var world_sync
var respawn_flow
var hud_flow
var match_end_flow
var player
var _last_status: String = ""
var _last_respawn_cooldown: Variant = null

func configure(world_sync_ref, respawn_flow_ref, hud_flow_ref, match_end_flow_ref, player_ref) -> void:
	world_sync = world_sync_ref
	respawn_flow = respawn_flow_ref
	hud_flow = hud_flow_ref
	match_end_flow = match_end_flow_ref
	player = player_ref

func reset() -> void:
	_last_status = ""
	_last_respawn_cooldown = null

func apply_state(state: Dictionary) -> void:
	if hud_flow == null:
		return
	var world_value: Dictionary = state.get("world", {}) if state.get("world", {}) is Dictionary else {}
	var session_value: Dictionary = state.get("session", {}) if state.get("session", {}) is Dictionary else {}
	var world_ships: Dictionary = world_value.get("ships", {}) if world_value.get("ships", {}) is Dictionary else {}
	var player_lifecycle: Dictionary = session_value.get("player_lifecycle", {}) if session_value.get("player_lifecycle", {}) is Dictionary else {}
	var self_id := str(state.get("self_id", ""))
	var status := PlayerLifecycle.status_for(player_lifecycle, self_id)
	if status != _last_status and status == PlayerLifecycle.STATUS_ELIMINATED:
		_apply_eliminated(session_value, self_id)
	_last_status = status
	if status != PlayerLifecycle.STATUS_ACTIVE or respawn_flow == null:
		return
	var stale: bool = match_end_flow != null and match_end_flow.has_method("has_stale_dead_presentation") and match_end_flow.has_stale_dead_presentation()
	if !respawn_flow.should_restore_alive_hud(world_ships, player_lifecycle, self_id, player, stale):
		return
	if world_sync != null:
		world_sync.clear_view_target_player()
	hud_flow.set_alive()
	if match_end_flow != null and match_end_flow.has_method("handle_alive_restored"):
		match_end_flow.handle_alive_restored()
	respawn_flow.clear_awaiting_confirmation()

func apply_lane_state(world_lane_state, session_lane_state, self_id: String) -> void:
	if hud_flow == null or world_lane_state == null or session_lane_state == null or self_id == "":
		return
	if hud_flow.hidden_for_match_over or hud_flow.is_game_over:
		return
	var player_lifecycle: Dictionary = session_lane_state.player_lifecycle if session_lane_state.player_lifecycle is Dictionary else {}
	var status := PlayerLifecycle.status_for(player_lifecycle, self_id)
	if status == PlayerLifecycle.STATUS_PENDING_RESPAWN:
		_apply_pending_respawn(session_lane_state, self_id)
		_last_status = status
		return
	if status != _last_status and status == PlayerLifecycle.STATUS_ELIMINATED:
		_apply_eliminated({"players": session_lane_state.player_sessions}, self_id)
	_last_status = status
	if status == PlayerLifecycle.STATUS_ELIMINATED:
		return
	if respawn_flow == null:
		return
	var world_ships: Dictionary = world_lane_state.ships if world_lane_state.ships is Dictionary else {}
	var stale: bool = hud_flow.has_method("has_dead_presentation") and hud_flow.has_dead_presentation()
	if !respawn_flow.should_restore_alive_hud(world_ships, player_lifecycle, self_id, player, stale):
		return
	if world_sync != null:
		world_sync.clear_view_target_player()
	hud_flow.clear_dead_presentation()
	if match_end_flow != null and match_end_flow.has_method("handle_alive_restored"):
		match_end_flow.handle_alive_restored()
	respawn_flow.clear_awaiting_confirmation()

func _apply_pending_respawn(session_lane_state, self_id: String) -> void:
	var player_sessions: Dictionary = session_lane_state.player_sessions if session_lane_state.player_sessions is Dictionary else {}
	var player_session = player_sessions.get(self_id, {})
	if player_session is Dictionary and player_session.has("lives") and hud_flow.has_method("apply_lives"):
		hud_flow.apply_lives(int(player_session.get("lives", 0)))
	var cooldown: float = 0.0
	if player_session is Dictionary:
		cooldown = float(player_session.get("respawn_cooldown", 0.0))
	var entering_pending: bool = _last_status != PlayerLifecycle.STATUS_PENDING_RESPAWN
	var cooldown_changed: bool = _last_respawn_cooldown == null or !is_equal_approx(float(_last_respawn_cooldown), cooldown)
	if entering_pending and player != null and player.has_method("stop_transient_effects"):
		player.stop_transient_effects()
	var has_dead_presentation: bool = hud_flow.has_method("has_dead_presentation") and hud_flow.has_dead_presentation()
	if entering_pending or cooldown_changed or !has_dead_presentation:
		hud_flow.set_dead(cooldown)
	_last_respawn_cooldown = cooldown


func _apply_eliminated(session_value: Dictionary, self_id: String) -> void:
	if player != null and player.has_method("stop_transient_effects"):
		player.stop_transient_effects()
	var players = session_value.get("players", {})
	var record = players.get(self_id, {}) if players is Dictionary else {}
	var lives := int(record.get("lives", 0)) if record is Dictionary else 0
	if hud_flow.has_method("apply_lives"):
		hud_flow.apply_lives(lives)
	if match_end_flow != null and match_end_flow.has_method("handle_local_player_eliminated"):
		match_end_flow.handle_local_player_eliminated(lives)
