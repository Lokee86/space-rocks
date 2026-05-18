extends Node2D

signal return_to_menu_requested

const Constants = preload("res://scripts/constants.gd")
const EffectsScript = preload("res://scripts/effects.gd")
const HudControllerScript = preload("res://scripts/ui/hud_controller.gd")
const NetworkClientScript = preload("res://scripts/network_client.gd")
const Packets = preload("res://scripts/packets.gd")
const WorldSyncScript = preload("res://scripts/world_sync.gd")

@onready var player: Player = $Player
@onready var bullets = $Bullets
@onready var asteroids: Node2D = $Asteroids

var respawn_requested := false
var has_received_state := false
var has_initial_spawn := false
var self_id := ""
var effects: Effects
var hud_controller: HudController
var network_client: NetworkClient
var room_id := ""
var world_sync: WorldSync


func set_room_id(value: String) -> void:
	room_id = value.strip_edges()

func _ready() -> void:
	network_client = NetworkClientScript.new()
	add_child(network_client)
	network_client.connected_to_server.connect(_on_network_connected)
	network_client.connection_closed.connect(_on_network_closed)
	network_client.packet_received.connect(_on_network_packet_received)
	network_client.packet_parse_failed.connect(_on_network_packet_parse_failed)

	world_sync = WorldSyncScript.new()
	world_sync.configure(self, player, bullets, asteroids)
	world_sync.bullet_spawned.connect(_on_world_bullet_spawned)

	hud_controller = HudControllerScript.new()
	hud_controller.configure(get_tree().current_scene)
	hud_controller.set_room_id(room_id)

	effects = EffectsScript.new()
	effects.configure(self, hud_controller.game_over_sound)

	DisplayServer.window_set_min_size(Vector2i(1280, 720))
	get_viewport().size_changed.connect(_send_client_config)

	network_client.connect_to_server(_websocket_url())


func _exit_tree() -> void:
	if network_client != null:
		network_client.begin_graceful_close()
	_clear_background_scroll_offset()


func _process(delta: float) -> void:
	network_client.poll()
	hud_controller.update(delta)
	if hud_controller.is_game_over && Input.is_action_just_pressed("OpenMenu"):
		await _close_network_connection()
		return_to_menu_requested.emit()
		return

	if network_client.is_connected_to_server():
		network_client.send_packet(player.get_input_packet())
		if hud_controller.can_respawn && !respawn_requested && Input.is_key_pressed(KEY_R):
			respawn_requested = true
			network_client.send_packet(Packets.respawn_packet())

	world_sync.interpolate(delta)
	_update_background_scroll_offset()


func _apply_state(data: Dictionary) -> void:
	if data.get(Packets.FIELD_TYPE, "") != Packets.TYPE_STATE:
		return

	self_id = data[Packets.FIELD_SELF_ID]
	var server_players: Dictionary = data[Packets.FIELD_PLAYERS]
	var server_bullets: Dictionary = data.get(Packets.FIELD_BULLETS, {})
	var server_asteroids: Dictionary = data.get(Packets.FIELD_ASTEROIDS, {})
	var server_events: Array = []
	var events_data = data.get(Packets.FIELD_EVENTS, [])
	if events_data is Array:
		server_events = events_data

	world_sync.apply_state(
		self_id,
		server_players,
		server_bullets,
		server_asteroids,
		has_received_state
	)
	_apply_events(server_events)
	has_received_state = true

	hud_controller.set_lives(int(data.get(Packets.FIELD_LIVES, Constants.PLAYER_STARTING_LIVES)))
	if server_players.has(self_id):
		has_initial_spawn = true
		hud_controller.set_score(int(server_players[self_id].get(Packets.FIELD_SCORE, 0)))
		if hud_controller.is_dead && respawn_requested:
			_set_alive_state()


func _on_network_connected() -> void:
	print("Connected!")
	_send_client_config()


func _on_network_closed() -> void:
	print("Closed")


func _on_network_packet_received(data: Dictionary) -> void:
	_apply_state(data)


func _on_network_packet_parse_failed(text: String) -> void:
	print("bad json: ", text)


func _on_world_bullet_spawned() -> void:
	player.play_laser_sound()


func _close_network_connection() -> void:
	if network_client != null:
		await network_client.close_gracefully()


func _update_background_scroll_offset() -> void:
	if !has_initial_spawn:
		return

	var shell := get_parent()
	if shell != null && shell.has_method("set_gameplay_scroll_offset"):
		shell.set_gameplay_scroll_offset(player.global_position)


func _clear_background_scroll_offset() -> void:
	var shell := get_parent()
	if shell != null && shell.has_method("clear_gameplay_scroll_offset"):
		shell.clear_gameplay_scroll_offset()


func _apply_events(server_events: Array) -> void:
	for event in server_events:
		if event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_BULLET_BLAST:
			effects.spawn_bullet_blast(Vector2(event[Packets.FIELD_X], event[Packets.FIELD_Y]))
		elif event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_SHIP_DEATH:
			if event[Packets.FIELD_PLAYER_ID] == self_id:
				_apply_self_death_event(event)
			effects.spawn_ship_death(Vector2(event[Packets.FIELD_X], event[Packets.FIELD_Y]))


func _apply_self_death_event(event: Dictionary) -> void:
	var lives := int(event.get(Packets.FIELD_LIVES, 0))
	hud_controller.set_lives(lives)
	if lives <= 0:
		_set_game_over_state()
		return

	_set_dead_state(float(event.get(Packets.FIELD_RESPAWN_DELAY, Constants.PLAYER_RESPAWN_DELAY)))


func _set_alive_state() -> void:
	respawn_requested = false
	hud_controller.set_alive()
	effects.reset_game_over_sound()


func _set_dead_state(respawn_delay: float) -> void:
	respawn_requested = false
	hud_controller.set_dead(respawn_delay)
	effects.stop_game_over_sound()


func _set_game_over_state() -> void:
	respawn_requested = false
	hud_controller.set_game_over()
	effects.play_game_over_sound_after_delay()


func _send_client_config() -> void:
	if network_client == null || !network_client.is_connected_to_server():
		return

	var visible_size := get_viewport_rect().size
	network_client.send_packet(Packets.client_config_packet(
		visible_size.x,
		visible_size.y
	))


func _websocket_url() -> String:
	if room_id == "":
		return "ws://localhost:8080/ws"

	return "ws://localhost:8080/ws?room_id=%s" % room_id.uri_encode()
