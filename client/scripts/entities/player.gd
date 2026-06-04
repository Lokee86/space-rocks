extends CharacterBody2D
class_name Player

const Constants = preload("res://scripts/generated/constants/constants.gd")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const AFTERBURNER_SCENE := preload("res://scenes/animations/blue_afterburner.tscn")
const PLAYER_HUE_SHIFT_SHADER := preload("res://shaders/player_hue_shift.gdshader")

@export var turn_left_action := &"turn_left"
@export var turn_right_action := &"turn_right"
@export var move_forward_action := &"move_forward"
@export var move_backward_action := &"move_backward"
@export var shoot_action := &"shoot"
@export var ship_visual_path: NodePath = ^"Sprite2D"

@onready var ship_visual: CanvasItem = get_node_or_null(ship_visual_path) as CanvasItem
@onready var afterburner_marker: Marker2D = $AfterburnerMarker

var afterburner: Node2D
var afterburner_sprite: AnimatedSprite2D
var afterburner_audio: AudioStreamPlayer2D
var afterburner_active := false
var ship_hue_material: ShaderMaterial
var player_hue := Constants.PLAYER_DEFAULT_HUE
var audio_flow := GameplayAudioFlow.new()


func _ready() -> void:
	_ensure_unique_ship_hue_material()
	set_player_hue(player_hue)

	afterburner = AFTERBURNER_SCENE.instantiate()
	afterburner.rotation_degrees = Constants.PLAYER_AFTERBURNER_ROTATION_DEGREES
	afterburner.scale = Vector2.ONE * Constants.PLAYER_AFTERBURNER_SCALE
	afterburner.visible = false
	afterburner_marker.add_child(afterburner)

	afterburner_sprite = afterburner.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D
	afterburner_audio = afterburner.get_node_or_null("AudioStreamPlayer2D") as AudioStreamPlayer2D
	if afterburner_audio != null && !afterburner_audio.finished.is_connected(_on_afterburner_audio_finished):
		afterburner_audio.finished.connect(_on_afterburner_audio_finished)


func get_input_packet() -> Dictionary:
	return Packets.input_packet(
		Input.is_action_pressed(move_forward_action),
		Input.is_action_pressed(move_backward_action),
		Input.is_action_pressed(turn_right_action),
		Input.is_action_pressed(turn_left_action),
		Input.is_action_pressed(shoot_action)
	)


func set_player_hue(hue: float) -> void:
	player_hue = fposmod(hue, 1.0)
	if _ensure_unique_ship_hue_material() == null:
		return

	ship_hue_material.set_shader_parameter("hue_shift", player_hue)


func set_afterburner_active(active: bool) -> void:
	if afterburner == null || afterburner_active == active:
		return

	afterburner_active = active
	afterburner.visible = active

	if active:
		if afterburner_sprite != null:
			afterburner_sprite.play("default")
		_play_afterburner_sound()
	else:
		if afterburner_sprite != null:
			afterburner_sprite.stop()
		_stop_afterburner_sound()


func set_remote_afterburner_visual_active(active: bool) -> void:
	if afterburner == null || afterburner_active == active:
		return

	afterburner_active = active
	afterburner.visible = active

	if active:
		if afterburner_sprite != null:
			afterburner_sprite.play("default")
	else:
		if afterburner_sprite != null:
			afterburner_sprite.stop()


func stop_transient_effects() -> void:
	set_afterburner_active(false)


func _play_afterburner_sound() -> void:
	if afterburner_audio == null:
		return
	if !afterburner_audio.playing:
		audio_flow.play_afterburner_sound(afterburner_audio)


func _stop_afterburner_sound() -> void:
	audio_flow.stop_afterburner_sound(afterburner_audio)


func _on_afterburner_audio_finished() -> void:
	if afterburner_active:
		_play_afterburner_sound()


func _ensure_unique_ship_hue_material() -> ShaderMaterial:
	if ship_hue_material != null:
		return ship_hue_material
	if ship_visual == null:
		return null

	var existing_material := ship_visual.material as ShaderMaterial
	if existing_material != null && existing_material.shader == PLAYER_HUE_SHIFT_SHADER:
		ship_hue_material = existing_material.duplicate() as ShaderMaterial
	else:
		ship_hue_material = ShaderMaterial.new()
		ship_hue_material.shader = PLAYER_HUE_SHIFT_SHADER

	ship_visual.material = ship_hue_material
	return ship_hue_material
