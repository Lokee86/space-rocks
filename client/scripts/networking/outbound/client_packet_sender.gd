extends RefCounted

const LobbyClientPackets = preload("res://scripts/networking/outbound/lobby_client_packets.gd")
const GameplayClientPackets = preload("res://scripts/networking/outbound/gameplay_client_packets.gd")
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


# Lobby
func send_create_room_request(trace_id := "") -> void:
	send_packet(LobbyClientPackets.create_room_request_packet(trace_id))


func send_join_room_request(room_code: String, trace_id := "") -> void:
	send_packet(LobbyClientPackets.join_room_request_packet(room_code, trace_id))


func send_leave_room_request() -> void:
	send_packet(LobbyClientPackets.leave_room_request_packet())


func send_set_ready_request(is_ready: bool) -> void:
	send_packet(LobbyClientPackets.set_ready_request_packet(is_ready))


func send_start_game_request() -> void:
	send_packet(LobbyClientPackets.start_game_request_packet())


func send_add_bot_request() -> void:
	send_packet(LobbyClientPackets.add_bot_request_packet())


func send_remove_room_member_request(player_id: String) -> void:
	send_packet(LobbyClientPackets.remove_room_member_request_packet(player_id))


func send_start_single_player_request(local_profile_id := "", trace_id := "") -> void:
	send_packet(LobbyClientPackets.start_single_player_request_packet(local_profile_id, trace_id))


func send_return_to_lobby_request() -> void:
	send_packet(LobbyClientPackets.return_to_lobby_request_packet())
