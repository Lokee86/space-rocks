extends GutTest

const DebugShapeIdResolver := preload("res://scripts/devtools/hitboxes/debug_shape_id_resolver.gd")


func test_resolver_returns_player_shape_id() -> void:
	assert_eq(DebugShapeIdResolver.player_shape_id({"ship_type": ""}), "player:v_wing")
	assert_eq(DebugShapeIdResolver.player_shape_id({"ship_type": "v_wing"}), "player:v_wing")


func test_resolver_returns_asteroid_shape_id() -> void:
	assert_eq(DebugShapeIdResolver.asteroid_shape_id({"variant": 2}), "asteroid:2")


func test_resolver_returns_bullet_shape_id() -> void:
	assert_eq(DebugShapeIdResolver.bullet_shape_id({}), "bullet")


func test_resolver_returns_pickup_shape_id() -> void:
	assert_eq(DebugShapeIdResolver.pickup_shape_id({"type": "1_up"}), "pickup:1_up")
