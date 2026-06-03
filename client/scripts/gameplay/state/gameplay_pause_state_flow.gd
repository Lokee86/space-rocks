extends RefCounted
class_name GameplayPauseStateFlow

const PlayerPauseStatePacketReader = preload("res://scripts/gameplay/state/player_pause_state_packet_reader.gd")
const PlayerPauseStateTracker = preload("res://scripts/gameplay/state/player_pause_state_tracker.gd")

var player_pause_state_tracker := PlayerPauseStateTracker.new()


func tracker() -> PlayerPauseStateTracker:
	return player_pause_state_tracker


func reset() -> void:
	player_pause_state_tracker.reset()


func apply_packet(packet: Dictionary) -> void:
	if !PlayerPauseStatePacketReader.is_player_pause_state(packet):
		return
	var state := PlayerPauseStatePacketReader.read(packet)
	player_pause_state_tracker.apply_state(state)
