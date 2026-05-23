extends Node2D

const Constants = preload("res://scripts/constants/constants.gd")
const GAME_LOOP_SCENE := preload("res://scenes/game_loop.tscn")
const MAIN_MENU_SCENE := preload("res://scenes/ui/main_menu.tscn")
const MULTIPLAYER_LOBBY_SCENE := preload("res://scenes/ui/dialogs/multiplayer_lobby.tscn")
const NetworkClientScript = preload("res://scripts/networking/network_client.gd")
const Packets = preload("res://scripts/networking/packets.gd")
const BACKGROUND_DRIFT := Vector2(18.0, 8.0)
const FOREGROUND_DRIFT := Vector2(42.0, 18.0)
const MULTIPLAYER_WS_URL := "ws://localhost:8080/ws"
const ROOM_ERROR_MESSAGES := {
	"room_not_found": "Room was not found.",
	"room_closed": "Room is closed.",
	"room_in_game": "Room is already in game.",
	"room_full": "Room is full.",
	"already_in_room": "Already in a room.",
	"not_in_room": "Not in a room.",
	"invalid_room_code": "Room code is invalid.",
	"not_ready": "Not all players are ready.",
	"invalid_room_state": "Room is not available.",
}

enum SessionMode {
	SINGLE_PLAYER,
	MULTIPLAYER,
}

@onready var background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground
@onready var canvas_layer: CanvasLayer = $CanvasLayer
@onready var main_menu: Control = $CanvasLayer/MainMenu

var game_loop: Node
var lobby_network_client: NetworkClient
var multiplayer_lobby: Control
var gameplay_scroll_offset := Vector2.ZERO
var drift_time := 0.0
var session_mode: SessionMode = SessionMode.SINGLE_PLAYER
var create_room_request_pending := false
var pending_join_room_code := ""
var latest_room_snapshot: Dictionary = {}
var current_room_code := ""
var current_room_state := ""
var local_room_member_id := ""
var room_members: Array = []
var room_ready_states: Dictionary = {}
var room_max_players := 0


func _ready() -> void:
	DisplayServer.window_set_min_size(Vector2i(Constants.WINDOW_MIN_SIZE))
	DisplayServer.window_set_max_size(Vector2i(Constants.WINDOW_MAX_SIZE))
	_clamp_window_size()
	_connect_main_menu()


func _process(delta: float) -> void:
	if lobby_network_client != null:
		lobby_network_client.poll()

	drift_time += delta
	_update_layer_shader(
		background,
		(BACKGROUND_DRIFT * drift_time) + (gameplay_scroll_offset * Constants.BACKGROUND_PARALLAX)
	)
	_update_layer_shader(
		foreground_background,
		Constants.FOREGROUND_BACKGROUND_OFFSET +
			(FOREGROUND_DRIFT * drift_time) +
			(gameplay_scroll_offset * Constants.FOREGROUND_BACKGROUND_PARALLAX)
	)


func _update_layer_shader(layer: TextureRect, scroll_offset: Vector2) -> void:
	if layer == null:
		return

	var background_material := layer.material as ShaderMaterial
	if background_material == null:
		return

	background_material.set_shader_parameter("scroll_offset", scroll_offset)


func _clamp_window_size() -> void:
	var current_size := DisplayServer.window_get_size()
	var clamped_size := Vector2i(
		clampi(current_size.x, int(Constants.WINDOW_MIN_SIZE.x), int(Constants.WINDOW_MAX_SIZE.x)),
		clampi(current_size.y, int(Constants.WINDOW_MIN_SIZE.y), int(Constants.WINDOW_MAX_SIZE.y))
	)
	if clamped_size != current_size:
		DisplayServer.window_set_size(clamped_size)


func _start_single_player() -> void:
	session_mode = SessionMode.SINGLE_PLAYER
	_start_game("")


func _start_multiplayer(room_id: String) -> void:
	session_mode = SessionMode.MULTIPLAYER
	_start_game(room_id)


func _create_multiplayer_room() -> void:
	print("[game_shell] create room requested")
	session_mode = SessionMode.MULTIPLAYER
	_show_multiplayer_lobby()
	_begin_create_room_flow()


func _join_multiplayer_room(room_code: String) -> void:
	print("[game_shell] join requested room_code=", room_code)
	session_mode = SessionMode.MULTIPLAYER
	_show_multiplayer_lobby()
	_begin_join_room_flow(room_code)


func _start_game(room_id: String) -> void:
	if game_loop != null:
		return

	game_loop = GAME_LOOP_SCENE.instantiate()
	if game_loop.has_method("set_room_id"):
		game_loop.set_room_id(room_id)
	if game_loop.has_signal("return_to_menu_requested"):
		game_loop.return_to_menu_requested.connect(_return_to_main_menu)
	add_child(game_loop)

	if main_menu != null:
		main_menu.queue_free()
		main_menu = null


func _create_multiplayer_game_loop() -> void:
	if game_loop != null:
		return

	game_loop = GAME_LOOP_SCENE.instantiate()
	if game_loop.has_method("set_room_id"):
		game_loop.set_room_id(current_room_code)
	if lobby_network_client != null && game_loop.has_method("set_network_client"):
		game_loop.set_network_client(lobby_network_client)
	if game_loop.has_signal("return_to_menu_requested"):
		game_loop.return_to_menu_requested.connect(_return_to_main_menu)
	add_child(game_loop)


func _return_to_main_menu() -> void:
	print("[game_shell] returning to main menu")
	clear_gameplay_scroll_offset()
	if game_loop != null:
		game_loop.queue_free()
		game_loop = null
	_hide_multiplayer_lobby()

	if main_menu == null:
		main_menu = MAIN_MENU_SCENE.instantiate()
		canvas_layer.add_child(main_menu)
		_connect_main_menu()


func _connect_main_menu() -> void:
	if main_menu != null && main_menu.has_signal("single_player_pressed"):
		if !main_menu.single_player_pressed.is_connected(_start_single_player):
			main_menu.single_player_pressed.connect(_start_single_player)
	if main_menu != null && main_menu.has_signal("multiplayer_create_requested"):
		if !main_menu.multiplayer_create_requested.is_connected(_create_multiplayer_room):
			main_menu.multiplayer_create_requested.connect(_create_multiplayer_room)
	if main_menu != null && main_menu.has_signal("multiplayer_join_requested"):
		if !main_menu.multiplayer_join_requested.is_connected(_join_multiplayer_room):
			main_menu.multiplayer_join_requested.connect(_join_multiplayer_room)
	if main_menu != null && main_menu.has_signal("multiplayer_room_requested"):
		if !main_menu.multiplayer_room_requested.is_connected(_start_multiplayer):
			main_menu.multiplayer_room_requested.connect(_start_multiplayer)


func _show_multiplayer_lobby() -> void:
	if multiplayer_lobby != null && is_instance_valid(multiplayer_lobby):
		return

	multiplayer_lobby = MULTIPLAYER_LOBBY_SCENE.instantiate() as Control
	if multiplayer_lobby.has_method("set_fake_data_enabled"):
		multiplayer_lobby.set_fake_data_enabled(false)
	canvas_layer.add_child(multiplayer_lobby)

	if multiplayer_lobby.has_signal("ready_requested"):
		if !multiplayer_lobby.ready_requested.is_connected(_request_ready_toggle):
			multiplayer_lobby.ready_requested.connect(_request_ready_toggle)

	if multiplayer_lobby.has_signal("start_game_requested"):
		if !multiplayer_lobby.start_game_requested.is_connected(_request_start_game):
			multiplayer_lobby.start_game_requested.connect(_request_start_game)

	if multiplayer_lobby.has_signal("leave_requested"):
		if !multiplayer_lobby.leave_requested.is_connected(_request_leave_room):
			multiplayer_lobby.leave_requested.connect(_request_leave_room)
	else:
		push_error("Multiplayer lobby is missing leave_requested signal.")

	if multiplayer_lobby.has_method("set_status"):
		multiplayer_lobby.set_status("Connecting...")

	if main_menu != null:
		main_menu.queue_free()
		main_menu = null


func _hide_multiplayer_lobby() -> void:
	if multiplayer_lobby == null:
		return
	if is_instance_valid(multiplayer_lobby):
		multiplayer_lobby.queue_free()
	multiplayer_lobby = null


func handle_network_packet(data: Dictionary) -> bool:
	if data.get(Packets.FIELD_TYPE, "") == Packets.TYPE_ROOM_ERROR:
		_handle_room_error_packet(data)
		return true
	if data.get(Packets.FIELD_TYPE, "") == Packets.TYPE_ROOM_SNAPSHOT:
		_store_room_snapshot(data)
		return true

	return false


func _begin_create_room_flow() -> void:
	create_room_request_pending = true
	pending_join_room_code = ""
	_set_lobby_status("Connecting...")
	_ensure_lobby_network_client()

	if lobby_network_client.is_connected_to_server():
		_send_pending_create_room_request()
		return

	var err := lobby_network_client.connect_to_server(MULTIPLAYER_WS_URL)
	if err != OK:
		create_room_request_pending = false
		_set_lobby_status("Could not connect to server.")


func _begin_join_room_flow(room_code: String) -> void:
	print("[game_shell] begin join flow room_code=", room_code)
	pending_join_room_code = room_code.strip_edges()
	create_room_request_pending = false
	_set_lobby_status("Connecting...")
	_ensure_lobby_network_client()

	if pending_join_room_code == "":
		_set_lobby_status("Enter a room code to join.")
		return

	if lobby_network_client.is_connected_to_server():
		_send_pending_join_room_request()
		return

	var err := lobby_network_client.connect_to_server(MULTIPLAYER_WS_URL)
	if err != OK:
		pending_join_room_code = ""
		_set_lobby_status("Could not connect to server.")


func _ensure_lobby_network_client() -> void:
	if lobby_network_client != null:
		return

	lobby_network_client = NetworkClientScript.new()
	add_child(lobby_network_client)
	lobby_network_client.connected_to_server.connect(_on_lobby_network_connected)
	lobby_network_client.connection_closed.connect(_on_lobby_network_closed)
	lobby_network_client.packet_received.connect(_on_lobby_network_packet_received)
	lobby_network_client.packet_parse_failed.connect(_on_lobby_network_packet_parse_failed)


func _on_lobby_network_connected() -> void:
	print("[multiplayer] websocket connected/open")
	_send_pending_create_room_request()
	_send_pending_join_room_request()


func _on_lobby_network_closed() -> void:
	create_room_request_pending = false
	pending_join_room_code = ""
	_set_lobby_status("Connection closed.")


func _on_lobby_network_packet_received(data: Dictionary) -> void:
	handle_network_packet(data)


func _on_lobby_network_packet_parse_failed(_text: String) -> void:
	_set_lobby_status("Received invalid server data.")


func _send_pending_create_room_request() -> void:
	if !create_room_request_pending:
		return
	create_room_request_pending = false
	print("[multiplayer] CreateRoomRequest sent")
	lobby_network_client.send_create_room_request()


func _send_pending_join_room_request() -> void:
	if pending_join_room_code == "":
		return
	var room_code := pending_join_room_code
	pending_join_room_code = ""
	print("[multiplayer] JoinRoomRequest sent room_code=", room_code)
	lobby_network_client.send_join_room_request(room_code)


func _request_ready_toggle() -> void:
	print("[game_shell] ready requested")
	if lobby_network_client == null || !lobby_network_client.is_connected_to_server():
		_set_lobby_status("Not connected to server.")
		return

	lobby_network_client.send_set_ready_request(!_local_member_ready())


func _request_start_game() -> void:
	print("[game_shell] start game requested")
	if lobby_network_client == null || !lobby_network_client.is_connected_to_server():
		_set_lobby_status("Not connected to server.")
		return
	if !_room_state_is_lobby() || !_local_member_ready():
		_set_lobby_status("Ready up before starting.")
		return

	lobby_network_client.send_start_game_request()


func _request_leave_room() -> void:
	print("[game_shell] leave requested")

	if lobby_network_client == null || !lobby_network_client.is_connected_to_server():
		print("[game_shell] leave requested without connected lobby network")
		_return_to_main_menu()
		return

	print("[game_shell] sending LeaveRoomRequest")
	if lobby_network_client.has_method("send_leave_room_request"):
		lobby_network_client.send_leave_room_request()
	else:
		push_error("NetworkClient is missing send_leave_room_request().")


func _handle_room_error_packet(data: Dictionary) -> void:
	var message := _room_error_message(data)
	if _show_multiplayer_dialog_error(message):
		return

	if multiplayer_lobby != null && is_instance_valid(multiplayer_lobby):
		if multiplayer_lobby.has_method("set_status"):
			multiplayer_lobby.set_status(message)


func _show_multiplayer_dialog_error(message: String) -> bool:
	if main_menu == null || !is_instance_valid(main_menu):
		return false
	if !main_menu.has_method("show_multiplayer_error"):
		return false

	return bool(main_menu.show_multiplayer_error(message))


func _room_error_message(data: Dictionary) -> String:
	var message := str(data.get(Packets.FIELD_MESSAGE, "")).strip_edges()
	if message != "":
		return message

	var error_code := str(data.get(Packets.FIELD_ERROR_CODE, "")).strip_edges()
	return str(ROOM_ERROR_MESSAGES.get(error_code, "Room request failed."))


func _set_lobby_status(text: String) -> void:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		return
	if multiplayer_lobby.has_method("set_status"):
		multiplayer_lobby.set_status(text)


func _store_room_snapshot(data: Dictionary) -> void:
	current_room_code = str(data.get(Packets.FIELD_ROOM_CODE, "")).strip_edges()
	current_room_state = str(data.get(Packets.FIELD_ROOM_STATE, "")).strip_edges()
	local_room_member_id = str(data.get(Packets.FIELD_LOCAL_MEMBER_ID, "")).strip_edges()
	room_members = _room_snapshot_members(data.get(Packets.FIELD_MEMBERS, []))
	room_ready_states = _room_snapshot_ready_states(room_members)
	room_max_players = int(data.get(Packets.FIELD_MAX_PLAYERS, 0))
	latest_room_snapshot = {
		Packets.FIELD_ROOM_CODE: current_room_code,
		Packets.FIELD_ROOM_STATE: current_room_state,
		Packets.FIELD_LOCAL_MEMBER_ID: local_room_member_id,
		Packets.FIELD_MEMBERS: room_members.duplicate(true),
		"ready_states": room_ready_states.duplicate(true),
		Packets.FIELD_MAX_PLAYERS: room_max_players,
	}
	_update_lobby_room_labels()
	_update_lobby_member_rows()
	_update_lobby_control_state()


func _room_snapshot_members(raw_members: Variant) -> Array:
	var stored_members := []
	if !(raw_members is Array):
		return stored_members

	for raw_member in raw_members:
		if raw_member is Dictionary:
			stored_members.append(raw_member.duplicate(true))
		else:
			stored_members.append(raw_member)

	return stored_members


func _room_snapshot_ready_states(members: Array) -> Dictionary:
	var ready_states := {}
	for member in members:
		if !(member is Dictionary):
			continue

		var member_id := str(member.get(Packets.FIELD_MEMBER_ID, member.get(Packets.FIELD_ID, ""))).strip_edges()
		if member_id == "":
			continue

		ready_states[member_id] = bool(member.get(Packets.FIELD_READY, false))

	return ready_states


func _update_lobby_room_labels() -> void:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		return

	if multiplayer_lobby.has_method("set_room_code"):
		multiplayer_lobby.set_room_code(current_room_code)
	if multiplayer_lobby.has_method("set_status"):
		multiplayer_lobby.set_status(_room_status_text())


func _update_lobby_member_rows() -> void:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		return
	if !multiplayer_lobby.has_method("set_members"):
		return

	multiplayer_lobby.set_members(room_members)


func _update_lobby_control_state() -> void:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		return

	var local_ready := _local_member_ready()
	if multiplayer_lobby.has_method("set_local_ready"):
		multiplayer_lobby.set_local_ready(local_ready)
	if multiplayer_lobby.has_method("set_start_enabled"):
		multiplayer_lobby.set_start_enabled(_room_state_is_lobby() && local_ready)


func _local_member_ready() -> bool:
	if local_room_member_id == "":
		return false

	return bool(room_ready_states.get(local_room_member_id, false))


func _room_state_is_lobby() -> bool:
	return current_room_state == "Lobby" || current_room_state == "lobby"


func _room_status_text() -> String:
	var status := current_room_state
	if status == "":
		status = "Unknown"

	if room_max_players > 0:
		return "%s (%d/%d)" % [status, room_members.size(), room_max_players]

	return status


func set_gameplay_scroll_offset(offset: Vector2) -> void:
	gameplay_scroll_offset = offset


func clear_gameplay_scroll_offset() -> void:
	gameplay_scroll_offset = Vector2.ZERO