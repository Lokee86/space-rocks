extends RefCounted
class_name GameplayAudioFlow

var game_over_sound: AudioStreamPlayer


func configure(hud: Control) -> void:
	if hud == null:
		game_over_sound = null
		return
	game_over_sound = hud.get_node_or_null("%GameOverSound") as AudioStreamPlayer


func play_bullet_blast_sound(sound: AudioStreamPlayer2D) -> void:
	if sound == null:
		return
	sound.play()


func play_pickup_collected_sound(sound: AudioStreamPlayer2D) -> void:
	if sound == null:
		return
	sound.play()


func play_pickup_spawned_sound(sound: AudioStreamPlayer2D) -> void:
	if sound == null:
		return
	sound.play()


func play_projectile_firing_sound(sound: AudioStreamPlayer2D, parent: Node) -> void:
	if sound == null:
		return
	if parent == null:
		return
	var detached_sound := sound.duplicate() as AudioStreamPlayer2D
	if detached_sound == null:
		return
	parent.add_child(detached_sound)
	detached_sound.global_position = sound.global_position
	detached_sound.play()
	detached_sound.finished.connect(detached_sound.queue_free)


func play_ship_death_sound(sound: AudioStreamPlayer2D) -> void:
	if sound == null:
		return
	sound.play()


func play_torpedo_explosion_sound(sound: AudioStreamPlayer2D) -> void:
	if sound == null:
		return
	sound.play()


func play_afterburner_sound(sound: AudioStreamPlayer2D) -> void:
	if sound == null:
		return
	sound.play()


func stop_afterburner_sound(sound: AudioStreamPlayer2D) -> void:
	if sound == null:
		return
	sound.stop()


func play_game_over_sound() -> void:
	if game_over_sound == null:
		return
	game_over_sound.play()


func has_game_over_sound() -> bool:
	return is_instance_valid(game_over_sound)


func stop_game_over_sound() -> void:
	if !is_instance_valid(game_over_sound):
		return
	game_over_sound.stop()

