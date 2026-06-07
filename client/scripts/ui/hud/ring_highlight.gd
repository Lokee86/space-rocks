extends Control

@export var radius_ratio: float = 0.5
@export var arc_degrees: float = 3.0
@export var width: float = 1.5
@export var spin_degrees_per_second: float = 180.0
@export var highlight_color: Color = Color(1.0, 0.65, 0.95, 0.65)

@export var tail_degrees: float = 45.0
@export var tail_segments: int = 18
@export var tail_width_multiplier: float = 2.0
@export var tail_alpha_multiplier: float = 0.45

@export var inner_glow_alpha: float = 0.24
@export var outer_glow_alpha: float = 0.10
@export var inner_glow_width_multiplier: float = 3.0
@export var outer_glow_width_multiplier: float = 6.0

@export var second_highlight_offset_degrees: float = 180.0

var _angle: float = 0.0


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	resized.connect(_on_resized)


func _process(delta: float) -> void:
	_angle += deg_to_rad(spin_degrees_per_second) * delta
	_angle = fposmod(_angle, TAU)
	queue_redraw()


func _draw() -> void:
	var center := size * 0.5
	var radius: float = min(size.x, size.y) * radius_ratio
	var second_offset := deg_to_rad(second_highlight_offset_degrees)

	_draw_spinner_arc(center, radius, _angle, 1.0)
	_draw_spinner_arc(center, radius, _angle + second_offset, 1.0)


func _draw_spinner_arc(center: Vector2, radius: float, head_angle: float, direction: float) -> void:
	var half_arc := deg_to_rad(arc_degrees) * 0.5
	var tail_angle := deg_to_rad(tail_degrees)

	for i in range(tail_segments, 0, -1):
		var age := float(i) / float(tail_segments)
		var fade := pow(1.0 - age, 2.0)
		var segment_angle := head_angle - direction * tail_angle * age

		var tail_color := Color(
			highlight_color.r,
			highlight_color.g,
			highlight_color.b,
			highlight_color.a * fade * tail_alpha_multiplier
		)

		draw_arc(
			center,
			radius,
			segment_angle - half_arc,
			segment_angle + half_arc,
			16,
			Color(
				highlight_color.r,
				highlight_color.g,
				highlight_color.b,
				outer_glow_alpha * fade
			),
			width * outer_glow_width_multiplier,
			true
		)

		draw_arc(
			center,
			radius,
			segment_angle - half_arc,
			segment_angle + half_arc,
			16,
			tail_color,
			width * tail_width_multiplier,
			true
		)

	draw_arc(
		center,
		radius,
		head_angle - half_arc,
		head_angle + half_arc,
		32,
		Color(highlight_color.r, highlight_color.g, highlight_color.b, outer_glow_alpha),
		width * outer_glow_width_multiplier,
		true
	)

	draw_arc(
		center,
		radius,
		head_angle - half_arc,
		head_angle + half_arc,
		32,
		Color(highlight_color.r, highlight_color.g, highlight_color.b, inner_glow_alpha),
		width * inner_glow_width_multiplier,
		true
	)

	draw_arc(
		center,
		radius,
		head_angle - half_arc,
		head_angle + half_arc,
		32,
		highlight_color,
		width,
		true
	)


func _on_resized() -> void:
	queue_redraw()