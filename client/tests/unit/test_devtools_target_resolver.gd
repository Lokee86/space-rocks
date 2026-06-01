extends GutTest

const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")


func test_resolve_explicit_player_target_wins() -> void:
	var resolved := DevtoolsTargetResolver.resolve("player-2", "player-3", "player-1")

	assert_eq(resolved, "player-2")


func test_resolve_game_target_uses_canonical_game_target() -> void:
	var resolved := DevtoolsTargetResolver.resolve(
		DevtoolsTargetResolver.TARGET_GAME,
		"player-3",
		"player-1"
	)

	assert_eq(resolved, "player-3")


func test_resolve_game_target_falls_back_to_local_player_when_canonical_empty() -> void:
	var resolved := DevtoolsTargetResolver.resolve(
		DevtoolsTargetResolver.TARGET_GAME,
		"",
		"player-1"
	)

	assert_eq(resolved, "player-1")


func test_resolve_empty_selected_falls_back_to_canonical_then_local() -> void:
	var resolved_with_canonical := DevtoolsTargetResolver.resolve("", "player-3", "player-1")
	assert_eq(resolved_with_canonical, "player-3")

	var resolved_with_local := DevtoolsTargetResolver.resolve("", "", "player-1")
	assert_eq(resolved_with_local, "player-1")


func test_resolve_empty_only_when_all_inputs_empty() -> void:
	var resolved := DevtoolsTargetResolver.resolve("", "", "")

	assert_eq(resolved, "")
