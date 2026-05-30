extends RefCounted
class_name DevSpawnPacketBuilder


const TYPE_DEBUG_SPAWN_ENTITY := "debug_spawn_entity"
const FIELD_TYPE := "type"
const FIELD_ENTITY_TYPE := "entity_type"
const FIELD_X := "x"
const FIELD_Y := "y"
const FIELD_HAS_DIRECTION := "has_direction"
const FIELD_DIRECTION_X := "direction_x"
const FIELD_DIRECTION_Y := "direction_y"
const FIELD_TARGET_PLAYER_ID := "target_player_id"
const ENTITY_PLAYER := "player"
const ENTITY_ASTEROID := "asteroid"
const ENTITY_BULLET := "bullet"


static func entity_type_for_action(action_name) -> String:
	match String(action_name):
		"spawn_player":
			return ENTITY_PLAYER
		"spawn_asteroid":
			return ENTITY_ASTEROID
		"spawn_bullet":
			return ENTITY_BULLET
		_:
			return ""


static func build_from_placement_result(result: Dictionary) -> Dictionary:
	var action_name = result.get("action_name", StringName())
	var entity_type := entity_type_for_action(action_name)
	if entity_type == "":
		return {}

	var server_position: Vector2 = result.get("server_position", Vector2.ZERO)
	var direction: Vector2 = result.get("direction", Vector2.ZERO)
	var has_direction := bool(result.get("has_direction", false))
	var target_player_id := str(result.get("target_player_id", ""))

	var packet := {
		FIELD_TYPE: TYPE_DEBUG_SPAWN_ENTITY,
		FIELD_ENTITY_TYPE: entity_type,
		FIELD_X: server_position.x,
		FIELD_Y: server_position.y,
		FIELD_HAS_DIRECTION: has_direction,
		FIELD_DIRECTION_X: direction.x,
		FIELD_DIRECTION_Y: direction.y
	}
	if target_player_id != "":
		packet[FIELD_TARGET_PLAYER_ID] = target_player_id
	return packet
