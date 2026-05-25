extends GutTest

const SpectateCycleViewPolicy := preload("res://scripts/gameplay/spectate/spectate_cycle_view_policy.gd")


func test_cycle_view_unavailable_in_single_player() -> void:
	assert_false(
		SpectateCycleViewPolicy.is_cycle_view_available("SinglePlayer", "InGame", true, true)
	)


func test_cycle_view_unavailable_when_not_spectating() -> void:
	assert_false(
		SpectateCycleViewPolicy.is_cycle_view_available("Multiplayer", "InGame", true, false)
	)


func test_cycle_view_unavailable_when_room_is_game_over() -> void:
	assert_false(
		SpectateCycleViewPolicy.is_cycle_view_available("Multiplayer", "GameOver", true, true)
	)


func test_cycle_view_available_for_multiplayer_game_over_spectating_active_room() -> void:
	assert_true(
		SpectateCycleViewPolicy.is_cycle_view_available("Multiplayer", "InGame", true, true)
	)
