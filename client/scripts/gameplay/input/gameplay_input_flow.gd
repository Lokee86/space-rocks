extends RefCounted
class_name GameplayInputFlow

var connection_service
var player
var menu_flow
var has_received_gameplay_state := false


func configure(connection_service_ref, player_ref, menu_flow_ref) -> void:
	connection_service = connection_service_ref
	player = player_ref
	menu_flow = menu_flow_ref


func reset() -> void:
	has_received_gameplay_state = false


func mark_gameplay_state_received() -> void:
	has_received_gameplay_state = true


func process() -> void:
	if !has_received_gameplay_state:
		return
	if player == null:
		return
	if connection_service == null:
		return
	if menu_flow != null && menu_flow.is_gameplay_paused:
		return

	connection_service.send_input_packet(player.get_input_packet())
