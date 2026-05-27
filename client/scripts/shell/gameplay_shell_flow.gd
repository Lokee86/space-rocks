extends RefCounted
class_name GameplayShellFlow

const GameplayRuntimeContext = preload("res://scripts/gameplay/session/gameplay_runtime_context.gd")
const GameplayStatePacketReader = preload("res://scripts/gameplay/session/gameplay_state_packet_reader.gd")
const GameplayRuntimeTickFlow = preload("res://scripts/shell/gameplay_runtime_tick_flow.gd")
const GameplayDevtoolsContext = preload("res://scripts/devtools/gameplay_devtools_context.gd")
const GameplaySpectateContext = preload("res://scripts/gameplay/spectate/gameplay_spectate_context.gd")
const Packets = preload("res://scripts/networking/packets/packets.gd")

signal gameplay_started
signal quit_to_main_menu_requested
signal return_to_lobby_requested

var player
var runtime_context
var connection_service
var hud_flow
var menu_flow
var background_flow
var runtime_tick_flow
var devtools_context
var spectate_context
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
	if menu_flow != null && menu_flow.has_signal("spectate_requested"):
		var spectate_callable := Callable(self, "_on_spectate_requested")
		if !menu_flow.spectate_requested.is_connected(spectate_callable):
			menu_flow.spectate_requested.connect(spectate_callable)
	background_flow = background_flow_ref
	runtime_context = GameplayRuntimeContext.new()
	runtime_context.configure_world(game_owner, player_ref, bullets, asteroids)
	runtime_context.configure_events(
		game_owner,
		game_over_sound,
		hud_flow,
		menu_flow
	)
	runtime_context.configure_respawn(connection_service, hud_flow)
	runtime_tick_flow = GameplayRuntimeTickFlow.new()
	runtime_tick_flow.configure(hud_flow)
	runtime_context.configure_pause_input(menu_flow)
	devtools_context = GameplayDevtoolsContext.new()
	devtools_context.configure(connection_service)
	spectate_context = GameplaySpectateContext.new()
	spectate_context.configure(menu_flow, spectate_menu_state, runtime_context.world_sync)
	runtime_context.configure_input(connection_service, player, menu_flow)


func reset() -> void:
	has_received_state = false
	if runtime_context != null:
		runtime_context.reset()
	if hud_flow != null:
		hud_flow.reset()
	if menu_flow != null && menu_flow.has_method("reset"):
		menu_flow.reset()
	if background_flow != null:
		background_flow.clear()
	if runtime_tick_flow != null:
		runtime_tick_flow.reset()
	if devtools_context != null:
		devtools_context.reset()
	if spectate_context != null:
		spectate_context.reset()


func apply_gameplay_state(packet: Dictionary) -> void:
	if runtime_context == null:
		return

	var is_first_gameplay_state := !has_received_state
	var state := GameplayStatePacketReader.read(packet)
	runtime_context.mark_input_gameplay_state_received()
	if background_flow != null:
		background_flow.mark_gameplay_state_received()
	if hud_flow != null:
		hud_flow.show_gameplay()
		if state["has_lives"]:
			hud_flow.apply_lives(state["lives"])
		var server_players: Dictionary = state["server_players"]
		var self_id: String = state["self_id"]
		if server_players.has(self_id):
			var self_state: Dictionary = server_players[self_id]
			hud_flow.apply_score(int(self_state.get(Packets.FIELD_SCORE, 0)))
	runtime_context.apply_world_state(state, has_received_state)
	runtime_context.apply_respawn_alive_restore(state, menu_flow)
	runtime_context.apply_server_events(state)
	has_received_state = true
	if is_first_gameplay_state:
		gameplay_started.emit()


func configure_spectate_menu_state(spectate_menu_state_ref) -> void:
	spectate_menu_state = spectate_menu_state_ref
	if spectate_context != null:
		spectate_context.configure(menu_flow, spectate_menu_state, runtime_context.world_sync)


func current_camera() -> Camera2D:
	if runtime_context == null:
		return null
	return runtime_context.current_camera()


func remote_player_visual_positions() -> Dictionary:
	if runtime_context == null:
		return {}
	return runtime_context.remote_player_visual_positions()


func process(_delta: float) -> void:
	if runtime_context != null:
		runtime_context.process(_delta)
	if runtime_tick_flow != null:
		runtime_tick_flow.process(_delta)
	runtime_context.process_pause_input(has_received_state)
	if devtools_context != null:
		devtools_context.process(has_received_state)
	if spectate_context != null:
		spectate_context.process()
	if background_flow != null:
		background_flow.process()
	runtime_context.process_respawn(has_received_state)
	runtime_context.process_input()


func _on_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()


func _on_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()


func _on_spectate_requested() -> void:
	if spectate_context != null:
		spectate_context.begin_spectating()
