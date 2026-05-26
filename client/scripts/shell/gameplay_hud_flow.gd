extends RefCounted
class_name GameplayHudFlow

var hud: Control
var is_dead := false
var is_game_over := false
var can_respawn := false
var respawn_countdown_remaining := 0.0
var respawn_timer_template := ""


func configure(hud_ref: Control) -> void:
	hud = hud_ref
	var respawn_timer_label := _respawn_timer_label()
	if respawn_timer_label != null:
		respawn_timer_template = respawn_timer_label.text
	set_alive()


func show_gameplay() -> void:
	if hud == null:
		return

	hud.show()
	_hide_hud_child("RoomID")


func reset() -> void:
	if hud != null:
		set_alive()
		hud.hide()


func update(delta: float) -> void:
	if !is_dead || is_game_over || can_respawn:
		return

	respawn_countdown_remaining = maxf(respawn_countdown_remaining - delta, 0.0)
	if respawn_countdown_remaining <= 0.0:
		_make_respawn_available()
		return

	var respawn_timer_label := _respawn_timer_label()
	if respawn_timer_label != null:
		respawn_timer_label.text = _respawn_timer_text(respawn_countdown_remaining)


func can_request_respawn() -> bool:
	return is_dead && !is_game_over && can_respawn


func set_alive() -> void:
	is_dead = false
	is_game_over = false
	can_respawn = false
	respawn_countdown_remaining = 0.0
	_hide_hud_child("CenterContainer/VBoxContainer2")
	_hide_hud_child("CenterContainer/GameOverContainer")


func set_dead(respawn_delay: float) -> void:
	is_dead = true
	is_game_over = false
	can_respawn = false
	respawn_countdown_remaining = maxf(respawn_delay, 0.0)
	_show_hud_child("CenterContainer/VBoxContainer2")
	_hide_hud_child("CenterContainer/GameOverContainer")

	if respawn_countdown_remaining <= 0.0:
		_make_respawn_available()
		return

	var respawn_timer_label := _respawn_timer_label()
	if respawn_timer_label != null:
		respawn_timer_label.text = _respawn_timer_text(respawn_countdown_remaining)
		respawn_timer_label.show()
	_hide_hud_child("CenterContainer/VBoxContainer2/RespawnTimer/RespawnTell")


func set_game_over() -> void:
	is_dead = false
	is_game_over = true
	can_respawn = false
	respawn_countdown_remaining = 0.0
	_hide_hud_child("CenterContainer/VBoxContainer2")
	_show_hud_child("CenterContainer/GameOverContainer")


func apply_score(score: int) -> void:
	var score_label := _get_hud_child("MarginContainer/HBoxContainer/MarginContainer/Score") as Label
	if score_label != null:
		score_label.text = "SCORE: %d" % score


func apply_lives(lives: int) -> void:
	var lives_label := _get_hud_child(
		"MarginContainer/HBoxContainer/LivesContainer/MarginContainer/LivesCount"
	) as Label
	if lives_label != null:
		lives_label.text = "%d x " % lives


func _hide_hud_child(path: NodePath) -> void:
	var child := _get_hud_child(path) as CanvasItem
	if child != null:
		child.hide()


func _show_hud_child(path: NodePath) -> void:
	var child := _get_hud_child(path) as CanvasItem
	if child != null:
		child.show()


func _respawn_timer_label() -> Label:
	return _get_hud_child("CenterContainer/VBoxContainer2/RespawnTimer") as Label


func _respawn_timer_text(respawn_delay: float) -> String:
	var template := respawn_timer_template
	if template == "":
		template = "Respawn in X"
	return template.replace("X", str(ceili(respawn_delay)))


func _make_respawn_available() -> void:
	can_respawn = true
	var respawn_timer_label := _respawn_timer_label()
	if respawn_timer_label != null:
		respawn_timer_label.text = ""
	_show_hud_child("CenterContainer/VBoxContainer2/RespawnTimer/RespawnTell")


func _get_hud_child(path: NodePath) -> Node:
	if hud == null:
		return null
	return hud.get_node_or_null(path)
