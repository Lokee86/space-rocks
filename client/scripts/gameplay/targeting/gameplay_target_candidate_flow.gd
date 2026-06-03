extends RefCounted
class_name GameplayTargetCandidateFlow

const TARGET_PLAYER_PICK_RADIUS := 32.0
const TARGET_ASTEROID_BASE_PICK_RADIUS := 32.0
const TARGET_BULLET_PICK_RADIUS := 12.0

var world_sync


func configure(world_sync_ref) -> void:
	world_sync = world_sync_ref


func target_visual_candidates() -> Array:
	var candidates: Array = []
	if world_sync == null:
		return candidates

	var player_positions: Dictionary = world_sync.player_target_positions()
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

	var asteroid_positions: Dictionary = world_sync.asteroid_target_positions()
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

	var bullet_positions: Dictionary = world_sync.bullet_target_positions()
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
