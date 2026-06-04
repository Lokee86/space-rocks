extends GutTest

const Constants := preload("res://scripts/generated/constants/constants.gd")
const VisualSyncPositions := preload("res://legacy/player_render/visual_sync_positions.gd")


func test_relative_to_local_visual_keeps_cross_edge_target_near_local_visual() -> void:
	assert_eq(
		VisualSyncPositions.relative_to_local_visual(
			Vector2(656.0, 320.0 - Constants.WORLD_HEIGHT),
			Vector2(656.0, 320.0),
			Vector2(676.0, Constants.WORLD_HEIGHT - 24.0)
		),
		Vector2(676.0, -24.0 - Constants.WORLD_HEIGHT)
	)


func test_world_copy_mismatch_detects_half_world_horizontal_jump() -> void:
	assert_true(
		VisualSyncPositions.is_world_copy_mismatch(
			Vector2(12.0, 200.0),
			Vector2(Constants.WORLD_WIDTH - 12.0, 200.0)
		)
	)


func test_world_copy_mismatch_ignores_nearby_positions() -> void:
	assert_false(
		VisualSyncPositions.is_world_copy_mismatch(
			Vector2(100.0, 200.0),
			Vector2(140.0, 220.0)
		)
	)

