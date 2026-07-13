extends Node2D
class_name TorpedoPresentation


func reset_from_pool() -> void:
	modulate = Color.WHITE
	rotation = 0.0
	scale = Vector2.ONE
	visible = false


func reset_for_pool() -> void:
	rotation = 0.0
	scale = Vector2.ONE
	visible = false