extends GutTest

const WeaponDisplayScene = preload("res://scenes/ui/weapon_displays/weapon_display.tscn")


func test_apply_weapon_display_state_shows_torpedo_icon_child() -> void:
	var display := WeaponDisplayScene.instantiate() as WeaponDisplay
	add_child_autofree(display)

	display.apply_weapon_display_state({
		"weapon_id": "torpedo",
	})

	var weapon_icon := display.get_node("Sprite2D/WeaponIcon")
	var torpedo_icon := display.get_node("Sprite2D/WeaponIcon/torpedo") as CanvasItem

	assert_not_null(weapon_icon)
	assert_not_null(torpedo_icon)
	assert_true(torpedo_icon.visible)


func test_real_scene_exposes_typed_weapon_display_child_contract() -> void:
	var display := WeaponDisplayScene.instantiate() as WeaponDisplay
	add_child_autofree(display)

	assert_true(display is WeaponDisplay)
	assert_true(display.get_node("%AmmoLabel") is Label)
	assert_true(display.get_node("%RingHighlight") is RingHighlight)
	assert_true(display.get_node("%ReadySweepHighlight") is ReadySweepHighlight)
	assert_true(display.get_node("%ReadyFlash") is AnimatedSprite2D)
	assert_true(display.get_node("%CooldownOverlay") is CooldownOverlay)
	assert_true(display.get_node("Sprite2D") is Sprite2D)
	assert_true(display.get_node("Sprite2D/WeaponIcon") is Node2D)


func test_cooldown_overlay_finished_is_relayed_through_weapon_display() -> void:
	var display := WeaponDisplayScene.instantiate() as WeaponDisplay
	add_child_autofree(display)
	watch_signals(display)

	display.get_node("%CooldownOverlay").cooldown_finished.emit()

	assert_signal_emitted(display, "cooldown_finished")
