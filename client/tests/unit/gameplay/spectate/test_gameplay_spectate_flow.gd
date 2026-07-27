extends GutTest

const GameplaySpectateFlowScript := preload("res://scripts/gameplay/spectate/gameplay_spectate_flow.gd")
const SpectateMenuStateScript := preload("res://scripts/gameplay/spectate/spectate_menu_state.gd")


class FakeWorldSync:
	extends RefCounted

	var selected_targets: Array[String] = []
	var focused_targets: Array[String] = []
	var cleared := false
	var focus_succeeds := true

	func set_view_target_player(player_id: String) -> void:
		selected_targets.append(player_id)

	func focus_camera_on_player(player_id: String) -> bool:
		focused_targets.append(player_id)
		return focus_succeeds

	func clear_view_target_player() -> void:
		cleared = true


class FakeConnectionService:
	extends RefCounted

	var packets: Array[Dictionary] = []

	func send_packet(packet: Dictionary, _trace_id: String = "") -> void:
		packets.append(packet.duplicate(true))


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


func test_spectate_target_requests_are_sent_and_missing_target_focus_retries() -> void:
	var menu_state := SpectateMenuStateScript.new()
	menu_state.apply_gameplay_state({
		"overlay": {"self_id": "player-1"},
		"world": {"ships": {"player-2": {}}},
		"session": {"player_lifecycle": {"player-2": "active"}},
	})
	var world_sync := FakeWorldSync.new()
	world_sync.focus_succeeds = false
	var connection := FakeConnectionService.new()
	var flow := GameplaySpectateFlowScript.new()
	flow.configure(null, menu_state, world_sync, connection)

	flow.begin_spectating()
	assert_true(flow.is_spectating)
	assert_eq(connection.packets.size(), 1)
	assert_eq(connection.packets[0].get("type"), "set_view_target_request")
	assert_eq(connection.packets[0].get("view_target_player_id"), "player-2")

	flow.process()
	assert_eq(world_sync.focused_targets, ["player-2", "player-2"])

	flow.reset()
	assert_eq(connection.packets.size(), 2)
	assert_eq(connection.packets[1].get("type"), "clear_view_target_request")
	assert_true(world_sync.cleared)
