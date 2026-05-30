extends RefCounted

const LobbyPacketReader := preload("res://scripts/lobby/lobby_packet_reader.gd")
const LobbySessionState := preload("res://scripts/lobby/lobby_session_state.gd")

var lobby_state: LobbySessionState = LobbySessionState.new()


func apply_room_snapshot(packet: Dictionary) -> String:
	lobby_state.apply_snapshot(
		LobbyPacketReader.room_code(packet),
		LobbyPacketReader.room_state(packet),
		LobbyPacketReader.local_player_id(packet),
		LobbyPacketReader.owner_id(packet),
		LobbyPacketReader.max_players(packet),
		LobbyPacketReader.members(packet)
	)
	return lobby_state.summary()


func current_state() -> LobbySessionState:
	return lobby_state


func clear() -> void:
	lobby_state.clear()
