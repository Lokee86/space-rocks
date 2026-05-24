extends RefCounted
class_name HudController

var score_label: Label
var lives_label: Label
var death_overlay: Control
var game_over_overlay: Control
var game_menu: GameMenu
var game_over_sound: AudioStreamPlayer
var cycle_view_label: Label
var respawn_timer_label: Label
var respawn_tell_label: Label
var room_id_label: Label
var respawn_timer_template := "Respawn in X"
var is_dead := false
var is_game_over := false
var is_suspended := false
var can_respawn := false
var respawn_countdown_remaining := 0.0
var room_id := ""
var is_multiplayer_session := false
var cycle_view_available := false


func configure(scene: Node) -> void:
	score_label = _find_label(scene, "Score")
	lives_label = _find_label(scene, "LivesCount")
	death_overlay = _find_message_container(scene, "YouDied")
	game_over_overlay = _find_label_container(scene, "GameOver")
	game_menu = _find_game_menu(scene)
	game_over_sound = _find_audio_stream_player(scene, "GameOverSound")
	cycle_view_label = _find_label(scene, "CycleView")
	respawn_timer_label = _find_label(death_overlay, "RespawnTimer")
	respawn_tell_label = _find_label(death_overlay, "RespawnTell")
	room_id_label = _find_label(scene, "RoomID")
	if respawn_timer_label != null:
		respawn_timer_template = respawn_timer_label.text

	set_score(0)
	set_lives(0)
	set_room_id("")
	set_alive()


func update(delta: float) -> void:
	if is_suspended:
		return
	if !is_dead || can_respawn || respawn_countdown_remaining <= 0:
		return

	respawn_countdown_remaining = max(0.0, respawn_countdown_remaining - delta)
	_update_respawn_timer_label()
	if respawn_countdown_remaining == 0:
		can_respawn = true
		if respawn_timer_label != null:
			if respawn_tell_label != null && respawn_timer_label.is_ancestor_of(respawn_tell_label):
				respawn_timer_label.text = ""
			else:
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


func set_room_id(room_id: String) -> void:
	self.room_id = room_id.strip_edges()
	_update_room_id_label()


func set_session_mode(value) -> void:
	is_multiplayer_session = str(value).strip_edges().to_lower() == "multiplayer"
	_update_room_id_label()


func _update_room_id_label() -> void:
	if room_id_label == null:
		return

	room_id_label.visible = is_multiplayer_session && room_id != ""
	if room_id != "":
		room_id_label.text = "ROOMID: %s" % room_id


func set_suspended(suspended: bool) -> void:
	is_suspended = suspended


func get_game_menu() -> GameMenu:
	return game_menu


func show_game_menu() -> void:
	if game_menu == null:
		return

	game_menu.visible = true
	_update_cycle_view_visibility()


func hide_game_menu() -> void:
	if game_menu == null:
		return

	game_menu.visible = false
	_update_cycle_view_visibility()


func is_game_menu_visible() -> bool:
	return game_menu != null && game_menu.visible


func set_alive() -> void:
	is_suspended = false
	is_dead = false
	is_game_over = false
	can_respawn = false
	cycle_view_available = false
	respawn_countdown_remaining = 0.0
	if death_overlay != null:
		death_overlay.visible = false
	if game_over_overlay != null:
		game_over_overlay.visible = false
	hide_game_menu()


func set_dead(respawn_delay: float) -> void:
	is_suspended = false
	is_dead = true
	is_game_over = false
	can_respawn = false
	cycle_view_available = false
	respawn_countdown_remaining = respawn_delay
	if death_overlay != null:
		death_overlay.visible = true
	if game_over_overlay != null:
		game_over_overlay.visible = false
	hide_game_menu()
	if respawn_timer_label != null:
		respawn_timer_label.text = respawn_timer_template
		respawn_timer_label.visible = true
	if respawn_tell_label != null:
		respawn_tell_label.visible = false
	_update_respawn_timer_label()


func set_game_over() -> void:
	is_suspended = false
	is_dead = true
	is_game_over = true
	can_respawn = false
	respawn_countdown_remaining = 0.0
	if death_overlay != null:
		death_overlay.visible = false
	if game_over_overlay != null:
		game_over_overlay.visible = true
	_update_cycle_view_visibility()


func set_cycle_view_available(available: bool) -> void:
	cycle_view_available = available
	_update_cycle_view_visibility()


func _update_respawn_timer_label() -> void:
	if respawn_timer_label == null:
		return

	var seconds_remaining := int(ceil(respawn_countdown_remaining))
	respawn_timer_label.text = respawn_timer_template.replace("X", str(seconds_remaining))


func _update_cycle_view_visibility() -> void:
	if cycle_view_label == null:
		return

	cycle_view_label.visible = (
		is_multiplayer_session &&
		is_game_over &&
		cycle_view_available &&
		!is_game_menu_visible()
	)


func _find_label(scene: Node, node_name: String) -> Label:
	if scene == null:
		return null

	return scene.find_child(node_name, true, false) as Label


func _find_message_container(scene: Node, label_name: String) -> Control:
	if scene == null:
		return null

	var label := scene.find_child(label_name, true, false) as Label
	if label == null:
		return null

	var container := label.get_parent()
	while container != null && container.get_parent() != null && container.get_parent().name != "CenterContainer":
		container = container.get_parent()

	return container as Control


func _find_label_container(scene: Node, label_name: String) -> Control:
	if scene == null:
		return null

	var label := scene.find_child(label_name, true, false) as Label
	if label == null:
		return null

	return label.get_parent() as Control


func _find_game_menu(scene: Node) -> GameMenu:
	if scene == null:
		return null

	return scene.find_child("GameMenu", true, false) as GameMenu


func _find_audio_stream_player(scene: Node, node_name: String) -> AudioStreamPlayer:
	if scene == null:
		return null

	var audio := scene.find_child(node_name, true, false) as AudioStreamPlayer
	if audio != null:
		return audio

	return null
