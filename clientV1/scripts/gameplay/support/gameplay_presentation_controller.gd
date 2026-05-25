extends RefCounted

const GameBackgroundScrollScript = preload("res://scripts/gameplay/support/game_background_scroll.gd")
const OffscreenIndicatorControllerScript = preload("res://scripts/gameplay/support/offscreen_indicator_controller.gd")

var background_scroll
var offscreen_indicator_controller


func configure(offscreen_indicators, gameplay_camera: Camera2D) -> void:
	background_scroll = GameBackgroundScrollScript.new()
	offscreen_indicator_controller = OffscreenIndicatorControllerScript.new()
	offscreen_indicator_controller.configure(offscreen_indicators, gameplay_camera)


func update_background_scroll(parent, has_initial_spawn: bool, is_spectating: bool, camera_follow, player: Node2D) -> void:
	background_scroll.update_scroll_offset(
		parent,
		has_initial_spawn,
		is_spectating,
		camera_follow,
		player
	)


func update_offscreen_indicators(remote_positions: Dictionary, remote_hues: Dictionary) -> void:
	offscreen_indicator_controller.update_indicators(remote_positions, remote_hues)


func clear_background_scroll(parent) -> void:
	background_scroll.clear_scroll_offset(parent)
