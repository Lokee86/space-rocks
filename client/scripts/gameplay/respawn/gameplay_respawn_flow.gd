extends RefCounted
class_name GameplayRespawnFlow

const PlayerLifecycle = preload("res://scripts/gameplay/lifecycle/player_lifecycle.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

var connection_service
var hud_flow
var awaiting_respawn_confirmation := false
var _logged_respawn_send := false
var _logged_respawn_blocked := {}


func configure(connection_service_ref, hud_flow_ref) -> void:
	connection_service = connection_service_ref
	hud_flow = hud_flow_ref


func reset() -> void:
	awaiting_respawn_confirmation = false
	_logged_respawn_send = false
	_logged_respawn_blocked.clear()


func request_respawn(required_lane_baselines_synced: bool) -> void:
	if !required_lane_baselines_synced:
		_log_respawn_blocked_once("readiness false")
		return
	if connection_service == null:
		_log_respawn_blocked_once("connection_service null")
		return
	if hud_flow == null:
		_log_respawn_blocked_once("hud_flow null")
		return
	if !hud_flow.can_request_respawn():
		_log_respawn_blocked_once("can_request_respawn false")
		return

	if !_logged_respawn_send:
		_logged_respawn_send = true
		ClientLogger.network_info("sending respawn request to network client")
	connection_service.send_respawn_request()
	mark_awaiting_confirmation()


func mark_awaiting_confirmation() -> void:
	awaiting_respawn_confirmation = true
	ClientLogger.network_info("respawn awaiting confirmation marked")


func clear_awaiting_confirmation() -> void:
	awaiting_respawn_confirmation = false


func is_awaiting_confirmation() -> bool:
	return awaiting_respawn_confirmation


func should_restore_alive_hud(
	world_ships: Dictionary,
	player_lifecycle: Dictionary,
	self_id: String,
	player,
	has_stale_dead_presentation := false
) -> bool:
	if !awaiting_respawn_confirmation && !has_stale_dead_presentation:
		return false

	if !PlayerLifecycle.is_player_active(player_lifecycle, self_id):
		return false

	if !world_ships.has(self_id):
		return false

	var self_state = world_ships[self_id]
	var has_valid_server_state := false
	if self_state is Dictionary:
		var self_state_dictionary: Dictionary = self_state
		has_valid_server_state = !self_state_dictionary.is_empty()

	var player_visible := false
	if player != null and player.get("visible") != null:
		player_visible = bool(player.get("visible"))

	return player_visible or has_valid_server_state


func _log_respawn_blocked_once(reason: String) -> void:
	if _logged_respawn_blocked.has(reason):
		return
	_logged_respawn_blocked[reason] = true
	ClientLogger.network_info("respawn request blocked: %s" % reason)