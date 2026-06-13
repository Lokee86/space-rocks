extends GutTest

const GameplayDeathFlow = preload("res://scripts/gameplay/events/gameplay_death_flow.gd")


class FakeHudFlow:
	var last_lives := -1
	var dead_calls := 0
	var game_over_calls := 0

	func apply_lives(lives) -> void:
		last_lives = lives

	func set_dead(respawn_delay) -> void:
		dead_calls += 1

	func set_game_over() -> void:
		game_over_calls += 1


class FakeMatchEndFlow:
	var calls := 0
	var last_event := {}

	func handle_local_player_eliminated(event: Dictionary) -> void:
		calls += 1
		last_event = event


func test_apply_self_death_event_keeps_respawn_behavior_for_lives_above_zero() -> void:
	var death_flow := GameplayDeathFlow.new()
	var hud_flow := FakeHudFlow.new()

	death_flow.configure(hud_flow, null, null)

	death_flow.apply_self_death_event({"lives": 2, "respawn_delay": 1.5})

	assert_eq(hud_flow.last_lives, 2)
	assert_eq(hud_flow.dead_calls, 1)
	assert_eq(hud_flow.game_over_calls, 0)


func test_apply_self_death_event_delegates_final_death_to_match_end_flow() -> void:
	var death_flow := GameplayDeathFlow.new()
	var hud_flow := FakeHudFlow.new()
	var match_end_flow := FakeMatchEndFlow.new()
	var event := {"lives": 0, "reason": "eliminated"}

	death_flow.configure(hud_flow, match_end_flow, null)

	death_flow.apply_self_death_event(event)

	assert_eq(hud_flow.last_lives, 0)
	assert_eq(hud_flow.game_over_calls, 0)
	assert_eq(match_end_flow.calls, 1)
	assert_eq(match_end_flow.last_event, event)
