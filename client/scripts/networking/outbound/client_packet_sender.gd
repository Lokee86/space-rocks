extends RefCounted

const NetworkClient = preload("res://scripts/networking/network_client.gd")
const LobbyClientPackets = preload("res://scripts/networking/outbound/lobby_client_packets.gd")
const GameplayClientPackets = preload("res://scripts/networking/outbound/gameplay_client_packets.gd")
const DevtoolsClientPackets = preload("res://scripts/networking/outbound/devtools_client_packets.gd")
const TelemetryClientPackets = preload("res://scripts/networking/outbound/telemetry_client_packets.gd")

var network_client: NetworkClient


func _init(client: NetworkClient = null) -> void:
	network_client = client


func configure(client: NetworkClient) -> void:
	network_client = client


func send_packet(packet: Dictionary) -> void:
	if network_client != null:
		network_client.send_raw_packet(packet)


func send_input_packet(packet: Dictionary) -> void:
	send_packet(packet)


# Gameplay
func send_respawn_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(GameplayClientPackets.respawn_packet())


func send_pause_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(GameplayClientPackets.pause_request_packet())


func send_set_target_player_request(target_player_id: String) -> void:
	if network_client != null:
		network_client.send_raw_packet(GameplayClientPackets.set_target_player_request_packet(target_player_id))


func send_select_target_at_position_request(x, y, target_kind, target_id) -> void:
	if network_client != null:
		network_client.send_raw_packet(GameplayClientPackets.select_target_at_position_request_packet(x, y, target_kind, target_id))


func send_clear_target_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(GameplayClientPackets.clear_target_request_packet())


# Devtools
func send_debug_kill_player_request(target_scope: String = "", target_player_id: String = "") -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.debug_kill_player_packet(target_scope, target_player_id))


func send_debug_kill_target_player_request(target_player_id: String, target_scope: String = "") -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.debug_kill_target_player_packet(target_player_id, target_scope))


func send_toggle_debug_invincible_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.toggle_debug_invincible_packet())


func send_toggle_debug_invincible_target_player_request(target_player_id: String) -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.toggle_debug_invincible_target_player_packet(target_player_id))


func send_toggle_debug_infinite_lives_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.toggle_debug_infinite_lives_packet())


func send_toggle_debug_infinite_lives_target_player_request(target_player_id: String) -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.toggle_debug_infinite_lives_target_player_packet(target_player_id))


func send_toggle_debug_freeze_world_request(freeze_target: String = "") -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.toggle_debug_freeze_world_target_packet(freeze_target))


func send_toggle_debug_freeze_player_request(target_scope: String = "", target_player_id: String = "") -> void:
	if network_client != null:
		if target_scope == "" && target_player_id == "":
			network_client.send_raw_packet(DevtoolsClientPackets.toggle_debug_freeze_player_packet())
		else:
			network_client.send_raw_packet(DevtoolsClientPackets.toggle_debug_freeze_player_target_player_packet(target_player_id))


func send_debug_set_score_request(_target_scope: String, target_player_id: String, score: int) -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.debug_set_score_packet(target_player_id, score))


func send_debug_add_score_request(_target_scope: String, target_player_id: String, amount: int) -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.debug_add_score_packet(target_player_id, amount))


func send_debug_set_lives_request(_target_scope: String, target_player_id: String, lives: int) -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.debug_set_lives_packet(target_player_id, lives))


func send_debug_add_lives_request(_target_scope: String, target_player_id: String, amount: int) -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.debug_add_lives_packet(target_player_id, amount))


func send_debug_clear_bullets_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.debug_clear_bullets_packet())


func send_debug_clear_asteroids_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(DevtoolsClientPackets.debug_clear_asteroids_packet())


# Lobby
func send_create_room_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(LobbyClientPackets.create_room_request_packet())


func send_join_room_request(room_code: String) -> void:
	if network_client != null:
		network_client.send_raw_packet(LobbyClientPackets.join_room_request_packet(room_code))


func send_leave_room_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(LobbyClientPackets.leave_room_request_packet())


func send_set_ready_request(is_ready: bool) -> void:
	if network_client != null:
		network_client.send_raw_packet(LobbyClientPackets.set_ready_request_packet(is_ready))


func send_start_game_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(LobbyClientPackets.start_game_request_packet())


func send_start_single_player_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(LobbyClientPackets.start_single_player_request_packet())


func send_return_to_lobby_request() -> void:
	if network_client != null:
		network_client.send_raw_packet(LobbyClientPackets.return_to_lobby_request_packet())


# Telemetry
func send_telemetry_ping(sequence: int, client_sent_msec: int) -> void:
	if network_client != null:
		network_client.send_raw_packet(TelemetryClientPackets.telemetry_ping_packet(sequence, client_sent_msec))
