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


func play_laser_sound(sound: AudioStreamPlayer2D) -> void:
	if sound == null:
		return
	sound.play()


func play_ship_death_sound(sound: AudioStreamPlayer2D) -> void:
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
