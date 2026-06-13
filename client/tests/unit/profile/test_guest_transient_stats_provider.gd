extends GutTest

const GuestTransientStatsProvider := preload("res://scripts/profile/guest_transient_stats_provider.gd")


func test_load_stats_starts_zero() -> void:
	var provider := GuestTransientStatsProvider.new()

	assert_eq(provider.load_stats(), {
		"total_score": 0,
		"high_score": 0,
		"ship_deaths": 0,
		"games_played": 0,
		"wins": 0,
	})


func test_apply_match_result_updates_games_score_high_score_ship_deaths_and_wins() -> void:
	var provider := GuestTransientStatsProvider.new()

	var stats := provider.apply_match_result(125, 3, true)

	assert_eq(stats, {
		"total_score": 125,
		"high_score": 125,
		"ship_deaths": 3,
		"games_played": 1,
		"wins": 1,
	})


func test_apply_match_result_high_score_only_increases() -> void:
	var provider := GuestTransientStatsProvider.new()

	provider.apply_match_result(125, 3, true)
	var stats := provider.apply_match_result(80, 1, false)

	assert_eq(stats, {
		"total_score": 205,
		"high_score": 125,
		"ship_deaths": 4,
		"games_played": 2,
		"wins": 1,
	})


func test_load_stats_returns_copy_not_mutable_internal_reference() -> void:
	var provider := GuestTransientStatsProvider.new()
	var stats := provider.load_stats()

	stats["total_score"] = 999
	stats["wins"] = 7

	assert_eq(provider.load_stats(), {
		"total_score": 0,
		"high_score": 0,
		"ship_deaths": 0,
		"games_played": 0,
		"wins": 0,
	})


func test_clear_resets_stats() -> void:
	var provider := GuestTransientStatsProvider.new()

	provider.apply_match_result(125, 3, true)
	provider.clear()

	assert_eq(provider.load_stats(), {
		"total_score": 0,
		"high_score": 0,
		"ship_deaths": 0,
		"games_played": 0,
		"wins": 0,
	})
