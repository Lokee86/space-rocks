extends GutTest

const DevtoolsStateContext := preload("res://scripts/devtools/context/devtools_state_context.gd")


func test_has_lane_baseline_sync_tracks_true_and_false() -> void:
	var context := DevtoolsStateContext.new()

	context.set_has_lane_baseline_sync(true)
	assert_true(context.has_lane_baseline_sync())

	context.set_has_lane_baseline_sync(false)
	assert_false(context.has_lane_baseline_sync())


func test_set_local_player_id_stores_id() -> void:
	var context := DevtoolsStateContext.new()

	context.set_local_player_id("Player-1")

	assert_eq(context.get_local_player_id(), "Player-1")


func test_set_game_target_player_stores_kind_id_and_player_id() -> void:
	var context := DevtoolsStateContext.new()

	context.set_game_target("player", "Player-2")

	assert_eq(context.get_game_target_kind(), "player")
	assert_eq(context.get_game_target_id(), "Player-2")
	assert_eq(context.get_game_target_player_id(), "Player-2")


func test_set_game_target_non_player_clears_game_target_player_id() -> void:
	var context := DevtoolsStateContext.new()

	context.set_game_target("asteroid", "asteroid-1")

	assert_eq(context.get_game_target_kind(), "asteroid")
	assert_eq(context.get_game_target_id(), "asteroid-1")
	assert_eq(context.get_game_target_player_id(), "")


func test_reset_game_target_clears_target_fields() -> void:
	var context := DevtoolsStateContext.new()
	context.set_game_target("player", "Player-2")

	context.reset_game_target()

	assert_eq(context.get_game_target_kind(), "")
	assert_eq(context.get_game_target_id(), "")
	assert_eq(context.get_game_target_player_id(), "")
