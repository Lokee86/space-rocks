extends RefCounted
class_name GameplayEffects


const Constants = preload("res://scripts/generated/constants/constants.gd")
const BULLET_BLAST_SCENE := preload("res://scenes/animations/bullet_blast.tscn")
const PICKUP_COLLECT_SCENE := preload("res://scenes/pickups/pickup_collect.tscn")
const SHIP_DEATH_SCENE := preload("res://scenes/animations/ship_death.tscn")
const TORPEDO_EXPLOSION_SCENE := preload("res://scenes/animations/torpedo_explosion.tscn")
const TORPEDO_EXPLOSION_SCALE_SOURCE_FRAME := 5
const EFFECT_CLEANUP_STARTED_META := &"effect_cleanup_started"

var owner_node: Node2D
var audio_flow := GameplayAudioFlow.new()
var game_over_sound_played := false
var game_over_sound_token := 0


func configure(game_owner: Node2D, hud: Control) -> void:
	owner_node = game_owner
	audio_flow.configure(hud)


func reset_game_over_sound() -> void:
	game_over_sound_played = false
	stop_game_over_sound()


func stop_game_over_sound() -> void:
	game_over_sound_token += 1
	audio_flow.stop_game_over_sound()


func play_game_over_sound_after_delay() -> void:
	if Constants.GAME_OVER_SOUND_DELAY <= 0:
		_play_game_over_sound()
		return

	var token := game_over_sound_token
	owner_node.get_tree().create_timer(Constants.GAME_OVER_SOUND_DELAY).timeout.connect(
		_on_game_over_sound_delay_timeout.bind(token)
	)


func spawn_bullet_blast(event_position: Vector2) -> void:
	var blast_node := BULLET_BLAST_SCENE.instantiate()
	blast_node.global_position = event_position
	blast_node.z_index = Constants.EFFECT_Z_INDEX
	owner_node.add_child(blast_node)

	var sprite := blast_node.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D
	var sound := blast_node.get_node_or_null("AsteroidDestroyed") as AudioStreamPlayer2D
	if sprite == null || sound == null:
		blast_node.queue_free()
		return

	sprite.animation_finished.connect(_hide_effect_sprite.bind(sprite))
	sound.finished.connect(_queue_free_effect_node.bind(blast_node))

	sprite.play("bullet_blast")
	audio_flow.play_bullet_blast_sound(sound)

	var sound_length := Constants.BULLET_BLAST_MIN_SOUND_LENGTH
	if sound.stream != null:
		sound_length = max(sound.stream.get_length(), sound_length)
	var blast_ref: WeakRef = weakref(blast_node)
	owner_node.get_tree().create_timer(sound_length + Constants.BULLET_BLAST_CLEANUP_PADDING).timeout.connect(
		func() -> void:
			var node := blast_ref.get_ref() as Node
			if node != null and is_instance_valid(node):
				node.queue_free()
	)


func spawn_pickup_collected(event_position: Vector2) -> void:
	var effect_node := PICKUP_COLLECT_SCENE.instantiate()
	effect_node.global_position = event_position
	effect_node.z_index = Constants.PICKUP_Z_INDEX + 1
	owner_node.add_child(effect_node)

	var particles := effect_node.get_node_or_null("GPUParticles2D") as GPUParticles2D
	var sound := effect_node.get_node_or_null("AudioStreamPlayer2D") as AudioStreamPlayer2D
	if particles == null || sound == null:
		effect_node.queue_free()
		return

	particles.restart()
	particles.emitting = true
	audio_flow.play_pickup_collected_sound(sound)

	var cleanup_time := particles.lifetime
	if sound.stream != null:
		cleanup_time = max(cleanup_time, sound.stream.get_length())
	var effect_ref: WeakRef = weakref(effect_node)
	owner_node.get_tree().create_timer(cleanup_time + 0.1).timeout.connect(
		func() -> void:
			var node := effect_ref.get_ref() as Node
			if node != null and is_instance_valid(node):
				node.queue_free()
	)


func spawn_ship_death(event_position: Vector2) -> void:
	var death_node := SHIP_DEATH_SCENE.instantiate()
	death_node.global_position = event_position
	death_node.z_index = Constants.EFFECT_Z_INDEX
	owner_node.add_child(death_node)

	var sprite := death_node.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D
	var sound := death_node.get_node_or_null("ShipDeath") as AudioStreamPlayer2D
	if sprite == null || sound == null:
		death_node.queue_free()
		return

	sprite.animation_finished.connect(_hide_effect_sprite.bind(sprite))
	sound.finished.connect(_queue_free_effect_node_once.bind(death_node))

	sprite.frame = 0
	sprite.frame_progress = 0.0
	sprite.play("default")
	audio_flow.play_ship_death_sound(sound)

	var sound_length := 0.0
	if sound.stream != null:
		sound_length = sound.stream.get_length()
	if sound_length > 0:
		var death_ref: WeakRef = weakref(death_node)
		owner_node.get_tree().create_timer(sound_length + Constants.SHIP_DEATH_CLEANUP_PADDING).timeout.connect(
			func() -> void:
				var node := death_ref.get_ref() as Node
				if node != null and is_instance_valid(node):
					_queue_free_effect_node_once(node)
		)


func spawn_torpedo_explosion(event_position: Vector2) -> void:
	var explosion_node := TORPEDO_EXPLOSION_SCENE.instantiate()
	explosion_node.global_position = event_position
	explosion_node.z_index = Constants.EFFECT_Z_INDEX
	owner_node.add_child(explosion_node)

	var sprite := explosion_node.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D
	if sprite == null:
		explosion_node.queue_free()
		return
	var sound := explosion_node.get_node_or_null("TorpedoExplosionSound") as AudioStreamPlayer2D

	sprite.frame = 0
	sprite.frame_progress = 0.0
	_scale_sprite_to_diameter(sprite, _torpedo_explosion_diameter(), TORPEDO_EXPLOSION_SCALE_SOURCE_FRAME)
	sprite.play("torpedo_explosion")
	sprite.animation_finished.connect(_queue_free_effect_node_once.bind(explosion_node))
	if sound != null:
		sound.finished.connect(_queue_free_effect_node_once.bind(explosion_node))
		audio_flow.play_torpedo_explosion_sound(sound)


func _play_game_over_sound() -> void:
	if audio_flow.has_game_over_sound() && !game_over_sound_played:
		game_over_sound_played = true
		audio_flow.play_game_over_sound()


func _torpedo_explosion_diameter() -> float:
	return float(Constants.TORPEDO_RADIAL_ZONE_COUNT * Constants.TORPEDO_RADIAL_ZONE_WIDTH) * 2.0


func _scale_sprite_to_diameter(sprite: AnimatedSprite2D, target_diameter: float, source_frame: int) -> void:
	if sprite == null:
		return
	if sprite.sprite_frames == null:
		return
	if target_diameter <= 0:
		return
	var frame_count := sprite.sprite_frames.get_frame_count(sprite.animation)
	if source_frame < 0 || source_frame >= frame_count:
		return
	var texture := sprite.sprite_frames.get_frame_texture(sprite.animation, source_frame)
	if texture == null:
		return
	var source_diameter: float = max(texture.get_width(), texture.get_height())
	if source_diameter <= 0:
		return
	sprite.scale = Vector2.ONE * (target_diameter / source_diameter)


func _hide_effect_sprite(sprite: AnimatedSprite2D) -> void:
	if is_instance_valid(sprite):
		sprite.visible = false


func _queue_free_effect_node(effect_node: Node) -> void:
	if is_instance_valid(effect_node):
		effect_node.queue_free()


func _queue_free_effect_node_once(effect_node: Node) -> void:
	if !is_instance_valid(effect_node):
		return
	if effect_node.get_meta(EFFECT_CLEANUP_STARTED_META, false):
		return
	effect_node.set_meta(EFFECT_CLEANUP_STARTED_META, true)
	effect_node.queue_free()


func _on_game_over_sound_delay_timeout(token: int) -> void:
	if token == game_over_sound_token:
		_play_game_over_sound()
