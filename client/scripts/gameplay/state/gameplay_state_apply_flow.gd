extends RefCounted
class_name GameplayStateApplyFlow

const GameplayStateApplyResultScript = preload("res://scripts/gameplay/state/gameplay_state_apply_result.gd")

var input_context
var devtools_context
var hud_flow
var runtime_context
var menu_flow


func configure(input_context_ref, devtools_context_ref, hud_flow_ref, runtime_context_ref, menu_flow_ref) -> void:
	input_context = input_context_ref
	devtools_context = devtools_context_ref
	hud_flow = hud_flow_ref
	runtime_context = runtime_context_ref
	menu_flow = menu_flow_ref


func apply_state(state: Dictionary, has_received_state: bool) -> GameplayStateApplyResult:
	var result: GameplayStateApplyResult = GameplayStateApplyResultScript.new()
	var is_first_gameplay_state := !has_received_state
	if devtools_context != null:
		devtools_context.apply_gameplay_state(state)
	if input_context != null:
		input_context.mark_gameplay_state_received()
	if hud_flow != null:
		hud_flow.apply_gameplay_state_summary(state)
	if runtime_context != null:
		runtime_context.apply_world_state(state, has_received_state)
		runtime_context.apply_respawn_alive_restore(state, menu_flow)
		runtime_context.apply_server_events(state)
	result.has_received_state = true
	result.started_gameplay = is_first_gameplay_state
	return result
