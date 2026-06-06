extends RefCounted


static func player_shape_id(player_state: Dictionary) -> String:
	var ship_type = str(player_state.get("ship_type", ""))
	if ship_type == "":
		ship_type = "v_wing"
	return "player:" + ship_type


static func asteroid_shape_id(asteroid_state: Dictionary) -> String:
	return "asteroid:" + str(int(asteroid_state.get("variant", 0)))


static func bullet_shape_id(_bullet_state: Dictionary) -> String:
	return "bullet"


static func pickup_shape_id(pickup_state: Dictionary) -> String:
	return "pickup:" + str(pickup_state.get("type", ""))
