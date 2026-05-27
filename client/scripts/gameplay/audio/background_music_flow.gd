extends AudioStreamPlayer
class_name BackgroundMusicFlow


func _ready() -> void:
	start_music()


func start_music() -> void:
	if not playing:
		play()


func stop_music() -> void:
	stop()


func ensure_music_playing() -> void:
	if not playing:
		play()
