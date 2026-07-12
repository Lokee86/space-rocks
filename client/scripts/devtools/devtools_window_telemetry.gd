extends RefCounted
class_name DevtoolsWindowTelemetry

const TELEMETRY_SOURCE_PLAYERS := "players"
const TELEMETRY_SOURCE_PLAYER_WORLD_STATES := "player_world_states"

var local_telemetry_select: OptionButton
var target_telemetry_select: OptionButton
var local_player_telemetry_text: Label
var target_telemetry_text: Label


func configure(
	local_select: OptionButton,
	target_select: OptionButton,
	local_text: Label,
	target_text: Label
) -> void:
	local_telemetry_select = local_select
	target_telemetry_select = target_select
	local_player_telemetry_text = local_text
	target_telemetry_text = target_text
	_initialize_source_select(local_telemetry_select)
	_initialize_source_select(target_telemetry_select)


func local_source() -> String:
	return _selected_metadata_as_string(local_telemetry_select)


func target_source() -> String:
	return _selected_metadata_as_string(target_telemetry_select)


func set_sources(local_source_name: String, target_source_name: String) -> void:
	_select_source(local_telemetry_select, local_source_name)
	_select_source(target_telemetry_select, target_source_name)


func refresh_local_state(state: Dictionary) -> void:
	local_player_telemetry_text.text = _format_state(state, false, "", "")


func refresh_target_state(target_kind: String, target_id: String, state: Dictionary) -> void:
	target_telemetry_text.text = _format_state(state, true, target_kind, target_id)


func _format_state(state: Dictionary, include_target: bool, target_kind: String, target_id: String) -> String:
	if include_target and (target_kind == "" or target_id == ""):
		return "—"
	if !include_target and state.is_empty():
		return "—"

	var lines: Array[String] = []
	if include_target:
		lines.append("target_kind: %s" % target_kind)
		lines.append("target_id: %s" % target_id)
		if state.is_empty():
			lines.append("state: —")
			return "\n".join(lines)
		lines.append("")

	var keys := state.keys()
	keys.sort()
	for key in keys:
		lines.append("%s: %s" % [str(key), _format_value(state.get(key))])
	return "\n".join(lines)


func _format_value(value) -> String:
	if value is Array or value is Dictionary:
		return JSON.stringify(value)
	if value is float:
		return "%.4f" % snappedf(value, 0.0001)
	return str(value)


func _selected_metadata_as_string(select: OptionButton) -> String:
	var selected_index := select.get_selected()
	if selected_index < 0:
		return ""
	return str(select.get_item_metadata(selected_index))


func _initialize_source_select(select: OptionButton) -> void:
	select.clear()
	select.add_item(TELEMETRY_SOURCE_PLAYERS)
	select.set_item_metadata(0, TELEMETRY_SOURCE_PLAYERS)
	select.add_item(TELEMETRY_SOURCE_PLAYER_WORLD_STATES)
	select.set_item_metadata(1, TELEMETRY_SOURCE_PLAYER_WORLD_STATES)
	select.select(0)


func _select_source(select: OptionButton, source: String) -> void:
	var selected_index := 0
	for index in range(select.get_item_count()):
		if str(select.get_item_metadata(index)) == source:
			selected_index = index
			break
	select.select(selected_index)
