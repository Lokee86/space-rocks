extends RefCounted
class_name GameplayInputFlow

const ClientLogger := preload("res://scripts/logging/logger.gd")

var connection_service
var player
var menu_flow
var _logged_input_packet := false


func configure(connection_service_ref, player_ref, menu_flow_ref) -> void:
	connection_service = connection_service_ref
	player = player_ref
	menu_flow = menu_flow_ref


func reset() -> void:
	_logged_input_packet = false


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
	if !_logged_input_packet:
		_logged_input_packet = true
		ClientLogger.network_event(
			ClientLogger.LEVEL_INFO,
			"first_input_packet_sent",
			"First input packet sent",
			{
				"packet_type": str(input_packet.get("type", "")),
				"forward": bool(input_packet.get("forward", false)),
				"back": bool(input_packet.get("back", false)),
				"left": bool(input_packet.get("left", false)),
				"right": bool(input_packet.get("right", false)),
				"primary_fire": bool(input_packet.get("primary_fire", false)),
				"secondary_fire": bool(input_packet.get("secondary_fire", false)),
			}
		)

	connection_service.send_input_packet(input_packet)
