extends GutTest

const PickupScript := preload("res://scripts/entities/pickup.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")


func test_apply_lifespan_state_stores_positive_values() -> void:
	var pickup = _create_pickup()

	pickup.apply_lifespan_state(2.5, 12.0)

	assert_eq(pickup.get("lifespan_age_seconds"), 2.5)
	assert_eq(pickup.get("lifespan_seconds"), 12.0)
	assert_true(pickup.get("has_lifespan_state"))


func test_collision_radius_returns_circle_shape_radius() -> void:
	var pickup = _create_pickup(true)

	assert_eq(pickup.collision_radius(), 30.0)


func test_collision_radius_returns_zero_without_collision_shape() -> void:
	var pickup = _create_pickup(false)

	assert_eq(pickup.collision_radius(), 0.0)


func test_pickup_outside_eol_window_stays_visible() -> void:
	var pickup = _create_pickup()
	pickup.apply_lifespan_state(2.0, 12.0)
	pickup.set("elapsed", 0.0)

	pickup._process(0.0)

	assert_true(pickup.sprite.visible)
	assert_true(pickup.glow_sprite.visible)


func test_pickup_inside_eol_window_can_hide_during_blink_cycle() -> void:
	var pickup = _create_pickup()
	pickup.apply_lifespan_state(11.99, 12.0)
	pickup.set("elapsed", 0.0)

	pickup._process(0.05)

	assert_false(pickup.sprite.visible)
	assert_false(pickup.glow_sprite.visible)
	pickup.queue_free()


func _create_pickup(include_collision_shape := false):
	var pickup = PickupScript.new()
	var sprite := Sprite2D.new()
	sprite.name = "Sprite2D"
	var glow_sprite := Sprite2D.new()
	glow_sprite.name = "GlowSprite2D"
	pickup.add_child(sprite)
	pickup.add_child(glow_sprite)

	if include_collision_shape:
		var collision_shape := CollisionShape2D.new()
		collision_shape.name = "CollisionShape2D"
		collision_shape.shape = CircleShape2D.new()
		collision_shape.shape.radius = 30.0
		pickup.add_child(collision_shape)

	add_child(pickup)
	pickup._ready()
	return pickup
