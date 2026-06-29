extends RefCounted
class_name GameplayProcessFlow

var runtime_context
var server_hitbox_overlay_flow
var runtime_tick_flow
var devtools_context
var input_context
var spectate_context


func configure(runtime_context_ref, server_hitbox_overlay_flow_ref, runtime_tick_flow_ref, devtools_context_ref, input_context_ref, spectate_context_ref) -> void:
	runtime_context = runtime_context_ref
	server_hitbox_overlay_flow = server_hitbox_overlay_flow_ref
	runtime_tick_flow = runtime_tick_flow_ref
	devtools_context = devtools_context_ref
	input_context = input_context_ref
	spectate_context = spectate_context_ref


func process(delta: float, required_lane_baselines_synced: bool) -> void:
	if runtime_context != null:
		runtime_context.process(delta)
	if server_hitbox_overlay_flow != null:
		server_hitbox_overlay_flow.process()
	if runtime_tick_flow != null:
		runtime_tick_flow.process(delta)
	if devtools_context != null:
		devtools_context.process(required_lane_baselines_synced)
	if input_context != null:
		input_context.process(required_lane_baselines_synced)
	if spectate_context != null:
		spectate_context.process()

