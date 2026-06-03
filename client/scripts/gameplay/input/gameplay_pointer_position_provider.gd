extends RefCounted
class_name GameplayPointerPositionProvider

var game_owner: Node2D
var runtime_context


func configure(game_owner_ref, runtime_context_ref) -> void:
	game_owner = game_owner_ref
	runtime_context = runtime_context_ref


func mouse_visual_position() -> Vector2:
	if runtime_context == null:
		return Vector2.ZERO
	if runtime_context.current_camera() == null:
		return Vector2.ZERO
	if game_owner == null:
		return Vector2.ZERO
	return game_owner.get_global_mouse_position()


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	if runtime_context == null:
		return visual_position
	return runtime_context.server_position_for_visual_position(visual_position)
