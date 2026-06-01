extends RefCounted
class_name TargetVisualCandidate

var target_kind := ""
var target_id := ""
var visual_position := Vector2.ZERO
var server_position := Vector2.ZERO
var pick_radius := 0.0
var visible := true
var pick_rank := 0

func is_valid() -> bool:
	return target_kind != "" and target_id != "" and pick_radius > 0.0
