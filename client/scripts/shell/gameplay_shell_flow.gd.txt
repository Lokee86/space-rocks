extends RefCounted
class_name GameplayShellFlow

signal gameplay_started
signal quit_to_main_menu_requested
signal return_to_lobby_requested

const WorldSyncScript = preload("res://scripts/world/world_sync.gd")
const GameplayStatePacketReader = preload("res://scripts/gameplay/session/gameplay_state_packet_reader.gd")
const GameplayEventFlow = preload("res://scripts/shell/gameplay_event_flow.gd")
const Packets = preload("res://scripts/networking/packets/packets.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")

var world_sync
var connection_service
var player
var hud_flow
var menu_flow
var background_flow
var event_flow
var spectate_menu_state
var has_received_state := false
var awaiting_respawn_confirmation := false
var pending_open_menu_before_spawn := false
var is_spectating := false
var debug_invincible_enabled := false
var debug_invincible_toggle_was_pressed := false


func configure(
	connection_service_ref,
	game_owner: Node2D,
	player_ref: Player,
	bullets: Node2D,
	asteroids: Node2D,
	hud_flow_ref,
	menu_flow_ref,
	background_flow_ref,
	game_over_sound: AudioStreamPlayer
) -> void:
	connection_service = connection_service_ref
	player = player_ref
	hud_flow = hud_flow_ref
	menu_flow = menu_flow_ref
	if menu_flow != null && menu_flow.has_signal("quit_to_main_menu_requested"):
		var quit_callable := Callable(self, "_on_quit_to_main_menu_requested")
		if !menu_flow.quit_to_main_menu_requested.is_connected(quit_callable):
			menu_flow.quit_to_main_menu_requested.connect(quit_callable)
	if menu_flow != null && menu_flow.has_signal("return_to_lobby_requested"):
		var return_to_lobby_callable := Callable(self, "_on_return_to_lobby_requested")
		if !menu_flow.return_to_lobby_requested.is_connected(return_to_lobby_callable):
			menu_flow.return_to_lobby_requested.connect(return_to_lobby_callable)
	if menu_flow != null && menu_flow.has_signal("spectate_requested"):
		var spectate_callable := Callable(self, "_on_spectate_requested")
		if !menu_flow.spectate_requested.is_connected(spectate_callable):
			menu_flow.spectate_requested.connect(spectate_callable)
	background_flow = background_flow_ref
	world_sync = WorldSyncScript.new()
	world_sync.configure(game_owner, player_ref, bullets, asteroids)
	world_sync.bullet_spawned.connect(_on_bullet_spawned)
	event_flow = GameplayEventFlow.new()
	event_flow.configure(
		game_owner,
		game_over_sound,
		Callable(world_sync, "visual_position_for_server_position")
	)
	event_flow.self_death_event.connect(_on_self_death_event)


func reset() -> void:
	has_received_state = false
	awaiting_respawn_confirmation = false
	pending_open_menu_before_spawn = false
	is_spectating = false
	if player != null:
		player.hide()
	if world_sync != null:
		world_sync.reset()
	if hud_flow != null:
		hud_flow.reset()
	if menu_flow != null && menu_flow.has_method("reset"):
		menu_flow.reset()
	if background_flow != null:
		background_flow.clear()
	if event_flow != null:
		event_flow.reset()


func apply_gameplay_state(packet: Dictionary) -> void:
	var is_first_gameplay_state := !has_received_state
	var state := GameplayStatePacketReader.read(packet)
	if hud_flow != null:
		hud_flow.show_gameplay()
		if state["has_lives"]:
			hud_flow.apply_lives(state["lives"])
		var server_players: Dictionary = state["server_players"]
		var self_id: String = state["self_id"]
		if server_players.has(self_id):
			var self_state: Dictionary = server_players[self_id]
			hud_flow.apply_score(int(self_state.get(Packets.FIELD_SCORE, 0)))
	world_sync.apply_state(
		state["self_id"],
		state["server_players"],
		state["server_bullets"],
		state["server_asteroids"],
		has_received_state
	)
	if hud_flow != null && _should_restore_alive_hud(state):
		hud_flow.set_alive()
		if menu_flow != null:
			menu_flow.set_alive()
		awaiting_respawn_confirmation = false
	if event_flow != null:
		event_flow.apply_server_events(state["server_events"], state["self_id"])
	has_received_state = true
	if pending_open_menu_before_spawn:
		pending_open_menu_before_spawn = false
		if menu_flow != null:
			menu_flow.open_live_pause_from_request(true)
	if is_first_gameplay_state:
		gameplay_started.emit()


func configure_spectate_menu_state(spectate_menu_state_ref) -> void:
	spectate_menu_state = spectate_menu_state_ref


func current_camera() -> Camera2D:
	if player == null:
		return null
	return player.get_node_or_null("Camera2D") as Camera2D


func remote_player_visual_positions() -> Dictionary:
	if world_sync == null:
		return {}
	return world_sync.get_remote_player_visual_positions()


func remote_player_hues() -> Dictionary:
	if world_sync == null:
		return {}
	return world_sync.get_remote_player_hues()


func process(delta: float) -> void:
	if world_sync != null:
		world_sync.interpolate(delta)
	if background_flow != null && has_received_state && player != null && player.visible:
		background_flow.set_scroll_reference(player.global_position)
	if hud_flow != null:
		hud_flow.update(delta)
		if (
			has_received_state
			&& connection_service != null
			&& hud_flow.can_request_respawn()
			&& Input.is_action_just_pressed("Respawn")
		):
			connection_service.send_respawn_request()
			awaiting_respawn_confirmation = true

	_update_local_player_presentation()

	if is_spectating && Input.is_action_just_pressed("OpenMenu"):
		if menu_flow != null:
			menu_flow.show_spectating_menu()
		_handle_debug_invincible_toggle()
		return

	if menu_flow != null:
		if !has_received_state && Input.is_action_just_pressed("OpenMenu"):
			pending_open_menu_before_spawn = true
		elif has_received_state:
			menu_flow.handle_open_menu_pressed(has_received_state)

	_handle_debug_invincible_toggle()

	if (
		is_spectating
		&& Input.is_action_just_pressed("SwitchCamera")
		&& spectate_menu_state != null
		&& world_sync != null
	):
		var target_id: String = spectate_menu_state.cycle_next_target()
		if !target_id.is_empty():
			world_sync.focus_camera_on_player(target_id)

	if (
		has_received_state
		&& player != null
		&& connection_service != null
		&& (menu_flow == null || !menu_flow.is_gameplay_paused)
	):
		connection_service.send_input_packet(player.get_input_packet())


func _on_bullet_spawned() -> void:
	if player != null:
		player.play_laser_sound()


func _on_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()


func _on_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()


func _handle_debug_invincible_toggle() -> void:
	var toggle_pressed := Input.is_key_pressed(KEY_1)
	if !has_received_state || connection_service == null:
		debug_invincible_toggle_was_pressed = toggle_pressed
		return
	if !toggle_pressed:
		debug_invincible_toggle_was_pressed = false
		return
	if debug_invincible_toggle_was_pressed:
		return

	debug_invincible_toggle_was_pressed = true
	debug_invincible_enabled = !debug_invincible_enabled
	connection_service.send_packet(Packets.toggle_debug_invincible_packet())
	ClientLogger.game_debug("Debug invincibility toggled: %s" % debug_invincible_enabled)


func _on_spectate_requested() -> void:
	if spectate_menu_state == null || world_sync == null:
		return

	var target_id: String = spectate_menu_state.begin_spectating()
	if !target_id.is_empty() && world_sync.focus_camera_on_player(target_id):
		is_spectating = true


func _update_local_player_presentation() -> void:
	if !has_received_state || player == null || !player.visible:
		return

	player.set_afterburner_active(Input.is_action_pressed(player.move_forward_action))


func _on_self_death_event(event: Dictionary) -> void:
	awaiting_respawn_confirmation = false
	if hud_flow == null:
		return

	var lives := int(event.get(Packets.FIELD_LIVES, 0))
	hud_flow.apply_lives(lives)
	if lives <= 0:
		hud_flow.set_game_over()
		if menu_flow != null:
			menu_flow.set_game_over()
		if event_flow != null && event_flow.has_method("play_game_over_sound_after_delay"):
			event_flow.play_game_over_sound_after_delay()
		return

	var respawn_delay := 0.0
	if event.has(Packets.FIELD_RESPAWN_DELAY):
		respawn_delay = float(event[Packets.FIELD_RESPAWN_DELAY])
	hud_flow.set_dead(respawn_delay)


func _should_restore_alive_hud(state: Dictionary) -> bool:
	if !awaiting_respawn_confirmation:
		return false

	if state["has_lives"] && int(state["lives"]) <= 0:
		return false

	var server_players: Dictionary = state["server_players"]
	var self_id: String = state["self_id"]
	if !server_players.has(self_id):
		return false

	var self_state = server_players[self_id]
	var has_valid_server_state := false
	if self_state is Dictionary:
		var self_state_dictionary: Dictionary = self_state
		has_valid_server_state = !self_state_dictionary.is_empty()
	return (player != null && player.visible) || has_valid_server_state
