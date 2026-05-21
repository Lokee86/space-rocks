extends GutTest

const Constants := preload("res://scripts/constants/constants.gd")
const WorldWrap := preload("res://scripts/world_wrap.gd")


func test_wrap_position_wraps_right_to_left() -> void:
	assert_eq(
		WorldWrap.wrap_position(Vector2(Constants.WORLD_WIDTH + 5.0, 100.0)),
		Vector2(5.0, 100.0)
	)


func test_wrap_position_wraps_left_to_right() -> void:
	assert_eq(
		WorldWrap.wrap_position(Vector2(-5.0, 100.0)),
		Vector2(Constants.WORLD_WIDTH - 5.0, 100.0)
	)


func test_shortest_delta_crosses_horizontal_edge() -> void:
	assert_eq(
		WorldWrap.shortest_delta(Vector2(Constants.WORLD_WIDTH - 5.0, 100.0), Vector2(5.0, 100.0)),
		Vector2(10.0, 0.0)
	)


func test_shortest_delta_crosses_vertical_edge() -> void:
	assert_eq(
		WorldWrap.shortest_delta(Vector2(100.0, Constants.WORLD_HEIGHT - 3.0), Vector2(100.0, 3.0)),
		Vector2(0.0, 6.0)
	)


func test_visual_position_relative_to_keeps_cross_edge_target_near_reference() -> void:
	assert_eq(
		WorldWrap.visual_position_relative_to(Vector2(Constants.WORLD_WIDTH - 5.0, 100.0), Vector2(5.0, 100.0)),
		Vector2(Constants.WORLD_WIDTH + 5.0, 100.0)
	)
