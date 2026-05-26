extends RefCounted
class_name PlayerHuePresenter

const Constants = preload("res://scripts/constants/constants.gd")

const REMOTE_PLAYER_HUES := [
	Constants.REMOTE_PLAYER_HUE_ZERO,
	Constants.REMOTE_PLAYER_HUE_ONE,
	Constants.REMOTE_PLAYER_HUE_TWO,
	Constants.REMOTE_PLAYER_HUE_THREE,
	Constants.REMOTE_PLAYER_HUE_FOUR,
	Constants.REMOTE_PLAYER_HUE_FIVE,
	Constants.REMOTE_PLAYER_HUE_SIX,
	Constants.REMOTE_PLAYER_HUE_SEVEN,
]

var remote_player_hues := {}


func reset() -> void:
	remote_player_hues.clear()


func apply_local_player_hue(player: Player) -> void:
	if player == null:
		return
	player.set_player_hue(Constants.LOCAL_PLAYER_DEFAULT_HUE)


func apply_remote_player_hue(player_id: String, remote_player: Player) -> void:
	if remote_player == null:
		return

	var hue := remote_hue_for_player(player_id)
	remote_player_hues[player_id] = hue
	remote_player.set_player_hue(hue)


func remove_player(player_id: String) -> void:
	remote_player_hues.erase(player_id)


func remote_player_hues_without(current_self_id: String) -> Dictionary:
	var hues := {}
	for player_id in remote_player_hues:
		if player_id == current_self_id:
			continue
		hues[player_id] = remote_player_hues[player_id]
	return hues


func remote_hue_for_player(player_id: String) -> float:
	if REMOTE_PLAYER_HUES.is_empty():
		return Constants.REMOTE_PLAYER_FALLBACK_HUE

	var start_index := player_id_hash(player_id) % REMOTE_PLAYER_HUES.size()
	for offset in REMOTE_PLAYER_HUES.size():
		var hue: float = REMOTE_PLAYER_HUES[(start_index + offset) % REMOTE_PLAYER_HUES.size()]
		if !hues_similar(hue, Constants.LOCAL_PLAYER_DEFAULT_HUE):
			return hue
	return Constants.REMOTE_PLAYER_FALLBACK_HUE


func hues_similar(
	first_hue: float,
	second_hue: float,
	tolerance := Constants.REMOTE_PLAYER_HUE_SIMILARITY_TOLERANCE
) -> bool:
	var distance := abs(fposmod(first_hue, 1.0) - fposmod(second_hue, 1.0))
	return min(distance, 1.0 - distance) < tolerance


func player_id_hash(player_id: String) -> int:
	var hash_value: int = 2166136261
	for index in range(player_id.length()):
		hash_value = int((hash_value ^ player_id.unicode_at(index)) * 16777619) & 0x7fffffff
	return hash_value
