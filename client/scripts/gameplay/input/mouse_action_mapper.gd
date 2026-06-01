extends RefCounted
class_name MouseActionMapper

const MouseActionNames = preload("res://scripts/gameplay/input/mouse_action_names.gd")

static func action_for_event(event: InputEvent, has_pending_context: bool) -> StringName:
	var mouse_button_event := event as InputEventMouseButton
	if mouse_button_event != null and mouse_button_event.pressed:
		if mouse_button_event.button_index == MOUSE_BUTTON_LEFT:
			if has_pending_context:
				return MouseActionNames.SPAWN_ENTITY
			return MouseActionNames.SELECT_TARGET
		if mouse_button_event.button_index == MOUSE_BUTTON_RIGHT:
			if has_pending_context:
				return MouseActionNames.CANCEL_ACTION
			return MouseActionNames.DESELECT_TARGET

	if event.is_action_pressed(MouseActionNames.CANCEL_ACTION_INPUT):
		if has_pending_context:
			return MouseActionNames.CANCEL_ACTION
		return MouseActionNames.NONE

	return MouseActionNames.NONE
