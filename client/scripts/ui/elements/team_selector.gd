extends HBoxContainer
class_name TeamSelector

signal team_selected(team_id: String)

const TeamPresentation := preload("res://scripts/teams/team_presentation.gd")

@onready var swatch: ColorRect = %TeamSwatch
@onready var option_button = %TeamOptionButton
var _updating := false


func _ready() -> void:
	option_button.item_selected.connect(_on_item_selected)


func configure(team_id: String, team_ids: Array, editable: bool) -> void:
	_updating = true
	var items := []
	for candidate in team_ids:
		var candidate_id := str(candidate)
		items.append({"label": TeamPresentation.display_name(candidate_id), "value": candidate_id})
	option_button.replace_items(items, team_id)
	option_button.disabled = not editable
	option_button.mouse_filter = Control.MOUSE_FILTER_STOP if editable else Control.MOUSE_FILTER_IGNORE
	_apply_swatch(team_id)
	_updating = false


func _on_item_selected(_index: int) -> void:
	var team_id := str(option_button.selected_value(""))
	_apply_swatch(team_id)
	if not _updating:
		team_selected.emit(team_id)


func _apply_swatch(team_id: String) -> void:
	if swatch != null:
		swatch.color = TeamPresentation.color(team_id)
