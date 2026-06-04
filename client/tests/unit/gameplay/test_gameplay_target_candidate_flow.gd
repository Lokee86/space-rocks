extends GutTest

const GameplayTargetCandidateFlow = preload("res://scripts/gameplay/targeting/gameplay_target_candidate_flow.gd")


class FakeTargetPositionSource:
	var player_position_entries := {}
	var asteroid_position_entries := {}
	var bullet_position_entries := {}

	func player_positions() -> Dictionary:
		return player_position_entries

	func asteroid_positions() -> Dictionary:
		return asteroid_position_entries

	func bullet_positions() -> Dictionary:
		return bullet_position_entries


func _candidate_map(candidates: Array) -> Dictionary:
	var result := {}
	for candidate in candidates:
		result[candidate.target_kind + ":" + candidate.target_id] = candidate
	return result


func test_target_visual_candidates_returns_empty_array_when_target_position_source_is_null() -> void:
	var flow := GameplayTargetCandidateFlow.new()

	assert_eq(flow.target_visual_candidates(), [])


func test_target_visual_candidates_builds_player_asteroid_and_bullet_candidates() -> void:
	var target_position_source := FakeTargetPositionSource.new()
	target_position_source.player_position_entries = {
		"player-1": {
			"visual_position": Vector2(10, 20),
			"server_position": Vector2(30, 40),
		}
	}
	target_position_source.asteroid_position_entries = {
		"asteroid-1": {
			"visual_position": Vector2(50, 60),
			"server_position": Vector2(70, 80),
			"visual_scale": 1.5,
		}
	}
	target_position_source.bullet_position_entries = {
		"bullet-1": {
			"visual_position": Vector2(90, 100),
			"server_position": Vector2(110, 120),
		}
	}

	var flow := GameplayTargetCandidateFlow.new()
	flow.configure(target_position_source)

	var candidates := flow.target_visual_candidates()
	assert_eq(candidates.size(), 3)

	var by_key := _candidate_map(candidates)
	assert_true(by_key.has("player:player-1"))
	assert_true(by_key.has("asteroid:asteroid-1"))
	assert_true(by_key.has("bullet:bullet-1"))

	var player_candidate = by_key["player:player-1"]
	assert_eq(player_candidate.target_kind, "player")
	assert_eq(player_candidate.target_id, "player-1")
	assert_eq(player_candidate.visual_position, Vector2(10, 20))
	assert_eq(player_candidate.server_position, Vector2(30, 40))
	assert_eq(player_candidate.pick_radius, 32.0)

	var asteroid_candidate = by_key["asteroid:asteroid-1"]
	assert_eq(asteroid_candidate.target_kind, "asteroid")
	assert_eq(asteroid_candidate.target_id, "asteroid-1")
	assert_eq(asteroid_candidate.visual_position, Vector2(50, 60))
	assert_eq(asteroid_candidate.server_position, Vector2(70, 80))
	assert_eq(asteroid_candidate.pick_radius, 48.0)

	var bullet_candidate = by_key["bullet:bullet-1"]
	assert_eq(bullet_candidate.target_kind, "bullet")
	assert_eq(bullet_candidate.target_id, "bullet-1")
	assert_eq(bullet_candidate.visual_position, Vector2(90, 100))
	assert_eq(bullet_candidate.server_position, Vector2(110, 120))
	assert_eq(bullet_candidate.pick_radius, 12.0)


func test_target_visual_candidates_skips_malformed_entries() -> void:
	var target_position_source := FakeTargetPositionSource.new()
	target_position_source.player_position_entries = {
		"missing_visual": {
			"server_position": Vector2(1, 2),
		},
		"missing_server": {
			"visual_position": Vector2(3, 4),
		},
	}
	target_position_source.asteroid_position_entries = {
		"asteroid-ok": {
			"visual_position": Vector2(5, 6),
			"server_position": Vector2(7, 8),
		},
		"asteroid-bad": {
			"visual_position": Vector2(9, 10),
		},
	}
	target_position_source.bullet_position_entries = {
		"bullet-bad": {
			"server_position": Vector2(11, 12),
		},
	}

	var flow := GameplayTargetCandidateFlow.new()
	flow.configure(target_position_source)

	var candidates := flow.target_visual_candidates()
	assert_eq(candidates.size(), 1)
	assert_eq(candidates[0].target_kind, "asteroid")
	assert_eq(candidates[0].target_id, "asteroid-ok")
