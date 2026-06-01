extends GutTest

const MouseActionMapper = preload("res://scripts/gameplay/input/mouse_action_mapper.gd")
const MouseActionNames = preload("res://scripts/gameplay/input/mouse_action_names.gd")

func _left_click_pressed_event() -> InputEventMouseButton:
	var event := InputEventMouseButton.new()
	event.button_index = MOUSE_BUTTON_LEFT
	event.pressed = true
	return event

func _right_click_pressed_event() -> InputEventMouseButton:
	var event := InputEventMouseButton.new()
	event.button_index = MOUSE_BUTTON_RIGHT
	event.pressed = true
	return event

func _cancel_action_pressed_event() -> InputEventAction:
	var event := InputEventAction.new()
	event.action = MouseActionNames.CANCEL_ACTION_INPUT
	event.pressed = true
	return event

func test_left_click_without_pending_context_returns_select_target() -> void:
	var action: StringName = MouseActionMapper.action_for_event(_left_click_pressed_event(), false)
	assert_eq(action, MouseActionNames.SELECT_TARGET)

func test_left_click_with_pending_context_returns_spawn_entity() -> void:
	var action: StringName = MouseActionMapper.action_for_event(_left_click_pressed_event(), true)
	assert_eq(action, MouseActionNames.SPAWN_ENTITY)

func test_right_click_without_pending_context_returns_deselect_target() -> void:
	var action: StringName = MouseActionMapper.action_for_event(_right_click_pressed_event(), false)
	assert_eq(action, MouseActionNames.DESELECT_TARGET)

func test_right_click_with_pending_context_returns_cancel_action() -> void:
	var action: StringName = MouseActionMapper.action_for_event(_right_click_pressed_event(), true)
	assert_eq(action, MouseActionNames.CANCEL_ACTION)

func test_escape_with_pending_context_returns_cancel_action() -> void:
	var action: StringName = MouseActionMapper.action_for_event(_cancel_action_pressed_event(), true)
	assert_eq(action, MouseActionNames.CANCEL_ACTION)

func test_escape_without_pending_context_returns_none() -> void:
	var action: StringName = MouseActionMapper.action_for_event(_cancel_action_pressed_event(), false)
	assert_eq(action, MouseActionNames.NONE)
