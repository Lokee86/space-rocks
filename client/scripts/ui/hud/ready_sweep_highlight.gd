extends ColorRect
class_name ReadySweepHighlight

@export var sweep_duration: float = 0.35
@export var sweep_start: float = -0.5
@export var sweep_end: float = 1.5
@export var sweep_width: float = 0.08
@export var sweep_softness: float = 0.12
@export var sweep_alpha: float = 0.75
@export var sweep_angle: float = 0.75
@export var badge_radius: float = 0.5
@export var badge_softness: float = 0.02

var _active_tween: Tween


func _ready() -> void:
	hide()
	_apply_shader_parameters()


func play() -> void:
	var shader_material := material as ShaderMaterial
	if shader_material == null:
		return

	if _active_tween != null and is_instance_valid(_active_tween):
		_active_tween.kill()

	show()
	_apply_shader_parameters()
	_set_sweep_progress(sweep_start)

	_active_tween = create_tween()
	_active_tween.tween_method(Callable(self, "_set_sweep_progress"), sweep_start, sweep_end, sweep_duration)
	_active_tween.finished.connect(
		func() -> void:
			_set_sweep_progress(sweep_start)
			hide()
			_active_tween = null
	)


func _set_sweep_progress(progress: float) -> void:
	var shader_material := material as ShaderMaterial
	if shader_material == null:
		return

	shader_material.set_shader_parameter("sweep_progress", progress)


func _apply_shader_parameters() -> void:
	var shader_material := material as ShaderMaterial
	if shader_material == null:
		return

	shader_material.set_shader_parameter("sweep_width", sweep_width)
	shader_material.set_shader_parameter("sweep_softness", sweep_softness)
	shader_material.set_shader_parameter("sweep_alpha", sweep_alpha)
	shader_material.set_shader_parameter("sweep_angle", sweep_angle)
	shader_material.set_shader_parameter("badge_radius", badge_radius)
	shader_material.set_shader_parameter("badge_softness", badge_softness)
