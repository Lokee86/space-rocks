extends RefCounted

var targeting_context = null
var pending_action: Callable = Callable()
var pending_cancel: Callable = Callable()

func configure(targeting_context_ref) -> void:
	targeting_context = targeting_context_ref

func has_pending_context() -> bool:
	return not pending_action.is_null() and not pending_cancel.is_null()

func clear_pending_context() -> void:
	pending_action = Callable()
	pending_cancel = Callable()

func set_pending_context(action: Callable, cancel: Callable) -> void:
	pending_action = action
	pending_cancel = cancel

func handle_input_event(event: InputEvent) -> bool:
	var action = MouseActionMapper.action_for_event(event, has_pending_context())
	if action == MouseActionNames.SPAWN_ENTITY:
		if pending_action != null and not pending_action.is_null():
			pending_action.call()
		return true
	if action == MouseActionNames.CANCEL_ACTION:
		if pending_cancel != null and not pending_cancel.is_null():
			pending_cancel.call()
		clear_pending_context()
		return true
	if action == MouseActionNames.SELECT_TARGET:
		if targeting_context == null:
			return false
		targeting_context.select_target()
		return true
	if action == MouseActionNames.DESELECT_TARGET:
		if targeting_context != null:
			targeting_context.deselect_target()
		return true

	return false

