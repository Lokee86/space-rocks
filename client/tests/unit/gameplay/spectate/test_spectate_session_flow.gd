extends GutTest

const SpectateSessionFlowScript := preload("res://scripts/gameplay/spectate/spectate_session_flow.gd")
const SpectateMenuStateScript := preload("res://scripts/gameplay/spectate/spectate_menu_state.gd")


class FakeMenuFlow:
	extends RefCounted

	var configured_state = null

	func configure_spectate_menu_state(state) -> void:
		configured_state = state


func test_apply_gameplay_state_reports_only_target_set_changes() -> void:
	var menu_flow := FakeMenuFlow.new()
	var menu_state := SpectateMenuStateScript.new()
	var flow := SpectateSessionFlowScript.new()
	flow.configure(menu_flow, null, menu_state)
	var initial_state := {
		"overlay": {"self_id": "player-1"},
		"world": {"ships": {"player-2": {}}},
		"session": {"player_lifecycle": {"player-2": "active"}},
	}

	assert_true(flow.apply_gameplay_state(initial_state))
	assert_false(flow.apply_gameplay_state(initial_state))

	var changed_state := initial_state.duplicate(true)
	changed_state["world"]["ships"]["player-3"] = {}
	changed_state["session"]["player_lifecycle"]["player-3"] = "active"
	assert_true(flow.apply_gameplay_state(changed_state))
