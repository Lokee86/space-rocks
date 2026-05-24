extends Node2D

signal return_to_menu_requested

const EffectsScript = preload("res://scripts/effects.gd")
const CameraFollowScript = preload("res://scripts/camera_follow.gd")
const HudControllerScript = preload("res://scripts/ui/hud_controller.gd")
const NetworkClientScript = preload("res://scripts/networking/network_client.gd")
const Packets = preload("res://scripts/networking/packets.gd")
const SpectateTargetsScript = preload("res://scripts/spectate_targets.gd")
const WorldSyncScript = preload("res://scripts/networking/world_sync.gd")
const RESPAWN_RETRY_SECONDS := 0.25

@onready var player: Player = $Player
@onready var bullets = $Bullets
@onready var asteroids: Node2D = $Asteroids
@onready var offscreen_indicators = get_node_or_null("CanvasLayer/HUD/OffscreenIndicators")
@onready var gameplay_camera := player.get_node_or_null("Camera2D") as Camera2D

var respawn_retry_remaining := 0.0
var awaiting_respawn_confirmation := false
var has_received_state := false
var has_initial_spawn := false
var is_gameplay_paused := false
var open_menu_input_armed := false
var debug_invincible_input_armed := true
var debug_infinite_lives_input_armed := true
var debug_freeze_world_input_armed := true
var debug_freeze_player_input_armed := true
var self_id := ""
var current_spectate_target_id := ""
var effects: Effects
var camera_follow
var game_menu: GameMenu
var injected_network_client: NetworkClient
var hud_controller: HudController
var is_spectating := false
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


func _ready() -> void:
	_setup_network_client()

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
	_connect_game_menu_signals()

	effects = EffectsScript.new()
	effects.configure(self, hud_controller.game_over_sound)

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
	respawn_retry_remaining = max(0.0, respawn_retry_remaining - delta)
	_update_open_menu_input_armed()
	_handle_debug_input()
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
	player_lifecycle = _player_lifecycle_from_state(data)
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


func _player_lifecycle_from_state(data: Dictionary) -> Dictionary:
	var lifecycle_data = data.get(Packets.FIELD_PLAYER_LIFECYCLE, {})
	if !(lifecycle_data is Dictionary):
		return {}

	var lifecycle := {}
	for player_id in lifecycle_data.keys():
		lifecycle[str(player_id)] = str(lifecycle_data[player_id])
	return lifecycle


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
	return current_room_state == "InGame" || current_room_state == "in_game"


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


func _close_network_connection() -> void:
	if network_client != null:
		await network_client.close_gracefully()


func _update_background_scroll_offset() -> void:
	if !has_initial_spawn:
		return

	var shell := get_parent()
	if shell != null && shell.has_method("set_gameplay_scroll_offset"):
		if is_spectating && camera_follow != null && camera_follow.camera != null:
			shell.set_gameplay_scroll_offset(camera_follow.camera.global_position)
		else:
			shell.set_gameplay_scroll_offset(player.global_position)


func _update_player_afterburner() -> void:
	player.set_afterburner_active(
		network_client.is_connected_to_server() &&
			has_initial_spawn &&
			!is_gameplay_paused &&
			player.visible &&
			Input.is_action_pressed(player.move_forward_action)
	)


func _update_offscreen_indicators() -> void:
	if offscreen_indicators == null || gameplay_camera == null:
		return

	offscreen_indicators.update_indicators(
		world_sync.get_remote_player_visual_positions(),
		gameplay_camera,
		world_sync.get_remote_player_hues()
	)


func _clear_background_scroll_offset() -> void:
	var shell := get_parent()
	if shell != null && shell.has_method("clear_gameplay_scroll_offset"):
		shell.clear_gameplay_scroll_offset()


func _apply_events(server_events: Array) -> void:
	for event in server_events:
		if event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_BULLET_BLAST:
			var event_position := Vector2(event[Packets.FIELD_X], event[Packets.FIELD_Y])
			effects.spawn_bullet_blast(world_sync.visual_position_for_server_position(event_position))
		elif event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_SHIP_DEATH:
			if event[Packets.FIELD_PLAYER_ID] == self_id:
				_apply_self_death_event(event)
			var event_position := Vector2(event[Packets.FIELD_X], event[Packets.FIELD_Y])
			effects.spawn_ship_death(world_sync.visual_position_for_server_position(event_position))


func _apply_self_death_event(event: Dictionary) -> void:
	var lives := int(event.get(Packets.FIELD_LIVES, 0))
	hud_controller.set_lives(lives)
	if lives <= 0:
		_set_game_over_state()
		return

	if event.has(Packets.FIELD_RESPAWN_DELAY):
		_set_dead_state(float(event[Packets.FIELD_RESPAWN_DELAY]))
	else:
		push_warning("Ship death event missing respawn delay")
		_set_dead_state(0.0)


func _set_alive_state() -> void:
	awaiting_respawn_confirmation = false
	_stop_spectating(false)
	_resume_gameplay_pause_if_needed()
	open_menu_input_armed = false
	_hide_game_menu()
	hud_controller.set_alive()
	_refresh_cycle_view_hint()
	effects.reset_game_over_sound()


func _set_dead_state(respawn_delay: float) -> void:
	awaiting_respawn_confirmation = false
	_stop_spectating(false)
	_resume_gameplay_pause_if_needed()
	open_menu_input_armed = false
	_hide_game_menu()
	player.set_afterburner_active(false)
	hud_controller.set_dead(respawn_delay)
	_refresh_cycle_view_hint()
	effects.stop_game_over_sound()


func _set_game_over_state() -> void:
	awaiting_respawn_confirmation = false
	_resume_gameplay_pause_if_needed()
	open_menu_input_armed = false
	_hide_game_menu()
	player.set_afterburner_active(false)
	hud_controller.set_game_over()
	_show_game_menu()
	_refresh_cycle_view_hint()
	effects.play_game_over_sound_after_delay()


func _handle_open_menu_pressed() -> bool:
	if !Input.is_action_just_pressed("OpenMenu"):
		return false
	if _should_block_open_menu_for_game_over():
		return false
	if !open_menu_input_armed && !hud_controller.is_game_over:
		return false

	if _is_game_menu_open():
		_close_game_menu()
	else:
		_open_game_menu()
	return true


func _open_game_menu() -> void:
	_show_game_menu()
	if _can_pause_server_gameplay():
		_set_gameplay_paused(true)


func _close_game_menu() -> void:
	if is_gameplay_paused:
		_set_gameplay_paused(false)
	else:
		_hide_game_menu()
		hud_controller.set_suspended(false)


func _can_pause_server_gameplay() -> bool:
	return network_client.is_connected_to_server() && has_initial_spawn && !hud_controller.is_dead


func _set_gameplay_paused(paused: bool) -> void:
	if is_gameplay_paused == paused:
		if !paused:
			_hide_game_menu()
			hud_controller.set_suspended(false)
		return

	is_gameplay_paused = paused
	hud_controller.set_suspended(paused)
	if paused:
		player.set_afterburner_active(false)
		network_client.send_packet(Packets.pause_player_packet())
	else:
		_hide_game_menu()
		network_client.send_packet(Packets.resume_player_packet())


func _resume_gameplay_pause_if_needed() -> void:
	if is_gameplay_paused:
		_set_gameplay_paused(false)
	else:
		hud_controller.set_suspended(false)


func _update_open_menu_input_armed() -> void:
	if open_menu_input_armed || !has_initial_spawn:
		return
	if !Input.is_action_pressed("OpenMenu"):
		open_menu_input_armed = true


func _return_to_menu_after_network_close() -> void:
	_hide_game_menu()
	await _close_network_connection()
	return_to_menu_requested.emit()


func _show_game_menu() -> void:
	game_menu = hud_controller.get_game_menu()
	if game_menu == null:
		return

	_refresh_game_menu_state()
	_connect_game_menu_signals()
	hud_controller.show_game_menu()


func _refresh_game_menu_state() -> void:
	if game_menu == null:
		return
	if game_menu.has_method("configure_for_state"):
		game_menu.configure_for_state(
			session_mode,
			_is_game_over(),
			current_room_state,
			_has_spectate_targets()
		)


func _connect_game_menu_signals() -> void:
	if game_menu == null:
		return

	if game_menu.has_signal("lobby_requested"):
		if !game_menu.lobby_requested.is_connected(_on_game_menu_lobby_requested):
			game_menu.lobby_requested.connect(_on_game_menu_lobby_requested)
	if game_menu.has_signal("spectate_requested"):
		if !game_menu.spectate_requested.is_connected(_on_game_menu_spectate_requested):
			game_menu.spectate_requested.connect(_on_game_menu_spectate_requested)
	if !game_menu.resume_requested.is_connected(_on_game_menu_resume_requested):
		game_menu.resume_requested.connect(_on_game_menu_resume_requested)
	if !game_menu.quit_requested.is_connected(_on_game_menu_quit_requested):
		game_menu.quit_requested.connect(_on_game_menu_quit_requested)


func _is_game_menu_open() -> bool:
	return hud_controller != null && hud_controller.is_game_menu_visible()


func _hide_game_menu() -> void:
	if hud_controller == null:
		return

	hud_controller.hide_game_menu()


func _on_game_menu_resume_requested() -> void:
	_close_game_menu()


func _on_game_menu_lobby_requested() -> void:
	if !_is_multiplayer_session():
		return
	if network_client == null || !network_client.is_connected_to_server():
		return

	network_client.send_return_to_lobby_request()


func _on_game_menu_spectate_requested() -> void:
	if !_is_multiplayer_session() || _is_room_game_over():
		return
	if !_start_spectating():
		_show_game_menu()


func _on_game_menu_quit_requested() -> void:
	is_gameplay_paused = false
	hud_controller.set_suspended(false)
	if _is_multiplayer_session() && network_client != null && network_client.is_connected_to_server():
		network_client.send_leave_room_request()
	_return_to_menu_after_network_close()


func _is_multiplayer_session() -> bool:
	return session_mode.strip_edges().to_lower() == "multiplayer"


func _is_game_over() -> bool:
	if hud_controller != null && hud_controller.is_game_over:
		return true

	return _is_multiplayer_session() && _is_room_game_over()


func _is_room_game_over() -> bool:
	return current_room_state.strip_edges().replace("_", "").to_lower() == "gameover"


func _has_spectate_targets() -> bool:
	if world_sync == null:
		return false

	return SpectateTargetsScript.select_target(
		self_id,
		"",
		world_sync.get_remote_player_visual_positions(),
		player_lifecycle
	) != ""


func _start_spectating() -> bool:
	var remote_positions := _remote_player_visual_positions()
	current_spectate_target_id = SpectateTargetsScript.select_target(
		self_id,
		current_spectate_target_id,
		remote_positions,
		player_lifecycle
	)
	if current_spectate_target_id == "":
		is_spectating = false
		return false

	is_spectating = true
	_hide_game_menu()
	_update_spectate_camera()
	_refresh_cycle_view_hint()
	return true


func _stop_spectating(show_game_over_menu: bool) -> void:
	if !is_spectating && current_spectate_target_id == "":
		return

	is_spectating = false
	current_spectate_target_id = ""
	if camera_follow != null:
		camera_follow.follow_local_player()
	if show_game_over_menu && hud_controller != null && hud_controller.is_game_over:
		_show_game_menu()
	_refresh_cycle_view_hint()


func _update_spectate_camera() -> void:
	if !is_spectating:
		return

	var remote_positions := _remote_player_visual_positions()
	current_spectate_target_id = SpectateTargetsScript.select_target(
		self_id,
		current_spectate_target_id,
		remote_positions,
		player_lifecycle
	)
	if current_spectate_target_id == "":
		_stop_spectating(true)
		return

	if camera_follow != null:
		camera_follow.follow_visual_position(remote_positions[current_spectate_target_id])


func _handle_spectate_input() -> void:
	if !is_spectating:
		return
	if !Input.is_action_just_pressed("SwitchCamera"):
		return

	_cycle_spectate_target()


func _cycle_spectate_target() -> void:
	if !is_spectating:
		return

	var remote_positions := _remote_player_visual_positions()
	current_spectate_target_id = SpectateTargetsScript.cycle_target(
		self_id,
		current_spectate_target_id,
		remote_positions,
		player_lifecycle
	)
	if current_spectate_target_id == "":
		_stop_spectating(true)
		return

	if camera_follow != null:
		camera_follow.follow_visual_position(remote_positions[current_spectate_target_id])


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
		is_spectating &&
		!_is_room_game_over()
	)


func _should_block_open_menu_for_game_over() -> bool:
	return !_is_multiplayer_session() && _is_game_over()


func _send_gameplay_input_if_active() -> void:
	if is_gameplay_paused:
		return

	network_client.send_packet(player.get_input_packet())
	if hud_controller.can_respawn && Input.is_key_pressed(KEY_R) && respawn_retry_remaining <= 0.0:
		respawn_retry_remaining = RESPAWN_RETRY_SECONDS
		awaiting_respawn_confirmation = true
		network_client.send_packet(Packets.respawn_packet())


func _handle_debug_input() -> void:
	if !Input.is_key_pressed(KEY_F1) && !Input.is_key_pressed(KEY_F2) && !Input.is_key_pressed(KEY_F3) && !Input.is_key_pressed(KEY_F4):
		debug_invincible_input_armed = true
		debug_infinite_lives_input_armed = true
		debug_freeze_world_input_armed = true
		debug_freeze_player_input_armed = true
		return
	if !network_client.is_connected_to_server():
		return
	if Input.is_key_pressed(KEY_F1) && debug_invincible_input_armed:
		debug_invincible_input_armed = false
		network_client.send_packet(Packets.toggle_debug_invincible_packet())
	if Input.is_key_pressed(KEY_F2) && debug_infinite_lives_input_armed:
		debug_infinite_lives_input_armed = false
		network_client.send_packet(Packets.toggle_debug_infinite_lives_packet())
	if Input.is_key_pressed(KEY_F3) && debug_freeze_world_input_armed:
		debug_freeze_world_input_armed = false
		network_client.send_packet(Packets.toggle_debug_freeze_world_packet())
	if Input.is_key_pressed(KEY_F4) && debug_freeze_player_input_armed:
		debug_freeze_player_input_armed = false
		network_client.send_packet(Packets.toggle_debug_freeze_player_packet())


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
