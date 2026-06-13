extends RefCounted
class_name GuestTransientStatsProvider

# Legacy compatibility only. Profile reads now come from the data-handler endpoint,
# so this provider is no longer the source of truth for profile stats.

var stats := {
	"total_score": 0,
	"high_score": 0,
	"ship_deaths": 0,
	"games_played": 0,
	"wins": 0,
}


func load_stats() -> Dictionary:
	return stats.duplicate(true)


func apply_match_result(score: int, ship_deaths: int, won: bool) -> Dictionary:
	stats["games_played"] += 1
	stats["total_score"] += score
	stats["high_score"] = max(stats["high_score"], score)
	stats["ship_deaths"] += ship_deaths
	if won:
		stats["wins"] += 1
	return load_stats()


func clear() -> void:
	stats = {
		"total_score": 0,
		"high_score": 0,
		"ship_deaths": 0,
		"games_played": 0,
		"wins": 0,
	}
