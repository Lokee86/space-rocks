extends Control
class_name LoadoutDialog

signal submit_requested(selection: Dictionary)
signal close_requested

@onready var panel: Control = %Panel
@onready var readout: LoadoutReadout = %LoadoutReadout


func _ready() -> void:
	readout.submit_requested.connect(_on_readout_submit_requested)
	readout.close_requested.connect(_on_readout_close_requested)
	get_viewport().size_changed.connect(_update_panel_size)
	_update_panel_size()
	readout.apply_button.call_deferred("grab_focus")


func configure(build_options: Dictionary, loadout_selection: Dictionary) -> void:
	if not is_node_ready():
		await ready
	readout.configure(build_options, loadout_selection)


func _update_panel_size() -> void:
	if panel == null:
		return
	var viewport_size := get_viewport_rect().size
	panel.custom_minimum_size = Vector2(
		minf(maxf(viewport_size.x - 48.0, 0.0), 900.0),
		minf(maxf(viewport_size.y - 48.0, 0.0), 640.0)
	)


func _unhandled_input(event: InputEvent) -> void:
	if event.is_action_pressed("ui_cancel"):
		get_viewport().set_input_as_handled()
		close_requested.emit()


func _on_readout_submit_requested(selection: Dictionary) -> void:
	submit_requested.emit(selection)


func _on_readout_close_requested() -> void:
	close_requested.emit()
