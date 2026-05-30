extends RefCounted
class_name DevRespawnPacketBuilder


const TYPE_DEBUG_RESPAWN_PLAYER := "debug_respawn_player"
const FIELD_TYPE := "type"
const FIELD_TARGET_PLAYER_ID := "target_player_id"

static func build_for_target_player(target_player_id: String) -> Dictionary:
	if target_player_id == "":
		return {}

	return {
		FIELD_TYPE: TYPE_DEBUG_RESPAWN_PLAYER,
		FIELD_TARGET_PLAYER_ID: target_player_id,
	}
