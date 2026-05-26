extends RefCounted
class_name GameplayShellFlow

signal gameplay_started

const WorldSyncScript = preload("res://scripts/world/world_sync.gd")
const GameplayStatePacketReader = preload("res://scripts/gameplay/session/gameplay_state_packet_reader.gd")
const GameplayEventFlow = preload("res://scripts/shell/gameplay_event_flow.gd")
const Packets = preload("res://scripts/networking/packets/packets.gd")

var world_sync
var connection_service
var player
var hud_flow
var background_flow
var event_flow
var has_received_state := false


func configure(
	connection_service_ref,
	game_owner: Node2D,
	player_ref: Player,
	bullets: Node2D,
	asteroids: Node2D,
	hud_flow_ref,
	background_flow_ref,
	game_over_sound: AudioStreamPlayer
) -> void:
	connection_service = connection_service_ref
	player = player_ref
	hud_flow = hud_flow_ref
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
	if player != null:
		player.hide()
	if hud_flow != null:
		hud_flow.reset()
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
	if event_flow != null:
		event_flow.apply_server_events(state["server_events"], state["self_id"])
	has_received_state = true
	if is_first_gameplay_state:
		gameplay_started.emit()


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

	if has_received_state && player != null && connection_service != null:
		connection_service.send_input_packet(player.get_input_packet())


func _on_bullet_spawned() -> void:
	if player != null:
		player.play_laser_sound()


func _on_self_death_event(event: Dictionary) -> void:
	if hud_flow == null:
		return

	var lives := int(event.get(Packets.FIELD_LIVES, 0))
	hud_flow.apply_lives(lives)
	if lives <= 0:
		hud_flow.set_game_over()
		return

	var respawn_delay := 0.0
	if event.has(Packets.FIELD_RESPAWN_DELAY):
		respawn_delay = float(event[Packets.FIELD_RESPAWN_DELAY])
	hud_flow.set_dead(respawn_delay)


func _should_restore_alive_hud(state: Dictionary) -> bool:
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
