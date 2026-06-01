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


func test_empty_canonical_target_falls_back_to_local_player() -> void:
	var controller := DevtoolsWindowController.new()
	var connection := FakeConnectionService.new()
	controller.configure_kill_player_routing(
		connection,
		"player-1",
		DevtoolsTargetResolver.TARGET_KIND_PLAYER,
		""
	)

	controller._on_kill_player_requested(DevtoolsTargetResolver.TARGET_GAME)

	assert_eq(connection.kill_self_calls, 1)
	assert_eq(connection.kill_target_calls, 0)
