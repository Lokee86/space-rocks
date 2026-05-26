extends RefCounted
class_name GameplayRuntimeTickFlow

var hud_flow


func configure(hud_flow_ref) -> void:
	hud_flow = hud_flow_ref


func process(delta: float) -> void:
	if hud_flow != null:
		hud_flow.update(delta)


func reset() -> void:
	pass
