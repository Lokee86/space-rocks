extends RefCounted
class_name GameplayTargetCandidateFlow

const TARGET_PLAYER_PICK_RADIUS := 32.0
const TARGET_PICKUP_PICK_RADIUS := 32.0
const TARGET_ASTEROID_BASE_PICK_RADIUS := 32.0
const TARGET_BULLET_PICK_RADIUS := 12.0

var target_position_source


func configure(target_position_source_ref) -> void:
	target_position_source = target_position_source_ref


func target_visual_candidates() -> Array:
	var candidates: Array = []
	if target_position_source == null:
		return candidates

	var player_positions: Dictionary = target_position_source.player_positions()
	for player_id in player_positions.keys():
		var position_entry = player_positions[player_id]
		if not (position_entry is Dictionary):
			continue
		if not position_entry.has("visual_position"):
			continue
		if not position_entry.has("server_position"):
			continue

		var candidate := TargetVisualCandidate.new()
		candidate.target_kind = "player"
		candidate.target_id = String(player_id)
		candidate.visual_position = position_entry["visual_position"]
		candidate.server_position = position_entry["server_position"]
		candidate.pick_radius = TARGET_PLAYER_PICK_RADIUS
		candidates.append(candidate)

	var pickup_positions: Dictionary = target_position_source.pickup_positions()
	for pickup_id in pickup_positions.keys():
		var position_entry = pickup_positions[pickup_id]
		if not (position_entry is Dictionary):
			continue
		if not position_entry.has("visual_position"):
			continue
		if not position_entry.has("server_position"):
			continue

		var pickup_candidate := TargetVisualCandidate.new()
		pickup_candidate.target_kind = "pickup"
		pickup_candidate.target_id = String(pickup_id)
		pickup_candidate.visual_position = position_entry["visual_position"]
		pickup_candidate.server_position = position_entry["server_position"]
		pickup_candidate.pick_radius = TargetPickRadiusResolver.pickup_radius(position_entry, TARGET_PICKUP_PICK_RADIUS)
		candidates.append(pickup_candidate)

	var asteroid_positions: Dictionary = target_position_source.asteroid_positions()
	for asteroid_id in asteroid_positions.keys():
		var position_entry = asteroid_positions[asteroid_id]
		if not (position_entry is Dictionary):
			continue
		if not position_entry.has("visual_position"):
			continue
		if not position_entry.has("server_position"):
			continue

		var visual_scale := 1.0
		if position_entry.has("visual_scale"):
			visual_scale = float(position_entry["visual_scale"])

		var asteroid_candidate := TargetVisualCandidate.new()
		asteroid_candidate.target_kind = "asteroid"
		asteroid_candidate.target_id = String(asteroid_id)
		asteroid_candidate.visual_position = position_entry["visual_position"]
		asteroid_candidate.server_position = position_entry["server_position"]
		asteroid_candidate.pick_radius = TARGET_ASTEROID_BASE_PICK_RADIUS * visual_scale
		candidates.append(asteroid_candidate)

	var bullet_positions: Dictionary = target_position_source.projectile_positions()
	for bullet_id in bullet_positions.keys():
		var position_entry = bullet_positions[bullet_id]
		if not (position_entry is Dictionary):
			continue
		if not position_entry.has("visual_position"):
			continue
		if not position_entry.has("server_position"):
			continue

		var bullet_candidate := TargetVisualCandidate.new()
		bullet_candidate.target_kind = "bullet"
		bullet_candidate.target_id = String(bullet_id)
		bullet_candidate.visual_position = position_entry["visual_position"]
		bullet_candidate.server_position = position_entry["server_position"]
		bullet_candidate.pick_radius = TARGET_BULLET_PICK_RADIUS
		candidates.append(bullet_candidate)

	return candidates
