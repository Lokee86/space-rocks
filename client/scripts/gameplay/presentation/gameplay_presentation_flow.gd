extends RefCounted
class_name GameplayPresentationFlow

const OSIndicatorController = preload("res://scripts/gameplay/presentation/os_indicator_controller.gd")

var hud: Control
var camera_provider: Callable
var remote_positions_provider: Callable
var remote_hues_provider: Callable
var os_indicator_controller := OSIndicatorController.new()


func configure(
	hud_ref: Control,
	camera_provider_ref: Callable,
	remote_positions_provider_ref: Callable,
	remote_hues_provider_ref: Callable
) -> void:
	hud = hud_ref
	camera_provider = camera_provider_ref
	remote_positions_provider = remote_positions_provider_ref
	remote_hues_provider = remote_hues_provider_ref
	os_indicator_controller.configure(hud)


func reset() -> void:
	os_indicator_controller.reset()


func process(_delta: float) -> void:
	if (
		camera_provider.is_null()
		|| remote_positions_provider.is_null()
		|| remote_hues_provider.is_null()
	):
		return

	var camera = camera_provider.call()
	var positions = remote_positions_provider.call()
	var hues = remote_hues_provider.call()
	os_indicator_controller.update_indicators(camera, positions, hues)
