extends RefCounted

var pause_states := {}


func reset() -> void:
	pause_states.clear()


func apply_state(state: Dictionary) -> void:
	var player_id := String(state.get("player_id", ""))
	if player_id.is_empty():
		return
	pause_states[player_id] = bool(state.get("paused", false))


func is_paused(player_id: String) -> bool:
	if !pause_states.has(player_id):
		return false
	return bool(pause_states[player_id])

