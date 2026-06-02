extends GutTest

const GameplayInputContext = preload("res://scripts/gameplay/input/gameplay_input_context.gd")

class FakeMouseActionFlow:
	var call_count := 0
	var last_event: InputEvent = null
	var return_value := false

	func handle_input_event(event: InputEvent) -> bool:
		call_count += 1
		last_event = event
		return return_value

func _left_click_pressed_event() -> InputEventMouseButton:
	var event := InputEventMouseButton.new()
	event.button_index = MOUSE_BUTTON_LEFT
	event.pressed = true
	return event

func test_handle_unhandled_input_returns_false_when_gameplay_state_not_received() -> void:
	var input_context := GameplayInputContext.new()
	var fake_mouse_flow := FakeMouseActionFlow.new()
	input_context.mouse_action_flow = fake_mouse_flow

	var consumed := input_context.handle_unhandled_input(_left_click_pressed_event(), false)

	assert_false(consumed)
	assert(fake_mouse_flow.call_count == 0)

func test_handle_unhandled_input_delegates_to_mouse_action_flow_when_gameplay_state_received() -> void:
	var input_context := GameplayInputContext.new()
	var fake_mouse_flow := FakeMouseActionFlow.new()
	fake_mouse_flow.return_value = true
	input_context.mouse_action_flow = fake_mouse_flow
	var event := _left_click_pressed_event()

	var consumed := input_context.handle_unhandled_input(event, true)

	assert_true(consumed)
	assert_eq(fake_mouse_flow.call_count, 1)
	assert_eq(fake_mouse_flow.last_event, event)
