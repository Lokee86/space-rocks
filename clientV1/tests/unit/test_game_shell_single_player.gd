extends GutTest

const GAME_SHELL_SCENE := preload("res://scenes/game.tscn")
const NetworkClientScript := preload("res://scripts/networking/network_client.gd")
const Packets := preload("res://scripts/networking/packets.gd")


func test_single_player_ignores_non_ingame_room_states() -> void:
	var shell := await _ready_shell()
	shell.session_mode = 0
	_add_injected_network_client(shell)

	for state in ["Lobby", "Starting", "GameOver"]:
		shell.handle_network_packet(_room_snapshot(state))
		assert_null(shell.game_loop, "single-player should not enter gameplay for %s" % state)


func test_single_player_enters_gameplay_from_ingame_snapshot_without_ready() -> void:
	var shell := await _ready_shell()
	shell.session_mode = 0
	var client := _add_injected_network_client(shell)

	shell.handle_network_packet(_room_snapshot("InGame"))

	assert_not_null(shell.game_loop)
	assert_eq(shell.current_room_state, "InGame")
	assert_eq(shell.current_room_code, "ABC123")
	assert_eq(shell.lobby_network_client, null)
	assert_eq(shell.game_loop.injected_network_client, client)


func test_single_player_enters_gameplay_from_ingame_state_changed() -> void:
	var shell := await _ready_shell()
	shell.session_mode = 0
	var client := _add_injected_network_client(shell)

	shell.handle_network_packet({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_STATE_CHANGED,
		Packets.FIELD_ROOM_CODE: "ABC123",
		Packets.FIELD_ROOM_STATE: "InGame",
	})

	assert_not_null(shell.game_loop)
	assert_eq(shell.current_room_state, "InGame")
	assert_eq(shell.current_room_code, "ABC123")
	assert_eq(shell.lobby_network_client, null)
	assert_eq(shell.game_loop.injected_network_client, client)


func test_multiplayer_ingame_transition_still_uses_multiplayer_mode() -> void:
	var shell := await _ready_shell()
	shell.session_mode = 1
	var client := _add_injected_network_client(shell)

	shell.handle_network_packet(_room_snapshot("InGame"))

	assert_not_null(shell.game_loop)
	assert_eq(shell.game_loop.session_mode, "Multiplayer")
	assert_eq(shell.game_loop.injected_network_client, client)


func _ready_shell() -> Node:
	var shell := GAME_SHELL_SCENE.instantiate()
	add_child_autofree(shell)
	await get_tree().process_frame
	return shell


func _add_injected_network_client(shell: Node) -> NetworkClient:
	var client := NetworkClientScript.new()
	shell.add_child(client)
	shell.lobby_network_client = client
	return client


func _room_snapshot(state: String) -> Dictionary:
	return {
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_CODE: "ABC123",
		Packets.FIELD_ROOM_STATE: state,
		Packets.FIELD_LOCAL_MEMBER_ID: "session-1",
		Packets.FIELD_MEMBERS: [
			{
				Packets.FIELD_MEMBER_ID: "session-1",
				Packets.FIELD_READY: false,
				Packets.FIELD_CONNECTED: true,
			},
		],
		Packets.FIELD_MAX_PLAYERS: 8,
	}
