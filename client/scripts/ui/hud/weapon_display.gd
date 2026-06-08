extends Control
class_name WeaponDisplay

const AMMO_POLICY_LIMITED := "limited"


func apply_weapon_display_state(state: Dictionary) -> void:
	var weapon_id := str(state.get("weapon_id", ""))
	var ammo_policy := str(state.get("ammo_policy", ""))
	var ammo_remaining := int(state.get("ammo_remaining", 0))
	var cooldown_remaining := float(state.get("cooldown_remaining", 0.0))
	var cooldown_total := float(state.get("cooldown_total", 0.0))

	apply_weapon_presentation(weapon_id)
	apply_ammo_state(ammo_policy, ammo_remaining)
	apply_cooldown_state(cooldown_remaining, cooldown_total)


func apply_weapon_presentation(weapon_id: String) -> void:
	var weapon_icon := _weapon_icon()
	if weapon_icon == null:
		return

	for child in weapon_icon.get_children():
		if child is CanvasItem:
			(child as CanvasItem).hide()

	var weapon_node := weapon_icon.get_node_or_null(weapon_id)
	if weapon_node is CanvasItem:
		(weapon_node as CanvasItem).show()


func apply_ammo_state(ammo_policy: String, ammo_remaining: int) -> void:
	var ammo_label := get_node_or_null("%AmmoLabel") as Label
	if ammo_label == null:
		return

	if ammo_policy == AMMO_POLICY_LIMITED:
		ammo_label.show()
		ammo_label.text = "x%d" % ammo_remaining
	else:
		ammo_label.hide()


func apply_cooldown_state(cooldown_remaining: float, cooldown_total: float) -> void:
	var ring_highlight := get_node_or_null("%RingHighlight") as CanvasItem
	if ring_highlight != null:
		if cooldown_remaining > 0.0:
			ring_highlight.hide()
		else:
			ring_highlight.show()

	var cooldown_overlay := get_node_or_null("%CooldownOverlay") as Control
	if cooldown_overlay == null:
		return

	if cooldown_overlay.has_method("apply_cooldown"):
		cooldown_overlay.apply_cooldown(cooldown_remaining, cooldown_total)
		return

	if cooldown_remaining <= 0.0:
		if cooldown_overlay.has_method("clear_countdown"):
			cooldown_overlay.clear_countdown()
		return

	if cooldown_overlay.has_method("sync_countdown"):
		cooldown_overlay.sync_countdown(cooldown_remaining)
	elif cooldown_overlay.has_method("start_countdown"):
		cooldown_overlay.start_countdown(cooldown_remaining)


func play_ready_effects() -> void:
	var ring_highlight := get_node_or_null("%RingHighlight") as CanvasItem
	if ring_highlight != null:
		ring_highlight.show()

	var ready_sweep := get_node_or_null("%ReadySweepHighlight")
	if ready_sweep != null and ready_sweep.has_method("play"):
		ready_sweep.play()

	var ready_flash := get_node_or_null("%ReadyFlash") as AnimatedSprite2D
	if ready_flash == null:
		return

	var callback := Callable(self, "_on_ready_flash_animation_finished").bind(ready_flash)
	if not ready_flash.animation_finished.is_connected(callback):
		ready_flash.animation_finished.connect(callback)

	ready_flash.show()
	ready_flash.stop()
	ready_flash.frame = 0
	ready_flash.play()


func _weapon_icon() -> Node:
	var weapon_icon := get_node_or_null("%WeaponIcon")
	if weapon_icon != null:
		return weapon_icon

	return get_node_or_null("Sprite2D/WeaponIcon")


func _on_ready_flash_animation_finished(ready_flash: AnimatedSprite2D) -> void:
	if ready_flash == null or not is_instance_valid(ready_flash):
		return

	ready_flash.stop()
	ready_flash.hide()
