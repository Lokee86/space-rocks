extends RefCounted

signal placement_completed(result: Dictionary)
signal placement_cancelled(action_name: StringName)

const ACTION_NAME := &"spawn_bullet"
const DRAG_DIRECTION_THRESHOLD := 8.0

var mouse_world_position = null
var active := false
var placement_context := {}
var placement_origin_result := {}


func configure(mouse_world_position_ref) -> void:
	mouse_world_position = mouse_world_position_ref


func begin(context: Dictionary = {}) -> void:
	active = true
	placement_context = context.duplicate(true)
	placement_origin_result = {}


func cancel() -> void:
	if not is_active():
		return
	active = false
	placement_context = {}
	placement_origin_result = {}
	placement_cancelled.emit(ACTION_NAME)


func is_active() -> bool:
	return active


func handle_unhandled_input(event: InputEvent) -> bool:
	if not is_active():
		return false

	if event.is_action_pressed(MouseActionNames.CANCEL_ACTION_INPUT):
		cancel()
		return true

	if event.is_action_pressed(MouseActionNames.SPAWN_ENTITY_INPUT):
		var position_result = mouse_world_position.current_position()
		if position_result.valid:
			placement_origin_result = position_result.duplicate(true)
		else:
			placement_origin_result = {}
		return true

	if event.is_action_released(MouseActionNames.SPAWN_ENTITY_INPUT):
		if placement_origin_result.get("valid", false):
			var current_position_result = mouse_world_position.current_position()
			if current_position_result.valid:
				var current_visual_position: Vector2 = current_position_result.get("visual_position", Vector2.ZERO)
				var origin_visual_position: Vector2 = placement_origin_result.get("visual_position", Vector2.ZERO)
				var drag_visual: Vector2 = current_visual_position - origin_visual_position
				if drag_visual.length() > DRAG_DIRECTION_THRESHOLD:
					var result := {
						"action_name": ACTION_NAME,
						"server_position": placement_origin_result.server_position,
						"visual_position": placement_origin_result.visual_position,
						"has_direction": true,
						"direction": drag_visual.normalized()
					}
					if placement_context.has("target_player_id"):
						result["target_player_id"] = placement_context["target_player_id"]
					placement_completed.emit(result)
		active = false
		placement_context = {}
		placement_origin_result = {}
		return true

	return false


func process(_delta: float) -> void:
	return

