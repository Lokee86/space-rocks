class_name DevtoolsHitboxTemplateCatalog
extends RefCounted

const PlayerScene := preload("res://scenes/player.tscn")
const AsteroidScene := preload("res://scenes/asteroid.tscn")
const BulletScene := preload("res://scenes/bullet.tscn")

var player_polygon_cache: PackedVector2Array = PackedVector2Array()
var asteroid_polygons_by_variant: Dictionary = {}
var bullet_polygon_cache: PackedVector2Array = PackedVector2Array()


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
	if asteroid_scene != null and asteroid_scene.has_method("set_asteroid_variant"):
		asteroid_scene.call("set_asteroid_variant", variant)

	var polygon := PackedVector2Array()
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
