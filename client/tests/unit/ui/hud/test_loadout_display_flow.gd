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
