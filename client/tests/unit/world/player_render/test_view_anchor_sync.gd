extends GutTest

const ViewAnchorSyncScript := preload("res://scripts/world/player_render/view_anchor_sync.gd")


func test_update_and_conversions_use_anchor_position() -> void:
	var sync = ViewAnchorSyncScript.new()

	sync.update_from_anchor_server_position(Vector2(100, 200))

	assert_eq(sync.visual_position(), Vector2(100, 200))
	assert_eq(sync.server_position(), Vector2(100, 200))
	assert_eq(
		sync.visual_position_for_server_position(Vector2(120, 130)),
		Vector2(120, 130)
	)
	assert_eq(
		sync.server_position_for_visual_position(Vector2(120, 130)),
		Vector2(120, 130)
	)
