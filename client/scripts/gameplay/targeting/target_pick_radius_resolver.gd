extends RefCounted
class_name TargetPickRadiusResolver


static func pickup_radius(position_entry: Dictionary, fallback: float) -> float:
	var node = position_entry.get("node", null)
	if node != null and is_instance_valid(node) and node.has_method("collision_radius"):
		var radius = float(node.collision_radius())
		if radius > 0.0:
			return radius

	return fallback
