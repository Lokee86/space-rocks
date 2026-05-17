extends RefCounted
class_name HudController

const Constants = preload("res://scripts/constants.gd")

var score_label: Label
var lives_label: Label
var death_overlay: Control
var game_over_overlay: Control
var game_over_sound: AudioStreamPlayer
var respawn_timer_label: Label
var respawn_tell_label: Label
var respawn_timer_template := "Respawn in X"
var is_dead := false
var can_respawn := false
var respawn_countdown_remaining := 0.0


func configure(scene: Node) -> void:
	score_label = _find_label(scene, "Score")
	lives_label = _find_label(scene, "LivesCount")
	death_overlay = _find_control(scene, "DeathOverlay")
	game_over_overlay = _find_control(scene, "GameOverOverlay")
	game_over_sound = _find_game_over_sound()
	respawn_timer_label = _find_label(scene, "RespawnTimer")
	respawn_tell_label = _find_label(scene, "RespawnTell")
	if respawn_timer_label != null:
		respawn_timer_template = respawn_timer_label.text

	set_score(0)
	set_lives(Constants.PLAYER_STARTING_LIVES)
	set_alive()


func update(delta: float) -> void:
	if !is_dead || can_respawn || respawn_countdown_remaining <= 0:
		return

	respawn_countdown_remaining = max(0.0, respawn_countdown_remaining - delta)
	_update_respawn_timer_label()
	if respawn_countdown_remaining == 0:
		can_respawn = true
		if respawn_timer_label != null:
			respawn_timer_label.visible = false
		if respawn_tell_label != null:
			respawn_tell_label.visible = true


func set_score(score: int) -> void:
	if score_label == null:
		return

	score_label.text = "SCORE: %d" % score


func set_lives(lives: int) -> void:
	if lives_label == null:
		return

	lives_label.text = "%d x " % lives


func set_alive() -> void:
	is_dead = false
	can_respawn = false
	respawn_countdown_remaining = 0.0
	if death_overlay != null:
		death_overlay.visible = false
	if game_over_overlay != null:
		game_over_overlay.visible = false


func set_dead(respawn_delay: float) -> void:
	is_dead = true
	can_respawn = false
	respawn_countdown_remaining = respawn_delay
	if death_overlay != null:
		death_overlay.visible = true
	if game_over_overlay != null:
		game_over_overlay.visible = false
	if respawn_timer_label != null:
		respawn_timer_label.visible = true
	if respawn_tell_label != null:
		respawn_tell_label.visible = false
	_update_respawn_timer_label()


func set_game_over() -> void:
	is_dead = true
	can_respawn = false
	respawn_countdown_remaining = 0.0
	if death_overlay != null:
		death_overlay.visible = false
	if game_over_overlay != null:
		game_over_overlay.visible = true


func _update_respawn_timer_label() -> void:
	if respawn_timer_label == null:
		return

	var seconds_remaining := int(ceil(respawn_countdown_remaining))
	respawn_timer_label.text = respawn_timer_template.replace("X", str(seconds_remaining))


func _find_label(scene: Node, node_name: String) -> Label:
	if scene == null:
		return null

	return scene.find_child(node_name, true, false) as Label


func _find_control(scene: Node, node_name: String) -> Control:
	if scene == null:
		return null

	return scene.find_child(node_name, true, false) as Control


func _find_game_over_sound() -> AudioStreamPlayer:
	if game_over_overlay == null:
		return null

	return game_over_overlay.find_child("GameOverSound", true, false) as AudioStreamPlayer
