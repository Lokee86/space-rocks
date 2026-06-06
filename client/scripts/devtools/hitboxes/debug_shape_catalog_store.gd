extends RefCounted

var shapes: Dictionary = {}


func apply_catalog_state(state: Dictionary) -> void:
	var shapes_value = state.get("shapes", {})
	shapes = shapes_value if shapes_value is Dictionary else {}


func shape_for_id(shape_id: String) -> Dictionary:
	var shape = shapes.get(shape_id, {})
	return shape if shape is Dictionary else {}


func reset() -> void:
	shapes.clear()
