extends Control
class_name LoadoutReadout

signal submit_requested(selection: Dictionary)
signal close_requested

const WEAPON_POINT_PRIMARY_1 := "primary_1"
const MODULE_SLOTS := ["shield_mod", "armor_mod", "engine_mod", "utility_mod"]

@onready var ship_option: OptionButton = %ShipOption
@onready var primary_weapon_option: OptionButton = %PrimaryWeaponOption
@onready var module_options: Dictionary = {
	"shield_mod": %ShieldModuleOption,
	"armor_mod": %ArmorModuleOption,
	"engine_mod": %EngineModuleOption,
	"utility_mod": %UtilityModuleOption,
}
@onready var status_label: Label = %StatusLabel
@onready var apply_button: BaseButton = %ApplyButton
@onready var back_button: BaseButton = %BackButton

var _starting_ammo_by_point: Dictionary = {}


func _ready() -> void:
	apply_button.pressed.connect(_on_apply_pressed)
	back_button.pressed.connect(_on_back_pressed)
	_set_selector_empty(ship_option, "NO SHIPS AVAILABLE")
	_set_selector_empty(primary_weapon_option, "NO PRIMARY_1 WEAPONS AVAILABLE")
	for slot in MODULE_SLOTS:
		_set_selector_empty(module_options[slot], "NO " + slot.to_upper() + " AVAILABLE")


func configure(build_options: Dictionary, loadout_selection: Dictionary) -> void:
	var options := _nested_dictionary(build_options, "build_options")
	var selection := _nested_dictionary(loadout_selection, "loadout_selection")
	var selected_weapons := _dictionary_copy(selection.get("selected_weapons_by_point", {}))
	var selected_modules := _dictionary_copy(selection.get("selected_modules_by_slot", {}))
	_starting_ammo_by_point = _dictionary_copy(selection.get("starting_ammo_by_point", {}))

	_populate_selector(
		ship_option,
		_array_value(options, "eligible_ships"),
		"owned_ship_id",
		str(selection.get("selected_owned_ship_id", ""))
	)
	_populate_selector(
		primary_weapon_option,
		_weapon_options_for_point(_array_value(options, "eligible_weapons"), WEAPON_POINT_PRIMARY_1),
		"owned_weapon_id",
		str(selected_weapons.get(WEAPON_POINT_PRIMARY_1, ""))
	)
	for slot in MODULE_SLOTS:
		_populate_selector(
			module_options[slot],
			_module_options_for_slot(_array_value(options, "eligible_modules"), slot),
			"owned_module_id",
			str(selected_modules.get(slot, ""))
		)

	_set_status(selection)
	_apply_button_state()


func _on_apply_pressed() -> void:
	var selected_ship_id := _selected_id(ship_option)
	var selected_weapon_id := _selected_id(primary_weapon_option)
	if selected_ship_id == "" or selected_weapon_id == "":
		status_label.text = "ERROR: SELECT A SHIP AND PRIMARY_1 WEAPON"
		return

	var selected_weapons := {WEAPON_POINT_PRIMARY_1: selected_weapon_id}
	var selected_modules := {}
	for slot in MODULE_SLOTS:
		var module_id := _selected_id(module_options[slot])
		if module_id != "":
			selected_modules[slot] = module_id

	submit_requested.emit({
		"selected_owned_ship_id": selected_ship_id,
		"selected_weapons_by_point": selected_weapons,
		"selected_modules_by_slot": selected_modules,
		"starting_ammo_by_point": _starting_ammo_by_point.duplicate(true),
	})


func _on_back_pressed() -> void:
	close_requested.emit()


func _populate_selector(selector: OptionButton, options: Array, value_key: String, selected_id: String) -> void:
	selector.clear()
	for option in options:
		if not option is Dictionary:
			continue
		var option_id := str(option.get(value_key, ""))
		if option_id == "":
			continue
		var label := _option_label(option, option_id)
		selector.add_item(label)
		selector.set_item_metadata(selector.item_count - 1, option_id)

	if selector.item_count == 0:
		_set_selector_empty(selector, "NO OPTIONS AVAILABLE")
		return

	selector.disabled = false
	var selected_index := 0
	for index in range(selector.item_count):
		if str(selector.get_item_metadata(index)) == selected_id:
			selected_index = index
			break
	selector.select(selected_index)


func _option_label(option: Dictionary, fallback_id: String) -> String:
	var label := str(option.get("display_name", option.get("name", "")))
	if label.is_empty():
		for catalog_key in ["ship_id", "weapon_id", "module_id"]:
			label = str(option.get(catalog_key, ""))
			if not label.is_empty():
				break
	if label.is_empty():
		label = fallback_id
	return label.replace("_", " ").to_upper()


func _set_selector_empty(selector: OptionButton, message: String) -> void:
	selector.clear()
	selector.text = message
	selector.disabled = true


func _selected_id(selector: OptionButton) -> String:
	if selector.disabled or selector.selected < 0:
		return ""
	return str(selector.get_item_metadata(selector.selected))


func _apply_button_state() -> void:
	apply_button.disabled = _selected_id(ship_option) == "" or _selected_id(primary_weapon_option) == ""


func _set_status(selection: Dictionary) -> void:
	var message := str(selection.get("message", ""))
	var error_code := str(selection.get("error_code", ""))
	if message != "":
		status_label.text = message
	elif error_code != "":
		status_label.text = "ERROR: " + error_code
	else:
		status_label.text = "LOADOUT STATUS: READY"


func _weapon_options_for_point(options: Array, point: String) -> Array:
	var result: Array = []
	for option in options:
		if option is Dictionary and str(option.get("weapon_point", "")) == point:
			result.append(option)
	return result


func _module_options_for_slot(options: Array, slot: String) -> Array:
	var result: Array = []
	for option in options:
		if option is Dictionary and str(option.get("module_slot", "")) == slot:
			result.append(option)
	return result


func _nested_dictionary(value: Dictionary, key: String) -> Dictionary:
	var nested = value.get(key, value)
	return nested as Dictionary if nested is Dictionary else {}


func _dictionary_copy(value) -> Dictionary:
	return value.duplicate(true) if value is Dictionary else {}


func _array_value(value: Dictionary, key: String) -> Array:
	return value.get(key, []) if value.get(key, []) is Array else []
