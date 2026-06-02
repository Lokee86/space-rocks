extends RefCounted
class_name PlayerHuePresenter

const Constants = preload("res://scripts/constants/constants.gd")
const REMOTE_HUE_STEP := 0.12
const PLAYER_COLOR_POLICY_LOCAL_SELECTED := "local_selected"
const PLAYER_COLOR_POLICY_AUTO_DISTINCT := "auto_distinct"
const PLAYER_COLOR_POLICY_PLAYER_ID_ASSIGNED := "player_id_assigned"

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
var remote_player_order := []
var local_player_hue := Constants.PLAYER_DEFAULT_HUE


func reset() -> void:
	remote_player_hues.clear()
	remote_player_order.clear()
	local_player_hue = Constants.PLAYER_DEFAULT_HUE


func apply_local_player_hue(player: Player) -> void:
	if player == null:
		return
	local_player_hue = player.player_hue


func set_remote_player_order(remote_player_ids: Array) -> void:
	remote_player_order = remote_player_ids.duplicate()


func apply_remote_player_hue(player_id: String, remote_player: Player) -> void:
	if remote_player == null:
		return

	var hue := remote_hue_for_player(player_id)
	remote_player_hues[player_id] = hue
	remote_player.set_player_hue(hue)


func apply_os_indicator_hue(player_id: String, indicator: Control) -> void:
	if indicator == null:
		return

	var graphic := indicator.get_node_or_null("TextureRect") as CanvasItem
	if graphic == null:
		return

	var shader_material := graphic.material as ShaderMaterial
	if shader_material == null:
		return

	graphic.material = shader_material.duplicate() as ShaderMaterial
	(graphic.material as ShaderMaterial).set_shader_parameter("hue_shift", remote_hue_for_player(player_id))


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
	var player_color_policy := _player_color_policy()
	if player_color_policy != PLAYER_COLOR_POLICY_LOCAL_SELECTED \
			and player_color_policy != PLAYER_COLOR_POLICY_AUTO_DISTINCT \
			and player_color_policy != PLAYER_COLOR_POLICY_PLAYER_ID_ASSIGNED:
		player_color_policy = PLAYER_COLOR_POLICY_AUTO_DISTINCT

	if player_color_policy == PLAYER_COLOR_POLICY_PLAYER_ID_ASSIGNED:
		player_color_policy = PLAYER_COLOR_POLICY_AUTO_DISTINCT

	var slot_index := remote_player_order.find(player_id)
	if slot_index >= 0:
		var hue := local_player_hue + (float(slot_index + 1) * REMOTE_HUE_STEP)
		if hue > 1.0:
			hue -= 1.0
		return hue
	if remote_player_hues.has(player_id):
		return float(remote_player_hues[player_id])
	return Constants.REMOTE_PLAYER_FALLBACK_HUE


func _player_color_policy() -> String:
	return str(Constants.PLAYER_COLOR_POLICY)


func hues_similar(
		first_hue: float,
		second_hue: float,
		tolerance := Constants.REMOTE_PLAYER_HUE_SIMILARITY_TOLERANCE
) -> bool:
		var distance: float = abs(fposmod(first_hue, 1.0) - fposmod(second_hue, 1.0))
		return min(distance, 1.0 - distance) < tolerance


func player_id_hash(player_id: String) -> int:
	var hash_value: int = 2166136261
	for index in range(player_id.length()):
		hash_value = int((hash_value ^ player_id.unicode_at(index)) * 16777619) & 0x7fffffff
	return hash_value
