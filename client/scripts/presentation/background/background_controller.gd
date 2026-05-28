extends Node
class_name BackgroundController

const GameplayBackgroundFlow := preload("res://scripts/presentation/background/background_flow.gd")

var background_flow


func configure(
	repeated_background: TextureRect,
	repeated_foreground_background: TextureRect,
	repeated_planet_background: TextureRect,
	parallax_target: Node2D = null
) -> void:
	background_flow = GameplayBackgroundFlow.new()
	background_flow.configure(
		repeated_background,
		repeated_foreground_background,
		repeated_planet_background,
		parallax_target
	)


func set_parallax_target(parallax_target: Node2D) -> void:
	if background_flow != null:
		background_flow.set_parallax_target(parallax_target)


func _process(_delta: float) -> void:
	if background_flow != null:
		background_flow.process_frame()


func reset_background() -> void:
	if background_flow != null:
		background_flow.clear()