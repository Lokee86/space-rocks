extends RefCounted
class_name GameplayAudioFlow


func play_bullet_blast_sound(sound: AudioStreamPlayer2D) -> void:
	if sound == null:
		return
	sound.play()


func play_ship_death_sound(sound: AudioStreamPlayer2D) -> void:
	if sound == null:
		return
	sound.play()
