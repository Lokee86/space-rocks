extends RefCounted
class_name DevtoolsWindowTelemetry

const TELEMETRY_SOURCE_PLAYERS := "players"
const TELEMETRY_SOURCE_PLAYER_WORLD_STATES := "player_world_states"

var window

func configure(window_ref) -> void:
	window = window_ref

func initialize() -> void:
	_initialize_telemetry_source_select(window.local_telemetry_select)
	_initialize_telemetry_source_select(window.target_telemetry_select)
	if !window.local_telemetry_select.item_selected.is_connected(window._on_local_telemetry_select_item_selected):
		window.local_telemetry_select.item_selected.connect(window._on_local_telemetry_select_item_selected)
	if !window.target_telemetry_select.item_selected.is_connected(window._on_target_telemetry_select_item_selected):
		window.target_telemetry_select.item_selected.connect(window._on_target_telemetry_select_item_selected)

func refresh_local_player_state(state: Dictionary) -> void:
	if state.is_empty():
		window.local_player_telemetry_text.text = "—"
		return
	var keys := state.keys()
	keys.sort()
	var lines: Array[String] = []
	for key in keys:
		lines.append("%s: %s" % [str(key), _format_telemetry_value(state.get(key))])
	window.local_player_telemetry_text.text = "\n".join(lines)

func refresh_target_state(target_kind: String, target_id: String, state: Dictionary) -> void:
	if target_kind == "" or target_id == "":
		window.target_telemetry_text.text = "—"
		return
	var lines: Array[String] = ["target_kind: %s" % target_kind, "target_id: %s" % target_id]
	if state.is_empty():
		lines.append("state: —")
	else:
		lines.append("")
		var keys := state.keys()
		keys.sort()
		for key in keys:
			lines.append("%s: %s" % [str(key), _format_telemetry_value(state.get(key))])
	window.target_telemetry_text.text = "\n".join(lines)

func local_source() -> String:
	return _selected_metadata_as_string(window.local_telemetry_select)

func target_source() -> String:
	return _selected_metadata_as_string(window.target_telemetry_select)

func set_sources(local_source_name: String, target_source_name: String) -> void:
	_select_telemetry_source(window.local_telemetry_select, local_source_name)
	_select_telemetry_source(window.target_telemetry_select, target_source_name)

func _selected_metadata_as_string(select: OptionButton) -> String:
	var index := select.get_selected()
	return "" if index < 0 else str(select.get_item_metadata(index))

func _initialize_telemetry_source_select(select: OptionButton) -> void:
	select.clear()
	select.add_item("players")
	select.set_item_metadata(0, TELEMETRY_SOURCE_PLAYERS)
	select.add_item("player_world_states")
	select.set_item_metadata(1, TELEMETRY_SOURCE_PLAYER_WORLD_STATES)
	select.select(0)

func _select_telemetry_source(select: OptionButton, source: String) -> void:
	for index in range(select.get_item_count()):
		if str(select.get_item_metadata(index)) == source:
			select.select(index)
			return

func _format_telemetry_value(value) -> String:
	if value is Array or value is Dictionary:
		return JSON.stringify(value)
	if value is float:
		return "%.4f" % snappedf(value, 0.0001)
	return str(value)
