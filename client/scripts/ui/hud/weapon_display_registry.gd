extends RefCounted
class_name WeaponDisplayRegistry

const Constants = preload("res://scripts/generated/constants/constants.gd")
const TORPEDO_DISPLAY = preload("res://scenes/ui/weapon_displays/torpedo_display.tscn")

const BASIC_CANNON_ID := "basic_cannon"
const TORPEDO_ID := "torpedo"

static func definition_for_weapon(weapon_id: String) -> Dictionary:
	if weapon_id == TORPEDO_ID:
		return {
			"scene": TORPEDO_DISPLAY,
			"cooldown_total": Constants.TORPEDO_COOLDOWN,
		}
	return {}


static func is_displayable_weapon(weapon_id: String) -> bool:
	return not definition_for_weapon(weapon_id).is_empty()
