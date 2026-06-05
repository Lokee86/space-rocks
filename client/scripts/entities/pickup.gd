extends Node2D

@export var sprite_path: NodePath = NodePath("Sprite2D")
@export var glow_sprite_path: NodePath = NodePath("GlowSprite2D")
@export var spawn_sound_path: NodePath = NodePath("PickupSpawned")

@export var sprite_pulse_rate := 2.0
@export var sprite_pulse_amount := 0.06

@export var glow_pulse_rate := 1.15
@export var glow_pulse_amount := 0.18
@export var glow_alpha_min := 0.35
@export var glow_alpha_max := 0.85

var sprite: Sprite2D
var glow_sprite: Sprite2D
var spawn_sound: AudioStreamPlayer2D

var sprite_base_scale := Vector2.ONE
var glow_base_scale := Vector2.ONE
var glow_base_modulate := Color.WHITE
var elapsed := 0.0


func _ready() -> void:
	sprite = get_node_or_null(sprite_path) as Sprite2D
	glow_sprite = get_node_or_null(glow_sprite_path) as Sprite2D
	spawn_sound = get_node_or_null(spawn_sound_path) as AudioStreamPlayer2D

	if sprite != null:
		sprite_base_scale = sprite.scale

	if glow_sprite != null:
		glow_base_scale = glow_sprite.scale
		glow_base_modulate = glow_sprite.modulate


func play_spawn_sound(audio_flow) -> void:
	if audio_flow == null:
		return
	if spawn_sound == null:
		return
	if !audio_flow.has_method("play_pickup_spawned_sound"):
		return
	audio_flow.play_pickup_spawned_sound(spawn_sound)


func _process(delta: float) -> void:
	elapsed += delta

	if sprite != null:
		var sprite_pulse := _pulse(sprite_pulse_rate, sprite_pulse_amount)
		sprite.scale = sprite_base_scale * sprite_pulse

	if glow_sprite != null:
		var glow_pulse := _pulse(glow_pulse_rate, glow_pulse_amount)
		glow_sprite.scale = glow_base_scale * glow_pulse

		var alpha_weight := (sin(elapsed * TAU * glow_pulse_rate) + 1.0) * 0.5
		var next_modulate := glow_base_modulate
		next_modulate.a = lerpf(glow_alpha_min, glow_alpha_max, alpha_weight)
		glow_sprite.modulate = next_modulate


func _pulse(rate: float, amount: float) -> float:
	return 1.0 + sin(elapsed * TAU * rate) * amount
