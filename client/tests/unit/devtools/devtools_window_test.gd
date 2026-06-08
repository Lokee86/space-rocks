extends GutTest

const DevtoolsWindowScene := preload("res://scenes/devtools/devtools_window.tscn")


func test_freeze_buttons_emit_expected_freeze_targets() -> void:
	var window := DevtoolsWindowScene.instantiate()
	add_child_autofree(window)

	var emitted_targets: Array = []
	window.toggle_freeze_world_requested.connect(func(freeze_target: String) -> void:
		emitted_targets.append(freeze_target)
	)

	var freeze_asteroids_button := window.find_child("FreezeAsteroidsButton", true, false) as Button
	var freeze_bullets_button := window.find_child("FreezeBulletsButton", true, false) as Button
	var freeze_spawns_button := window.find_child("FreezeSpawnsButton", true, false) as Button
	var freeze_collisions_button := window.find_child("FreezeCollisionsButton", true, false) as Button
	var freeze_world_button := window.find_child("FreezeWorldButton", true, false) as Button

	assert_not_null(freeze_asteroids_button)
	assert_not_null(freeze_bullets_button)
	assert_not_null(freeze_spawns_button)
	assert_not_null(freeze_collisions_button)
	assert_not_null(freeze_world_button)

	freeze_asteroids_button.pressed.emit()
	freeze_bullets_button.pressed.emit()
	freeze_spawns_button.pressed.emit()
	freeze_collisions_button.pressed.emit()
	freeze_world_button.pressed.emit()

	assert_eq(emitted_targets.size(), 5)
	assert_eq(emitted_targets[0], "asteroids")
	assert_eq(emitted_targets[1], "bullets")
	assert_eq(emitted_targets[2], "spawns")
	assert_eq(emitted_targets[3], "collisions")
	assert_eq(emitted_targets[4], "")


func test_set_debug_status_updates_granular_freeze_labels() -> void:
	var window := DevtoolsWindowScene.instantiate()
	add_child_autofree(window)

	window.set_debug_status({
		"world_frozen": false,
		"asteroids_frozen": true,
		"bullets_frozen": true,
		"spawning_frozen": true,
		"collisions_frozen": false,
	})

	var world_status_label := window.find_child("WorldFrozenStatusLabel", true, false) as Label
	var asteroids_status_label := window.find_child("FreezeAsteroidsStatusLabel", true, false) as Label
	var bullets_status_label := window.find_child("FreezeBulletsStatusLabel", true, false) as Label
	var spawns_status_label := window.find_child("FreezeSpawnsStatusLabel", true, false) as Label
	var collisions_status_label := window.find_child("FreezeCollisionsStatusLabel", true, false) as Label

	assert_not_null(world_status_label)
	assert_not_null(asteroids_status_label)
	assert_not_null(bullets_status_label)
	assert_not_null(spawns_status_label)
	assert_not_null(collisions_status_label)

	assert_eq(asteroids_status_label.text, "Active")
	assert_eq(bullets_status_label.text, "Active")
	assert_eq(spawns_status_label.text, "Active")
	assert_eq(collisions_status_label.text, "Inactive")
	assert_eq(world_status_label.text, "Inactive")


func test_pickup_select_includes_catalog_types_and_defaults_to_1_up() -> void:
	var window := DevtoolsWindowScene.instantiate()
	add_child_autofree(window)

	var pickup_select := window.find_child("PickupSelect", true, false) as OptionButton

	assert_not_null(pickup_select)
	assert_gt(pickup_select.get_item_count(), 0)

	var pickup_types: Array[String] = []
	for index in range(pickup_select.get_item_count()):
		pickup_types.append(str(pickup_select.get_item_metadata(index)))

	assert_true(pickup_types.has("1_up"))
	assert_true(pickup_types.has("torpedo"))

	var selected_index := pickup_select.get_selected()
	assert_gte(selected_index, 0)
	assert_eq(str(pickup_select.get_item_metadata(selected_index)), "1_up")
