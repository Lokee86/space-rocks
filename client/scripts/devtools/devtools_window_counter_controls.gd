extends RefCounted
class_name DevtoolsWindowCounterControls

var window

func configure(window_ref) -> void:
	window = window_ref

func initialize() -> void:
	if !window.set_score_button.pressed.is_connected(_on_set_score_button_pressed):
		window.set_score_button.pressed.connect(_on_set_score_button_pressed)
	if !window.add_score_button.pressed.is_connected(_on_add_score_button_pressed):
		window.add_score_button.pressed.connect(_on_add_score_button_pressed)
	if !window.set_lives_button.pressed.is_connected(_on_set_lives_button_pressed):
		window.set_lives_button.pressed.connect(_on_set_lives_button_pressed)
	if !window.add_lives_button.pressed.is_connected(_on_add_lives_button_pressed):
		window.add_lives_button.pressed.connect(_on_add_lives_button_pressed)

func refresh_targets(rows: Array) -> void:
	_refresh_target_option(window.set_score_select, rows)
	_refresh_target_option(window.add_score_select, rows)
	_refresh_target_option(window.set_lives_select, rows)
	_refresh_target_option(window.add_lives_select, rows)

func _on_set_score_button_pressed() -> void:
	_emit_counter_request(window.set_score_amount, window.set_score_select, window.set_score_requested)

func _on_add_score_button_pressed() -> void:
	_emit_counter_request(window.add_score_amount, window.add_score_select, window.add_score_requested)

func _on_set_lives_button_pressed() -> void:
	_emit_counter_request(window.set_lives_amount, window.set_lives_select, window.set_lives_requested)

func _on_add_lives_button_pressed() -> void:
	_emit_counter_request(window.add_lives_amount, window.add_lives_select, window.add_lives_requested)

func _emit_counter_request(input: LineEdit, select: OptionButton, signal_ref: Signal) -> void:
	var value_text := input.text.strip_edges()
	if !value_text.is_valid_int():
		return
	var target_player_id := _selected_player_id(select)
	if target_player_id == "":
		return
	signal_ref.emit(target_player_id, int(value_text))

func _selected_player_id(select: OptionButton) -> String:
	var selected_index := select.get_selected()
	if selected_index < 0:
		return ""
	return str(select.get_item_metadata(selected_index))

func _refresh_target_option(select: OptionButton, rows: Array) -> void:
	var previous_player_id := _selected_player_id(select)
	select.clear()
	var selected_index := -1
	for row in rows:
		if !(row is Dictionary):
			continue
		var label := str(row.get("label", ""))
		var player_id := str(row.get("player_id", ""))
		select.add_item(label)
		var item_index := select.get_item_count() - 1
		select.set_item_metadata(item_index, player_id)
		if player_id == previous_player_id:
			selected_index = item_index
	if selected_index >= 0:
		select.select(selected_index)
