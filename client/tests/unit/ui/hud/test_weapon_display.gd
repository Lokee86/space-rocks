extends GutTest

const WeaponDisplayScene = preload("res://scenes/ui/weapon_displays/weapon_display.tscn")


func test_apply_weapon_display_state_shows_torpedo_icon_child() -> void:
	var display := WeaponDisplayScene.instantiate()
	add_child_autofree(display)

	display.apply_weapon_display_state({
		"weapon_id": "torpedo",
	})

	var weapon_icon := display.get_node("Sprite2D/WeaponIcon")
	var torpedo_icon := display.get_node("Sprite2D/WeaponIcon/torpedo") as CanvasItem

	assert_not_null(weapon_icon)
	assert_not_null(torpedo_icon)
	assert_true(torpedo_icon.visible)
