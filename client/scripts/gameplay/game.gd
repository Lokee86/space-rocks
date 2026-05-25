extends Node2D

signal return_to_menu_requested

const EffectsScript = preload("res://scripts/gameplay/effects.gd")
const CameraFollowScript = preload("res://scripts/camera/camera_follow.gd")
const DebugInputControllerScript = preload("res://scripts/gameplay/support/debug_input_controller.gd")
const GameBackgroundScrollScript = preload("res://scripts/gameplay/support/game_background_scroll.gd")
const GameplayEventControllerScript = preload("res://scripts/gameplay/events/gameplay_event_controller.gd")
const GameplayLifecycleControllerScript = preload("res://scripts/gameplay/gameplay_lifecycle_controller.gd")
const GameplayMenuControllerScript = preload("res://scripts/gameplay/menu/gameplay_menu_controller.gd")
const HudControllerScript = preload("res://scripts/ui/hud/hud_controller.gd")
const NetworkClientScript = preload("res://scripts/networking/network_client.gd")
const OffscreenIndicatorControllerScript = preload("res://scripts/gameplay/support/offscreen_indicator_controller.gd")
const Packets = preload("res://scripts/networking/packets.gd")
const SpectateTargetsScript = preload("res://scripts/gameplay/spectate_targets.gd")
const WorldSyncScript = preload("res://scripts/networking/world_sync.gd")
const RoomState = preload("res://scripts/session/room_state.gd")
const PlayerLifecycle = preload("res://scripts/gameplay/player_lifecycle.gd")
const SpectateControllerScript = preload("res://scripts/gameplay/spectate/spectate_controller.gd")

@onready var player: Player = $Player
@onready var bullets = $Bullets
@onready var asteroids: Node2D = $Asteroids
@onready var offscreen_indicators = get_node_or_null("CanvasLayer/HUD/OffscreenIndicators")
@onready var gameplay_camera := player.get_node_or_null("Camera2D") as Camera2D

var respawn_retry_remaining: float:
	get:
		return _gameplay_lifecycle_controller().respawn_retry_remaining
	set(value):
		_gameplay_lifecycle_controller().respawn_retry_remaining = value
var awaiting_respawn_confirmation: bool:
	get:
		return _gameplay_lifecycle_controller().is_awaiting_respawn_confirmation()
	set(value):
		_gameplay_lifecycle_controller().set_awaiting_respawn_confirmation(value)
var has_received_state := false
var has_initial_spawn := false
var is_gameplay_paused: bool:
	get:
		return _gameplay_menu_controller().is_gameplay_paused
	set(value):
		_gameplay_menu_controller().is_gameplay_paused = value
var open_menu_input_armed: bool:
	get:
		return _gameplay_menu_controller().open_menu_input_armed
	set(value):
		_gameplay_menu_controller().open_menu_input_armed = value
var self_id := ""
var current_spectate_target_id: String:
	get:
		return _spectate_controller().current_target_id()
	set(value):
		_spectate_controller().set_current_target_id(value)
var debug_input_controller
var background_scroll
var gameplay_event_controller
var gameplay_lifecycle_controller
var gameplay_menu_controller
var offscreen_indicator_controller
var effects: Effects
var camera_follow
var game_menu: GameMenu
var injected_network_client: NetworkClient
var hud_controller: HudController
var is_spectating: bool:
	get:
		return _spectate_controller().is_active()
	set(value):
		_spectate_controller().set_active(value)
var spectate_controller
var network_client: NetworkClient
var room_id := ""
var current_room_state := ""
var session_mode := "SinglePlayer"
var preserve_network_on_exit := false
var player_lifecycle := {}
var world_sync: WorldSync


func set_room_id(value: String) -> void:
	room_id = value.strip_edges()


func set_network_client(existing_network_client: NetworkClient) -> void:
	injected_network_client = existing_network_client


func set_session_mode(value) -> void:
	session_mode = str(value)


func _session_mode() -> String:
	return session_mode


func _current_room_state() -> String:
	return current_room_state


func _ready() -> void:
	_setup_network_client()

	debug_input_controller = DebugInputControllerScript.new()
	background_scroll = GameBackgroundScrollScript.new()
	offscreen_indicator_controller = OffscreenIndicatorControllerScript.new()
	offscreen_indicator_controller.configure(offscreen_indicators, gameplay_camera)
	_spectate_controller()

	world_sync = WorldSyncScript.new()
	world_sync.configure(self, player, bullets, asteroids)
	world_sync.bullet_spawned.connect(_on_world_bullet_spawned)

	camera_follow = CameraFollowScript.new()
	camera_follow.configure(gameplay_camera)

	hud_controller = HudControllerScript.new()
	hud_controller.configure(get_tree().current_scene)
	hud_controller.set_session_mode(session_mode)
	hud_controller.set_room_id(room_id)
	game_menu = hud_controller.get_game_menu()
	_gameplay_menu_controller()
	_connect_game_menu_signals()

	effects = EffectsScript.new()
	effects.configure(self, hud_controller.game_over_sound)
	gameplay_event_controller = GameplayEventControllerScript.new()
	gameplay_event_controller.configure(
		effects,
		Callable(world_sync, "visual_position_for_server_position")
	)
	_gameplay_lifecycle_controller()

	get_viewport().size_changed.connect(_send_client_config)

	if injected_network_client == null:
		network_client.connect_to_server(_websocket_url())
	elif network_client.is_connected_to_server():
		_send_client_config()


func _exit_tree() -> void:
	if network_client != null && !preserve_network_on_exit:
		network_client.begin_graceful_close()
	_clear_background_scroll_offset()


func _process(delta: float) -> void:
	if network_client != null:
		network_client.poll()
	if network_client == null:
		return
	hud_controller.update(delta)
	_gameplay_lifecycle_controller().tick_respawn_retry(delta)
	_update_open_menu_input_armed()
	debug_input_controller.handle_input(network_client)
	_handle_spectate_input()
	if _handle_open_menu_pressed():
		return

	if network_client.is_connected_to_server():
		_send_gameplay_input_if_active()

	_update_player_afterburner()
	world_sync.interpolate(delta)
	_update_spectate_camera()
	_update_offscreen_indicators()
	_update_background_scroll_offset()


func _apply_state(data: Dictionary) -> void:
	if data.get(Packets.FIELD_TYPE, "") != Packets.TYPE_STATE:
		return
	if !_can_process_gameplay_packets():
		return

	self_id = data[Packets.FIELD_SELF_ID]
	var server_players: Dictionary = data[Packets.FIELD_PLAYERS]
	player_lifecycle = PlayerLifecycle.from_state(data)
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

	if data.has(Packets.FIELD_LIVES):
		hud_controller.set_lives(int(data[Packets.FIELD_LIVES]))
	else:
		push_warning("State packet missing lives")
	if server_players.has(self_id):
		has_initial_spawn = true
		hud_controller.set_score(int(server_players[self_id].get(Packets.FIELD_SCORE, 0)))
		if hud_controller.is_dead && awaiting_respawn_confirmation:
			_set_alive_state()


func _on_network_connected() -> void:
	print("Connected!")
	_send_client_config()


func _on_network_closed() -> void:
	print("Closed")


func _on_network_packet_received(data: Dictionary) -> void:
	var packet_type := str(data.get(Packets.FIELD_TYPE, ""))
	if packet_type == Packets.TYPE_ROOM_SNAPSHOT || packet_type == Packets.TYPE_ROOM_STATE_CHANGED:
		_store_room_state(data)
		var shell := get_parent()
		if shell != null && shell.has_method("handle_network_packet"):
			shell.handle_network_packet(data)
		return

	if packet_type == Packets.TYPE_ROOM_ERROR:
		var shell := get_parent()
		if shell != null && shell.has_method("handle_network_packet"):
			shell.handle_network_packet(data)
		return

	_apply_state(data)


func _on_network_packet_parse_failed(text: String) -> void:
	print("bad json: ", text)


func _store_room_state(data: Dictionary) -> void:
	current_room_state = str(data.get(Packets.FIELD_ROOM_STATE, current_room_state)).strip_edges()
	if _is_room_game_over():
		_stop_spectating(true)
	_refresh_game_menu_state()
	_refresh_cycle_view_hint()


func _is_room_in_game() -> bool:
	return RoomState.is_in_game(current_room_state)


func _can_process_gameplay_packets() -> bool:
	if current_room_state == "":
		return true

	return _is_room_in_game() || _is_room_game_over()


func _setup_network_client() -> void:
	if injected_network_client != null:
		network_client = injected_network_client
		preserve_network_on_exit = true
		if network_client.get_parent() != self:
			network_client.reparent(self)
	else:
		network_client = NetworkClientScript.new()
		add_child(network_client)

	if !network_client.connected_to_server.is_connected(_on_network_connected):
		network_client.connected_to_server.connect(_on_network_connected)
	if !network_client.connection_closed.is_connected(_on_network_closed):
		network_client.connection_closed.connect(_on_network_closed)
	if !network_client.packet_received.is_connected(_on_network_packet_received):
		network_client.packet_received.connect(_on_network_packet_received)
	if !network_client.packet_parse_failed.is_connected(_on_network_packet_parse_failed):
		network_client.packet_parse_failed.connect(_on_network_packet_parse_failed)


func release_network_client_for_lobby() -> NetworkClient:
	if network_client == null:
		return null

	if network_client.connected_to_server.is_connected(_on_network_connected):
		network_client.connected_to_server.disconnect(_on_network_connected)
	if network_client.connection_closed.is_connected(_on_network_closed):
		network_client.connection_closed.disconnect(_on_network_closed)
	if network_client.packet_received.is_connected(_on_network_packet_received):
		network_client.packet_received.disconnect(_on_network_packet_received)
	if network_client.packet_parse_failed.is_connected(_on_network_packet_parse_failed):
		network_client.packet_parse_failed.disconnect(_on_network_packet_parse_failed)

	var released_client := network_client
	preserve_network_on_exit = true
	set_process(false)
	network_client = null
	if released_client.get_parent() == self && get_parent() != null:
		released_client.reparent(get_parent())

	return released_client


func _on_world_bullet_spawned() -> void:
	player.play_laser_sound()


func _spectate_controller():
	if spectate_controller == null:
		spectate_controller = SpectateControllerScript.new()

	return spectate_controller


func _gameplay_lifecycle_controller():
	if gameplay_lifecycle_controller == null:
		gameplay_lifecycle_controller = GameplayLifecycleControllerScript.new()
		gameplay_lifecycle_controller.configure(
			hud_controller,
			effects,
			player,
			Callable(self, "_stop_spectating"),
			Callable(self, "_resume_gameplay_pause_if_needed"),
			Callable(self, "_hide_game_menu"),
			Callable(self, "_show_game_menu"),
			Callable(self, "_refresh_cycle_view_hint"),
			Callable(self, "_disarm_open_menu_input"),
			Callable(self, "_set_dead_state"),
			Callable(self, "_set_game_over_state")
		)

	return gameplay_lifecycle_controller


func _gameplay_menu_controller():
	if gameplay_menu_controller == null:
		gameplay_menu_controller = GameplayMenuControllerScript.new()
		gameplay_menu_controller.configure(
			hud_controller,
			network_client,
			player,
			Callable(self, "_session_mode"),
			Callable(self, "_is_game_over"),
			Callable(self, "_current_room_state"),
			Callable(self, "_is_room_game_over"),
			Callable(self, "_has_spectate_targets"),
			Callable(self, "_return_to_menu_after_network_close"),
			Callable(self, "_start_spectating")
		)

	return gameplay_menu_controller


func _close_network_connection() -> void:
	if network_client != null:
		await network_client.close_gracefully()


func _update_background_scroll_offset() -> void:
	background_scroll.update_scroll_offset(
		get_parent(),
		has_initial_spawn,
		is_spectating,
		camera_follow,
		player
	)


func _update_player_afterburner() -> void:
	player.set_afterburner_active(
		network_client.is_connected_to_server() &&
			has_initial_spawn &&
			!is_gameplay_paused &&
			player.visible &&
			Input.is_action_pressed(player.move_forward_action)
	)


func _update_offscreen_indicators() -> void:
	offscreen_indicator_controller.update_indicators(
		world_sync.get_remote_player_visual_positions(),
		world_sync.get_remote_player_hues()
	)


func _clear_background_scroll_offset() -> void:
	background_scroll.clear_scroll_offset(get_parent())


func _apply_events(server_events: Array) -> void:
	for event in server_events:
		if event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_BULLET_BLAST:
			gameplay_event_controller.apply_bullet_blast(event)
		elif event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_SHIP_DEATH:
			if event[Packets.FIELD_PLAYER_ID] == self_id:
				_apply_self_death_event(event)
			gameplay_event_controller.apply_ship_death(event)


func _apply_self_death_event(event: Dictionary) -> void:
	_gameplay_lifecycle_controller().apply_self_death_event(event)


func _set_alive_state() -> void:
	_gameplay_lifecycle_controller().set_alive_state()


func _clear_awaiting_respawn_confirmation() -> void:
	_gameplay_lifecycle_controller().clear_awaiting_respawn_confirmation()


func _disarm_open_menu_input() -> void:
	_gameplay_menu_controller().disarm_open_menu_input()


func _set_dead_state(respawn_delay: float) -> void:
	_gameplay_lifecycle_controller().set_dead_state(respawn_delay)


func _set_game_over_state() -> void:
	_gameplay_lifecycle_controller().set_game_over_state()


func _handle_open_menu_pressed() -> bool:
	return _gameplay_menu_controller().handle_open_menu_pressed(has_initial_spawn)


func _open_game_menu() -> void:
	_gameplay_menu_controller().open_game_menu(has_initial_spawn)


func _close_game_menu() -> void:
	_gameplay_menu_controller().close_game_menu()


func _can_pause_server_gameplay() -> bool:
	return _gameplay_menu_controller().can_pause_server_gameplay(has_initial_spawn)


func _set_gameplay_paused(paused: bool) -> void:
	_gameplay_menu_controller().set_gameplay_paused(paused)


func _resume_gameplay_pause_if_needed() -> void:
	if is_gameplay_paused:
		_set_gameplay_paused(false)
	else:
		hud_controller.set_suspended(false)


func _update_open_menu_input_armed() -> void:
	_gameplay_menu_controller().update_open_menu_input_armed(has_initial_spawn)


func _return_to_menu_after_network_close() -> void:
	_hide_game_menu()
	await _close_network_connection()
	return_to_menu_requested.emit()


func _show_game_menu() -> void:
	game_menu = hud_controller.get_game_menu()
	_gameplay_menu_controller().show_game_menu()


func _refresh_game_menu_state() -> void:
	_gameplay_menu_controller().refresh_game_menu_state(game_menu)


func _connect_game_menu_signals() -> void:
	_gameplay_menu_controller().connect_game_menu_signals(game_menu)


func _is_game_menu_open() -> bool:
	return _gameplay_menu_controller().is_game_menu_open()


func _hide_game_menu() -> void:
	_gameplay_menu_controller().hide_game_menu()


func _on_game_menu_resume_requested() -> void:
	_gameplay_menu_controller().on_resume_requested()


func _on_game_menu_lobby_requested() -> void:
	_gameplay_menu_controller().on_lobby_requested()


func _on_game_menu_spectate_requested() -> void:
	_gameplay_menu_controller().on_spectate_requested()


func _on_game_menu_quit_requested() -> void:
	_gameplay_menu_controller().on_quit_requested()


func _is_multiplayer_session() -> bool:
	return session_mode.strip_edges().to_lower() == "multiplayer"


func _is_game_over() -> bool:
	if hud_controller != null && hud_controller.is_game_over:
		return true

	return _is_multiplayer_session() && _is_room_game_over()


func _is_room_game_over() -> bool:
	return RoomState.is_game_over(current_room_state)


func _has_spectate_targets() -> bool:
	return _spectate_controller().has_targets(
		self_id,
		_remote_player_visual_positions(),
		player_lifecycle
	)


func _start_spectating() -> bool:
	return _spectate_controller().start_spectating(
		self_id,
		_remote_player_visual_positions(),
		player_lifecycle,
		Callable(self, "_hide_game_menu"),
		Callable(self, "_update_spectate_camera"),
		Callable(self, "_refresh_cycle_view_hint")
	)


func _stop_spectating(show_game_over_menu: bool) -> void:
	_spectate_controller().stop_spectating(
		show_game_over_menu,
		hud_controller != null && hud_controller.is_game_over,
		Callable(self, "_follow_local_player"),
		Callable(self, "_show_game_menu"),
		Callable(self, "_refresh_cycle_view_hint")
	)


func _follow_local_player() -> void:
	if camera_follow != null:
		camera_follow.follow_local_player()


func _update_spectate_camera() -> void:
	_spectate_controller().update_camera(
		self_id,
		_remote_player_visual_positions(),
		player_lifecycle,
		camera_follow,
		Callable(self, "_stop_spectating")
	)


func _handle_spectate_input() -> void:
	if !is_spectating:
		return
	if !Input.is_action_just_pressed("SwitchCamera"):
		return

	_cycle_spectate_target()


func _cycle_spectate_target() -> void:
	_spectate_controller().cycle_target(
		self_id,
		_remote_player_visual_positions(),
		player_lifecycle,
		Callable(self, "_stop_spectating"),
		Callable(self, "_follow_visual_position")
	)


func _follow_visual_position(visual_position: Vector2) -> void:
	if camera_follow != null:
		camera_follow.follow_visual_position(visual_position)


func _remote_player_visual_positions() -> Dictionary:
	if world_sync == null:
		return {}

	return world_sync.get_remote_player_visual_positions()


func _refresh_cycle_view_hint() -> void:
	if hud_controller == null:
		return

	hud_controller.set_session_mode(session_mode)
	hud_controller.set_cycle_view_available(
		_is_multiplayer_session() &&
		hud_controller.is_game_over &&
		_spectate_controller().is_active() &&
		!_is_room_game_over()
	)


func _should_block_open_menu_for_game_over() -> bool:
	return _gameplay_menu_controller().should_block_open_menu_for_game_over()


func _send_gameplay_input_if_active() -> void:
	if is_gameplay_paused:
		return

	network_client.send_packet(player.get_input_packet())
	_gameplay_lifecycle_controller().send_respawn_request_if_ready(network_client)


func _send_client_config() -> void:
	if network_client == null || !network_client.is_connected_to_server():
		return

	var visible_size := get_viewport_rect().size
	network_client.send_packet(Packets.client_config_packet(
		visible_size.x,
		visible_size.y
	))


func _websocket_url() -> String:
	return "ws://localhost:8080/ws"
