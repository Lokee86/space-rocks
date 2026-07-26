extends Node
class_name GameplayCameraController

const BASE_VISIBLE_WORLD_SIZE := Vector2(1280.0, 720.0)
const MIN_ZOOM := 0.5
const MAX_ZOOM := 2.0
const ZOOM_STEP := 0.1

var camera: Camera2D
var send_client_config := Callable()
var gameplay_active_provider := Callable()
var zoom_level := 1.0


func configure(
	camera_ref: Camera2D,
	send_client_config_ref: Callable,
	gameplay_active_provider_ref: Callable
) -> void:
	camera = camera_ref
	send_client_config = send_client_config_ref
	gameplay_active_provider = gameplay_active_provider_ref
	_apply_zoom()


func _unhandled_input(event: InputEvent) -> void:
	if !handle_input(event):
		return
	get_viewport().set_input_as_handled()


func handle_input(event: InputEvent) -> bool:
	if !_is_gameplay_active():
		return false
	if !(event is InputEventMouseButton) || !event.pressed:
		return false

	match event.button_index:
		MOUSE_BUTTON_WHEEL_UP:
			set_zoom_level(zoom_level + ZOOM_STEP)
			return true
		MOUSE_BUTTON_WHEEL_DOWN:
			set_zoom_level(zoom_level - ZOOM_STEP)
			return true
		_:
			return false


func set_zoom_level(value: float) -> void:
	var next_zoom := clampf(value, MIN_ZOOM, MAX_ZOOM)
	next_zoom = roundf(next_zoom / ZOOM_STEP) * ZOOM_STEP
	if is_equal_approx(next_zoom, zoom_level):
		return

	zoom_level = next_zoom
	_apply_zoom()
	if send_client_config.is_valid():
		send_client_config.call()


func current_zoom() -> float:
	return zoom_level


func visible_world_size() -> Vector2:
	return BASE_VISIBLE_WORLD_SIZE / zoom_level


func _apply_zoom() -> void:
	if camera != null:
		camera.zoom = Vector2.ONE * zoom_level


func _is_gameplay_active() -> bool:
	if !gameplay_active_provider.is_valid():
		return false
	return bool(gameplay_active_provider.call())
