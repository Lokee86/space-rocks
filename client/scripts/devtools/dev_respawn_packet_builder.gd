extends RefCounted
class_name DevRespawnPacketBuilder


const TYPE_DEBUG_RESPAWN_PLAYER := "debug_respawn_player"
const FIELD_TYPE := "type"
const FIELD_TARGET_PLAYER_ID := "target_player_id"
const FIELD_X := "x"
const FIELD_Y := "y"


static func build_from_placement_result(result: Dictionary) -> Dictionary:
	var target_player_id := str(result.get("target_player_id", ""))
	if target_player_id == "":
		return {}

	var server_position: Vector2 = result.get("server_position", Vector2.ZERO)
	return {
		FIELD_TYPE: TYPE_DEBUG_RESPAWN_PLAYER,
		FIELD_TARGET_PLAYER_ID: target_player_id,
		FIELD_X: server_position.x,
		FIELD_Y: server_position.y
	}
