extends RefCounted
class_name SpectateMenuState

const PlayerLifecycle := preload("res://scripts/gameplay/lifecycle/player_lifecycle.gd")

# Owns spectate availability/menu state. This is the future home for outside
# spectating rules; it does not render UI or send packets.

var self_id := ""
var current_target_id := ""
var player_lifecycle := {}
var world_ships := {}


func apply_gameplay_state(state: Dictionary) -> void:
	var overlay: Variant = state.get("overlay", {})
	var nested_self_id: String = str(overlay.get("self_id", "")) if overlay is Dictionary else ""
	self_id = str(state.get("self_id", nested_self_id))

	var session: Variant = state.get("session", {})
	var nested_lifecycle: Variant = session.get("player_lifecycle", {}) if session is Dictionary else {}
	var lifecycle: Variant = state.get("player_lifecycle", nested_lifecycle)
	if lifecycle is Dictionary:
		player_lifecycle = lifecycle.duplicate(true)
	else:
		player_lifecycle = {}

	var world: Variant = state.get("world", {})
	var nested_ships: Variant = world.get("ships", {}) if world is Dictionary else {}
	var ships: Variant = state.get("ships", nested_ships)
	if ships is Dictionary:
		world_ships = ships.duplicate(true)
	else:
		world_ships = {}


func reset() -> void:
	self_id = ""
	current_target_id = ""
	player_lifecycle.clear()
	world_ships.clear()


func spectate_target_ids() -> Array:
	var target_ids := []
	for player_id_value in world_ships.keys():
		var player_id := str(player_id_value)
		if player_id == self_id:
			continue

		var lifecycle_status := PlayerLifecycle.status_for(player_lifecycle, player_id)
		if !lifecycle_status.is_empty() and lifecycle_status != PlayerLifecycle.STATUS_ACTIVE:
			continue

		target_ids.append(player_id)
	target_ids.sort()
	return target_ids


func current_target() -> String:
	var target_ids := spectate_target_ids()
	if current_target_id in target_ids:
		return current_target_id
	if target_ids.is_empty():
		return ""
	return str(target_ids[0])


func begin_spectating() -> String:
	current_target_id = current_target()
	return current_target_id


func cycle_next_target() -> String:
	var target_ids := spectate_target_ids()
	if target_ids.is_empty():
		current_target_id = ""
		return current_target_id

	var current_index := target_ids.find(current_target_id)
	if current_index == -1:
		current_target_id = str(target_ids[0])
		return current_target_id

	current_target_id = str(target_ids[(current_index + 1) % target_ids.size()])
	return current_target_id


func has_spectate_targets() -> bool:
	return !self_id.is_empty() && !spectate_target_ids().is_empty()
