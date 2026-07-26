extends GutTest

const GameplaySpectateFlowScript := preload("res://scripts/gameplay/spectate/gameplay_spectate_flow.gd")
const SpectateMenuStateScript := preload("res://scripts/gameplay/spectate/spectate_menu_state.gd")


class FakeWorldSync:
	extends RefCounted

	var selected_targets: Array[String] = []
	var focused_targets: Array[String] = []

	func set_view_target_player(player_id: String) -> void:
		selected_targets.append(player_id)

	func focus_camera_on_player(player_id: String) -> bool:
		focused_targets.append(player_id)
		return true


func test_begin_and_cycle_spectating_updates_view_anchor_target() -> void:
	var menu_state := SpectateMenuStateScript.new()
	menu_state.apply_gameplay_state({
		"overlay": {"self_id": "player-1"},
		"world": {"ships": {"player-2": {}, "player-3": {}}},
		"session": {
			"player_lifecycle": {
				"player-2": "active",
				"player-3": "active",
			}
		},
	})
	var world_sync := FakeWorldSync.new()
	var flow := GameplaySpectateFlowScript.new()
	flow.configure(null, menu_state, world_sync)

	flow.begin_spectating()

	assert_true(flow.is_spectating)
	assert_eq(world_sync.selected_targets, ["player-2"])
	assert_eq(world_sync.focused_targets, ["player-2"])

	flow.request_cycle_target()

	assert_eq(world_sync.selected_targets, ["player-2", "player-3"])
	assert_eq(world_sync.focused_targets, ["player-2", "player-3"])
