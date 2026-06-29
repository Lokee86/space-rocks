extends RefCounted
class_name MouseActionMapper

static func action_for_event(event: InputEvent, has_pending_context: bool) -> StringName:
	if has_pending_context:
		if event.is_action_pressed(MouseActionNames.SPAWN_ENTITY_INPUT):
			return MouseActionNames.SPAWN_ENTITY
		if event.is_action_pressed(MouseActionNames.CANCEL_ACTION_INPUT):
			return MouseActionNames.CANCEL_ACTION
	else:
		if event.is_action_pressed(MouseActionNames.SELECT_TARGET_INPUT):
			return MouseActionNames.SELECT_TARGET
		if event.is_action_pressed(MouseActionNames.DESELECT_TARGET_INPUT):
			return MouseActionNames.DESELECT_TARGET

	if event.is_action_pressed(MouseActionNames.CANCEL_ACTION_INPUT):
		return MouseActionNames.NONE

	return MouseActionNames.NONE

