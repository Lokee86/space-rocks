extends Control
class_name WeaponDisplay

const AMMO_POLICY_LIMITED := "limited"

signal cooldown_finished

@onready var ammo_label: Label = %AmmoLabel
@onready var ring_highlight: RingHighlight = %RingHighlight
@onready var ready_sweep_highlight: ReadySweepHighlight = %ReadySweepHighlight
@onready var ready_flash: AnimatedSprite2D = %ReadyFlash
@onready var cooldown_overlay: CooldownOverlay = %CooldownOverlay
@onready var weapon_icon: Node2D = $Sprite2D/WeaponIcon


func _ready() -> void:
	cooldown_overlay.cooldown_finished.connect(_on_cooldown_finished)
	ready_flash.animation_finished.connect(_on_ready_flash_animation_finished)


func _on_cooldown_finished() -> void:
	cooldown_finished.emit()


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
	for child in weapon_icon.get_children():
		if child is CanvasItem:
			(child as CanvasItem).hide()

	var weapon_node := weapon_icon.get_node(weapon_id)
	if weapon_node is CanvasItem:
		(weapon_node as CanvasItem).show()


func apply_ammo_state(ammo_policy: String, ammo_remaining: int) -> void:
	if ammo_policy == AMMO_POLICY_LIMITED:
		ammo_label.show()
		ammo_label.text = "x%d" % ammo_remaining
	else:
		ammo_label.hide()


func apply_cooldown_state(cooldown_remaining: float, cooldown_total: float) -> void:
	if cooldown_remaining > 0.0:
		ring_highlight.hide()
	else:
		ring_highlight.show()

	cooldown_overlay.apply_cooldown(cooldown_remaining, cooldown_total)


func play_ready_effects() -> void:
	ring_highlight.show()
	ready_sweep_highlight.play()
	ready_flash.show()
	ready_flash.stop()
	ready_flash.frame = 0
	ready_flash.play()

func _on_ready_flash_animation_finished() -> void:
	ready_flash.stop()
	ready_flash.hide()
