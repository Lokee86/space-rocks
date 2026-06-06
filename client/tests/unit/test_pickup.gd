extends GutTest

const PickupScript := preload("res://scripts/entities/pickup.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")


func test_apply_lifespan_state_stores_positive_values() -> void:
	var pickup = _create_pickup()

	pickup.apply_lifespan_state(2.5, 12.0)

	assert_eq(pickup.get("lifespan_age_seconds"), 2.5)
	assert_eq(pickup.get("lifespan_seconds"), 12.0)
	assert_true(pickup.get("has_lifespan_state"))


func test_pickup_outside_eol_window_stays_visible() -> void:
	var pickup = _create_pickup()
	pickup.apply_lifespan_state(2.0, 12.0)
	pickup.set("elapsed", 0.0)

	pickup._process(0.0)

	assert_true(pickup.sprite.visible)
	assert_true(pickup.glow_sprite.visible)


func test_pickup_inside_eol_window_can_hide_during_blink_cycle() -> void:
	var pickup = _create_pickup()
	pickup.apply_lifespan_state(11.7, 12.0)
	pickup.set("elapsed", 0.6)

	pickup._process(0.0)

	assert_false(pickup.sprite.visible)
	assert_false(pickup.glow_sprite.visible)


func _create_pickup():
	var pickup = PickupScript.new()
	var sprite := Sprite2D.new()
	sprite.name = "Sprite2D"
	var glow_sprite := Sprite2D.new()
	glow_sprite.name = "GlowSprite2D"
	pickup.add_child(sprite)
	pickup.add_child(glow_sprite)
	add_child(pickup)
	pickup._ready()
	return pickup
