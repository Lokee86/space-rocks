extends Window

const ClientLogger = preload("res://scripts/logging/logger.gd")

signal toggle_invincible_requested(target_player_id: String)
signal toggle_infinite_lives_requested(target_player_id: String)
signal toggle_freeze_world_requested(freeze_target: String)
signal toggle_freeze_player_requested(target_player_id: String)
signal set_score_requested(target_player_id: String, score: int)
signal add_score_requested(target_player_id: String, amount: int)
signal set_lives_requested(target_player_id: String, lives: int)
signal add_lives_requested(target_player_id: String, amount: int)
signal clear_bullets_requested
signal clear_asteroids_requested
signal kill_player_requested(player_id: String)
signal spawn_asteroid_placement_requested
signal spawn_player_placement_requested(target_player_id: String)
signal spawn_bullet_placement_requested
signal respawn_player_placement_requested(target_player_id: String)
signal game_target_set_requested(target_player_id: String)
signal game_target_clear_requested()

@onready var invincible_button: Button = %InvincibleButton
@onready var infinite_lives_button: Button = %InfiniteLivesButton
@onready var freeze_world_button: Button = %FreezeWorldButton
@onready var freeze_asteroids_button: Button = %FreezeAsteroidsButton
@onready var freeze_bullets_button: Button = %FreezeBulletsButton
@onready var freeze_spawns_button: Button = %FreezeSpawnsButton
@onready var freeze_collisions_button: Button = %FreezeCollisionsButton
@onready var freeze_player_button: Button = %FreezePlayerButton
@onready var spawn_asteroid_button: Button = %SpawnAsteroidButton
@onready var spawn_player_button: Button = %SpawnPlayerButton
@onready var spawn_bullet_button: Button = %SpawnBulletButton
@onready var respawn_player_button: Button = %RespawnPlayerButton
@onready var spawn_player_select: OptionButton = %SpawnPlayerSelect
@onready var respawn_player_select: OptionButton = %RespawnPlayerSelect
@onready var invincible_status_select: OptionButton = %InvincibleStatusSelect
@onready var infinite_lives_select: OptionButton = %InfiniteLivesSelect
@onready var world_frozen_status_label: Label = %WorldFrozenStatusLabel
@onready var freeze_asteroids_status_label: Label = %FreezeAsteroidsStatusLabel
@onready var freeze_bullets_status_label: Label = %FreezeBulletsStatusLabel
@onready var freeze_spawns_status_label: Label = %FreezeSpawnsStatusLabel
@onready var freeze_collisions_status_label: Label = %FreezeCollisionsStatusLabel
@onready var player_frozen_select: OptionButton = %PlayerFrozenSelect
@onready var kill_player_button: Button = %KillPlayerButton
@onready var kill_player_select: OptionButton = %KillPlayerSelect
@onready var set_score_amount: LineEdit = %SetScoreAmount
@onready var set_score_select: OptionButton = %SetScoreSelect
@onready var set_score_button: Button = %SetScoreButton
@onready var add_score_amount: LineEdit = %AddScoreAmount
@onready var add_score_select: OptionButton = %AddScoreSelect
@onready var add_score_button: Button = %AddScoreButton
@onready var set_lives_amount: LineEdit = %SetLivesAmount
@onready var set_lives_select: OptionButton = %SetLivesSelect
@onready var set_lives_button: Button = %SetLivesButton
@onready var add_lives_amount: LineEdit = %AddLivesAmount
@onready var add_lives_select: OptionButton = %AddLivesSelect
@onready var add_lives_button: Button = %AddLivesButton
@onready var clear_bullets_button: Button = %ClearBulletsButton
@onready var clear_asteroids_button: Button = %ClearAsteroidsButton
@onready var game_target_select: OptionButton = %GameTargetSelect
@onready var set_game_target_button: Button = %SetGameTargetButton
@onready var clear_game_target_button: Button = %ClearGameTargetButton


func _ready() -> void:
	if !close_requested.is_connected(_on_close_requested):
		close_requested.connect(_on_close_requested)
	if !invincible_button.pressed.is_connected(_on_invincible_button_pressed):
		invincible_button.pressed.connect(_on_invincible_button_pressed)
	if !infinite_lives_button.pressed.is_connected(_on_infinite_lives_button_pressed):
		infinite_lives_button.pressed.connect(_on_infinite_lives_button_pressed)
	if !freeze_world_button.pressed.is_connected(_on_freeze_world_button_pressed):
		freeze_world_button.pressed.connect(_on_freeze_world_button_pressed)
	if !freeze_asteroids_button.pressed.is_connected(_on_freeze_asteroids_button_pressed):
		freeze_asteroids_button.pressed.connect(_on_freeze_asteroids_button_pressed)
	if !freeze_bullets_button.pressed.is_connected(_on_freeze_bullets_button_pressed):
		freeze_bullets_button.pressed.connect(_on_freeze_bullets_button_pressed)
	if !freeze_spawns_button.pressed.is_connected(_on_freeze_spawns_button_pressed):
		freeze_spawns_button.pressed.connect(_on_freeze_spawns_button_pressed)
	if !freeze_collisions_button.pressed.is_connected(_on_freeze_collisions_button_pressed):
		freeze_collisions_button.pressed.connect(_on_freeze_collisions_button_pressed)
	if !freeze_player_button.pressed.is_connected(_on_freeze_player_button_pressed):
		freeze_player_button.pressed.connect(_on_freeze_player_button_pressed)
	if !spawn_asteroid_button.pressed.is_connected(_on_spawn_asteroid_button_pressed):
		spawn_asteroid_button.pressed.connect(_on_spawn_asteroid_button_pressed)
	if !spawn_player_button.pressed.is_connected(_on_spawn_player_button_pressed):
		spawn_player_button.pressed.connect(_on_spawn_player_button_pressed)
	if !spawn_bullet_button.pressed.is_connected(_on_spawn_bullet_button_pressed):
		spawn_bullet_button.pressed.connect(_on_spawn_bullet_button_pressed)
	if !respawn_player_button.pressed.is_connected(_on_respawn_player_button_pressed):
		respawn_player_button.pressed.connect(_on_respawn_player_button_pressed)
	if !kill_player_button.pressed.is_connected(_on_kill_player_button_pressed):
		kill_player_button.pressed.connect(_on_kill_player_button_pressed)
	if !set_score_button.pressed.is_connected(_on_set_score_button_pressed):
		set_score_button.pressed.connect(_on_set_score_button_pressed)
	if !add_score_button.pressed.is_connected(_on_add_score_button_pressed):
		add_score_button.pressed.connect(_on_add_score_button_pressed)
	if !set_lives_button.pressed.is_connected(_on_set_lives_button_pressed):
		set_lives_button.pressed.connect(_on_set_lives_button_pressed)
	if !add_lives_button.pressed.is_connected(_on_add_lives_button_pressed):
		add_lives_button.pressed.connect(_on_add_lives_button_pressed)
	if !clear_bullets_button.pressed.is_connected(_on_clear_bullets_button_pressed):
		clear_bullets_button.pressed.connect(_on_clear_bullets_button_pressed)
	if !clear_asteroids_button.pressed.is_connected(_on_clear_asteroids_button_pressed):
		clear_asteroids_button.pressed.connect(_on_clear_asteroids_button_pressed)
	if !set_game_target_button.pressed.is_connected(_on_set_game_target_button_pressed):
		set_game_target_button.pressed.connect(_on_set_game_target_button_pressed)
	if !clear_game_target_button.pressed.is_connected(_on_clear_game_target_button_pressed):
		clear_game_target_button.pressed.connect(_on_clear_game_target_button_pressed)


func show_window() -> void:
	popup_centered()


func hide_window() -> void:
	hide()


func toggle_window() -> void:
	if visible:
		hide_window()
	else:
		show_window()


func set_debug_status(status: Dictionary) -> void:
	world_frozen_status_label.text = _active_inactive(status.get("world_frozen", false))
	freeze_asteroids_status_label.text = _active_inactive(status.get("asteroids_frozen", false))
	freeze_bullets_status_label.text = _active_inactive(status.get("bullets_frozen", false))
	freeze_spawns_status_label.text = _active_inactive(status.get("spawning_frozen", false))
	freeze_collisions_status_label.text = _active_inactive(status.get("collisions_frozen", false))


func refresh_invincible_targets(rows: Array) -> void:
	_refresh_target_option(invincible_status_select, rows, "", true, true)


func refresh_infinite_lives_targets(rows: Array) -> void:
	_refresh_target_option(infinite_lives_select, rows, "", true, true)


func refresh_player_frozen_targets(rows: Array) -> void:
	_refresh_target_option(player_frozen_select, rows, "", true, true)


func refresh_kill_player_targets(target_rows: Array) -> void:
	var previous_player_id := ""
	var previous_index := kill_player_select.get_selected()
	if previous_index >= 0:
		previous_player_id = str(kill_player_select.get_item_metadata(previous_index))

	kill_player_select.clear()

	var selected_index := -1
	for row in target_rows:
		if !(row is Dictionary):
			continue

		var label := str(row.get("label", ""))
		var player_id := str(row.get("player_id", ""))
		kill_player_select.add_item(label)
		var item_index := kill_player_select.get_item_count() - 1
		kill_player_select.set_item_metadata(item_index, player_id)
		if player_id == previous_player_id:
			selected_index = item_index

	if selected_index >= 0:
		kill_player_select.select(selected_index)


func refresh_spawn_player_slots(max_players: int) -> void:
	var previous_player_id := ""
	var previous_index := spawn_player_select.get_selected()
	if previous_index >= 0:
		previous_player_id = str(spawn_player_select.get_item_metadata(previous_index))

	spawn_player_select.clear()
	spawn_player_select.add_item("New Player")
	spawn_player_select.set_item_metadata(0, "")

	var selected_index := 0
	if max_players > 0:
		for index in range(max_players):
			var player_id := "Player-%d" % (index + 1)
			spawn_player_select.add_item(player_id)
			var item_index := spawn_player_select.get_item_count() - 1
			spawn_player_select.set_item_metadata(item_index, player_id)
			if player_id == previous_player_id:
				selected_index = item_index

	spawn_player_select.select(selected_index)


func refresh_respawn_player_targets(target_rows: Array) -> void:
	var previous_player_id := ""
	var previous_index := respawn_player_select.get_selected()
	if previous_index >= 0:
		previous_player_id = str(respawn_player_select.get_item_metadata(previous_index))

	respawn_player_select.clear()

	var selected_index := -1
	for row in target_rows:
		if !(row is Dictionary):
			continue

		var label := str(row.get("label", ""))
		var player_id := str(row.get("player_id", ""))
		respawn_player_select.add_item(label)
		var item_index := respawn_player_select.get_item_count() - 1
		respawn_player_select.set_item_metadata(item_index, player_id)
		if player_id == previous_player_id:
			selected_index = item_index

	if selected_index >= 0:
		respawn_player_select.select(selected_index)


func refresh_counter_player_targets(rows: Array) -> void:
	_refresh_target_option(set_score_select, rows, "", true, true)
	_refresh_target_option(add_score_select, rows, "", true, true)
	_refresh_target_option(set_lives_select, rows, "", true, true)
	_refresh_target_option(add_lives_select, rows, "", true, true)


func refresh_game_target_options(
	rows: Array,
	current_target_player_id: String = "",
	current_target_kind: String = "",
	current_target_id: String = ""
) -> void:
	var previous_player_id := ""
	var previous_index := game_target_select.get_selected()
	if previous_index >= 0:
		previous_player_id = str(game_target_select.get_item_metadata(previous_index))

	game_target_select.clear()
	game_target_select.add_item(_format_game_target_display(current_target_kind, current_target_id))
	game_target_select.set_item_metadata(0, "")

	var selected_index := -1
	for row in rows:
		if !(row is Dictionary):
			continue

		var label := str(row.get("label", ""))
		var player_id := str(row.get("player_id", ""))
		if player_id == DevtoolsTargetResolver.TARGET_GAME:
			continue
		game_target_select.add_item(label)
		var item_index := game_target_select.get_item_count() - 1
		game_target_select.set_item_metadata(item_index, player_id)
		if current_target_player_id != "" and player_id == current_target_player_id:
			selected_index = item_index
		elif current_target_player_id == "" and player_id == previous_player_id:
			selected_index = item_index

	if selected_index >= 0:
		game_target_select.select(selected_index)
	else:
		game_target_select.select(0)


func _on_close_requested() -> void:
	hide_window()


func _on_invincible_button_pressed() -> void:
	toggle_invincible_requested.emit(_selected_metadata_as_string(invincible_status_select))


func _on_infinite_lives_button_pressed() -> void:
	toggle_infinite_lives_requested.emit(_selected_metadata_as_string(infinite_lives_select))


func _on_freeze_world_button_pressed() -> void:
	toggle_freeze_world_requested.emit("")


func _on_freeze_asteroids_button_pressed() -> void:
	toggle_freeze_world_requested.emit("asteroids")


func _on_freeze_bullets_button_pressed() -> void:
	toggle_freeze_world_requested.emit("bullets")


func _on_freeze_spawns_button_pressed() -> void:
	toggle_freeze_world_requested.emit("spawns")


func _on_freeze_collisions_button_pressed() -> void:
	toggle_freeze_world_requested.emit("collisions")


func _on_freeze_player_button_pressed() -> void:
	toggle_freeze_player_requested.emit(_selected_metadata_as_string(player_frozen_select))


func _on_set_score_button_pressed() -> void:
	if !_line_edit_has_valid_int(set_score_amount):
		return
	var target_player_id: String = _selected_player_id(set_score_select)
	if target_player_id == "":
		return
	var score: int = _line_edit_int(set_score_amount)
	set_score_requested.emit(target_player_id, score)


func _on_add_score_button_pressed() -> void:
	if !_line_edit_has_valid_int(add_score_amount):
		return
	var target_player_id: String = _selected_player_id(add_score_select)
	if target_player_id == "":
		return
	var amount: int = _line_edit_int(add_score_amount)
	add_score_requested.emit(target_player_id, amount)


func _on_set_lives_button_pressed() -> void:
	if !_line_edit_has_valid_int(set_lives_amount):
		return
	var target_player_id: String = _selected_player_id(set_lives_select)
	if target_player_id == "":
		return
	var lives: int = _line_edit_int(set_lives_amount)
	set_lives_requested.emit(target_player_id, lives)


func _on_add_lives_button_pressed() -> void:
	if !_line_edit_has_valid_int(add_lives_amount):
		return
	var target_player_id: String = _selected_player_id(add_lives_select)
	if target_player_id == "":
		return
	var amount: int = _line_edit_int(add_lives_amount)
	add_lives_requested.emit(target_player_id, amount)


func _on_clear_bullets_button_pressed() -> void:
	clear_bullets_requested.emit()


func _on_clear_asteroids_button_pressed() -> void:
	clear_asteroids_requested.emit()


func _on_set_game_target_button_pressed() -> void:
	var target_player_id := _selected_metadata_as_string(game_target_select)
	if target_player_id == "":
		return
	game_target_set_requested.emit(target_player_id)


func _on_clear_game_target_button_pressed() -> void:
	game_target_clear_requested.emit()


func _on_kill_player_button_pressed() -> void:
	var selected_index := kill_player_select.get_selected()
	if selected_index < 0:
		return

	var selected_player_id := str(kill_player_select.get_item_metadata(selected_index))
	if selected_player_id == "":
		return

	kill_player_requested.emit(selected_player_id)


func _on_spawn_asteroid_button_pressed() -> void:
	spawn_asteroid_placement_requested.emit()


func _on_spawn_player_button_pressed() -> void:
	spawn_player_placement_requested.emit(_selected_metadata_as_string(spawn_player_select))


func _on_spawn_bullet_button_pressed() -> void:
	spawn_bullet_placement_requested.emit()


func _on_respawn_player_button_pressed() -> void:
	var target_player_id := _selected_metadata_as_string(respawn_player_select)
	ClientLogger.game_info("Devtools respawn button pressed")
	ClientLogger.game_info("Devtools respawn selected target_player_id='%s'" % target_player_id)
	respawn_player_placement_requested.emit(target_player_id)


func _active_inactive(value) -> String:
	if bool(value):
		return "Active"
	return "Inactive"


func _format_game_target_display(target_kind: String, target_id: String) -> String:
	if target_kind == "" or target_id == "":
		return "\u2014"
	return "%s:%s" % [target_kind, target_id]


func _selected_metadata_as_string(select: OptionButton) -> String:
	var selected_index := select.get_selected()
	if selected_index < 0:
		return ""
	return str(select.get_item_metadata(selected_index))


func _selected_player_id(select: OptionButton) -> String:
	return _selected_metadata_as_string(select)


func _line_edit_has_valid_int(input: LineEdit) -> bool:
	var text: String = input.text.strip_edges()
	return text.is_valid_int()


func _line_edit_int(input: LineEdit) -> int:
	var text: String = input.text.strip_edges()
	if !text.is_valid_int():
		return 0
	return int(text)


func _refresh_target_option(
	select: OptionButton,
	rows: Array,
	preferred_player_id: String = "",
	include_game_target: bool = true,
	default_to_game_target: bool = false
) -> void:
	var previous_player_id := ""
	var previous_index := select.get_selected()
	if previous_index >= 0:
		previous_player_id = str(select.get_item_metadata(previous_index))

	select.clear()

	var selected_index := -1
	var game_target_index := -1
	for row in rows:
		if !(row is Dictionary):
			continue

		var label := str(row.get("label", ""))
		var player_id := str(row.get("player_id", ""))
		if player_id == DevtoolsTargetResolver.TARGET_GAME and !include_game_target:
			continue
		select.add_item(label)
		var item_index := select.get_item_count() - 1
		select.set_item_metadata(item_index, player_id)
		if player_id == DevtoolsTargetResolver.TARGET_GAME:
			game_target_index = item_index
		if preferred_player_id != "" and player_id == preferred_player_id:
			selected_index = item_index
		elif preferred_player_id == "" and player_id == previous_player_id:
			if player_id != DevtoolsTargetResolver.TARGET_GAME:
				selected_index = item_index

	if selected_index < 0 and default_to_game_target and game_target_index >= 0:
		selected_index = game_target_index

	if selected_index >= 0:
		select.select(selected_index)
