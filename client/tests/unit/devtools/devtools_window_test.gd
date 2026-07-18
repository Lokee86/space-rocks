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


func test_measurement_buttons_route_scenario_start_stop_and_reset_requests() -> void:
	var window := DevtoolsWindowScene.instantiate()
	add_child_autofree(window)
	var started_labels: Array = []
	var stop_count := [0]
	var reset_count := [0]
	window.measurement_start_requested.connect(func(scenario_label: String) -> void:
		started_labels.append(scenario_label)
	)
	window.measurement_stop_requested.connect(func() -> void:
		stop_count[0] += 1
	)
	window.measurement_reset_requested.connect(func() -> void:
		reset_count[0] += 1
	)

	var scenario_input := window.find_child("MeasurementScenarioLabel", true, false) as LineEdit
	var start_button := window.find_child("MeasurementStartButton", true, false) as Button
	var stop_button := window.find_child("MeasurementStopButton", true, false) as Button
	var reset_button := window.find_child("MeasurementResetButton", true, false) as Button
	assert_not_null(scenario_input)
	assert_not_null(start_button)
	assert_not_null(stop_button)
	assert_not_null(reset_button)

	scenario_input.text = "  soak  "
	start_button.pressed.emit()
	window.refresh_measurement_state({"recording": true, "active_run_id": "run-1", "pending_request_ids": {}})
	stop_button.pressed.emit()
	reset_button.pressed.emit()

	assert_eq(started_labels, ["soak"])
	assert_eq(stop_count[0], 1)
	assert_eq(reset_count[0], 1)


func test_measurement_state_refresh_updates_status_run_id_and_export_fields() -> void:
	var window := DevtoolsWindowScene.instantiate()
	add_child_autofree(window)

	window.refresh_measurement_state({
		"recording": true,
		"active_run_id": "run-7",
		"pending_request_ids": {},
		"latest_export_result": {"success": true, "path": "user://measurements/run-7.json"},
		"last_tooling_error": {},
	})

	var status_label := window.find_child("MeasurementStatusLabel", true, false) as Label
	var run_id_label := window.find_child("MeasurementActiveRunIdLabel", true, false) as Label
	var export_label := window.find_child("MeasurementExportLabel", true, false) as Label
	var start_button := window.find_child("MeasurementStartButton", true, false) as Button
	var stop_button := window.find_child("MeasurementStopButton", true, false) as Button
	var reset_button := window.find_child("MeasurementResetButton", true, false) as Button
	assert_eq(status_label.text, "Status: Recording")
	assert_eq(run_id_label.text, "Active run: run-7")
	assert_eq(export_label.text, "Latest export: user://measurements/run-7.json")
	assert_true(start_button.disabled)
	assert_false(stop_button.disabled)

	window.refresh_measurement_state({
		"recording": false,
		"active_run_id": "",
		"pending_request_ids": {"start": "request-1"},
		"latest_export_result": {"success": false, "error": "disk full"},
		"last_tooling_error": {},
	})
	assert_eq(status_label.text, "Status: Starting")
	assert_eq(run_id_label.text, "Active run: —")
	assert_eq(export_label.text, "Latest export error: disk full")
	assert_true(reset_button.disabled)
