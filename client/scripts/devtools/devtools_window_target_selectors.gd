extends RefCounted
class_name DevtoolsWindowTargetSelectors

var invincible_status_select: OptionButton
var infinite_lives_select: OptionButton
var player_frozen_select: OptionButton
var kill_player_select: OptionButton
var respawn_player_select: OptionButton
var set_score_select: OptionButton
var add_score_select: OptionButton
var set_lives_select: OptionButton
var add_lives_select: OptionButton


func configure(
	invincible_select: OptionButton,
	infinite_lives_select_ref: OptionButton,
	player_frozen_select_ref: OptionButton,
	kill_player_select_ref: OptionButton,
	respawn_player_select_ref: OptionButton,
	set_score_select_ref: OptionButton,
	add_score_select_ref: OptionButton,
	set_lives_select_ref: OptionButton,
	add_lives_select_ref: OptionButton
) -> void:
	invincible_status_select = invincible_select
	infinite_lives_select = infinite_lives_select_ref
	player_frozen_select = player_frozen_select_ref
	kill_player_select = kill_player_select_ref
	respawn_player_select = respawn_player_select_ref
	set_score_select = set_score_select_ref
	add_score_select = add_score_select_ref
	set_lives_select = set_lives_select_ref
	add_lives_select = add_lives_select_ref


func refresh_invincible_targets(rows: Array) -> void:
	_refresh_target_option(invincible_status_select, rows)


func refresh_infinite_lives_targets(rows: Array) -> void:
	_refresh_target_option(infinite_lives_select, rows)


func refresh_player_frozen_targets(rows: Array) -> void:
	_refresh_target_option(player_frozen_select, rows)


func refresh_kill_player_targets(rows: Array) -> void:
	_refresh_target_option(kill_player_select, rows, false)


func refresh_respawn_player_targets(rows: Array) -> void:
	_refresh_target_option(respawn_player_select, rows, false)


func refresh_counter_player_targets(rows: Array) -> void:
	_refresh_target_option(set_score_select, rows)
	_refresh_target_option(add_score_select, rows)
	_refresh_target_option(set_lives_select, rows)
	_refresh_target_option(add_lives_select, rows)


func _refresh_target_option(select: OptionButton, rows: Array, default_to_game_target := true) -> void:
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
		select.add_item(label)
		var item_index := select.get_item_count() - 1
		select.set_item_metadata(item_index, player_id)
		if player_id == DevtoolsTargetResolver.TARGET_GAME:
			game_target_index = item_index
		if player_id == previous_player_id:
			selected_index = item_index

	if selected_index < 0 and default_to_game_target and game_target_index >= 0:
		selected_index = game_target_index
	if selected_index >= 0:
		select.select(selected_index)