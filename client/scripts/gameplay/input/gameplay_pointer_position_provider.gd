extends RefCounted
class_name GameplayPointerPositionProvider

var game_owner: Node2D
var server_position_for_visual_position_provider: Callable


func configure(game_owner_ref, server_position_for_visual_position_provider_ref: Callable) -> void:
	game_owner = game_owner_ref
	server_position_for_visual_position_provider = server_position_for_visual_position_provider_ref


func mouse_visual_position() -> Vector2:
	if game_owner == null:
		return Vector2.ZERO
	return game_owner.get_global_mouse_position()


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	if server_position_for_visual_position_provider.is_null():
		return visual_position
	return server_position_for_visual_position_provider.call(visual_position)
