extends RefCounted
class_name GameplayStateApplyFlow

const GameplayStateApplyResultScript = preload("res://scripts/gameplay/state/gameplay_state_apply_result.gd")
const GameplayWorldStateApplyFlowScript = preload("res://scripts/gameplay/runtime/gameplay_world_state_apply_flow.gd")

var input_context
var devtools_context
var hud_flow
var world_state_apply_flow
var event_lifecycle_flow
var alive_restore_flow
var gameplay_readiness


func configure(
	input_context_ref,
	devtools_context_ref,
	hud_flow_ref,
	world_sync_ref,
	event_lifecycle_flow_ref,
	alive_restore_flow_ref,
	gameplay_readiness_ref = null
) -> void:
	input_context = input_context_ref
	devtools_context = devtools_context_ref
	hud_flow = hud_flow_ref
	world_state_apply_flow = GameplayWorldStateApplyFlowScript.new()
	world_state_apply_flow.configure(world_sync_ref)
	event_lifecycle_flow = event_lifecycle_flow_ref
	alive_restore_flow = alive_restore_flow_ref
	gameplay_readiness = gameplay_readiness_ref


func apply_state(state: Dictionary, has_received_state: bool) -> GameplayStateApplyResult:
	var result: GameplayStateApplyResult = GameplayStateApplyResultScript.new()
	if devtools_context != null:
		devtools_context.apply_gameplay_state(state)
	if hud_flow != null:
		hud_flow.apply_gameplay_state_summary(state)
	if world_state_apply_flow != null:
		world_state_apply_flow.apply_world_state(state, has_received_state)
	if alive_restore_flow != null:
		alive_restore_flow.apply_state(state)
	if event_lifecycle_flow != null:
		event_lifecycle_flow.apply_server_events(state)
	if gameplay_readiness != null:
		gameplay_readiness.apply_legacy_state_compatibility_baseline()
	result.has_received_state = has_received_state if gameplay_readiness == null else gameplay_readiness.is_gameplay_ready()
	result.started_gameplay = result.has_received_state
	return result
