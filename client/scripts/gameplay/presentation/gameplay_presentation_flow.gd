extends RefCounted
class_name GameplayPresentationFlow

var hud: Control
var camera_provider: Callable
var remote_positions_provider: Callable
var os_indicator_controller := OSIndicatorController.new()
var local_player_presentation_controller := LocalPlayerPresentationController.new()


func configure(
	hud_ref: Control,
	player_ref,
	camera_provider_ref: Callable,
	remote_positions_provider_ref: Callable
) -> void:
	hud = hud_ref
	camera_provider = camera_provider_ref
	remote_positions_provider = remote_positions_provider_ref
	os_indicator_controller.configure(hud)
	local_player_presentation_controller.configure(player_ref)


func reset() -> void:
	os_indicator_controller.reset()
	local_player_presentation_controller.reset()


func process(_delta: float, has_received_state: bool) -> void:
	local_player_presentation_controller.process(has_received_state)
	if (
		camera_provider.is_null()
		|| remote_positions_provider.is_null()
	):
		return

	var camera = camera_provider.call()
	var positions = remote_positions_provider.call()
	os_indicator_controller.update_indicators(camera, positions)
