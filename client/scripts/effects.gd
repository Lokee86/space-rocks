extends RefCounted
class_name Effects

const Constants = preload("res://scripts/constants.gd")
const BULLET_BLAST_SCENE := preload("res://scenes/animations/bullet_blast.tscn")
const SHIP_DEATH_SCENE := preload("res://scenes/animations/ship_death.tscn")
const EFFECT_Z_INDEX := 40

var owner_node: Node2D
var game_over_sound: AudioStreamPlayer
var game_over_sound_played := false
var game_over_sound_token := 0


func configure(game_owner: Node2D, game_over_audio: AudioStreamPlayer) -> void:
	owner_node = game_owner
	game_over_sound = game_over_audio


func reset_game_over_sound() -> void:
	game_over_sound_played = false
	stop_game_over_sound()


func stop_game_over_sound() -> void:
	game_over_sound_token += 1
	if game_over_sound != null:
		game_over_sound.stop()


func play_game_over_sound_after_delay() -> void:
	if Constants.GAME_OVER_SOUND_DELAY <= 0:
		_play_game_over_sound()
		return

	var token := game_over_sound_token
	owner_node.get_tree().create_timer(Constants.GAME_OVER_SOUND_DELAY).timeout.connect(func() -> void:
		if token == game_over_sound_token:
			_play_game_over_sound()
	)


func spawn_bullet_blast(event_position: Vector2) -> void:
	var blast_node := BULLET_BLAST_SCENE.instantiate()
	blast_node.global_position = event_position
	blast_node.z_index = EFFECT_Z_INDEX
	owner_node.add_child(blast_node)

	var sprite := blast_node.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D
	var sound := blast_node.get_node_or_null("AsteroidDestroyed") as AudioStreamPlayer2D
	if sprite == null || sound == null:
		blast_node.queue_free()
		return

	var free_blast := func() -> void:
		if is_instance_valid(blast_node):
			blast_node.queue_free()

	sprite.animation_finished.connect(func() -> void:
		sprite.visible = false
	)
	sound.finished.connect(free_blast)

	sprite.play("bullet_blast")
	sound.play()

	var sound_length := 1.0
	if sound.stream != null:
		sound_length = max(sound.stream.get_length(), sound_length)
	owner_node.get_tree().create_timer(sound_length + 0.25).timeout.connect(free_blast)


func spawn_ship_death(event_position: Vector2) -> void:
	var death_node := SHIP_DEATH_SCENE.instantiate()
	death_node.global_position = event_position
	death_node.z_index = EFFECT_Z_INDEX
	owner_node.add_child(death_node)

	var sprite := death_node.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D
	var sound := death_node.get_node_or_null("ShipDeath") as AudioStreamPlayer2D
	if sprite == null || sound == null:
		death_node.queue_free()
		return

	var death_finished := false
	var free_death := func() -> void:
		if death_finished:
			return
		death_finished = true
		if is_instance_valid(death_node):
			death_node.queue_free()

	sprite.animation_finished.connect(func() -> void:
		sprite.visible = false
	)
	sound.finished.connect(free_death)

	sprite.frame = 0
	sprite.frame_progress = 0.0
	sprite.play("default")
	sound.play()

	var sound_length := 0.0
	if sound.stream != null:
		sound_length = sound.stream.get_length()
	if sound_length > 0:
		owner_node.get_tree().create_timer(sound_length + 0.05).timeout.connect(free_death)


func _play_game_over_sound() -> void:
	if game_over_sound != null && !game_over_sound_played:
		game_over_sound_played = true
		game_over_sound.play()
