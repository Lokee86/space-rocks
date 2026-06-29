extends RefCounted
class_name ServerHitboxOverlayFlow

const Log := preload("res://scripts/logging/logger.gd")
const DebugShapeCatalogPacketReader := preload("res://scripts/devtools/hitboxes/debug_shape_catalog_packet_reader.gd")
const DebugShapeCatalogStore := preload("res://scripts/devtools/hitboxes/debug_shape_catalog_store.gd")
const DebugShapeIdResolver := preload("res://scripts/devtools/hitboxes/debug_shape_id_resolver.gd")

var overlay
var world_sync
var shape_catalog_store
var latest_debug_collision_bodies: Array = []
var latest_gameplay_state: Dictionary = {}
var _logged_waiting_for_shape_catalog := false
var _logged_waiting_for_gameplay_state := false
var _logged_entries_generated := false


func configure(game_owner: Node2D, world_sync_ref) -> void:
	world_sync = world_sync_ref
	overlay = game_owner.get_node_or_null("ServerHitboxOverlay") if game_owner != null else null
	shape_catalog_store = DebugShapeCatalogStore.new()
	reset()


func reset() -> void:
	latest_debug_collision_bodies = []
	latest_gameplay_state = {}
	_logged_waiting_for_shape_catalog = false
	_logged_waiting_for_gameplay_state = false
	_logged_entries_generated = false
	if shape_catalog_store != null:
		shape_catalog_store.reset()
	if overlay != null && is_instance_valid(overlay) and overlay.has_method("set_hitbox_entries"):
		overlay.set_hitbox_entries([])


func apply_gameplay_state(state: Dictionary) -> void:
	latest_gameplay_state = state


func apply_debug_shape_catalog_packet(packet: Dictionary) -> void:
	if shape_catalog_store == null:
		return
	var catalog_state = DebugShapeCatalogPacketReader.read(packet)
	shape_catalog_store.apply_catalog_state(catalog_state)


func process() -> void:
	if overlay == null || !is_instance_valid(overlay):
		return
	if !overlay.has_method("is_enabled") or !overlay.is_enabled():
		return
	if world_sync == null:
		return
	if latest_gameplay_state.is_empty():
		if !_logged_waiting_for_gameplay_state:
			_logged_waiting_for_gameplay_state = true
			Log.world_sync_debug("server hitbox overlay waiting for gameplay state")
		return
	if shape_catalog_store == null or shape_catalog_store.shape_count() == 0:
		if !_logged_waiting_for_shape_catalog:
			_logged_waiting_for_shape_catalog = true
			Log.world_sync_debug("server hitbox overlay waiting for shape catalog")
		return

	var draw_entries: Array = []
	var players_value = latest_gameplay_state.get("server_players", {})
	if players_value is Dictionary:
		for player_id in players_value:
			var player_state_value = players_value[player_id]
			if !(player_state_value is Dictionary):
				continue

			var shape_id = DebugShapeIdResolver.player_shape_id(player_state_value)
			var shape_definition = shape_catalog_store.shape_for_id(shape_id) if shape_catalog_store != null else {}
			if shape_definition.is_empty():
				continue

			var x = float(player_state_value.get("x", 0.0))
			var y = float(player_state_value.get("y", 0.0))
			var rotation = float(player_state_value.get("rotation", 0.0))
			var points = _shape_definition_visual_points(shape_definition, x, y, rotation, 1.0)
			if points.is_empty():
				continue

			draw_entries.append({
				"kind": "player",
				"id": shape_id,
				"points": points,
			})

	var asteroids_value = latest_gameplay_state.get("server_asteroids", {})
	if asteroids_value is Dictionary:
		for asteroid_id in asteroids_value:
			var asteroid_state_value = asteroids_value[asteroid_id]
			if !(asteroid_state_value is Dictionary):
				continue

			var shape_id = DebugShapeIdResolver.asteroid_shape_id(asteroid_state_value)
			var shape_definition = shape_catalog_store.shape_for_id(shape_id) if shape_catalog_store != null else {}
			if shape_definition.is_empty():
				shape_id = "asteroid:0"
				shape_definition = shape_catalog_store.shape_for_id(shape_id) if shape_catalog_store != null else {}
			if shape_definition.is_empty():
				continue

			var x = float(asteroid_state_value.get("x", 0.0))
			var y = float(asteroid_state_value.get("y", 0.0))
			var rotation = float(asteroid_state_value.get("rotation", 0.0))
			var scale = float(asteroid_state_value.get("scale", 1.0))
			var points = _shape_definition_visual_points(shape_definition, x, y, rotation, scale)
			if points.is_empty():
				continue

			draw_entries.append({
				"kind": "asteroid",
				"id": shape_id,
				"points": points,
			})

	var bullets_value = latest_gameplay_state.get("server_bullets", {})
	if bullets_value is Dictionary:
		for bullet_id in bullets_value:
			var bullet_state_value = bullets_value[bullet_id]
			if !(bullet_state_value is Dictionary):
				continue

			var shape_id = DebugShapeIdResolver.bullet_shape_id(bullet_state_value)
			var shape_definition = shape_catalog_store.shape_for_id(shape_id) if shape_catalog_store != null else {}
			if shape_definition.is_empty():
				continue

			var x = float(bullet_state_value.get("x", 0.0))
			var y = float(bullet_state_value.get("y", 0.0))
			var rotation = float(bullet_state_value.get("rotation", 0.0))
			var points = _shape_definition_visual_points(shape_definition, x, y, rotation, 1.0)
			if points.is_empty():
				continue

			draw_entries.append({
				"kind": "bullet",
				"id": shape_id,
				"points": points,
			})

	var pickups_value = latest_gameplay_state.get("server_pickups", {})
	if pickups_value is Dictionary:
		for pickup_id in pickups_value:
			var pickup_state_value = pickups_value[pickup_id]
			if !(pickup_state_value is Dictionary):
				continue

			var shape_id = DebugShapeIdResolver.pickup_shape_id(pickup_state_value)
			var shape_definition = shape_catalog_store.shape_for_id(shape_id) if shape_catalog_store != null else {}
			if shape_definition.is_empty():
				continue

			var x = float(pickup_state_value.get("x", 0.0))
			var y = float(pickup_state_value.get("y", 0.0))
			var points = _shape_definition_visual_points(shape_definition, x, y, 0.0, 1.0)
			if points.is_empty():
				continue

			draw_entries.append({
				"kind": "pickup",
				"id": shape_id,
				"points": points,
			})

	overlay.set_hitbox_entries(draw_entries)


func _shape_definition_visual_points(shape_definition: Dictionary, x: float, y: float, rotation: float, scale: float) -> PackedVector2Array:
	if world_sync == null:
		return PackedVector2Array()

	var points_value = shape_definition.get("points", [])
	if !(points_value is Array):
		return PackedVector2Array()

	var draw_points := PackedVector2Array()
	for point_value in points_value:
		if !(point_value is Dictionary):
			continue

		var point_x_value = point_value.get("x", null)
		var point_y_value = point_value.get("y", null)
		if point_x_value == null || point_y_value == null:
			continue

		var local_point := Vector2(float(point_x_value), float(point_y_value))
		local_point *= scale
		local_point = local_point.rotated(rotation)
		var server_position := Vector2(x, y) + local_point
		draw_points.append(world_sync.visual_position_for_server_position(server_position))

	return draw_points
