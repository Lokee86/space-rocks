extends RefCounted
class_name ServerHitboxOverlayFlow

var overlay
var world_sync
var latest_debug_collision_bodies: Array = []


func configure(game_owner: Node2D, world_sync_ref) -> void:
	world_sync = world_sync_ref
	overlay = game_owner.get_node_or_null("ServerHitboxOverlay") if game_owner != null else null
	reset()


func reset() -> void:
	latest_debug_collision_bodies = []
	if overlay != null && is_instance_valid(overlay) and overlay.has_method("set_hitbox_entries"):
		overlay.set_hitbox_entries([])


func apply_gameplay_state(state: Dictionary) -> void:
	var collision_bodies_value = state.get("debug_collision_bodies", [])
	if collision_bodies_value is Array:
		latest_debug_collision_bodies = collision_bodies_value
	else:
		latest_debug_collision_bodies = []


func process() -> void:
	if overlay == null || !is_instance_valid(overlay):
		return
	if !overlay.has_method("is_enabled") or !overlay.is_enabled():
		return
	if world_sync == null:
		return

	var draw_entries: Array = []
	for entry in latest_debug_collision_bodies:
		if !(entry is Dictionary):
			continue

		var points_value = entry.get("points", [])
		if !(points_value is Array):
			continue

		var draw_points := PackedVector2Array()
		for point in points_value:
			if !(point is Dictionary):
				continue
			var server_point := Vector2(float(point.get("x", 0.0)), float(point.get("y", 0.0)))
			draw_points.append(world_sync.visual_position_for_server_position(server_point))

		draw_entries.append({
			"kind": str(entry.get("kind", "")),
			"id": str(entry.get("id", "")),
			"points": draw_points,
		})

	overlay.set_hitbox_entries(draw_entries)
