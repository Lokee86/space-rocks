extends RefCounted

const LobbyClientPackets = preload("res://scripts/networking/outbound/lobby_client_packets.gd")
const GameplayClientPackets = preload("res://scripts/networking/outbound/gameplay_client_packets.gd")
const DevtoolsClientPackets = preload("res://scripts/networking/outbound/devtools_client_packets.gd")
const TelemetryClientPackets = preload("res://scripts/networking/outbound/telemetry_client_packets.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

var network_client: NetworkClient:
	set(value):
		network_client = value
		_missing_network_client_reported = false
var _missing_network_client_reported := false


func _init(client: NetworkClient = null) -> void:
	network_client = client


func configure(client: NetworkClient) -> void:
	network_client = client


func _can_send() -> bool:
	if network_client != null:
		return true
	if !_missing_network_client_reported:
		_missing_network_client_reported = true
		ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_CLIENT_DEPENDENCY_UNAVAILABLE,
			"",
			{},
			{
				"subsystem": "networking_outbound",
				"dependency": "network_client",
				"failure_mode": "not_configured",
			}
		)
	return false


func send_packet(packet: Dictionary, trace_id: String = "") -> void:
	if _can_send():
		network_client.send_raw_packet(packet, trace_id)


func send_input_packet(packet: Dictionary, trace_id: String = "") -> void:
	send_packet(packet, trace_id)

func send_resync_request(match_id: String, lane: String, baseline_id: String, sequence, reason: String) -> void:
	send_packet(Packets.resync_request_packet(match_id, lane, baseline_id, sequence, reason))


func send_authenticate_request(token: String, trace_id: String = "") -> void:
	if token.is_empty():
		return
	send_packet(Packets.authenticate_request_packet(token, trace_id), trace_id)


func send_webrtc_offer(description_type: String, sdp: String) -> void:
	send_packet({
			"type": "webrtc_offer",
			"description_type": description_type,
			"sdp": sdp,
		})


func send_webrtc_ice_candidate(media: String, index: int, name: String) -> void:
	send_packet({
			"type": "webrtc_ice_candidate",
			"media": media,
			"index": index,
			"name": name,
		})


func send_webrtc_smoke(smoke_id: String, origin: String, message: String) -> void:
	send_packet({
			"type": "webrtc_smoke",
			"smoke_id": smoke_id,
			"origin": origin,
			"message": message,
		})


func send_webrtc_failed(error_code: String, message: String) -> void:
	send_packet({
			"type": "webrtc_failed",
			"error_code": error_code,
			"message": message,
		})


# Gameplay
func send_respawn_request() -> void:
	if _can_send():
		send_packet(GameplayClientPackets.respawn_packet())


func send_pause_request() -> void:
	send_packet(GameplayClientPackets.pause_request_packet())


func send_set_target_player_request(target_id: String) -> void:
	send_packet(GameplayClientPackets.set_target_player_request_packet("player", target_id))


func send_select_target_at_position_request(x, y, target_kind, target_id) -> void:
	send_packet(GameplayClientPackets.select_target_at_position_request_packet(x, y, target_kind, target_id))


func send_clear_target_request() -> void:
	send_packet(GameplayClientPackets.clear_target_request_packet())


# Devtools
func send_debug_kill_player_request(target_scope: String = "", target_player_id: String = "") -> void:
	send_packet(DevtoolsClientPackets.debug_kill_player_packet(target_scope, target_player_id))


func send_debug_kill_target_player_request(target_player_id: String, target_scope: String = "") -> void:
	send_packet(DevtoolsClientPackets.debug_kill_target_player_packet(target_player_id, target_scope))


func send_toggle_debug_invincible_request() -> void:
	send_packet(DevtoolsClientPackets.toggle_debug_invincible_packet())


func send_toggle_debug_invincible_target_player_request(target_player_id: String) -> void:
	send_packet(DevtoolsClientPackets.toggle_debug_invincible_target_player_packet(target_player_id))


func send_toggle_debug_infinite_lives_request() -> void:
	send_packet(DevtoolsClientPackets.toggle_debug_infinite_lives_packet())


func send_toggle_debug_infinite_lives_target_player_request(target_player_id: String) -> void:
	send_packet(DevtoolsClientPackets.toggle_debug_infinite_lives_target_player_packet(target_player_id))


func send_toggle_debug_freeze_world_request(freeze_target: String = "") -> void:
	send_packet(DevtoolsClientPackets.toggle_debug_freeze_world_target_packet(freeze_target))


func send_toggle_debug_freeze_player_request(target_scope: String = "", target_player_id: String = "") -> void:
	if target_scope == "" && target_player_id == "":
		send_packet(DevtoolsClientPackets.toggle_debug_freeze_player_packet())
	else:
		send_packet(DevtoolsClientPackets.toggle_debug_freeze_player_target_player_packet(target_player_id))


func send_debug_set_score_request(_target_scope: String, target_player_id: String, score: int) -> void:
	send_packet(DevtoolsClientPackets.debug_set_score_packet(target_player_id, score))


func send_debug_add_score_request(_target_scope: String, target_player_id: String, amount: int) -> void:
	send_packet(DevtoolsClientPackets.debug_add_score_packet(target_player_id, amount))


func send_debug_set_lives_request(_target_scope: String, target_player_id: String, lives: int) -> void:
	send_packet(DevtoolsClientPackets.debug_set_lives_packet(target_player_id, lives))


func send_debug_add_lives_request(_target_scope: String, target_player_id: String, amount: int) -> void:
	send_packet(DevtoolsClientPackets.debug_add_lives_packet(target_player_id, amount))


func send_debug_clear_bullets_request() -> void:
	send_packet(DevtoolsClientPackets.debug_clear_bullets_packet())


func send_debug_clear_asteroids_request() -> void:
	send_packet(DevtoolsClientPackets.debug_clear_asteroids_packet())


# Lobby
func send_create_room_request() -> void:
	send_packet(LobbyClientPackets.create_room_request_packet())


func send_join_room_request(room_code: String) -> void:
	send_packet(LobbyClientPackets.join_room_request_packet(room_code))


func send_leave_room_request() -> void:
	send_packet(LobbyClientPackets.leave_room_request_packet())


func send_set_ready_request(is_ready: bool) -> void:
	send_packet(LobbyClientPackets.set_ready_request_packet(is_ready))


func send_start_game_request() -> void:
	send_packet(LobbyClientPackets.start_game_request_packet())


func send_start_single_player_request(local_profile_id := "") -> void:
	send_packet(LobbyClientPackets.start_single_player_request_packet(local_profile_id))


func send_return_to_lobby_request() -> void:
	send_packet(LobbyClientPackets.return_to_lobby_request_packet())


# Telemetry
func send_telemetry_ping(sequence: int, client_sent_msec: int) -> void:
	send_packet(TelemetryClientPackets.telemetry_ping_packet(sequence, client_sent_msec))
