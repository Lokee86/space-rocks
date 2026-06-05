class_name DevtoolsHitboxTemplateCatalog
extends RefCounted

const PlayerScene := preload("res://scenes/player.tscn")
const AsteroidScene := preload("res://scenes/asteroid.tscn")
const BulletScene := preload("res://scenes/bullet.tscn")
const PickupOneUpScene := preload("res://scenes/pickups/1_up.tscn")

var player_polygon_cache: PackedVector2Array = PackedVector2Array()
var asteroid_polygons_by_variant: Dictionary = {}
var bullet_polygon_cache: PackedVector2Array = PackedVector2Array()
var pickup_polygons_by_type: Dictionary = {}


func player_polygon() -> PackedVector2Array:
	if !player_polygon_cache.is_empty():
		return player_polygon_cache.duplicate()

	var player_scene := PlayerScene.instantiate()
	var polygon := PackedVector2Array()
	var collision_polygon := player_scene.get_node_or_null("CollisionPolygon2D") as CollisionPolygon2D
	if collision_polygon != null:
		polygon = collision_polygon.polygon
	player_scene.free()

	player_polygon_cache = polygon.duplicate()
	return player_polygon_cache.duplicate()


func asteroid_polygon(variant: int) -> PackedVector2Array:
	var resolved_variant := variant
	if asteroid_polygons_by_variant.has(resolved_variant):
		return (asteroid_polygons_by_variant[resolved_variant] as PackedVector2Array).duplicate()

	var asteroid_scene := AsteroidScene.instantiate()
	var collision_variants := asteroid_scene.get_node_or_null("CollisionVariants") as Node2D
	if collision_variants != null and collision_variants.get_child_count() > 0:
		resolved_variant = wrapi(variant, 0, collision_variants.get_child_count())

	var polygon := PackedVector2Array()
	if collision_variants != null and collision_variants.get_child_count() > 0:
		var variant_collision := collision_variants.get_child(resolved_variant) as CollisionPolygon2D
		if variant_collision != null:
			polygon = variant_collision.polygon

	if polygon.is_empty():
		var collision_polygon := asteroid_scene.get_node_or_null("CollisionPolygon2D") as CollisionPolygon2D
		if collision_polygon != null:
			polygon = collision_polygon.polygon
	asteroid_scene.free()

	if polygon.is_empty():
		return PackedVector2Array()

	asteroid_polygons_by_variant[resolved_variant] = polygon.duplicate()
	return (asteroid_polygons_by_variant[resolved_variant] as PackedVector2Array).duplicate()


func bullet_polygon() -> PackedVector2Array:
	if !bullet_polygon_cache.is_empty():
		return bullet_polygon_cache.duplicate()

	var half_width := 3.0
	var half_height := 12.0
	bullet_polygon_cache = PackedVector2Array([
		Vector2(-half_width, -half_height),
		Vector2(half_width, -half_height),
		Vector2(half_width, half_height),
		Vector2(-half_width, half_height),
	])
	return bullet_polygon_cache.duplicate()


func pickup_polygon(pickup_type: String) -> PackedVector2Array:
	if pickup_polygons_by_type.has(pickup_type):
		return (pickup_polygons_by_type[pickup_type] as PackedVector2Array).duplicate()

	var polygon := PackedVector2Array()
	if pickup_type == "1_up":
		var pickup_scene := PickupOneUpScene.instantiate()
		var collision_shape := pickup_scene.get_node_or_null("CollisionShape2D") as CollisionShape2D
		if collision_shape != null and collision_shape.shape is CircleShape2D:
			polygon = _circle_polygon(collision_shape.shape as CircleShape2D, collision_shape.position)
		pickup_scene.free()

	if polygon.is_empty():
		return PackedVector2Array()

	pickup_polygons_by_type[pickup_type] = polygon.duplicate()
	return (pickup_polygons_by_type[pickup_type] as PackedVector2Array).duplicate()


func _circle_polygon(circle: CircleShape2D, offset: Vector2) -> PackedVector2Array:
	var points := PackedVector2Array()
	var point_count := 24
	for i in point_count:
		var angle := TAU * float(i) / float(point_count)
		points.append(offset + Vector2(cos(angle), sin(angle)) * circle.radius)
	return points
