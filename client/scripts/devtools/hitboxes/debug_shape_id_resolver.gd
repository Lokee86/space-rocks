extends RefCounted

const SHIP_SHAPE_BY_TYPE := {
	"v_wing": "v_wing",
	"v_wing_scout": "v_wing",
}


static func player_shape_id(player_state: Dictionary) -> String:
	var ship_type := str(player_state.get("ship_type", ""))
	var shape_id: String = SHIP_SHAPE_BY_TYPE.get(ship_type, "v_wing")
	return "player:" + shape_id


static func asteroid_shape_id(asteroid_state: Dictionary) -> String:
	return "asteroid:" + str(int(asteroid_state.get("variant", 0)))


static func bullet_shape_id(_bullet_state: Dictionary) -> String:
	return "bullet"


static func pickup_shape_id(pickup_state: Dictionary) -> String:
	return "pickup:" + str(pickup_state.get("type", ""))
