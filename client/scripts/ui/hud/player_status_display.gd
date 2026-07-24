extends VBoxContainer
class_name PlayerStatusDisplay

@onready var hull_label: Label = %HullLabel
@onready var shields_label: Label = %ShieldsLabel
@onready var modules_label: Label = %ModulesLabel


func apply_status(state: Dictionary) -> void:
	var health := _safe_int(state.get("health", null))
	var max_health := _safe_int(state.get("max_health", null))
	var shields := _safe_int(state.get("shields", null))
	var max_shields := _safe_int(state.get("max_shields", null))

	hull_label.text = _meter_text("HULL", health, max_health)
	shields_label.text = _meter_text("SHIELDS", shields, max_shields)
	modules_label.text = _modules_text(state)


func clear_status() -> void:
	hull_label.text = "HULL: --"
	shields_label.text = "SHIELDS: --"
	modules_label.text = "MODULES: NONE"


func _safe_int(value: Variant) -> int:
	if value is int or value is float:
		return int(value)
	return 0


func _meter_text(label: String, current: int, maximum: int) -> String:
	if maximum <= 0:
		return "%s: --" % label
	return "%s: %d/%d" % [label, clampi(current, 0, maximum), maximum]


func _modules_text(state: Dictionary) -> String:
	var equipped: Array[String] = []
	_append_module(equipped, "SHD", str(state.get("shield_module_id", "")))
	_append_module(equipped, "ARM", str(state.get("armor_module_id", "")))
	_append_module(equipped, "ENG", str(state.get("engine_module_id", "")))
	_append_module(equipped, "UTL", str(state.get("utility_module_id", "")))
	if equipped.is_empty():
		return "MODULES: NONE"
	return "MODULES: " + "  ".join(equipped)


func _append_module(equipped: Array[String], slot_label: String, module_id: String) -> void:
	if module_id.is_empty():
		return
	equipped.append("%s:%s" % [slot_label, module_id.replace("_", " ").to_upper()])
