extends GutTest

const MouseActionFlow = preload("res://scripts/gameplay/input/mouse_action_flow.gd")

class FakeTargetRequestFlow:
	var select_calls := 0
	var deselect_calls := 0

	func select_target() -> void:
		select_calls += 1

	func deselect_target() -> void:
		deselect_calls += 1

class FakePendingCallbacks:
	var action_calls := 0
	var cancel_calls := 0

	func record_action() -> void:
		action_calls += 1

	func record_cancel() -> void:
		cancel_calls += 1

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
	event.action = "CancelAction"
	event.pressed = true
	return event

func test_left_click_without_pending_context_calls_select_once_and_consumes() -> void:
	var fake_target_flow := FakeTargetRequestFlow.new()
	var flow := MouseActionFlow.new()
	flow.configure(fake_target_flow)

	var consumed := flow.handle_input_event(_left_click_pressed_event())

	assert_true(consumed)
	assert_eq(fake_target_flow.select_calls, 1)
	assert_eq(fake_target_flow.deselect_calls, 0)

func test_right_click_without_pending_context_calls_deselect_once_and_consumes() -> void:
	var fake_target_flow := FakeTargetRequestFlow.new()
	var flow := MouseActionFlow.new()
	flow.configure(fake_target_flow)

	var consumed := flow.handle_input_event(_right_click_pressed_event())

	assert_true(consumed)
	assert_eq(fake_target_flow.select_calls, 0)
	assert_eq(fake_target_flow.deselect_calls, 1)

func test_left_click_with_pending_context_calls_pending_action_not_target_select() -> void:
	var fake_target_flow := FakeTargetRequestFlow.new()
	var flow := MouseActionFlow.new()
	flow.configure(fake_target_flow)
	var callbacks := FakePendingCallbacks.new()

	flow.set_pending_context(
		Callable(callbacks, "record_action"),
		Callable(callbacks, "record_cancel")
	)

	var consumed := flow.handle_input_event(_left_click_pressed_event())

	assert_true(consumed)
	assert_eq(callbacks.action_calls, 1)
	assert_eq(callbacks.cancel_calls, 0)
	assert_eq(fake_target_flow.select_calls, 0)
	assert_eq(fake_target_flow.deselect_calls, 0)
	assert_true(flow.has_pending_context())

func test_right_click_with_pending_context_calls_pending_cancel_clears_context_and_not_target_deselect() -> void:
	var fake_target_flow := FakeTargetRequestFlow.new()
	var flow := MouseActionFlow.new()
	flow.configure(fake_target_flow)
	var callbacks := FakePendingCallbacks.new()

	flow.set_pending_context(
		Callable(callbacks, "record_action"),
		Callable(callbacks, "record_cancel")
	)

	var consumed := flow.handle_input_event(_right_click_pressed_event())

	assert_true(consumed)
	assert_eq(callbacks.action_calls, 0)
	assert_eq(callbacks.cancel_calls, 1)
	assert_eq(fake_target_flow.select_calls, 0)
	assert_eq(fake_target_flow.deselect_calls, 0)
	assert_false(flow.has_pending_context())

func test_escape_with_pending_context_calls_pending_cancel_and_clears_context() -> void:
	var fake_target_flow := FakeTargetRequestFlow.new()
	var flow := MouseActionFlow.new()
	flow.configure(fake_target_flow)
	var callbacks := FakePendingCallbacks.new()

	flow.set_pending_context(
		Callable(callbacks, "record_action"),
		Callable(callbacks, "record_cancel")
	)

	var consumed := flow.handle_input_event(_cancel_action_pressed_event())

	assert_true(consumed)
	assert_eq(callbacks.action_calls, 0)
	assert_eq(callbacks.cancel_calls, 1)
	assert_eq(fake_target_flow.select_calls, 0)
	assert_eq(fake_target_flow.deselect_calls, 0)
	assert_false(flow.has_pending_context())
