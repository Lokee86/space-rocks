extends RefCounted
class_name GameplayEventLifecycleFlow

var event_flow
var death_flow


func configure(
	game_owner: Node2D,
	hud: Control,
	hud_flow_ref,
	menu_flow_ref,
	player_ref,
	visual_position_for_server_position: Callable,
	event_flow_ref = null,
	death_flow_ref = null
) -> void:
	event_flow = event_flow_ref if event_flow_ref != null else GameplayEventFlow.new()
	event_flow.configure(
		game_owner,
		hud,
		visual_position_for_server_position
	)
	death_flow = death_flow_ref if death_flow_ref != null else GameplayDeathFlow.new()
	death_flow.configure(hud_flow_ref, menu_flow_ref, event_flow, player_ref)
	event_flow.self_death_event.connect(Callable(death_flow, "apply_self_death_event"))


func apply_server_events(state: Dictionary) -> void:
	if event_flow != null:
		event_flow.apply_server_events(state["server_events"], state["self_id"])


func reset() -> void:
	if event_flow != null:
		event_flow.reset()
