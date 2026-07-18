extends RefCounted
class_name DevtoolsWindowMeasurement

signal start_requested(scenario_label: String)
signal stop_requested
signal reset_requested

var scenario_label_input: LineEdit
var start_button: Button
var stop_button: Button
var reset_button: Button
var status_label: Label
var active_run_id_label: Label
var export_label: Label


func configure(window_ref) -> void:
	scenario_label_input = window_ref.find_child("MeasurementScenarioLabel", true, false) as LineEdit
	start_button = window_ref.find_child("MeasurementStartButton", true, false) as Button
	stop_button = window_ref.find_child("MeasurementStopButton", true, false) as Button
	reset_button = window_ref.find_child("MeasurementResetButton", true, false) as Button
	status_label = window_ref.find_child("MeasurementStatusLabel", true, false) as Label
	active_run_id_label = window_ref.find_child("MeasurementActiveRunIdLabel", true, false) as Label
	export_label = window_ref.find_child("MeasurementExportLabel", true, false) as Label
	if start_button != null and !start_button.pressed.is_connected(_on_start_pressed):
		start_button.pressed.connect(_on_start_pressed)
	if stop_button != null and !stop_button.pressed.is_connected(_on_stop_pressed):
		stop_button.pressed.connect(_on_stop_pressed)
	if reset_button != null and !reset_button.pressed.is_connected(_on_reset_pressed):
		reset_button.pressed.connect(_on_reset_pressed)
	refresh_state({})


func refresh_state(state: Dictionary) -> void:
	var pending_request_ids: Dictionary = state.get("pending_request_ids", {})
	var recording := bool(state.get("recording", false))
	var starting := pending_request_ids.has("start")
	var stopping := pending_request_ids.has("stop")
	if start_button != null:
		start_button.disabled = recording or starting
	if stop_button != null:
		stop_button.disabled = !recording or stopping
	if reset_button != null:
		reset_button.disabled = (
			pending_request_ids.has("start")
			or pending_request_ids.has("stop")
			or pending_request_ids.has("reset")
		)

	var status := "Idle"
	if starting:
		status = "Starting"
	elif stopping:
		status = "Stopping"
	elif recording:
		status = "Recording"
	var tooling_error: Dictionary = state.get("last_tooling_error", {})
	var error_message := str(tooling_error.get("message", ""))
	if !error_message.is_empty():
		status += " — %s" % error_message
	if status_label != null:
		status_label.text = "Status: %s" % status

	if active_run_id_label != null:
		var active_run_id := str(state.get("active_run_id", ""))
		active_run_id_label.text = "Active run: %s" % (active_run_id if !active_run_id.is_empty() else "—")

	if export_label != null:
		var export_result: Dictionary = state.get("latest_export_result", {})
		var export_path := str(export_result.get("path", ""))
		var export_error := str(export_result.get("error", ""))
		if !export_error.is_empty():
			export_label.text = "Latest export error: %s" % export_error
		elif !export_path.is_empty():
			export_label.text = "Latest export: %s" % export_path
		else:
			export_label.text = "Latest export: —"


func _on_start_pressed() -> void:
	if scenario_label_input == null:
		start_requested.emit("")
		return
	start_requested.emit(scenario_label_input.text.strip_edges())


func _on_stop_pressed() -> void:
	stop_requested.emit()


func _on_reset_pressed() -> void:
	reset_requested.emit()
