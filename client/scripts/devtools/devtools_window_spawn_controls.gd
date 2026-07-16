extends RefCounted
class_name DevtoolsWindowSpawnControls

var window

func configure(window_ref) -> void:
	window = window_ref

func initialize() -> void:
	if !window.spawn_asteroid_button.pressed.is_connected(_on_spawn_asteroid_button_pressed): window.spawn_asteroid_button.pressed.connect(_on_spawn_asteroid_button_pressed)
	if !window.spawn_pickup_button.pressed.is_connected(_on_spawn_pickup_button_pressed): window.spawn_pickup_button.pressed.connect(_on_spawn_pickup_button_pressed)
	if !window.spawn_player_button.pressed.is_connected(_on_spawn_player_button_pressed): window.spawn_player_button.pressed.connect(_on_spawn_player_button_pressed)
	if !window.spawn_bullet_button.pressed.is_connected(_on_spawn_bullet_button_pressed): window.spawn_bullet_button.pressed.connect(_on_spawn_bullet_button_pressed)
	if !window.respawn_player_button.pressed.is_connected(_on_respawn_player_button_pressed): window.respawn_player_button.pressed.connect(_on_respawn_player_button_pressed)
	_initialize_pickup_select()

func _on_spawn_asteroid_button_pressed() -> void: window.spawn_asteroid_placement_requested.emit()
func _on_spawn_pickup_button_pressed() -> void: window.spawn_pickup_placement_requested.emit(_selected_metadata_as_string(window.pickup_select))
func _on_spawn_player_button_pressed() -> void: window.spawn_player_placement_requested.emit(_selected_metadata_as_string(window.spawn_player_select))
func _on_spawn_bullet_button_pressed() -> void: window.spawn_bullet_placement_requested.emit()
func _on_respawn_player_button_pressed() -> void:
	var target_player_id := _selected_metadata_as_string(window.respawn_player_select)
	window.respawn_player_placement_requested.emit(target_player_id)

func _initialize_pickup_select() -> void:
	window.pickup_select.clear()
	var pickup_types := PickupPresentationCatalog.available_pickup_types()
	var selected_index := -1
	for pickup_type in pickup_types:
		window.pickup_select.add_item(pickup_type)
		var item_index: int = window.pickup_select.get_item_count() - 1
		window.pickup_select.set_item_metadata(item_index, pickup_type)
		if pickup_type == "1_up": selected_index = item_index
	if window.pickup_select.get_item_count() == 0: return
	window.pickup_select.select(0 if selected_index < 0 else selected_index)

func _selected_metadata_as_string(select: OptionButton) -> String:
	var index := select.get_selected()
	return "" if index < 0 else str(select.get_item_metadata(index))
