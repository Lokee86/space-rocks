extends RefCounted
class_name DevToolsSessionFlow

const DebugKillInputFlow := preload("res://scripts/gameplay/devtools/debug_kill_input_flow.gd")
const DevConnectionService := preload("res://scripts/devtools/dev_connection_service.gd")
const DebugMouseWorldPosition := preload("res://scripts/gameplay/devtools/debug_mouse_world_position.gd")
const DebugClickPlacementFlow := preload("res://scripts/gameplay/devtools/debug_click_placement_flow.gd")
const DebugContinuousBulletSpawnFlow := preload("res://scripts/gameplay/devtools/debug_continuous_bullet_spawn_flow.gd")

const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")


var connection_service
var scene_root: Node
var gameplay_shell_flow
var world_sync
var debug_kill_input_flow
var dev_connection_service
var debug_mouse_world_position
var debug_click_placement_flow
var debug_continuous_bullet_spawn_flow


func configure(connection_service_ref, scene_root_ref: Node, gameplay_shell_flow_ref, world_sync_ref) -> void:
	connection_service = connection_service_ref
	scene_root = scene_root_ref
	gameplay_shell_flow = gameplay_shell_flow_ref
	world_sync = world_sync_ref
	debug_kill_input_flow = DebugKillInputFlow.new()
	debug_kill_input_flow.configure()
	dev_connection_service = DevConnectionService.new()
	dev_connection_service.configure(connection_service)
	if scene_root is Node2D && world_sync != null:
		debug_mouse_world_position = DebugMouseWorldPosition.new()
		debug_mouse_world_position.configure(
			scene_root,
			Callable(world_sync, "server_position_for_visual_position")
		)
		debug_click_placement_flow = DebugClickPlacementFlow.new()
		debug_click_placement_flow.configure(debug_mouse_world_position)
		debug_click_placement_flow.placement_completed.connect(
			Callable(self, "_on_debug_click_placement_completed")
		)
		debug_click_placement_flow.placement_cancelled.connect(
			Callable(self, "_on_debug_click_placement_cancelled")
		)
		debug_continuous_bullet_spawn_flow = DebugContinuousBulletSpawnFlow.new()
		debug_continuous_bullet_spawn_flow.configure(debug_mouse_world_position)
		debug_continuous_bullet_spawn_flow.placement_completed.connect(
			Callable(self, "_on_debug_continuous_bullet_spawn_completed")
		)
		debug_continuous_bullet_spawn_flow.placement_cancelled.connect(
			Callable(self, "_on_debug_continuous_bullet_spawn_cancelled")
		)


func attach_to_gameplay_shell(gameplay_shell_flow_ref) -> void:
	if gameplay_shell_flow_ref == null:
		return
	if not gameplay_shell_flow_ref.has_method("configure_devtools_placement_request_route"):
		return
	gameplay_shell_flow_ref.configure_devtools_placement_request_route(
		Callable(self, "begin_debug_click_placement")
	)


func configure_kill_player_route(route: Callable) -> void:
	if debug_kill_input_flow != null:
		debug_kill_input_flow.configure_kill_player_route(route)


func process(delta: float) -> void:
	if debug_kill_input_flow != null:
		debug_kill_input_flow.process()
	if debug_continuous_bullet_spawn_flow != null:
		debug_continuous_bullet_spawn_flow.process(delta)


func handle_input(event: InputEvent) -> bool:
	if debug_continuous_bullet_spawn_flow != null and debug_continuous_bullet_spawn_flow.is_active():
		if debug_continuous_bullet_spawn_flow.handle_unhandled_input(event):
			return true

	if debug_click_placement_flow != null and debug_click_placement_flow.is_active():
		if debug_click_placement_flow.handle_unhandled_input(event):
			return true
	return false


func begin_debug_click_placement(action_name: StringName, placement_context: Dictionary = {}) -> void:
	if action_name == &"continuous_spawn_bullet":
		if debug_continuous_bullet_spawn_flow == null:
			return
		debug_continuous_bullet_spawn_flow.begin(placement_context)
		return
	if debug_click_placement_flow == null:
		return
	debug_click_placement_flow.begin(action_name, placement_context)


func reset() -> void:
	if debug_kill_input_flow != null and debug_kill_input_flow.has_method("reset"):
		debug_kill_input_flow.reset()
	if debug_mouse_world_position != null and debug_mouse_world_position.has_method("reset"):
		debug_mouse_world_position.reset()
	if debug_click_placement_flow != null and debug_click_placement_flow.has_method("reset"):
		debug_click_placement_flow.reset()
	if debug_click_placement_flow != null and debug_click_placement_flow.has_method("cancel"):
		debug_click_placement_flow.cancel()
	if debug_continuous_bullet_spawn_flow != null and debug_continuous_bullet_spawn_flow.has_method("reset"):
		debug_continuous_bullet_spawn_flow.reset()
	if debug_continuous_bullet_spawn_flow != null and debug_continuous_bullet_spawn_flow.has_method("cancel"):
		debug_continuous_bullet_spawn_flow.cancel()
	if dev_connection_service != null and dev_connection_service.has_method("reset"):
		dev_connection_service.reset()


func _on_debug_click_placement_completed(result: Dictionary) -> void:
	if gameplay_shell_flow != null && gameplay_shell_flow.has_method("handle_devtools_placement_result"):
		gameplay_shell_flow.handle_devtools_placement_result(result)


func _on_debug_click_placement_cancelled(_action_name: StringName) -> void:
	return


func _on_debug_continuous_bullet_spawn_cancelled(_action_name: StringName) -> void:
	return


func _on_debug_continuous_bullet_spawn_completed(result: Dictionary) -> void:
	if dev_connection_service == null || !dev_connection_service.has_method("send_begin_continuous_bullet_stream_from_placement_result"):
		return
	var operation_trace = result.get("_operation_trace", null)
	if not operation_trace is ClientOperationTrace:
		return
	var packet := DevSpawnPacketBuilder.build_continuous_bullet_stream_from_placement_result(result)
	if packet.is_empty():
		ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_DEVTOOLS_COMMAND_REJECTED,
			"",
			{"trace_id": operation_trace.trace_id()},
			{
				"command_type": DevSpawnPacketBuilder.TYPE_DEBUG_BEGIN_CONTINUOUS_BULLET_STREAM,
				"reason": "placement_result_invalid",
			}
		)
		return
	dev_connection_service.send_begin_continuous_bullet_stream_from_placement_result(result, operation_trace)
