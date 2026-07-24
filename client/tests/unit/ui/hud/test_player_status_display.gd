extends GutTest

const PlayerStatusDisplayScene := preload("res://scenes/ui/elements/player_status_display.tscn")

var display: PlayerStatusDisplay


func before_each() -> void:
	display = PlayerStatusDisplayScene.instantiate() as PlayerStatusDisplay
	add_child_autofree(display)
	await get_tree().process_frame


func test_apply_status_formats_hull_shields_and_modules() -> void:
	display.apply_status({
		"health": 73,
		"max_health": 100,
		"shields": 18,
		"max_shields": 40,
		"shield_module_id": "shield_booster",
		"armor_module_id": "",
		"engine_module_id": "afterburner",
		"utility_module_id": "",
	})

	assert_eq(display.hull_label.text, "HULL: 73/100")
	assert_eq(display.shields_label.text, "SHIELDS: 18/40")
	assert_eq(display.modules_label.text, "MODULES: SHD:SHIELD BOOSTER  ENG:AFTERBURNER")


func test_clear_status_restores_neutral_readout() -> void:
	display.apply_status({"health": 1, "max_health": 10, "shields": 0, "max_shields": 0})
	display.clear_status()

	assert_eq(display.hull_label.text, "HULL: --")
	assert_eq(display.shields_label.text, "SHIELDS: --")
	assert_eq(display.modules_label.text, "MODULES: NONE")
