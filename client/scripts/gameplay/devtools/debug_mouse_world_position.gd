extends RefCounted


var game_owner: Node2D
var server_position_provider: Callable


func configure(game_owner_ref: Node2D, server_position_provider_ref: Callable) -> void:
	game_owner = game_owner_ref
	server_position_provider = server_position_provider_ref


func current_position() -> Dictionary:
	if game_owner == null || server_position_provider.is_null():
		return {"valid": false}

	var visual_position: Vector2 = game_owner.get_global_mouse_position()
	var server_position_variant = server_position_provider.call(visual_position)
	if server_position_variant is Vector2:
		return {
			"valid": true,
			"visual_position": visual_position,
			"server_position": server_position_variant,
		}

	return {
		"valid": false,
		"visual_position": visual_position,
	}
