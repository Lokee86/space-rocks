extends RefCounted
class_name GameplayInputFlow

var connection_service
var player
var menu_flow


func configure(connection_service_ref, player_ref, menu_flow_ref) -> void:
	connection_service = connection_service_ref
	player = player_ref
	menu_flow = menu_flow_ref


func reset() -> void:
	pass


func process(required_lane_baselines_synced: bool) -> void:
	if !required_lane_baselines_synced:
		return
	if player == null:
		return
	if connection_service == null:
		return
	if menu_flow != null && menu_flow.is_gameplay_paused:
		return

	var input_packet = player.get_input_packet()
	connection_service.send_input_packet(input_packet)
