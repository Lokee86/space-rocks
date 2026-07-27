extends Control
class_name CooldownOverlay

signal cooldown_finished

const PACKET_RESTART_EPSILON := 0.001
const CORRECTION_WINDOW_SECONDS := 0.25

@export var overlay_size: Vector2 = Vector2(64.0, 64.0)
@export var radius_ratio: float = 0.43
@export var overlay_color: Color = Color(0.0, 0.0, 0.0, 0.55)
@export var wedge_segments: int = 64

@onready var cooldown_label: Label = $CooldownLabel

var _cooldown_total: float = 0.0
var _cooldown_remaining: float = 0.0
var _authoritative_remaining: float = 0.0
var _last_packet_remaining: float = 0.0
var _has_packet_sample := false


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE

	visible = false


func _process(delta: float) -> void:
	if _cooldown_remaining <= 0.0:
		return

	var previous_remaining: float = _cooldown_remaining
	var predicted_remaining: float = maxf(_cooldown_remaining - delta, 0.0)
	_authoritative_remaining = maxf(_authoritative_remaining - delta, 0.0)

	if _has_packet_sample:
		var correction_weight: float = 1.0 - exp(-delta / CORRECTION_WINDOW_SECONDS)
		var corrected_remaining: float = lerp(
			predicted_remaining,
			_authoritative_remaining,
			correction_weight
		)
		_cooldown_remaining = clampf(corrected_remaining, 0.0, previous_remaining)
	else:
		_cooldown_remaining = predicted_remaining

	if _cooldown_remaining <= 0.0:
		_hide_countdown(false)
		cooldown_finished.emit()
		return

	_update_label()
	queue_redraw()


func _draw() -> void:
	if _cooldown_remaining <= 0.0 or _cooldown_total <= 0.0:
		return

	var ratio := _cooldown_remaining / _cooldown_total
	var center := size * 0.5
	var radius: float = min(size.x, size.y) * radius_ratio

	_draw_cooldown_wedge(center, radius, ratio)


func start_countdown(seconds: float) -> void:
	_cooldown_total = max(seconds, 0.01)
	_cooldown_remaining = _cooldown_total
	_authoritative_remaining = _cooldown_remaining
	_last_packet_remaining = 0.0
	_has_packet_sample = false
	visible = true
	_update_label()
	queue_redraw()


func apply_cooldown(remaining: float, total: float) -> void:
	if remaining <= 0.0 or total <= 0.0:
		_last_packet_remaining = 0.0
		_has_packet_sample = true
		_hide_countdown(false)
		return

	var normalized_remaining: float = maxf(remaining, 0.0)
	var starts_new_countdown: bool = (
		not _has_packet_sample
		or normalized_remaining > _last_packet_remaining + PACKET_RESTART_EPSILON
	)

	_cooldown_total = max(total, normalized_remaining, 0.01)
	_authoritative_remaining = normalized_remaining
	_last_packet_remaining = normalized_remaining
	_has_packet_sample = true

	if starts_new_countdown:
		_cooldown_remaining = normalized_remaining

	if _cooldown_remaining <= 0.0:
		return

	visible = true
	_update_label()
	queue_redraw()


func sync_countdown(remaining: float) -> void:
	if remaining <= 0.0:
		clear_countdown()
		return

	_cooldown_total = max(_cooldown_total, remaining, 0.01)
	_cooldown_remaining = remaining
	_authoritative_remaining = remaining
	_last_packet_remaining = 0.0
	_has_packet_sample = false
	visible = true
	_update_label()
	queue_redraw()


func clear_countdown() -> void:
	_hide_countdown(true)


func _hide_countdown(reset_packet_tracking: bool) -> void:
	_cooldown_total = 0.0
	_cooldown_remaining = 0.0
	_authoritative_remaining = 0.0
	visible = false
	cooldown_label.text = ""
	if reset_packet_tracking:
		_last_packet_remaining = 0.0
		_has_packet_sample = false
	queue_redraw()


func _update_label() -> void:
	cooldown_label.text = "%.1f" % _cooldown_remaining


func _draw_cooldown_wedge(center: Vector2, radius: float, ratio: float) -> void:
	var points := PackedVector2Array()
	points.append(center)

	ratio = clamp(ratio, 0.0, 1.0)
	var start_angle := -PI * 0.5
	var sweep := -TAU * ratio
	var steps: int = max(3, int(wedge_segments * ratio))

	for i in range(steps + 1):
		var t := float(i) / float(steps)
		var angle := start_angle + sweep * t
		points.append(center + Vector2(cos(angle), sin(angle)) * radius)

	draw_colored_polygon(points, overlay_color)
