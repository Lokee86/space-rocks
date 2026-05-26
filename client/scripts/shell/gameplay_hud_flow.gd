extends RefCounted
class_name GameplayHudFlow

var hud: Control


func configure(hud_ref: Control) -> void:
	hud = hud_ref


func show_gameplay() -> void:
	if hud == null:
		return

	hud.show()
	_hide_hud_child("CenterContainer/VBoxContainer2")
	_hide_hud_child("CenterContainer/GameOverContainer")
	_hide_hud_child("RoomID")


func reset() -> void:
	if hud != null:
		hud.hide()


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


func _get_hud_child(path: NodePath) -> Node:
	if hud == null:
		return null
	return hud.get_node_or_null(path)
