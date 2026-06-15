extends CharacterBody2D

const AsteroidVariants = preload("res://scripts/generated/asteroids/asteroid_variants.gd")

@onready var sprite: Sprite2D = $Sprite2D
@onready var collision: CollisionPolygon2D = $CollisionPolygon2D
@onready var collision_variants: Node2D = $CollisionVariants

func set_asteroid_variant(index: int) -> void:
	var variant_count := AsteroidVariants.count()
	if variant_count <= 0:
		return

	var variant_index := wrapi(index, 0, variant_count)
	var texture_path := AsteroidVariants.texture_path_for_index(index)
	if texture_path != "":
		sprite.texture = load(texture_path) as Texture2D
	collision.disabled = false

	if collision_variants.get_child_count() == 0:
		return

	for child in collision_variants.get_children():
		var variant_collision := child as CollisionPolygon2D
		if variant_collision != null:
			variant_collision.disabled = true

	var shape_node := collision_variants.get_child(
		min(variant_index, collision_variants.get_child_count() - 1)
	) as CollisionPolygon2D
	if shape_node != null:
		collision.polygon = shape_node.polygon
