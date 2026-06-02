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


func test_resolve_player_target_scope_all_players_returns_all_players_scope_and_empty_player_id() -> void:
	var resolved := DevtoolsTargetResolver.resolve_player_target_scope(
		DevtoolsTargetResolver.TARGET_ALL_PLAYERS,
		DevtoolsTargetResolver.TARGET_KIND_PLAYER,
		"player-3",
		"player-1"
	)

	assert_eq(resolved["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS)
	assert_eq(resolved["target_player_id"], "")


func test_resolve_player_target_scope_explicit_player_returns_single_player_scope_and_target_id() -> void:
	var resolved := DevtoolsTargetResolver.resolve_player_target_scope(
		"player-2",
		DevtoolsTargetResolver.TARGET_KIND_PLAYER,
		"player-3",
		"player-1"
	)

	assert_eq(resolved["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER)
	assert_eq(resolved["target_player_id"], "player-2")


func test_resolve_player_target_scope_game_target_preserves_single_player_resolution() -> void:
	var resolved := DevtoolsTargetResolver.resolve_player_target_scope(
		DevtoolsTargetResolver.TARGET_GAME,
		DevtoolsTargetResolver.TARGET_KIND_PLAYER,
		"player-3",
		"player-1"
	)

	assert_eq(resolved["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER)
	assert_eq(resolved["target_player_id"], "player-3")


func test_resolve_player_target_scope_local_fallback_preserves_single_player_resolution() -> void:
	var resolved := DevtoolsTargetResolver.resolve_player_target_scope(
		"",
		"",
		"",
		"player-1"
	)

	assert_eq(resolved["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER)
	assert_eq(resolved["target_player_id"], "player-1")
