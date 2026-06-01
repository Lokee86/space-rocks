extends RefCounted
class_name TargetVisualPicker

const TARGET_KIND_PRIORITY_PLAYER := 4
const TARGET_KIND_PRIORITY_ENEMY := 3
const TARGET_KIND_PRIORITY_ASTEROID := 2
const TARGET_KIND_PRIORITY_BULLET := 1

static func pick(candidates: Array, mouse_visual_position: Vector2):
	var best_candidate = null
	for candidate in candidates:
		if candidate == null:
			continue
		if not candidate.is_valid():
			continue
		if not candidate.visible:
			continue
		if mouse_visual_position.distance_to(candidate.visual_position) > candidate.pick_radius:
			continue
		if best_candidate == null:
			best_candidate = candidate
			continue
		if candidate.pick_rank > best_candidate.pick_rank:
			best_candidate = candidate
			continue
		if candidate.pick_rank == best_candidate.pick_rank:
			var candidate_kind_priority := _target_kind_priority(candidate.target_kind)
			var best_kind_priority := _target_kind_priority(best_candidate.target_kind)
			if candidate_kind_priority > best_kind_priority:
				best_candidate = candidate
				continue
			if candidate_kind_priority == best_kind_priority and String(candidate.target_id) < String(best_candidate.target_id):
				best_candidate = candidate

	return best_candidate

static func _target_kind_priority(target_kind: String) -> int:
	match target_kind:
		"player":
			return TARGET_KIND_PRIORITY_PLAYER
		"enemy":
			return TARGET_KIND_PRIORITY_ENEMY
		"asteroid":
			return TARGET_KIND_PRIORITY_ASTEROID
		"bullet":
			return TARGET_KIND_PRIORITY_BULLET
		_:
			return 0
