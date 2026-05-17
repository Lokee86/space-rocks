extends CharacterBody2D

@onready var sprite: Sprite2D = $Sprite2D
@onready var collision: CollisionPolygon2D = $CollisionPolygon2D
@onready var collision_variants: Node2D = $CollisionVariants

var asteroid_textures := [
	preload("res://assets/asteroids/asteroid1.png"),
	preload("res://assets/asteroids/asteroid2.png"),
	preload("res://assets/asteroids/asteroid3.png"),
	preload("res://assets/asteroids/asteroid4.png"),
]


func set_asteroid_variant(index: int) -> void:
	var variant_index := wrapi(index, 0, asteroid_textures.size())
	sprite.texture = asteroid_textures[variant_index]

	if collision_variants.get_child_count() == 0:
		return

	var shape_node := collision_variants.get_child(
		min(variant_index, collision_variants.get_child_count() - 1)
	) as CollisionPolygon2D
	if shape_node != null:
		collision.polygon = shape_node.polygon
