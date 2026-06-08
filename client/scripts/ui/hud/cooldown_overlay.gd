extends Control

signal cooldown_finished

@export var overlay_size: Vector2 = Vector2(64.0, 64.0)
@export var radius_ratio: float = 0.43
@export var overlay_color: Color = Color(0.0, 0.0, 0.0, 0.55)
@export var wedge_segments: int = 64

@onready var cooldown_label: Label = $CooldownLabel

var _cooldown_total: float = 0.0
var _cooldown_remaining: float = 0.0


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE

	visible = false


func _process(delta: float) -> void:
	if _cooldown_remaining <= 0.0:
		return

	_cooldown_remaining = max(_cooldown_remaining - delta, 0.0)

	if _cooldown_remaining <= 0.0:
		clear_countdown()
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
	visible = true
	_update_label()
	queue_redraw()


func apply_cooldown(remaining: float, total: float) -> void:
	if remaining <= 0.0 or total <= 0.0:
		clear_countdown()
		return

	_cooldown_total = max(total, remaining, 0.01)
	_cooldown_remaining = max(remaining, 0.0)
	visible = true
	_update_label()
	queue_redraw()


func sync_countdown(remaining: float) -> void:
	if remaining <= 0.0:
		clear_countdown()
		return

	_cooldown_total = max(_cooldown_total, remaining, 0.01)
	_cooldown_remaining = remaining
	visible = true
	_update_label()
	queue_redraw()


func clear_countdown() -> void:
	_cooldown_total = 0.0
	_cooldown_remaining = 0.0
	visible = false
	cooldown_label.text = ""
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
