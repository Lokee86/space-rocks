extends RefCounted
class_name GameplayRespawnFlow

const PlayerLifecycle = preload("res://scripts/gameplay/lifecycle/player_lifecycle.gd")

var connection_service
var hud_flow: GameplayHudFlow = null
var awaiting_respawn_confirmation := false


func configure(connection_service_ref, hud_flow_ref: GameplayHudFlow) -> void:
	connection_service = connection_service_ref
	hud_flow = hud_flow_ref


func reset() -> void:
	awaiting_respawn_confirmation = false


func request_respawn(required_lane_baselines_synced: bool) -> void:
	if !required_lane_baselines_synced:
		return
	if connection_service == null:
		return
	if hud_flow == null:
		return
	if !hud_flow.can_request_respawn():
		return

	connection_service.send_respawn_request()
	mark_awaiting_confirmation()


func mark_awaiting_confirmation() -> void:
	awaiting_respawn_confirmation = true


func clear_awaiting_confirmation() -> void:
	awaiting_respawn_confirmation = false


func is_awaiting_confirmation() -> bool:
	return awaiting_respawn_confirmation


func should_restore_alive_hud(
	world_ships: Dictionary,
	player_lifecycle: Dictionary,
	self_id: String,
	player: Player,
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

	var player_visible := player != null and player.visible

	return player_visible or has_valid_server_state
