extends GutTest

const LoadoutReadoutScene := preload("res://scenes/ui/transmission_displays/loadout_readout.tscn")

var readout: LoadoutReadout


func before_each() -> void:
	readout = LoadoutReadoutScene.instantiate() as LoadoutReadout
	add_child_autofree(readout)
	await get_tree().process_frame


func test_configure_selects_authoritative_fallback_options() -> void:
	readout.configure({
		"eligible_ships": [{"owned_ship_id": "ship-owned-1", "ship_id": "v_wing"}],
		"eligible_weapons": [{"owned_weapon_id": "weapon-owned-1", "weapon_id": "pulse", "weapon_point": "primary_1"}],
		"eligible_modules": [{"owned_module_id": "module-owned-1", "module_id": "shield_booster", "module_slot": "shield_mod"}],
	}, {
		"selected_owned_ship_id": "ship-owned-1",
		"selected_weapons_by_point": {"primary_1": "weapon-owned-1"},
		"selected_modules_by_slot": {"shield_mod": "module-owned-1"},
		"starting_ammo_by_point": {"primary_1": 24},
		"valid": true,
	})

	assert_eq(readout.ship_option.item_count, 1)
	assert_eq(readout.primary_weapon_option.item_count, 1)
	assert_eq(readout.module_options["shield_mod"].item_count, 1)
	assert_false(readout.apply_button.disabled)
	assert_eq(readout.status_label.text, "LOADOUT STATUS: READY")


func test_apply_emits_selected_owned_instance_ids() -> void:
	var submissions: Array = []
	readout.submit_requested.connect(func(selection: Dictionary) -> void:
		submissions.append(selection)
	)
	readout.configure({
		"eligible_ships": [{"owned_ship_id": "ship-owned-1", "ship_id": "v_wing"}],
		"eligible_weapons": [{"owned_weapon_id": "weapon-owned-1", "weapon_id": "pulse", "weapon_point": "primary_1"}],
		"eligible_modules": [],
	}, {
		"selected_owned_ship_id": "ship-owned-1",
		"selected_weapons_by_point": {"primary_1": "weapon-owned-1"},
		"selected_modules_by_slot": {},
		"starting_ammo_by_point": {"primary_1": 12},
		"valid": true,
	})

	readout._on_apply_pressed()

	assert_eq(submissions.size(), 1)
	assert_eq(submissions[0]["selected_owned_ship_id"], "ship-owned-1")
	assert_eq(submissions[0]["selected_weapons_by_point"]["primary_1"], "weapon-owned-1")
	assert_eq(submissions[0]["starting_ammo_by_point"]["primary_1"], 12)


func test_empty_options_disable_apply_without_errors() -> void:
	readout.configure({}, {"valid": false, "error_code": "loadout_unavailable"})

	assert_true(readout.ship_option.disabled)
	assert_true(readout.primary_weapon_option.disabled)
	assert_true(readout.apply_button.disabled)
	assert_eq(readout.status_label.text, "ERROR: loadout_unavailable")
