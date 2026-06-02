extends RefCounted
class_name DevRespawnPacketBuilder


const TYPE_DEBUG_RESPAWN_PLAYER := "debug_respawn_player"
const FIELD_TYPE := "type"
const FIELD_TARGET_SCOPE := "target_scope"
const FIELD_TARGET_PLAYER_ID := "target_player_id"

static func build(target_scope: String, target_player_id: String) -> Dictionary:
	var packet := {
		FIELD_TYPE: TYPE_DEBUG_RESPAWN_PLAYER,
	}

	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS:
		packet[FIELD_TARGET_SCOPE] = target_scope
		return packet

	if target_player_id == "":
		return {}

	if target_scope != "":
		packet[FIELD_TARGET_SCOPE] = target_scope
	packet[FIELD_TARGET_PLAYER_ID] = target_player_id
	return packet
