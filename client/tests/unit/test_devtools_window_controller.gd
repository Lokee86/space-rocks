extends GutTest

const DevtoolsWindowController := preload("res://scripts/devtools/devtools_window_controller.gd")
const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")


func test_effective_target_explicit_selected_player_wins() -> void:
	var controller := DevtoolsWindowController.new()
	controller.self_player_id = "player-1"
	controller.latest_game_target_player_id = "player-2"

	var resolved: String = controller._effective_target("player-3")

	assert_eq(resolved, "player-3")


func test_effective_target_game_target_resolves_to_canonical_target() -> void:
	var controller := DevtoolsWindowController.new()
	controller.self_player_id = "player-1"
	controller.latest_game_target_player_id = "player-2"

	var resolved: String = controller._effective_target(DevtoolsTargetResolver.TARGET_GAME)

	assert_eq(resolved, "player-2")


func test_effective_target_empty_canonical_falls_back_to_local_player() -> void:
	var controller := DevtoolsWindowController.new()
	controller.self_player_id = "player-1"
	controller.latest_game_target_player_id = ""

	var resolved: String = controller._effective_target(DevtoolsTargetResolver.TARGET_GAME)

	assert_eq(resolved, "player-1")
