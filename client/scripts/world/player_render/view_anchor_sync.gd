extends RefCounted
class_name ViewAnchorSync

# This API owns the active ViewAnchor/render-anchor mapping.

const LegacyLocalVisualSyncScript = preload("res://legacy/player_render/local_visual_sync.gd")

var legacy_sync


func _init() -> void:
	legacy_sync = LegacyLocalVisualSyncScript.new()


func reset() -> void:
	legacy_sync = LegacyLocalVisualSyncScript.new()


func update_from_anchor_server_position(server_position: Vector2) -> void:
	legacy_sync.update_from_server_position(server_position)


func server_position() -> Vector2:
	return legacy_sync.server_position()


func visual_position() -> Vector2:
	return legacy_sync.visual_position()


func visual_position_for_server_position(server_position: Vector2) -> Vector2:
	return legacy_sync.visual_position_for_server_position(server_position)


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	return legacy_sync.server_position_for_visual_position(visual_position)
