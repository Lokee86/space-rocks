extends RefCounted
class_name LoadoutDisplayFlow

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

const SLOT_PRIMARY := "primary"
const SLOT_SECONDARY := "secondary"

var hud: Control
var loadout_container: HBoxContainer
var display_nodes := {}
var displayed_weapon_ids := {}
var previous_cooldown_remaining := {}
var ready_effect_played_for_cooldown := {}


func configure(hud_ref: Control) -> void:
	hud = hud_ref
	loadout_container = hud.get_node_or_null("%LoadoutContainer") as HBoxContainer
	display_nodes = {
		SLOT_PRIMARY: null,
		SLOT_SECONDARY: null,
	}
	displayed_weapon_ids = {
		SLOT_PRIMARY: "",
		SLOT_SECONDARY: "",
	}
	previous_cooldown_remaining = {
		SLOT_PRIMARY: 0.0,
		SLOT_SECONDARY: 0.0,
	}
	ready_effect_played_for_cooldown = {
		SLOT_PRIMARY: true,
		SLOT_SECONDARY: true,
	}


func clear() -> void:
	for slot in [SLOT_PRIMARY, SLOT_SECONDARY]:
		_clear_slot(slot)


func apply_player_state(player_state: Dictionary) -> void:
	_apply_slot({
		"slot": SLOT_PRIMARY,
		"weapon_id": str(player_state.get(Packets.FIELD_PRIMARY_WEAPON_ID, "")),
		"ammo_policy": str(player_state.get(Packets.FIELD_PRIMARY_AMMO_POLICY, "")),
		"ammo_remaining": _int_or_default(player_state.get(Packets.FIELD_PRIMARY_AMMO_REMAINING, 0), 0),
		"cooldown_remaining": _float_or_default(player_state.get(Packets.FIELD_PRIMARY_COOLDOWN_REMAINING, 0.0), 0.0),
	})
	_apply_slot({
		"slot": SLOT_SECONDARY,
		"weapon_id": str(player_state.get(Packets.FIELD_SECONDARY_WEAPON_ID, "")),
		"ammo_policy": str(player_state.get(Packets.FIELD_SECONDARY_AMMO_POLICY, "")),
		"ammo_remaining": _int_or_default(player_state.get(Packets.FIELD_SECONDARY_AMMO_REMAINING, 0), 0),
		"cooldown_remaining": _float_or_default(player_state.get(Packets.FIELD_SECONDARY_COOLDOWN_REMAINING, 0.0), 0.0),
	})


func _clear_slot(slot: String) -> void:
	var display_node: WeaponDisplay = display_nodes.get(slot, null)
	if display_node != null and is_instance_valid(display_node):
		display_node.queue_free()
	display_nodes[slot] = null
	displayed_weapon_ids[slot] = ""
	previous_cooldown_remaining[slot] = 0.0
	ready_effect_played_for_cooldown[slot] = true

func _int_or_default(value, default_value: int) -> int:
	if value == null:
		return default_value
	return int(value)


func _float_or_default(value, default_value: float) -> float:
	if value == null:
		return default_value
	return float(value)


func _ensure_display_for_slot(slot: String, weapon_id: String, scene: PackedScene) -> WeaponDisplay:
	var display_node: WeaponDisplay = display_nodes.get(slot, null)
	if displayed_weapon_ids.get(slot, "") == weapon_id and is_instance_valid(display_node):
		_connect_cooldown_finished(display_node)
		return display_node

	_clear_slot(slot)
	if loadout_container == null or scene == null:
		return null

	var new_node := scene.instantiate()
	if not new_node is WeaponDisplay:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Weapon display scene must instantiate its presentation root",
		{},
		{
			"subsystem": "hud",
			"failure_mode": "wrong_scene_root",
			"resource_kind": "scene",
			"expected_type": "WeaponDisplay",
			"actual_type": new_node.get_class(),
			"resource_path": scene.resource_path,
		}
	)
		new_node.queue_free()
		return null

	var new_display := new_node as WeaponDisplay
	loadout_container.add_child(new_display)
	display_nodes[slot] = new_display
	displayed_weapon_ids[slot] = weapon_id
	_connect_cooldown_finished(new_display)
	return new_display


func _apply_display_state(display: WeaponDisplay, slot_state: Dictionary, _cooldown_total: float) -> void:
	var slot := str(slot_state.get("slot", ""))
	var cooldown_total: float = float(_cooldown_total)
	var ammo_remaining: int = int(slot_state.get("ammo_remaining", 0))
	var cooldown_remaining: float = float(slot_state.get("cooldown_remaining", 0.0))
	var previous_remaining: float = float(previous_cooldown_remaining.get(slot, 0.0))
	var display_state := {
		"weapon_id": str(slot_state.get("weapon_id", "")),
		"ammo_policy": str(slot_state.get("ammo_policy", "")),
		"ammo_remaining": ammo_remaining,
		"cooldown_remaining": cooldown_remaining,
		"cooldown_total": cooldown_total,
	}

	if cooldown_remaining > 0.0:
		ready_effect_played_for_cooldown[slot] = false

	display.apply_weapon_display_state(display_state)

	if previous_remaining > 0.0 and cooldown_remaining <= 0.0:
		if not bool(ready_effect_played_for_cooldown.get(slot, true)):
			display.play_ready_effects()
			ready_effect_played_for_cooldown[slot] = true

	previous_cooldown_remaining[slot] = cooldown_remaining


func _apply_slot(slot_state: Dictionary) -> void:
	var slot := str(slot_state.get("slot", ""))
	var weapon_id := str(slot_state.get("weapon_id", ""))
	if not WeaponDisplayRegistry.is_displayable_weapon(weapon_id):
		_clear_slot(slot)
		return

	var definition := WeaponDisplayRegistry.definition_for_weapon(weapon_id)
	var scene := definition.get("scene", null) as PackedScene
	var display := _ensure_display_for_slot(slot, weapon_id, scene)
	if display == null:
		return

	_apply_display_state(display, slot_state, float(definition.get("cooldown_total", 0.0)))


func _connect_cooldown_finished(display: WeaponDisplay) -> void:
	var callback := Callable(self, "_on_display_cooldown_finished").bind(display)
	if not display.cooldown_finished.is_connected(callback):
		display.cooldown_finished.connect(callback)


func _on_display_cooldown_finished(display: WeaponDisplay) -> void:
	if display == null or not is_instance_valid(display):
		return

	var slot := _slot_for_display(display)
	if slot != "":
		ready_effect_played_for_cooldown[slot] = true
	display.play_ready_effects()


func _slot_for_display(display: WeaponDisplay) -> String:
	for slot in [SLOT_PRIMARY, SLOT_SECONDARY]:
		var display_node: WeaponDisplay = display_nodes.get(slot, null)
		if display_node == display:
			return slot
	return ""
