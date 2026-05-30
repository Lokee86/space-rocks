extends RefCounted
class_name DevConnectionService

const DevSpawnPacketBuilder := preload("res://scripts/devtools/dev_spawn_packet_builder.gd")
const DevRespawnPacketBuilder := preload("res://scripts/devtools/dev_respawn_packet_builder.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

var connection_service


func configure(connection_service_ref) -> void:
	connection_service = connection_service_ref


func is_configured() -> bool:
	return connection_service != null


func send_spawn_from_placement_result(result: Dictionary) -> void:
	if connection_service == null:
		ClientLogger.game_warn("DevConnectionService: send spawn ignored, connection_service is null")
		return
	if result.is_empty():
		ClientLogger.game_warn("DevConnectionService: send spawn ignored, placement result is empty")
		return
	var packet: Dictionary = DevSpawnPacketBuilder.build_from_placement_result(result)
	if packet.is_empty():
		ClientLogger.game_warn("DevConnectionService: send spawn ignored, packet build returned empty")
		return
	if !connection_service.has_method("send_packet"):
		ClientLogger.game_warn("DevConnectionService: send spawn ignored, send_packet is unavailable")
		return
	connection_service.send_packet(packet)
	ClientLogger.game_info(
		"DevConnectionService: dev spawn packet sent entity_type=%s x=%s y=%s has_direction=%s"
		% [
			str(packet.get(DevSpawnPacketBuilder.FIELD_ENTITY_TYPE, "")),
			str(packet.get(DevSpawnPacketBuilder.FIELD_X, 0.0)),
			str(packet.get(DevSpawnPacketBuilder.FIELD_Y, 0.0)),
			str(packet.get(DevSpawnPacketBuilder.FIELD_HAS_DIRECTION, false))
		]
	)


func send_respawn_from_placement_result(result: Dictionary) -> void:
	if connection_service == null:
		ClientLogger.game_warn("DevConnectionService: send respawn ignored, connection_service is null")
		return
	var packet: Dictionary = DevRespawnPacketBuilder.build_from_placement_result(result)
	if packet.is_empty():
		ClientLogger.game_warn("DevConnectionService: send respawn ignored, packet build returned empty")
		return
	if !connection_service.has_method("send_packet"):
		ClientLogger.game_warn("DevConnectionService: send respawn ignored, send_packet is unavailable")
		return
	connection_service.send_packet(packet)
	ClientLogger.game_info(
		"DevConnectionService: dev respawn packet sent target_player_id=%s x=%s y=%s"
		% [
			str(packet.get(DevRespawnPacketBuilder.FIELD_TARGET_PLAYER_ID, "")),
			str(packet.get(DevRespawnPacketBuilder.FIELD_X, 0.0)),
			str(packet.get(DevRespawnPacketBuilder.FIELD_Y, 0.0))
		]
	)
