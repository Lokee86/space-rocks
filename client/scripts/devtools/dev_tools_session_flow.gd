extends RefCounted
class_name DevToolsSessionFlow

const DebugKillInputFlow := preload("res://scripts/gameplay/devtools/debug_kill_input_flow.gd")
const DevConnectionService := preload("res://scripts/devtools/dev_connection_service.gd")
const DebugMouseWorldPosition := preload("res://scripts/gameplay/devtools/debug_mouse_world_position.gd")
const DebugClickPlacementFlow := preload("res://scripts/gameplay/devtools/debug_click_placement_flow.gd")
const DebugContinuousBulletSpawnFlow := preload("res://scripts/gameplay/devtools/debug_continuous_bullet_spawn_flow.gd")

var connection_service
var scene_root: Node
var gameplay_shell_flow
var logger: Callable
var debug_kill_input_flow
var dev_connection_service
var debug_mouse_world_position
var debug_click_placement_flow
var debug_continuous_bullet_spawn_flow


func configure(connection_service_ref, scene_root_ref: Node, gameplay_shell_flow_ref, logger_callable: Callable) -> void:
	connection_service = connection_service_ref
	scene_root = scene_root_ref
	gameplay_shell_flow = gameplay_shell_flow_ref
	logger = logger_callable
	debug_kill_input_flow = DebugKillInputFlow.new()
	debug_kill_input_flow.configure(connection_service)
	dev_connection_service = DevConnectionService.new()
	dev_connection_service.configure(connection_service)
	if scene_root is Node2D:
		debug_mouse_world_position = DebugMouseWorldPosition.new()
		debug_mouse_world_position.configure(
			scene_root,
			Callable(gameplay_shell_flow, "server_position_for_visual_position")
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
	if !gameplay_shell_flow_ref.has_method("configure_debug_placement_route"):
		return
	gameplay_shell_flow_ref.configure_debug_placement_route(
		Callable(self, "begin_debug_click_placement")
	)


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


func _log(message: String) -> void:
	if logger != null and logger.is_valid():
		logger.call(message)


func _on_debug_click_placement_completed(result: Dictionary) -> void:
	_log(
		"Debug click placement completed: %s at %s has_direction=%s direction=%s"
		% [
			String(result.get("action_name", StringName())),
			str(result.get("server_position", Vector2.ZERO)),
			str(result.get("has_direction", false)),
			str(result.get("direction", Vector2.ZERO))
		]
	)
	if gameplay_shell_flow != null && gameplay_shell_flow.has_method("handle_debug_placement_result"):
		gameplay_shell_flow.handle_debug_placement_result(result)


func _on_debug_click_placement_cancelled(action_name: StringName) -> void:
	_log("Debug click placement cancelled: %s" % String(action_name))


func _on_debug_continuous_bullet_spawn_cancelled(action_name: StringName) -> void:
	_log("Debug continuous bullet spawn cancelled: %s" % String(action_name))


func _on_debug_continuous_bullet_spawn_completed(result: Dictionary) -> void:
	_log(
		"Debug continuous bullet stream placement completed: %s at %s has_direction=%s direction=%s"
		% [
			String(result.get("action_name", StringName())),
			str(result.get("server_position", Vector2.ZERO)),
			str(result.get("has_direction", false)),
			str(result.get("direction", Vector2.ZERO))
		]
	)
	if dev_connection_service != null && dev_connection_service.has_method("send_begin_continuous_bullet_stream_from_placement_result"):
		dev_connection_service.send_begin_continuous_bullet_stream_from_placement_result(result)
