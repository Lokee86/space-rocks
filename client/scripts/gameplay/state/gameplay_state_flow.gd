extends RefCounted

class_name GameplayStateFlow

const GameplayStatePacketReader := preload("res://scripts/gameplay/state/gameplay_state_packet_reader.gd")

var gameplay_composition
var has_received_gameplay_state := false


func configure(gameplay_composition_ref) -> void:
	gameplay_composition = gameplay_composition_ref


func handle_gameplay_state_packet(packet: Dictionary) -> void:
	has_received_gameplay_state = true
	var state = GameplayStatePacketReader.read(packet)
	if gameplay_composition != null:
		gameplay_composition.apply_gameplay_state(state)


func has_received_state() -> bool:
	return has_received_gameplay_state


func reset() -> void:
	has_received_gameplay_state = false
