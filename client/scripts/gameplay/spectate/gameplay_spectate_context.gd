extends RefCounted
class_name GameplaySpectateContext

const GameplaySpectateFlow = preload("res://scripts/gameplay/spectate/gameplay_spectate_flow.gd")

var spectate_flow


func configure(menu_flow_ref, spectate_menu_state_ref, world_sync_ref) -> void:
	if spectate_flow == null:
		spectate_flow = GameplaySpectateFlow.new()
	spectate_flow.configure(menu_flow_ref, spectate_menu_state_ref, world_sync_ref)


func reset() -> void:
	if spectate_flow != null:
		spectate_flow.reset()


func process() -> void:
	if spectate_flow != null:
		spectate_flow.process()


func begin_spectating() -> void:
	if spectate_flow != null:
		spectate_flow.begin_spectating()
