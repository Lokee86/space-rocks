extends RefCounted

var offscreen_indicators
var gameplay_camera: Camera2D


func configure(indicators_node, camera: Camera2D) -> void:
	offscreen_indicators = indicators_node
	gameplay_camera = camera


func update_indicators(remote_player_visual_positions: Dictionary, remote_player_hues: Dictionary) -> void:
	if offscreen_indicators == null || gameplay_camera == null:
		return

	offscreen_indicators.update_indicators(
		remote_player_visual_positions,
		gameplay_camera,
		remote_player_hues
	)
