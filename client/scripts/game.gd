extends Node2D

@export var background_parallax := 0.25
@export var foreground_background_parallax := 0.45
@export var foreground_background_offset := Vector2(480.0, 270.0)

@onready var player: Node2D = $Player
@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground


func _process(_delta: float) -> void:
	_update_layer_shader(repeated_background, background_parallax, Vector2.ZERO)
	_update_layer_shader(repeated_foreground_background, foreground_background_parallax, foreground_background_offset)


func _update_layer_shader(background: TextureRect, parallax: float, offset: Vector2) -> void:
	var material := background.material as ShaderMaterial
	if material == null:
		return

	material.set_shader_parameter(
		"scroll_offset",
		(player.global_position * parallax) + offset
	)