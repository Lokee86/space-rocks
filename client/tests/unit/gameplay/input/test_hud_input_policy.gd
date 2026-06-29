extends GutTest

const HudInputPolicy := preload("res://scripts/gameplay/input/hud_input_policy.gd")


class FakeViewport:
	extends RefCounted

	var hovered_control: Control

	func gui_get_hovered_control() -> Control:
		return hovered_control


func test_gameplay_ui_root_hover_returns_true() -> void:
	var policy := add_child_autofree(HudInputPolicy.new())
	var gameplay_ui_root := _create_gameplay_ui_root()
	var viewport: FakeViewport = _create_viewport(gameplay_ui_root)

	assert_true(
		policy.should_gameplay_ui_receive_mouse_event(
			_mouse_button_event(true),
			gameplay_ui_root,
			viewport
		)
	)


func test_gameplay_ui_descendant_hover_returns_true() -> void:
	var policy := add_child_autofree(HudInputPolicy.new())
	var gameplay_ui_root := _create_gameplay_ui_root()
	var descendant_button := Button.new()
	gameplay_ui_root.add_child(descendant_button)
	var viewport: FakeViewport = _create_viewport(descendant_button)

	assert_true(
		policy.should_gameplay_ui_receive_mouse_event(
			_mouse_button_event(true),
			gameplay_ui_root,
			viewport
		)
	)


func test_gameplay_ui_mouse_event_rejects_non_pressed_button() -> void:
	var policy := add_child_autofree(HudInputPolicy.new())
	var gameplay_ui_root := _create_gameplay_ui_root()
	var viewport: FakeViewport = _create_viewport(gameplay_ui_root)

	assert_false(
		policy.should_gameplay_ui_receive_mouse_event(
			_mouse_button_event(false),
			gameplay_ui_root,
			viewport
		)
	)


func test_gameplay_ui_mouse_event_rejects_null_gameplay_ui_root() -> void:
	var policy := add_child_autofree(HudInputPolicy.new())
	var viewport: FakeViewport = _create_viewport(null)

	assert_false(
		policy.should_gameplay_ui_receive_mouse_event(
			_mouse_button_event(true),
			null,
			viewport
		)
	)


func test_gameplay_ui_mouse_event_rejects_null_viewport() -> void:
	var policy := add_child_autofree(HudInputPolicy.new())
	var gameplay_ui_root := _create_gameplay_ui_root()

	assert_false(
		policy.should_gameplay_ui_receive_mouse_event(
			_mouse_button_event(true),
			gameplay_ui_root,
			null
		)
	)


func test_gameplay_ui_mouse_event_rejects_hover_outside_gameplay_ui() -> void:
	var policy := add_child_autofree(HudInputPolicy.new())
	var gameplay_ui_root := _create_gameplay_ui_root()
	var outside_root := Control.new()
	var outside_button := Button.new()
	outside_root.add_child(outside_button)
	var viewport: FakeViewport = _create_viewport(outside_button)

	add_child_autofree(outside_root)

	assert_false(
		policy.should_gameplay_ui_receive_mouse_event(
			_mouse_button_event(true),
			gameplay_ui_root,
			viewport
		)
	)


func _create_gameplay_ui_root() -> Control:
	var gameplay_ui_root := Control.new()
	add_child_autofree(gameplay_ui_root)
	return gameplay_ui_root


func _create_viewport(hovered_control: Control) -> FakeViewport:
	var viewport := FakeViewport.new()
	viewport.hovered_control = hovered_control
	return viewport


func _mouse_button_event(pressed: bool) -> InputEventMouseButton:
	var event := InputEventMouseButton.new()
	event.button_index = MOUSE_BUTTON_LEFT
	event.pressed = pressed
	return event
