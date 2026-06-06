extends Node2D

const Constants = preload("res://scripts/generated/constants/constants.gd")

@export var sprite_path: NodePath = NodePath("Sprite2D")
@export var glow_sprite_path: NodePath = NodePath("GlowSprite2D")
@export var spawn_sound_path: NodePath = NodePath("PickupSpawned")
@export var collision_shape_path: NodePath = NodePath("CollisionShape2D")

@export var sprite_pulse_rate := 2.0
@export var sprite_pulse_amount := 0.06
@export var eol_flash_slow_period := 0.6
@export var eol_flash_fast_period := 0.08

@export var glow_pulse_rate := 1.15
@export var glow_pulse_amount := 0.18
@export var glow_alpha_min := 0.35
@export var glow_alpha_max := 0.85

var sprite: Sprite2D
var glow_sprite: Sprite2D
var spawn_sound: AudioStreamPlayer2D
var collision_shape: CollisionShape2D

var sprite_base_scale := Vector2.ONE
var sprite_base_modulate := Color.WHITE
var glow_base_scale := Vector2.ONE
var glow_base_modulate := Color.WHITE
var elapsed := 0.0
var lifespan_age_seconds := 0.0
var lifespan_seconds := 0.0
var has_lifespan_state := false
var eol_blink_phase := 0.0


func _ready() -> void:
	sprite = get_node_or_null(sprite_path) as Sprite2D
	glow_sprite = get_node_or_null(glow_sprite_path) as Sprite2D
	spawn_sound = get_node_or_null(spawn_sound_path) as AudioStreamPlayer2D
	collision_shape = get_node_or_null(collision_shape_path) as CollisionShape2D

	if sprite != null:
		sprite_base_scale = sprite.scale
		sprite_base_modulate = sprite.modulate

	if glow_sprite != null:
		glow_base_scale = glow_sprite.scale
		glow_base_modulate = glow_sprite.modulate


func collision_radius() -> float:
	if collision_shape == null:
		return 0.0

	var circle_shape := collision_shape.shape as CircleShape2D
	if circle_shape == null:
		return 0.0

	return circle_shape.radius


func play_spawn_sound(audio_flow) -> void:
	if audio_flow == null:
		return
	if spawn_sound == null:
		return
	if !audio_flow.has_method("play_pickup_spawned_sound"):
		return
	audio_flow.play_pickup_spawned_sound(spawn_sound)


func apply_lifespan_state(age_seconds: float, total_lifespan_seconds: float) -> void:
	lifespan_age_seconds = max(age_seconds, 0.0)
	lifespan_seconds = max(total_lifespan_seconds, 0.0)
	has_lifespan_state = lifespan_seconds > 0.0
	if not has_lifespan_state:
		eol_blink_phase = 0.0


func _process(delta: float) -> void:
	elapsed += delta

	var warning_window: float = lifespan_seconds * Constants.PICKUP_EOL_FLASH
	var remaining: float = max(lifespan_seconds - lifespan_age_seconds, 0.0)
	var show_pickup: bool = true
	var blink_visible: bool = true

	if not has_lifespan_state or warning_window <= 0.0 or remaining > warning_window:
		show_pickup = true
		eol_blink_phase = 0.0
	else:
		var progress: float = 1.0 - clamp(remaining / warning_window, 0.0, 1.0)
		var blink_period: float = lerpf(eol_flash_slow_period, eol_flash_fast_period, progress)
		if blink_period > 0.0:
			eol_blink_phase += delta / blink_period
			blink_visible = fmod(eol_blink_phase, 1.0) < 0.5
		else:
			blink_visible = true
		show_pickup = blink_visible

	if sprite != null:
		sprite.visible = show_pickup
		if show_pickup:
			var sprite_pulse: float = _pulse(sprite_pulse_rate, sprite_pulse_amount)
			sprite.scale = sprite_base_scale * sprite_pulse
			sprite.modulate = sprite_base_modulate

	if glow_sprite != null:
		glow_sprite.visible = show_pickup
		if show_pickup:
			var glow_pulse: float = _pulse(glow_pulse_rate, glow_pulse_amount)
			glow_sprite.scale = glow_base_scale * glow_pulse

			var alpha_weight := (sin(elapsed * TAU * glow_pulse_rate) + 1.0) * 0.5
			var next_modulate := glow_base_modulate
			next_modulate.a = lerpf(glow_alpha_min, glow_alpha_max, alpha_weight)
			glow_sprite.modulate = next_modulate


func _pulse(rate: float, amount: float) -> float:
	return 1.0 + sin(elapsed * TAU * rate) * amount
