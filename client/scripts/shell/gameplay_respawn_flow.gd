extends RefCounted
class_name GameplayRespawnFlow

var connection_service
var hud_flow
var awaiting_respawn_confirmation := false


func configure(connection_service_ref, hud_flow_ref) -> void:
	connection_service = connection_service_ref
	hud_flow = hud_flow_ref


func reset() -> void:
	awaiting_respawn_confirmation = false


func process(_has_received_state: bool) -> void:
	pass


func request_respawn(has_received_state: bool) -> void:
	if !has_received_state || connection_service == null || hud_flow == null:
		return
	if !hud_flow.can_request_respawn():
		return

	connection_service.send_respawn_request()
	mark_awaiting_confirmation()


func mark_awaiting_confirmation() -> void:
	awaiting_respawn_confirmation = true


func clear_awaiting_confirmation() -> void:
	awaiting_respawn_confirmation = false


func should_restore_alive_hud(state: Dictionary, player) -> bool:
	if !awaiting_respawn_confirmation:
		return false

	if state["has_lives"] && int(state["lives"]) <= 0:
		return false

	var server_players: Dictionary = state["server_players"]
	var self_id: String = state["self_id"]
	if !server_players.has(self_id):
		return false

	var self_state = server_players[self_id]
	var has_valid_server_state := false
	if self_state is Dictionary:
		var self_state_dictionary: Dictionary = self_state
		has_valid_server_state = !self_state_dictionary.is_empty()
	return (player != null && player.visible) || has_valid_server_state
