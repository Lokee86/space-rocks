extends GutTest

const GAME_SHELL_SCENE := preload("res://scenes/game.tscn")
const MULTIPLAYER_LOBBY_SCENE := preload("res://scenes/ui/dialogs/multiplayer_lobby.tscn")
const Packets := preload("res://scripts/networking/packets.gd")


func test_player_rows_show_not_ready_ready_and_joining_statuses() -> void:
	var lobby := await _ready_lobby()

	lobby.set_members([
		{Packets.FIELD_MEMBER_ID: "new", Packets.FIELD_READY: false},
		{Packets.FIELD_MEMBER_ID: "ready", Packets.FIELD_READY: true, Packets.FIELD_CONNECTED: true},
		{Packets.FIELD_MEMBER_ID: "joining", Packets.FIELD_READY: false, Packets.FIELD_CONNECTED: false},
	])
	await get_tree().process_frame

	var rows := _player_rows(lobby)
	assert_eq(rows.size(), 3)
	_assert_row_status(rows[0], "Not Ready", false)
	_assert_row_status(rows[1], "Ready", true)
	_assert_row_status(rows[2], "Joining", false)


func test_start_game_button_tracks_all_connected_members_ready() -> void:
	var shell := await _ready_shell()
	shell.session_mode = 1
	shell._show_multiplayer_lobby()
	await get_tree().process_frame

	shell.handle_network_packet(_room_snapshot(true, false))
	await get_tree().process_frame
	assert_true(shell.multiplayer_lobby.start_game_button.disabled)

	shell.handle_network_packet(_room_snapshot(true, true))
	await get_tree().process_frame
	assert_false(shell.multiplayer_lobby.start_game_button.disabled)

	shell.handle_network_packet(_room_snapshot(true, false))
	await get_tree().process_frame
	assert_true(shell.multiplayer_lobby.start_game_button.disabled)


func _ready_lobby() -> Control:
	var lobby := MULTIPLAYER_LOBBY_SCENE.instantiate() as Control
	add_child_autofree(lobby)
	await get_tree().process_frame
	return lobby


func _ready_shell() -> Node:
	var shell := GAME_SHELL_SCENE.instantiate()
	add_child_autofree(shell)
	await get_tree().process_frame
	return shell


func _room_snapshot(local_ready: bool, remote_ready: bool) -> Dictionary:
	return {
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_CODE: "ABC123",
		Packets.FIELD_ROOM_STATE: "Lobby",
		Packets.FIELD_LOCAL_MEMBER_ID: "local",
		Packets.FIELD_MEMBERS: [
			{
				Packets.FIELD_MEMBER_ID: "local",
				Packets.FIELD_READY: local_ready,
				Packets.FIELD_CONNECTED: true,
			},
			{
				Packets.FIELD_MEMBER_ID: "remote",
				Packets.FIELD_READY: remote_ready,
				Packets.FIELD_CONNECTED: true,
			},
		],
		Packets.FIELD_MAX_PLAYERS: 8,
	}


func _player_rows(lobby: Control) -> Array:
	var container := lobby.find_child("PlayerListContainer", true, false) as Container
	return container.get_children()


func _assert_row_status(row: Node, status_text: String, ready: bool) -> void:
	var label := row.find_child("PlayerReadyLabel", true, false) as Label
	var ready_green := row.find_child("ReadyGreen", true, false) as CanvasItem
	var ready_red := row.find_child("ReadyRed", true, false) as CanvasItem

	assert_eq(label.text, status_text)
	assert_true(label.visible)
	assert_eq(ready_green.visible, ready)
	assert_eq(ready_red.visible, !ready)
