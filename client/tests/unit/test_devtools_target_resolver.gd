extends GutTest

const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")


func test_resolve_explicit_player_target_wins() -> void:
	var resolved := DevtoolsTargetResolver.resolve_player_target(
		"player-2",
		DevtoolsTargetResolver.TARGET_KIND_PLAYER,
		"player-3",
		"player-1"
	)

	assert_eq(resolved, "player-2")


func test_resolve_game_target_uses_canonical_game_target() -> void:
	var resolved := DevtoolsTargetResolver.resolve_player_target(
		DevtoolsTargetResolver.TARGET_GAME,
		DevtoolsTargetResolver.TARGET_KIND_PLAYER,
		"player-3",
		"player-1"
	)

	assert_eq(resolved, "player-3")


func test_resolve_explicit_game_target_with_no_target_resolves_empty_string() -> void:
	var resolved := DevtoolsTargetResolver.resolve_player_target(
		DevtoolsTargetResolver.TARGET_GAME,
		"",
		"",
		"player-1"
	)

	assert_eq(resolved, "")


func test_resolve_game_target_returns_empty_for_asteroid_canonical_target() -> void:
	var resolved := DevtoolsTargetResolver.resolve_player_target(
		DevtoolsTargetResolver.TARGET_GAME,
		DevtoolsTargetResolver.TARGET_KIND_ASTEROID,
		"asteroid-1",
		"player-1"
	)

	assert_eq(resolved, "")


func test_resolve_game_target_returns_empty_for_bullet_canonical_target() -> void:
	var resolved := DevtoolsTargetResolver.resolve_player_target(
		DevtoolsTargetResolver.TARGET_GAME,
		DevtoolsTargetResolver.TARGET_KIND_BULLET,
		"bullet-1",
		"player-1"
	)

	assert_eq(resolved, "")


func test_resolve_game_target_returns_empty_for_enemy_canonical_target() -> void:
	var resolved := DevtoolsTargetResolver.resolve_player_target(
		DevtoolsTargetResolver.TARGET_GAME,
		DevtoolsTargetResolver.TARGET_KIND_ENEMY,
		"enemy-1",
		"player-1"
	)

	assert_eq(resolved, "")
