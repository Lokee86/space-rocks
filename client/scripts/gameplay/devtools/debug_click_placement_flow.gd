extends RefCounted

signal placement_completed(result: Dictionary)
signal placement_cancelled(action_name: StringName)

const DRAG_DIRECTION_THRESHOLD := 8.0

var mouse_world_position = null
var active_action_name := StringName()
var placement_start_result := {}
var placement_context := {}


func configure(mouse_world_position_ref) -> void:
	mouse_world_position = mouse_world_position_ref


func begin(action_name: StringName, context: Dictionary = {}) -> void:
	active_action_name = action_name
	placement_context = context.duplicate(true)


func cancel() -> void:
	if not is_active():
		return
	var cancelled_action_name := active_action_name
	active_action_name = StringName()
	placement_start_result = {}
	placement_context = {}
	placement_cancelled.emit(cancelled_action_name)


func is_active() -> bool:
	return not active_action_name.is_empty()


func handle_unhandled_input(event: InputEvent) -> bool:
	if not is_active():
		return false

	if event.is_action_pressed("ui_cancel"):
		cancel()
		return true

	if event is InputEventMouseButton and event.pressed:
		var mouse_button_event := event as InputEventMouseButton
		if mouse_button_event.button_index == MOUSE_BUTTON_RIGHT:
			cancel()
			return true
		if mouse_button_event.button_index == MOUSE_BUTTON_LEFT:
			var position_result = mouse_world_position.current_position()
			if position_result.valid:
				placement_start_result = position_result
			return true
	if event is InputEventMouseButton and not event.pressed:
		var mouse_button_event := event as InputEventMouseButton
		if mouse_button_event.button_index == MOUSE_BUTTON_LEFT:
			if placement_start_result.get("valid", false):
				var has_direction := false
				var direction := Vector2.ZERO
				var release_result = mouse_world_position.current_position()
				if release_result.valid:
					var release_visual_position: Vector2 = release_result.get("visual_position", Vector2.ZERO)
					var start_visual_position: Vector2 = placement_start_result.get("visual_position", Vector2.ZERO)
					var drag_visual: Vector2 = release_visual_position - start_visual_position
					if drag_visual.length() > DRAG_DIRECTION_THRESHOLD:
						has_direction = true
						direction = drag_visual.normalized()
				var result := {
					"action_name": active_action_name,
					"server_position": placement_start_result.server_position,
					"visual_position": placement_start_result.visual_position,
					"has_direction": has_direction,
					"direction": direction
				}
				if placement_context.has("target_player_id"):
					result["target_player_id"] = placement_context["target_player_id"]
				placement_completed.emit(result)
				active_action_name = StringName()
				placement_start_result = {}
				placement_context = {}
				return true

	return false
