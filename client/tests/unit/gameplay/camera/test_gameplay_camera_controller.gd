extends GutTest

const GameplayCameraControllerScript := preload("res://scripts/gameplay/camera/gameplay_camera_controller.gd")

var controller
var camera: Camera2D
var sent_sizes: Array[Vector2] = []
var gameplay_active := true


func before_each() -> void:
	sent_sizes.clear()
	gameplay_active = true
	camera = Camera2D.new()
	controller = GameplayCameraControllerScript.new()
	add_child_autofree(camera)
	add_child_autofree(controller)
	controller.configure(
		camera,
		Callable(self, "_record_client_config"),
		Callable(self, "_is_gameplay_active")
	)


func test_mouse_wheel_up_zooms_in_and_reports_smaller_view() -> void:
	assert_true(controller.handle_input(_wheel_event(MOUSE_BUTTON_WHEEL_UP)))
	assert_almost_eq(controller.current_zoom(), 1.1, 0.0001)
	assert_almost_eq(camera.zoom.x, 1.1, 0.0001)
	assert_almost_eq(camera.zoom.y, 1.1, 0.0001)
	assert_eq(sent_sizes.size(), 1)
	assert_almost_eq(sent_sizes[0].x, 1280.0 / 1.1, 0.001)
	assert_almost_eq(sent_sizes[0].y, 720.0 / 1.1, 0.001)


func test_mouse_wheel_down_zooms_out_and_reports_larger_view() -> void:
	assert_true(controller.handle_input(_wheel_event(MOUSE_BUTTON_WHEEL_DOWN)))
	assert_almost_eq(controller.current_zoom(), 0.9, 0.0001)
	assert_gt(sent_sizes[0].x, 1280.0)
	assert_gt(sent_sizes[0].y, 720.0)


func test_zoom_is_clamped_between_half_and_double() -> void:
	controller.set_zoom_level(10.0)
	assert_almost_eq(controller.current_zoom(), 2.0, 0.0001)
	assert_eq(controller.visible_world_size(), Vector2(640.0, 360.0))

	controller.set_zoom_level(-10.0)
	assert_almost_eq(controller.current_zoom(), 0.5, 0.0001)
	assert_eq(controller.visible_world_size(), Vector2(2560.0, 1440.0))


func test_wheel_input_is_ignored_outside_gameplay() -> void:
	gameplay_active = false

	assert_false(controller.handle_input(_wheel_event(MOUSE_BUTTON_WHEEL_UP)))
	assert_almost_eq(controller.current_zoom(), 1.0, 0.0001)
	assert_true(sent_sizes.is_empty())


func _record_client_config() -> void:
	sent_sizes.append(controller.visible_world_size())


func _is_gameplay_active() -> bool:
	return gameplay_active


func _wheel_event(button_index: MouseButton) -> InputEventMouseButton:
	var event := InputEventMouseButton.new()
	event.button_index = button_index
	event.pressed = true
	return event
