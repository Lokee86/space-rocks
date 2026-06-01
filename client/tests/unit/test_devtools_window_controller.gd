extends GutTest

const DevtoolsWindowController := preload("res://scripts/devtools/devtools_window_controller.gd")
const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")

class FakeConnectionService:
	extends RefCounted

	var kill_self_calls := 0
	var kill_target_calls := 0
	var last_target_player_id := ""

	func send_debug_kill_player_request() -> void:
		kill_self_calls += 1

	func send_debug_kill_target_player_request(target_player_id: String) -> void:
		kill_target_calls += 1
		last_target_player_id = target_player_id

class FakeEffectReceiver:
	extends RefCounted

	var invincible_calls := 0
	var infinite_lives_calls := 0
	var freeze_player_calls := 0
	var set_score_calls := 0
	var add_score_calls := 0
	var set_lives_calls := 0
	var add_lives_calls := 0
	var last_target_player_id := ""

	func on_toggle_invincible_requested(target_player_id: String) -> void:
		invincible_calls += 1
		last_target_player_id = target_player_id

	func on_toggle_infinite_lives_requested(target_player_id: String) -> void:
		infinite_lives_calls += 1
		last_target_player_id = target_player_id

	func on_toggle_freeze_player_requested(target_player_id: String) -> void:
		freeze_player_calls += 1
		last_target_player_id = target_player_id

	func on_set_score_requested(target_player_id: String, _score: int) -> void:
		set_score_calls += 1
		last_target_player_id = target_player_id

	func on_add_score_requested(target_player_id: String, _amount: int) -> void:
		add_score_calls += 1
		last_target_player_id = target_player_id

	func on_set_lives_requested(target_player_id: String, _lives: int) -> void:
		set_lives_calls += 1
		last_target_player_id = target_player_id

	func on_add_lives_requested(target_player_id: String, _amount: int) -> void:
		add_lives_calls += 1
		last_target_player_id = target_player_id


func test_effective_target_explicit_selected_player_wins() -> void:
	var controller := DevtoolsWindowController.new()
	controller.self_player_id = "player-1"
	controller.game_target_kind = DevtoolsTargetResolver.TARGET_KIND_PLAYER
	controller.game_target_id = "player-2"

	var resolved: String = controller._effective_target("player-3")

	assert_eq(resolved, "player-3")


func test_player_canonical_target_resolves_and_sends_player_only_command() -> void:
	var controller := DevtoolsWindowController.new()
	var connection := FakeConnectionService.new()
	controller.configure_kill_player_routing(
		connection,
		"player-1",
		DevtoolsTargetResolver.TARGET_KIND_PLAYER,
		"player-2"
	)

	controller._on_kill_player_requested(DevtoolsTargetResolver.TARGET_GAME)

	assert_eq(connection.kill_target_calls, 1)
	assert_eq(connection.last_target_player_id, "player-2")
	assert_ne(connection.last_target_player_id, "player-1")
	assert_ne(connection.last_target_player_id, DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(connection.kill_self_calls, 0)


func test_asteroid_canonical_target_does_not_send_player_only_command() -> void:
	var controller := DevtoolsWindowController.new()
	var connection := FakeConnectionService.new()
	controller.configure_kill_player_routing(
		connection,
		"player-1",
		DevtoolsTargetResolver.TARGET_KIND_ASTEROID,
		"asteroid-1"
	)

	controller._on_kill_player_requested(DevtoolsTargetResolver.TARGET_GAME)

	assert_eq(connection.kill_target_calls, 0)
	assert_eq(connection.kill_self_calls, 0)


func test_bullet_canonical_target_does_not_send_player_only_command() -> void:
	var controller := DevtoolsWindowController.new()
	var connection := FakeConnectionService.new()
	controller.configure_kill_player_routing(
		connection,
		"player-1",
		DevtoolsTargetResolver.TARGET_KIND_BULLET,
		"bullet-1"
	)

	controller._on_kill_player_requested(DevtoolsTargetResolver.TARGET_GAME)

	assert_eq(connection.kill_target_calls, 0)
	assert_eq(connection.kill_self_calls, 0)


func test_explicit_game_target_with_empty_target_sends_no_kill_request() -> void:
	var controller := DevtoolsWindowController.new()
	var connection := FakeConnectionService.new()
	controller.configure_kill_player_routing(
		connection,
		"player-1",
		DevtoolsTargetResolver.TARGET_KIND_PLAYER,
		""
	)

	controller._on_kill_player_requested(DevtoolsTargetResolver.TARGET_GAME)

	assert_eq(connection.kill_self_calls, 0)
	assert_eq(connection.kill_target_calls, 0)


func test_explicit_game_target_without_active_player_target_sends_no_kill_request() -> void:
	var controller := DevtoolsWindowController.new()
	var connection := FakeConnectionService.new()
	controller.configure_kill_player_routing(
		connection,
		"player-1",
		"",
		""
	)

	controller._on_kill_player_requested(DevtoolsTargetResolver.TARGET_GAME)

	assert_eq(connection.kill_self_calls, 0)
	assert_eq(connection.kill_target_calls, 0)


func test_game_target_player_routes_player_only_effect_signals_to_game_target_player() -> void:
	var controller := DevtoolsWindowController.new()
	controller.configure_kill_player_routing(
		null,
		"player-1",
		DevtoolsTargetResolver.TARGET_KIND_PLAYER,
		"player-2"
	)
	var receiver := FakeEffectReceiver.new()
	controller.toggle_invincible_requested.connect(Callable(receiver, "on_toggle_invincible_requested"))
	controller.toggle_infinite_lives_requested.connect(Callable(receiver, "on_toggle_infinite_lives_requested"))
	controller.toggle_freeze_player_requested.connect(Callable(receiver, "on_toggle_freeze_player_requested"))
	controller.set_score_requested.connect(Callable(receiver, "on_set_score_requested"))
	controller.add_score_requested.connect(Callable(receiver, "on_add_score_requested"))
	controller.set_lives_requested.connect(Callable(receiver, "on_set_lives_requested"))
	controller.add_lives_requested.connect(Callable(receiver, "on_add_lives_requested"))

	controller._on_toggle_invincible_requested(DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(receiver.invincible_calls, 1)
	assert_eq(receiver.last_target_player_id, "player-2")

	controller._on_toggle_infinite_lives_requested(DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(receiver.infinite_lives_calls, 1)
	assert_eq(receiver.last_target_player_id, "player-2")

	controller._on_toggle_freeze_player_requested(DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(receiver.freeze_player_calls, 1)
	assert_eq(receiver.last_target_player_id, "player-2")

	controller._on_set_score_requested(DevtoolsTargetResolver.TARGET_GAME, 10)
	assert_eq(receiver.set_score_calls, 1)
	assert_eq(receiver.last_target_player_id, "player-2")

	controller._on_add_score_requested(DevtoolsTargetResolver.TARGET_GAME, 5)
	assert_eq(receiver.add_score_calls, 1)
	assert_eq(receiver.last_target_player_id, "player-2")

	controller._on_set_lives_requested(DevtoolsTargetResolver.TARGET_GAME, 3)
	assert_eq(receiver.set_lives_calls, 1)
	assert_eq(receiver.last_target_player_id, "player-2")

	controller._on_add_lives_requested(DevtoolsTargetResolver.TARGET_GAME, 1)
	assert_eq(receiver.add_lives_calls, 1)
	assert_eq(receiver.last_target_player_id, "player-2")


func test_game_target_asteroid_does_not_emit_player_only_effect_signals() -> void:
	var controller := DevtoolsWindowController.new()
	controller.configure_kill_player_routing(
		null,
		"player-1",
		DevtoolsTargetResolver.TARGET_KIND_ASTEROID,
		"asteroid-1"
	)
	var receiver := FakeEffectReceiver.new()
	controller.toggle_invincible_requested.connect(Callable(receiver, "on_toggle_invincible_requested"))
	controller.toggle_infinite_lives_requested.connect(Callable(receiver, "on_toggle_infinite_lives_requested"))
	controller.toggle_freeze_player_requested.connect(Callable(receiver, "on_toggle_freeze_player_requested"))
	controller.set_score_requested.connect(Callable(receiver, "on_set_score_requested"))
	controller.add_score_requested.connect(Callable(receiver, "on_add_score_requested"))
	controller.set_lives_requested.connect(Callable(receiver, "on_set_lives_requested"))
	controller.add_lives_requested.connect(Callable(receiver, "on_add_lives_requested"))

	controller._on_toggle_invincible_requested(DevtoolsTargetResolver.TARGET_GAME)
	controller._on_toggle_infinite_lives_requested(DevtoolsTargetResolver.TARGET_GAME)
	controller._on_toggle_freeze_player_requested(DevtoolsTargetResolver.TARGET_GAME)
	controller._on_set_score_requested(DevtoolsTargetResolver.TARGET_GAME, 10)
	controller._on_add_score_requested(DevtoolsTargetResolver.TARGET_GAME, 5)
	controller._on_set_lives_requested(DevtoolsTargetResolver.TARGET_GAME, 3)
	controller._on_add_lives_requested(DevtoolsTargetResolver.TARGET_GAME, 1)

	assert_eq(receiver.invincible_calls, 0)
	assert_eq(receiver.infinite_lives_calls, 0)
	assert_eq(receiver.freeze_player_calls, 0)
	assert_eq(receiver.set_score_calls, 0)
	assert_eq(receiver.add_score_calls, 0)
	assert_eq(receiver.set_lives_calls, 0)
	assert_eq(receiver.add_lives_calls, 0)
