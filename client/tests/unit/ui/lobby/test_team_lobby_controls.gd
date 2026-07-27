extends GutTest

const LobbyScene := preload("res://scenes/ui/dialogs/multiplayer_lobby.tscn")


func test_owner_assigned_custom_room_allows_owner_to_edit_every_team() -> void:
	var lobby := await _create_lobby()
	lobby.apply_lobby_state(
		"ROOM", "lobby", "Player-1", "Player-1", 8,
		[
			{"player_id": "Player-1", "ready": false, "connected": true, "team_id": "team_1"},
			{"player_id": "Player-2", "ready": false, "connected": true, "team_id": "team_2"},
		],
		"custom", "owner_assigned", 0, false, false
	)

	var rows := (lobby.get_node("%PlayerListContainer") as Container).get_children()
	assert_eq(rows.size(), 2)
	assert_false(rows[0].team_selector.get_node("%TeamOptionButton").disabled)
	assert_false(rows[1].team_selector.get_node("%TeamOptionButton").disabled)


func test_player_selected_room_only_allows_local_assignment() -> void:
	var lobby := await _create_lobby()
	lobby.apply_lobby_state(
		"ROOM", "lobby", "Player-2", "Player-1", 8,
		[
			{"player_id": "Player-1", "ready": false, "connected": true, "team_id": "team_1"},
			{"player_id": "Player-2", "ready": false, "connected": true, "team_id": "team_2"},
		],
		"custom", "player_selected", 0, false, false
	)

	var rows := (lobby.get_node("%PlayerListContainer") as Container).get_children()
	assert_true(rows[0].team_selector.get_node("%TeamOptionButton").disabled)
	assert_false(rows[1].team_selector.get_node("%TeamOptionButton").disabled)


func test_ffa_hides_team_column_and_adaptive_roster_scrolls() -> void:
	var lobby := await _create_lobby()
	lobby.apply_lobby_state("ROOM", "lobby", "Player-1", "Player-1", 8, [{"player_id": "Player-1", "ready": true}], "ffa")

	assert_false((lobby.get_node("%TeamHeader") as Control).visible)
	assert_false((lobby.get_node("%PlayerListContainer").get_child(0).team_selector as Control).visible)
	var player_scroll := lobby.find_child("PlayerScroll", true, false) as ScrollContainer
	assert_not_null(player_scroll)
	assert_eq(player_scroll.horizontal_scroll_mode, ScrollContainer.SCROLL_MODE_AUTO)
	assert_true(lobby.find_child("Actions", true, false) is HFlowContainer)


func _create_lobby() -> Control:
	var lobby := LobbyScene.instantiate()
	add_child_autofree(lobby)
	await get_tree().process_frame
	return lobby
