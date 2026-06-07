extends RefCounted
class_name LoadoutDisplayFlow

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const WeaponDisplayRegistry = preload("res://scripts/ui/hud/weapon_display_registry.gd")

const SLOT_PRIMARY := "primary"
const SLOT_SECONDARY := "secondary"
const AMMO_POLICY_LIMITED := "limited"

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
		"ammo_remaining": int(player_state.get(Packets.FIELD_PRIMARY_AMMO_REMAINING, 0)),
		"cooldown_remaining": float(player_state.get(Packets.FIELD_PRIMARY_COOLDOWN_REMAINING, 0.0)),
	})
	_apply_slot({
		"slot": SLOT_SECONDARY,
		"weapon_id": str(player_state.get(Packets.FIELD_SECONDARY_WEAPON_ID, "")),
		"ammo_policy": str(player_state.get(Packets.FIELD_SECONDARY_AMMO_POLICY, "")),
		"ammo_remaining": int(player_state.get(Packets.FIELD_SECONDARY_AMMO_REMAINING, 0)),
		"cooldown_remaining": float(player_state.get(Packets.FIELD_SECONDARY_COOLDOWN_REMAINING, 0.0)),
	})


func _clear_slot(slot: String) -> void:
	var display_node: Node = display_nodes.get(slot, null)
	if display_node != null and is_instance_valid(display_node):
		display_node.queue_free()
	display_nodes[slot] = null
	displayed_weapon_ids[slot] = ""
	previous_cooldown_remaining[slot] = 0.0
	ready_effect_played_for_cooldown[slot] = true


func _ensure_display_for_slot(slot: String, weapon_id: String, scene: PackedScene) -> Node:
	var display_node: Node = display_nodes.get(slot, null)
	if displayed_weapon_ids.get(slot, "") == weapon_id and is_instance_valid(display_node):
		_connect_cooldown_finished(display_node)
		return display_node

	_clear_slot(slot)
	if loadout_container == null or scene == null:
		return null

	var new_node := scene.instantiate()
	loadout_container.add_child(new_node)
	display_nodes[slot] = new_node
	displayed_weapon_ids[slot] = weapon_id
	_connect_cooldown_finished(new_node)
	return new_node


func _apply_display_state(display: Node, slot_state: Dictionary, cooldown_total: float) -> void:
	var slot := str(slot_state.get("slot", ""))
	var ammo_remaining: int = int(slot_state.get("ammo_remaining", 0))
	var cooldown_remaining: float = float(slot_state.get("cooldown_remaining", 0.0))
	var previous_remaining: float = float(previous_cooldown_remaining.get(slot, 0.0))

	var ammo_label := display.get_node_or_null("%AmmoLabel") as Label
	if ammo_label != null:
		if str(slot_state.get("ammo_policy", "")) == AMMO_POLICY_LIMITED:
			ammo_label.visible = true
			ammo_label.text = "x%d" % ammo_remaining
		else:
			ammo_label.visible = false

	var ring_highlight := display.get_node_or_null("%RingHighlight") as CanvasItem
	if ring_highlight != null:
		if cooldown_remaining > 0.0:
			ring_highlight.hide()
		else:
			ring_highlight.show()

	var cooldown_overlay := display.get_node_or_null("%CooldownOverlay") as Control
	if cooldown_overlay != null:
		if cooldown_remaining <= 0.0:
			if cooldown_overlay.has_method("clear_countdown"):
				cooldown_overlay.clear_countdown()
		else:
			if previous_remaining <= 0.0:
				ready_effect_played_for_cooldown[slot] = false
			if cooldown_overlay.has_method("sync_countdown"):
				cooldown_overlay.sync_countdown(cooldown_remaining)
			elif cooldown_overlay.has_method("start_countdown"):
				cooldown_overlay.start_countdown(cooldown_remaining)

	if previous_remaining > 0.0 and cooldown_remaining <= 0.0:
		if not bool(ready_effect_played_for_cooldown.get(slot, true)):
			_play_ready_effects_for_display(display)
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


func _connect_cooldown_finished(display: Node) -> void:
	var cooldown_overlay := display.get_node_or_null("%CooldownOverlay")
	if cooldown_overlay == null or not cooldown_overlay.has_signal("cooldown_finished"):
		return

	var callback := Callable(self, "_on_display_cooldown_finished").bind(display)
	if not cooldown_overlay.cooldown_finished.is_connected(callback):
		cooldown_overlay.cooldown_finished.connect(callback)


func _on_display_cooldown_finished(display: Node) -> void:
	if display == null or not is_instance_valid(display):
		return

	var slot := _slot_for_display(display)
	if slot != "":
		ready_effect_played_for_cooldown[slot] = true
	_play_ready_effects_for_display(display)


func _play_ready_effects_for_display(display: Node) -> void:
	if display == null or not is_instance_valid(display):
		return

	var ring_highlight := display.get_node_or_null("%RingHighlight") as CanvasItem
	if ring_highlight != null:
		ring_highlight.show()
	_play_ready_sweep(display)
	_play_ready_flash(display)


func _slot_for_display(display: Node) -> String:
	for slot in [SLOT_PRIMARY, SLOT_SECONDARY]:
		var display_node: Node = display_nodes.get(slot, null)
		if display_node == display:
			return slot
	return ""


func _play_ready_sweep(display: Node) -> void:
	if display == null or not is_instance_valid(display):
		return

	var ready_sweep := display.get_node_or_null("%ReadySweepHighlight")
	if ready_sweep == null:
		return

	if not ready_sweep.has_method("play"):
		return

	ready_sweep.play()


func _play_ready_flash(display: Node) -> void:
	if display == null or not is_instance_valid(display):
		return

	var ready_flash := display.get_node_or_null("%ReadyFlash") as AnimatedSprite2D
	if ready_flash == null:
		return

	var callback := Callable(self, "_on_ready_flash_animation_finished").bind(ready_flash)
	if not ready_flash.animation_finished.is_connected(callback):
		ready_flash.animation_finished.connect(callback)

	ready_flash.show()
	ready_flash.stop()
	ready_flash.frame = 0
	ready_flash.play()


func _on_ready_flash_animation_finished(ready_flash: AnimatedSprite2D) -> void:
	if ready_flash == null or not is_instance_valid(ready_flash):
		return

	ready_flash.stop()
	ready_flash.hide()
