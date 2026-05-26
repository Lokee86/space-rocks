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
var event_flow
var has_received_state := false


func configure(
	connection_service_ref,
	game_owner: Node2D,
	player_ref: Player,
	bullets: Node2D,
	asteroids: Node2D,
	hud_flow_ref,
	game_over_sound: AudioStreamPlayer
) -> void:
	connection_service = connection_service_ref
	player = player_ref
	hud_flow = hud_flow_ref
	world_sync = WorldSyncScript.new()
	world_sync.configure(game_owner, player_ref, bullets, asteroids)
	event_flow = GameplayEventFlow.new()
	event_flow.configure(
		game_owner,
		game_over_sound,
		Callable(world_sync, "visual_position_for_server_position")
	)


func reset() -> void:
	has_received_state = false
	if player != null:
		player.hide()
	if hud_flow != null:
		hud_flow.reset()
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
	if event_flow != null:
		event_flow.apply_server_events(state["server_events"], state["self_id"])
	has_received_state = true
	if is_first_gameplay_state:
		gameplay_started.emit()


func process(delta: float) -> void:
	if world_sync != null:
		world_sync.interpolate(delta)

	if has_received_state && player != null && connection_service != null:
		connection_service.send_input_packet(player.get_input_packet())
