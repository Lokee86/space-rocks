extends Control

@export var radius_ratio: float = 0.5
@export var arc_degrees: float = 3.0
@export var width: float = 1.5
@export var spin_degrees_per_second: float = 180.0
@export var highlight_color: Color = Color(1.0, 0.65, 0.95, 0.65)

func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	pivot_offset = size * 0.5
	resized.connect(_on_resized)


func _process(delta: float) -> void:
	rotation += deg_to_rad(spin_degrees_per_second) * delta


func _draw() -> void:
	var center := size * 0.5
	var radius:float = min(size.x, size.y) * radius_ratio
	var half_arc := deg_to_rad(arc_degrees) * 0.5

	# Faint wider glow behind the highlight.
	draw_arc(
		center,
		radius,
		-half_arc,
		half_arc,
		32,
		Color(highlight_color.r, highlight_color.g, highlight_color.b, 0.18),
		width * 2.5,
		true
	)

	# Sharp foreground arc.
	draw_arc(
		center,
		radius,
		-half_arc,
		half_arc,
		32,
		highlight_color,
		width,
		true
	)


func _on_resized() -> void:
	pivot_offset = size * 0.5
	queue_redraw()