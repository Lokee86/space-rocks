extends GutTest

const CameraFollowScript := preload("res://scripts/camera_follow.gd")

var player: Node2D
var camera: Camera2D
var camera_follow


func before_each() -> void:
	player = Node2D.new()
	player.global_position = Vector2(100.0, 150.0)
	add_child(player)

	camera = Camera2D.new()
	player.add_child(camera)

	camera_follow = CameraFollowScript.new()
	camera_follow.configure(camera)


func after_each() -> void:
	camera_follow = null
	if player != null:
		player.free()
		player = null


func test_follow_visual_position_places_camera_at_explicit_position() -> void:
	camera_follow.follow_visual_position(Vector2(320.0, 240.0))

	assert_eq(camera.global_position, Vector2(320.0, 240.0))


func test_follow_local_player_restores_camera_to_player_origin() -> void:
	camera_follow.follow_visual_position(Vector2(320.0, 240.0))

	camera_follow.follow_local_player()

	assert_eq(camera.position, Vector2.ZERO)
	assert_eq(camera.global_position, player.global_position)


func test_missing_camera_is_safe() -> void:
	var empty_follow := CameraFollowScript.new()

	empty_follow.follow_visual_position(Vector2(1.0, 2.0))
	empty_follow.follow_local_player()

	pass_test("missing camera calls are safe")
