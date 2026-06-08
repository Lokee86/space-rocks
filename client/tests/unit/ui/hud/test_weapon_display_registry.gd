extends GutTest

const WeaponDisplayRegistry = preload("res://scripts/ui/hud/weapon_display_registry.gd")


func test_basic_cannon_is_not_displayable() -> void:
	assert_false(WeaponDisplayRegistry.is_displayable_weapon("basic_cannon"))


func test_empty_string_is_not_displayable() -> void:
	assert_false(WeaponDisplayRegistry.is_displayable_weapon(""))


func test_unknown_weapon_is_not_displayable() -> void:
	assert_false(WeaponDisplayRegistry.is_displayable_weapon("unknown_weapon"))


func test_torpedo_is_displayable() -> void:
	assert_true(WeaponDisplayRegistry.is_displayable_weapon("torpedo"))


func test_torpedo_definition_has_scene() -> void:
	var definition := WeaponDisplayRegistry.definition_for_weapon("torpedo")

	assert_true(definition.has("scene"))
	assert_not_null(definition["scene"])
	assert_eq((definition["scene"] as PackedScene).resource_path, "res://scenes/ui/weapon_displays/weapon_display.tscn")


func test_torpedo_definition_has_positive_cooldown_total() -> void:
	var definition := WeaponDisplayRegistry.definition_for_weapon("torpedo")

	assert_true(definition.has("cooldown_total"))
	assert_gt(float(definition["cooldown_total"]), 0.0)
