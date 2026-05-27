extends GutTest

const Constants := preload("res://scripts/constants/constants.gd")
const LocalVisualSyncScript := preload("res://scripts/world/local_visual_sync.gd")


func test_first_update_initializes_server_and_visual_positions() -> void:
	var sync := LocalVisualSyncScript.new()

	sync.update_from_server_position(Vector2(100.0, 120.0))

	assert_true(sync.is_initialized())
	assert_eq(sync.server_position(), Vector2(100.0, 120.0))
	assert_eq(sync.visual_position(), Vector2(100.0, 120.0))


func test_subsequent_wrapped_update_advances_visual_position_continuously() -> void:
	var sync := LocalVisualSyncScript.new()
	sync.update_from_server_position(Vector2(Constants.WORLD_WIDTH - 5.0, 100.0))

	sync.update_from_server_position(Vector2(5.0, 100.0))

	assert_eq(sync.server_position(), Vector2(5.0, 100.0))
	assert_eq(sync.visual_position(), Vector2(Constants.WORLD_WIDTH + 5.0, 100.0))


func test_visual_position_for_server_position_uses_local_visual_reference() -> void:
	var sync := LocalVisualSyncScript.new()
	sync.update_from_server_position(Vector2(Constants.WORLD_WIDTH - 5.0, 100.0))

	assert_eq(
		sync.visual_position_for_server_position(Vector2(5.0, 100.0)),
		Vector2(Constants.WORLD_WIDTH + 5.0, 100.0)
	)


func test_visual_position_for_server_position_returns_server_position_before_init() -> void:
	var sync := LocalVisualSyncScript.new()

	assert_eq(
		sync.visual_position_for_server_position(Vector2(5.0, 100.0)),
		Vector2(5.0, 100.0)
	)
