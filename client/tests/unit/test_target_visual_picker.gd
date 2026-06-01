extends GutTest

const TargetVisualCandidate = preload("res://scripts/gameplay/input/target_visual_candidate.gd")
const TargetVisualPicker = preload("res://scripts/gameplay/input/target_visual_picker.gd")

func _candidate(
	target_kind: String = "player",
	target_id: String = "target-1",
	visual_position: Vector2 = Vector2.ZERO,
	pick_radius: float = 10.0,
	visible: bool = true
) -> TargetVisualCandidate:
	var candidate := TargetVisualCandidate.new()
	candidate.target_kind = target_kind
	candidate.target_id = target_id
	candidate.visual_position = visual_position
	candidate.pick_radius = pick_radius
	candidate.visible = visible
	return candidate

func test_pick_empty_candidate_list_returns_null() -> void:
	var result = TargetVisualPicker.pick([], Vector2.ZERO)
	assert_null(result)

func test_pick_invisible_candidate_is_ignored() -> void:
	var result = TargetVisualPicker.pick([_candidate("player", "p1", Vector2.ZERO, 10.0, false)], Vector2.ZERO)
	assert_null(result)

func test_pick_invalid_candidate_is_ignored() -> void:
	var result = TargetVisualPicker.pick([_candidate("", "p1", Vector2.ZERO, 10.0, true)], Vector2.ZERO)
	assert_null(result)

func test_pick_out_of_radius_candidate_is_ignored() -> void:
	var result = TargetVisualPicker.pick([_candidate("player", "p1", Vector2.ZERO, 5.0, true)], Vector2(100, 100))
	assert_null(result)

func test_pick_higher_pick_rank_wins() -> void:
	var low_rank := _candidate("player", "low", Vector2.ZERO, 10.0, true)
	low_rank.pick_rank = 1
	var high_rank := _candidate("asteroid", "high", Vector2.ZERO, 10.0, true)
	high_rank.pick_rank = 2

	var result = TargetVisualPicker.pick([low_rank, high_rank], Vector2.ZERO)
	assert_eq(result, high_rank)

func test_pick_player_beats_asteroid_when_rank_ties() -> void:
	var asteroid := _candidate("asteroid", "asteroid-1", Vector2.ZERO, 10.0, true)
	asteroid.pick_rank = 5
	var player := _candidate("player", "player-1", Vector2.ZERO, 10.0, true)
	player.pick_rank = 5

	var result = TargetVisualPicker.pick([asteroid, player], Vector2.ZERO)
	assert_eq(result, player)

func test_pick_asteroid_beats_bullet_when_rank_ties() -> void:
	var bullet := _candidate("bullet", "bullet-1", Vector2.ZERO, 10.0, true)
	bullet.pick_rank = 3
	var asteroid := _candidate("asteroid", "asteroid-1", Vector2.ZERO, 10.0, true)
	asteroid.pick_rank = 3

	var result = TargetVisualPicker.pick([bullet, asteroid], Vector2.ZERO)
	assert_eq(result, asteroid)

func test_pick_lower_target_id_wins_when_rank_and_kind_tie() -> void:
	var later := _candidate("player", "target-b", Vector2.ZERO, 10.0, true)
	later.pick_rank = 7
	var earlier := _candidate("player", "target-a", Vector2.ZERO, 10.0, true)
	earlier.pick_rank = 7

	var result = TargetVisualPicker.pick([later, earlier], Vector2.ZERO)
	assert_eq(result, earlier)
