extends Control
class_name LoadoutDialog

signal submit_requested(selection: Dictionary)
signal close_requested

@onready var readout: LoadoutReadout = %LoadoutReadout


func _ready() -> void:
	readout.submit_requested.connect(_on_readout_submit_requested)
	readout.close_requested.connect(_on_readout_close_requested)
	readout.apply_button.call_deferred("grab_focus")


func configure(build_options: Dictionary, loadout_selection: Dictionary) -> void:
	if not is_node_ready():
		await ready
	readout.configure(build_options, loadout_selection)


func _unhandled_input(event: InputEvent) -> void:
	if event.is_action_pressed("ui_cancel"):
		get_viewport().set_input_as_handled()
		close_requested.emit()


func _on_readout_submit_requested(selection: Dictionary) -> void:
	submit_requested.emit(selection)


func _on_readout_close_requested() -> void:
	close_requested.emit()
