extends RefCounted
class_name AsteroidSync

var asteroids_layer: Node2D


func configure(layer: Node2D) -> void:
	asteroids_layer = layer
