extends GutTest

const HudScene = preload("res://scenes/ui/hud.tscn")
const LoadoutDisplayFlow = preload("res://scripts/ui/hud/loadout_display_flow.gd")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")

var _hud: Control
var _flow := LoadoutDisplayFlow.new()


func before_each() -> void:
	_hud = HudScene.instantiate()
	add_child_autofree(_hud)
	_flow = LoadoutDisplayFlow.new()
	_flow.configure(_hud)


func test_basic_cannon_primary_creates_no_loadout_display_child() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_PRIMARY_WEAPON_ID: "basic_cannon",
	}))

	assert_eq(_loadout_container().get_child_count(), 0)


func test_torpedo_secondary_creates_one_loadout_display_child() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
	}))

	assert_eq(_loadout_container().get_child_count(), 1)


func test_torpedo_display_contains_ready_sweep_highlight() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
	}))

	var ready_sweep := _display_ready_sweep_highlight()
	assert_not_null(ready_sweep)
	assert_true(ready_sweep.has_method("play"))
	assert_false(ready_sweep.visible)


func test_torpedo_display_contains_ready_flash() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
	}))

	var ready_flash := _display_ready_flash()
	assert_not_null(ready_flash)


func test_torpedo_display_uses_generic_weapon_display_scene() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
	}))

	var display := _loadout_container().get_child(0)

	assert_eq(display.scene_file_path, "res://scenes/ui/weapon_displays/weapon_display.tscn")


func test_torpedo_with_infinite_ammo_policy_hides_ammo_label() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_AMMO_POLICY: "infinite",
		Packets.FIELD_SECONDARY_AMMO_REMAINING: 7,
	}))

	var ammo_label := _display_ammo_label()
	assert_not_null(ammo_label)
	assert_false(ammo_label.visible)


func test_torpedo_with_limited_ammo_policy_shows_ammo_label() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_AMMO_POLICY: "limited",
		Packets.FIELD_SECONDARY_AMMO_REMAINING: 7,
	}))

	var ammo_label := _display_ammo_label()
	assert_not_null(ammo_label)
	assert_true(ammo_label.visible)


func test_limited_ammo_updates_ammo_label_text() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_AMMO_POLICY: "limited",
		Packets.FIELD_SECONDARY_AMMO_REMAINING: 3,
	}))

	var ammo_label := _display_ammo_label()
	assert_not_null(ammo_label)
	assert_eq(ammo_label.text, "x3")


func test_torpedo_with_cooldown_remaining_hides_ring_highlight() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 2.0,
	}))

	var ring_highlight := _display_ring_highlight()
	assert_not_null(ring_highlight)
	assert_false(ring_highlight.visible)


func test_torpedo_with_cooldown_remaining_on_first_display_creation_shows_cooldown_overlay() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 5.0,
	}))

	var cooldown_overlay := _display_cooldown_overlay()
	assert_not_null(cooldown_overlay)
	assert_true(cooldown_overlay.visible)
	assert_eq(_display_cooldown_label().text, "5.0")


func test_active_cooldown_state_keeps_cooldown_overlay_visible() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 5.0,
	}))
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 4.0,
	}))

	var cooldown_overlay := _display_cooldown_overlay()
	assert_not_null(cooldown_overlay)
	assert_true(cooldown_overlay.visible)
	assert_eq(_display_cooldown_label().text, "4.0")


func test_larger_active_cooldown_remaining_syncs_overlay_instead_of_leaving_it_stale() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 3.0,
	}))
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 5.0,
	}))

	assert_eq(_display_cooldown_label().text, "5.0")
	assert_true(_display_cooldown_overlay().visible)


func test_torpedo_with_no_cooldown_shows_ring_highlight() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 0.0,
	}))

	var ring_highlight := _display_ring_highlight()
	assert_not_null(ring_highlight)
	assert_true(ring_highlight.visible)


func test_cooldown_finished_signal_makes_ring_highlight_visible() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 2.0,
	}))

	var ring_highlight := _display_ring_highlight()
	assert_not_null(ring_highlight)
	assert_false(ring_highlight.visible)

	_display_cooldown_overlay().cooldown_finished.emit()

	assert_true(ring_highlight.visible)


func test_first_cooldown_ready_transition_triggers_ready_effects() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 2.0,
	}))

	var ready_flash := _display_ready_flash()
	var ready_sweep := _display_ready_sweep_highlight()
	assert_not_null(ready_flash)
	assert_not_null(ready_sweep)
	assert_false(ready_flash.visible)
	assert_false(ready_sweep.visible)

	_display_cooldown_overlay().clear_countdown()
	_display_cooldown_overlay().cooldown_finished.emit()

	assert_true(ready_flash.visible or ready_flash.is_playing())
	assert_true(ready_sweep.visible)


func test_cooldown_ready_transition_clears_overlay_and_only_plays_ready_effects_once() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 2.0,
	}))
	var cooldown_overlay := _display_cooldown_overlay()
	cooldown_overlay.clear_countdown()
	cooldown_overlay.cooldown_finished.emit()

	var ready_sweep := _display_ready_sweep_highlight()
	var ready_flash := _display_ready_flash()
	assert_false(_display_cooldown_overlay().visible)
	assert_true(ready_sweep.visible)
	assert_true(ready_flash.visible or ready_flash.is_playing())

	ready_sweep.hide()
	ready_flash.hide()
	ready_flash.stop()

	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 0.0,
	}))

	assert_false(ready_sweep.visible)
	assert_false(ready_flash.visible)


func test_play_ready_sweep_does_not_error_when_ready_sweep_exists() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
	}))

	_flow._play_ready_sweep(_loadout_container().get_child(0))

	assert_not_null(_display_ready_sweep_highlight())


func test_play_ready_flash_does_not_error_when_ready_sweep_exists() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
	}))

	_flow._play_ready_flash(_loadout_container().get_child(0))

	assert_not_null(_display_ready_flash())


func test_play_ready_flash_returns_gracefully_when_ready_sweep_is_missing() -> void:
	var fake_display := Control.new()
	add_child_autofree(fake_display)

	_flow._play_ready_flash(fake_display)

	assert_true(true)


func test_switching_from_torpedo_to_unknown_weapon_clears_the_display() -> void:
	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "torpedo",
	}))
	assert_eq(_loadout_container().get_child_count(), 1)

	_flow.apply_player_state(_player_state({
		Packets.FIELD_SECONDARY_WEAPON_ID: "unknown_weapon",
	}))
	await get_tree().process_frame

	assert_eq(_loadout_container().get_child_count(), 0)


func _loadout_container() -> HBoxContainer:
	return _hud.get_node("%LoadoutContainer") as HBoxContainer


func _display_ammo_label() -> Label:
	var display := _loadout_container().get_child(0)
	return display.get_node("%AmmoLabel") as Label


func _display_ring_highlight() -> CanvasItem:
	var display := _loadout_container().get_child(0)
	return display.get_node("%RingHighlight") as CanvasItem


func _display_cooldown_overlay() -> Control:
	var display := _loadout_container().get_child(0)
	return display.get_node("%CooldownOverlay") as Control


func _display_cooldown_label() -> Label:
	return _display_cooldown_overlay().get_node("CooldownLabel") as Label


func _display_ready_sweep_highlight() -> CanvasItem:
	var display := _loadout_container().get_child(0)
	return display.get_node("%ReadySweepHighlight") as CanvasItem


func _display_ready_flash() -> AnimatedSprite2D:
	var display := _loadout_container().get_child(0)
	return display.get_node("%ReadyFlash") as AnimatedSprite2D


func _player_state(fields: Dictionary) -> Dictionary:
	var state := {
		Packets.FIELD_PRIMARY_WEAPON_ID: "",
		Packets.FIELD_PRIMARY_AMMO_POLICY: "",
		Packets.FIELD_PRIMARY_AMMO_REMAINING: 0,
		Packets.FIELD_PRIMARY_COOLDOWN_REMAINING: 0.0,
		Packets.FIELD_SECONDARY_WEAPON_ID: "",
		Packets.FIELD_SECONDARY_AMMO_POLICY: "",
		Packets.FIELD_SECONDARY_AMMO_REMAINING: 0,
		Packets.FIELD_SECONDARY_COOLDOWN_REMAINING: 0.0,
	}

	for key in fields.keys():
		state[key] = fields[key]

	return state
