extends RefCounted
class_name GameplayHudFlow

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")

var hud: Control
var hidden_for_match_over := false
var is_dead := false
var is_game_over := false
var can_respawn := false
var current_score := 0
var respawn_countdown_remaining := 0.0
var respawn_timer_template := ""
var loadout_display_flow := LoadoutDisplayFlow.new()
var _logged_set_dead_diagnostics := false
var _logged_respawn_available := false


func configure(hud_ref: Control) -> void:
	hud = hud_ref
	loadout_display_flow.configure(hud)
	var respawn_timer_label := _respawn_timer_label()
	if respawn_timer_label != null:
		respawn_timer_template = respawn_timer_label.text
	set_alive()


func show_gameplay() -> void:
	if hud == null:
		return
	if hidden_for_match_over:
		hud.hide()
		return

	hud.show()
	_hide_hud_child("RoomID")


func apply_gameplay_state_summary(state: Dictionary) -> void:
	show_gameplay()
	if state.get("has_lives", false):
		apply_lives(int(state.get("lives", 0)))
	var player_sessions: Dictionary = state.get("player_sessions", {})
	var self_id: String = str(state.get("self_id", ""))
	var self_session_state = player_sessions.get(self_id, {})
	if self_session_state is Dictionary and self_session_state.has(Packets.FIELD_SCORE):
		apply_score(int(self_session_state.get(Packets.FIELD_SCORE, 0)))
	var server_players: Dictionary = state.get("server_players", {})
	var self_player_state = server_players.get(self_id, {})
	if self_player_state is Dictionary:
		loadout_display_flow.apply_player_state(self_player_state)


func apply_overlay_lane_state(overlay_lane_state) -> void:
	if overlay_lane_state == null:
		return

	show_gameplay()
	if overlay_lane_state.self_id != null:
		var self_id := str(overlay_lane_state.self_id)
		if self_id != "":
			loadout_display_flow.clear()
	if overlay_lane_state.lives != null:
		apply_lives(int(overlay_lane_state.lives))
	if overlay_lane_state.score != null:
		apply_score(int(overlay_lane_state.score))

	var player_state := {
		"primary_weapon_id": overlay_lane_state.primary_weapon_id,
		"secondary_weapon_id": overlay_lane_state.secondary_weapon_id,
		"primary_ammo_policy": overlay_lane_state.primary_ammo_policy,
		"secondary_ammo_policy": overlay_lane_state.secondary_ammo_policy,
		"primary_cooldown_remaining": overlay_lane_state.primary_cooldown_remaining,
		"secondary_cooldown_remaining": overlay_lane_state.secondary_cooldown_remaining,
		"primary_ammo_remaining": overlay_lane_state.primary_ammo_remaining,
		"secondary_ammo_remaining": overlay_lane_state.secondary_ammo_remaining,
	}
	loadout_display_flow.apply_player_state(player_state)

func apply_session_lane_state(session_lane_state, self_id := "") -> void:
	if session_lane_state == null:
		return

	show_gameplay()
	if self_id != "" and session_lane_state.player_sessions != null and session_lane_state.player_sessions.has(self_id):
		var self_session = session_lane_state.player_sessions[self_id]
		if self_session is Dictionary:
			if self_session.has(Packets.FIELD_SCORE):
				apply_score(int(self_session.get(Packets.FIELD_SCORE, 0)))
			if self_session.has(Packets.FIELD_LIVES):
				apply_lives(int(self_session.get(Packets.FIELD_LIVES, 0)))
			if self_session.has(Packets.FIELD_RESPAWN_COOLDOWN):
				var respawn_cooldown := float(self_session.get(Packets.FIELD_RESPAWN_COOLDOWN, 0.0))
				if respawn_cooldown > 0.0:
					set_dead(respawn_cooldown)

	if session_lane_state.total_asteroids != null:
		var total_asteroids := int(session_lane_state.total_asteroids)
		if total_asteroids >= 0:
			show_gameplay()
func reset() -> void:
	hidden_for_match_over = false
	_logged_set_dead_diagnostics = false
	_logged_respawn_available = false
	if hud != null:
		set_alive()
		loadout_display_flow.clear()
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


func hide_for_match_over() -> void:
	hidden_for_match_over = true
	if hud != null:
		hud.hide()


func clear_match_over_visibility_lock() -> void:
	hidden_for_match_over = false


func set_alive() -> void:
	is_dead = false
	is_game_over = false
	can_respawn = false
	current_score = 0
	respawn_countdown_remaining = 0.0
	_hide_hud_child("CenterContainer/VBoxContainer2")
	_hide_hud_child("CenterContainer/GameOverContainer")


func clear_dead_presentation() -> void:
	is_dead = false
	can_respawn = false
	respawn_countdown_remaining = 0.0
	_hide_hud_child("CenterContainer/VBoxContainer2")
	if !hidden_for_match_over:
		_hide_hud_child("CenterContainer/GameOverContainer")


func set_dead(respawn_delay: float) -> void:
	is_dead = true
	is_game_over = false
	can_respawn = false
	respawn_countdown_remaining = maxf(respawn_delay, 0.0)
	if !_logged_set_dead_diagnostics:
		print("GameplayHudFlow.set_dead: respawn_delay=%s can_respawn=%s" % [str(respawn_countdown_remaining), str(can_respawn)])
		_logged_set_dead_diagnostics = true
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
	_show_hud_child("CenterContainer/GameOverContainer/MarginContainer")


func apply_score(score_value: int) -> void:
	current_score = score_value
	var score_label := _get_hud_child("MarginContainer/HBoxContainer/MarginContainer/Score") as Label
	if score_label != null:
		score_label.text = "SCORE: %d" % score_value


func score() -> int:
	return current_score


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


func _has_dead_presentation() -> bool:
	return is_dead or can_respawn


func _make_respawn_available() -> void:
	can_respawn = true
	if !_logged_respawn_available:
		print("GameplayHudFlow._make_respawn_available: can_respawn=%s" % str(can_respawn))
		_logged_respawn_available = true
	var respawn_timer_label := _respawn_timer_label()
	if respawn_timer_label != null:
		respawn_timer_label.text = ""
	_show_hud_child("CenterContainer/VBoxContainer2/RespawnTimer/RespawnTell")


func _get_hud_child(path: NodePath) -> Node:
	if hud == null:
		return null
	return hud.get_node_or_null(path)
