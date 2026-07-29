extends GutTest

const LoadoutReadoutScene := preload("res://scenes/ui/transmission_displays/loadout_readout.tscn")

var readout: LoadoutReadout


func before_each() -> void:
	readout = LoadoutReadoutScene.instantiate() as LoadoutReadout
	add_child_autofree(readout)
	await get_tree().process_frame


func test_configure_selects_authoritative_loadout_options() -> void:
	readout.configure({
		"eligible_ships": [
			{"owned_ship_id": "ship-owned-1", "ship_id": "v_wing"},
			{"owned_ship_id": "ship-owned-2", "ship_id": "v_wing_scout"},
		],
		"eligible_weapons": [
			{"owned_weapon_id": "weapon-owned-1", "weapon_id": "pulse", "weapon_point": "primary_1"},
			{"owned_weapon_id": "weapon-owned-2", "weapon_id": "torpedo", "weapon_point": "secondary_1"},
		],
		"eligible_modules": [{"owned_module_id": "module-owned-1", "module_id": "shield_capacitor", "module_slot": "shield_mod"}],
	}, {
		"selected_owned_ship_id": "ship-owned-2",
		"selected_weapons_by_point": {"primary_1": "weapon-owned-1", "secondary_1": "weapon-owned-2"},
		"selected_modules_by_slot": {"shield_mod": "module-owned-1"},
		"starting_ammo_by_point": {"secondary_1": 3},
		"valid": true,
	})

	assert_eq(readout.ship_option.item_count, 2)
	assert_eq(readout.ship_option.get_item_metadata(readout.ship_option.selected), "ship-owned-2")
	assert_eq(readout.primary_weapon_option.item_count, 1)
	assert_eq(readout.secondary_weapon_option.item_count, 2)
	assert_eq(readout.secondary_weapon_option.get_item_metadata(readout.secondary_weapon_option.selected), "weapon-owned-2")
	assert_eq(readout.module_options["shield_mod"].item_count, 2)
	assert_eq(readout.module_options["shield_mod"].get_item_metadata(readout.module_options["shield_mod"].selected), "module-owned-1")
	assert_false(readout.apply_button.disabled)
	assert_eq(readout.status_label.text, "LOADOUT STATUS: READY")


func test_apply_emits_selected_optional_equipment() -> void:
	var submissions: Array = []
	readout.submit_requested.connect(func(selection: Dictionary) -> void:
		submissions.append(selection)
	)
	readout.configure({
		"eligible_ships": [{"owned_ship_id": "ship-owned-1", "ship_id": "v_wing"}],
		"eligible_weapons": [
			{"owned_weapon_id": "weapon-owned-1", "weapon_id": "pulse", "weapon_point": "primary_1"},
			{"owned_weapon_id": "weapon-owned-2", "weapon_id": "torpedo", "weapon_point": "secondary_1"},
		],
		"eligible_modules": [{"owned_module_id": "module-owned-1", "module_id": "engine_overdrive", "module_slot": "engine_mod"}],
	}, {
		"selected_owned_ship_id": "ship-owned-1",
		"selected_weapons_by_point": {"primary_1": "weapon-owned-1", "secondary_1": "weapon-owned-2"},
		"selected_modules_by_slot": {"engine_mod": "module-owned-1"},
		"starting_ammo_by_point": {"secondary_1": 3},
		"valid": true,
	})

	readout._on_apply_pressed()

	assert_eq(submissions.size(), 1)
	assert_eq(submissions[0]["selected_owned_ship_id"], "ship-owned-1")
	assert_eq(submissions[0]["selected_weapons_by_point"]["primary_1"], "weapon-owned-1")
	assert_eq(submissions[0]["selected_weapons_by_point"]["secondary_1"], "weapon-owned-2")
	assert_eq(submissions[0]["selected_modules_by_slot"]["engine_mod"], "module-owned-1")
	assert_eq(submissions[0]["starting_ammo_by_point"]["secondary_1"], 3)


func test_optional_selectors_can_submit_none() -> void:
	var submissions: Array = []
	readout.submit_requested.connect(func(selection: Dictionary) -> void:
		submissions.append(selection)
	)
	readout.configure({
		"eligible_ships": [{"owned_ship_id": "ship-owned-1", "ship_id": "v_wing"}],
		"eligible_weapons": [
			{"owned_weapon_id": "weapon-owned-1", "weapon_id": "pulse", "weapon_point": "primary_1"},
			{"owned_weapon_id": "weapon-owned-2", "weapon_id": "torpedo", "weapon_point": "secondary_1"},
		],
		"eligible_modules": [{"owned_module_id": "module-owned-1", "module_id": "shield_capacitor", "module_slot": "shield_mod"}],
	}, {
		"selected_owned_ship_id": "ship-owned-1",
		"selected_weapons_by_point": {"primary_1": "weapon-owned-1"},
		"selected_modules_by_slot": {},
		"valid": true,
	})

	readout._on_apply_pressed()

	assert_eq(submissions.size(), 1)
	assert_false(submissions[0]["selected_weapons_by_point"].has("secondary_1"))
	assert_true(submissions[0]["selected_modules_by_slot"].is_empty())


func test_empty_required_options_disable_apply_without_errors() -> void:
	readout.configure({}, {"valid": false, "error_code": "loadout_unavailable"})

	assert_true(readout.ship_option.disabled)
	assert_true(readout.primary_weapon_option.disabled)
	assert_false(readout.secondary_weapon_option.disabled)
	assert_true(readout.apply_button.disabled)
	assert_eq(readout.status_label.text, "ERROR: loadout_unavailable")
