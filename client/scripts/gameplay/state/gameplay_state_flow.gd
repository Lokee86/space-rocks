extends RefCounted
class_name GameplayStateFlow

const GameplayStatePacketReader := preload("res://scripts/gameplay/state/gameplay_state_packet_reader.gd")
const GameplayReadiness := preload("res://scripts/protocol/realtime/gameplay_readiness.gd")

var gameplay_composition
var gameplay_readiness := GameplayReadiness.new()


func configure(gameplay_composition_ref) -> void:
	gameplay_composition = gameplay_composition_ref


func set_gameplay_readiness(readiness: GameplayReadiness) -> void:
	if readiness == null:
		return
	gameplay_readiness = readiness


func handle_gameplay_state_packet(packet: Dictionary) -> void:
	var state = GameplayStatePacketReader.read(packet)
	if gameplay_composition != null:
		gameplay_composition.apply_gameplay_state(state)


func has_received_state() -> bool:
	return gameplay_readiness.is_gameplay_ready()


func reset() -> void:
	gameplay_readiness = GameplayReadiness.new()
