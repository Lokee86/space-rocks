extends RefCounted
class_name ServerHitboxOverlayFlow

var overlay
var runtime_context


func configure(game_owner: Node2D, runtime_context_ref) -> void:
	runtime_context = runtime_context_ref
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
	if runtime_context == null:
		return
	if !runtime_context.has_method("server_hitbox_draw_entries"):
		return
	overlay.set_hitbox_entries(runtime_context.server_hitbox_draw_entries())
