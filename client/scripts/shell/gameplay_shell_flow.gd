extends RefCounted
class_name GameplayShellFlow

const WorldSyncScript = preload("res://scripts/world/world_sync.gd")
const GameplayStatePacketReader = preload("res://scripts/gameplay/session/gameplay_state_packet_reader.gd")
const GameplayEventFlow = preload("res://scripts/shell/gameplay_event_flow.gd")
const GameplayDeathFlow = preload("res://scripts/shell/gameplay_death_flow.gd")
const GameplayRespawnFlow = preload("res://scripts/shell/gameplay_respawn_flow.gd")
const GameplayRuntimeTickFlow = preload("res://scripts/shell/gameplay_runtime_tick_flow.gd")
const GameplayPauseInputFlow = preload("res://scripts/shell/gameplay_pause_input_flow.gd")
const Packets = preload("res://scripts/networking/packets/packets.gd")

signal gameplay_started
signal quit_to_main_menu_requested
signal return_to_lobby_requested

var player
var world_sync
var connection_service
var hud_flow
var menu_flow
var background_flow
var event_flow
var death_flow
var respawn_flow
var runtime_tick_flow
var pause_input_flow
var spectate_menu_state
var has_received_state := false


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
	death_flow = GameplayDeathFlow.new()
	death_flow.configure(hud_flow, menu_flow, event_flow)
	event_flow.self_death_event.connect(Callable(death_flow, "apply_self_death_event"))
	respawn_flow = GameplayRespawnFlow.new()
	respawn_flow.configure(connection_service, hud_flow)
	runtime_tick_flow = GameplayRuntimeTickFlow.new()
	runtime_tick_flow.configure(hud_flow)
	pause_input_flow = GameplayPauseInputFlow.new()
	pause_input_flow.configure(menu_flow)


func reset() -> void:
	has_received_state = false
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
	if death_flow != null:
		death_flow.reset()
	if respawn_flow != null:
		respawn_flow.reset()
	if runtime_tick_flow != null:
		runtime_tick_flow.reset()
	if pause_input_flow != null:
		pause_input_flow.reset()


func apply_gameplay_state(packet: Dictionary) -> void:
	if world_sync == null:
		return

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
	if hud_flow != null && respawn_flow != null && respawn_flow.should_restore_alive_hud(state, player):
		hud_flow.set_alive()
		if menu_flow != null:
			menu_flow.set_alive()
		respawn_flow.clear_awaiting_confirmation()
	if event_flow != null:
		event_flow.apply_server_events(state["server_events"], state["self_id"])
	has_received_state = true
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


func process(_delta: float) -> void:
	if world_sync != null:
		world_sync.interpolate(_delta)
	if runtime_tick_flow != null:
		runtime_tick_flow.process(_delta)
	if pause_input_flow != null:
		pause_input_flow.process(has_received_state)
	if background_flow != null && has_received_state && player != null && player.visible:
		background_flow.set_scroll_reference(player.global_position)
	if respawn_flow != null:
		respawn_flow.process(has_received_state)
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
