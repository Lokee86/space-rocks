extends RefCounted
class_name ServerHitboxOverlayFlow

var overlay
var world_sync


func configure(game_owner: Node2D, world_sync_ref) -> void:
	world_sync = world_sync_ref
	overlay = game_owner.get_node_or_null("ServerHitboxOverlay") if game_owner != null else null
	reset()


func reset() -> void:
	if overlay != null && is_instance_valid(overlay) and overlay.has_method("set_hitbox_entries"):
		overlay.set_hitbox_entries([])


func process() -> void:
	if overlay == null || !is_instance_valid(overlay):
		return
	if !overlay.has_method("is_enabled") or !overlay.is_enabled():
		return
	if world_sync == null:
		return
	if !world_sync.has_method("server_hitbox_draw_entries"):
		return
	overlay.set_hitbox_entries(world_sync.server_hitbox_draw_entries())
