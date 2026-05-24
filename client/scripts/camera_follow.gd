extends RefCounted
class_name CameraFollow

var camera: Camera2D


func configure(gameplay_camera: Camera2D) -> void:
	camera = gameplay_camera


func follow_local_player() -> void:
	if camera == null:
		return

	camera.position = Vector2.ZERO


func follow_visual_position(visual_position: Vector2) -> void:
	if camera == null:
		return

	camera.global_position = visual_position
