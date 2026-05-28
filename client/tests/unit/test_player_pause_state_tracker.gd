extends GutTest

const PlayerPauseStateTracker := preload("res://scripts/gameplay/state/player_pause_state_tracker.gd")


func test_unknown_player_defaults_false() -> void:
	var tracker := PlayerPauseStateTracker.new()

	assert_false(tracker.is_paused("player-1"))


func test_apply_state_with_paused_true_marks_player_paused() -> void:
	var tracker := PlayerPauseStateTracker.new()

	tracker.apply_state({
		"player_id": "player-1",
		"paused": true,
	})

	assert_true(tracker.is_paused("player-1"))


func test_apply_state_with_paused_false_clears_player_pause() -> void:
	var tracker := PlayerPauseStateTracker.new()
	tracker.apply_state({
		"player_id": "player-1",
		"paused": true,
	})

	tracker.apply_state({
		"player_id": "player-1",
		"paused": false,
	})

	assert_false(tracker.is_paused("player-1"))


func test_reset_clears_tracked_state() -> void:
	var tracker := PlayerPauseStateTracker.new()
	tracker.apply_state({
		"player_id": "player-1",
		"paused": true,
	})

	tracker.reset()

	assert_false(tracker.is_paused("player-1"))


func test_empty_player_id_is_ignored() -> void:
	var tracker := PlayerPauseStateTracker.new()

	tracker.apply_state({
		"player_id": "",
		"paused": true,
	})

	assert_false(tracker.is_paused(""))
