extends GutTest

const AsteroidVariants := preload("res://scripts/generated/asteroids/asteroid_variants.gd")


func test_count_returns_eight() -> void:
	assert_eq(AsteroidVariants.count(), 8)


func test_texture_path_for_index_zero_uses_first_variant_texture() -> void:
	assert_true(AsteroidVariants.texture_path_for_index(0).ends_with("asteroid1.png"))


func test_texture_path_for_index_seven_uses_last_variant_texture() -> void:
	assert_true(AsteroidVariants.texture_path_for_index(7).ends_with("asteroid8.png"))


func test_texture_path_for_wrapped_index_eight_uses_first_variant_texture() -> void:
	assert_true(AsteroidVariants.texture_path_for_index(8).ends_with("asteroid1.png"))


func test_collision_shape_for_current_variants_is_asteroid_zero() -> void:
	for index in range(AsteroidVariants.count()):
		assert_eq(AsteroidVariants.collision_shape_for_index(index), "asteroid:0")


func test_current_variant_spawn_weights_are_one() -> void:
	for index in range(AsteroidVariants.count()):
		assert_eq(AsteroidVariants.timed_spawn_weight_for_index(index), 1.0)
		assert_eq(AsteroidVariants.fragment_spawn_weight_for_index(index), 1.0)
		assert_eq(AsteroidVariants.debug_spawn_weight_for_index(index), 1.0)
