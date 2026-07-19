extends RefCounted
class_name ClientMeasurementCoordinator

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ReportStateScript := preload("res://scripts/devtools/measurement/client_measurement_report_state.gd")

signal run_started(packet: Dictionary)
signal snapshot_received(report: Dictionary)
signal finalized(report: Dictionary)
signal error_received(error: Dictionary)
signal state_changed(state: Dictionary)

var connection_service
var measurement_context
var report_writer
var _report_state := ReportStateScript.new()

var request_ids: Dictionary:
	get:
		return _report_state.request_ids
var pending_request_ids: Dictionary:
	get:
		return _report_state.pending_request_ids
var active_run_id := ""
var recording := false
var latest_server_snapshot: Dictionary:
	get:
		return _report_state.latest_server_snapshot
var latest_server_report: Dictionary:
	get:
		return _report_state.latest_server_report
var latest_client_report: Dictionary:
	get:
		return _report_state.latest_client_report
var latest_combined_report: Dictionary:
	get:
		return _report_state.latest_combined_report
var latest_export_result: Dictionary = {}
var last_tooling_error: Dictionary:
	get:
		return _report_state.last_tooling_error

var _pending_start_metadata: Dictionary = {}
var _pending_scenario_label := ""
var _active_metadata: Dictionary = {}


func _init(connection_service_ref = null, measurement_context_ref = null, report_writer_ref = null) -> void:
	if connection_service_ref != null or measurement_context_ref != null or report_writer_ref != null:
		configure(connection_service_ref, measurement_context_ref, report_writer_ref)


func configure(connection_service_ref, measurement_context_ref = null, report_writer_ref = null) -> void:
	connection_service = connection_service_ref
	measurement_context = measurement_context_ref
	if report_writer_ref != null:
		configure_report_writer(report_writer_ref)
	_connect_connection_signal("measurement_started_received", Callable(self, "_on_measurement_started_received"))
	_connect_connection_signal("measurement_snapshot_received", Callable(self, "_on_measurement_snapshot_received"))
	_connect_connection_signal("measurement_stopped_received", Callable(self, "_on_measurement_stopped_received"))
	_connect_connection_signal("tooling_error_received", Callable(self, "_on_tooling_error_received"))


func configure_report_writer(report_writer_ref) -> void:
	report_writer = report_writer_ref
	var finalized_handler := Callable(self, "_on_finalized_report")
	if report_writer != null and !finalized.is_connected(finalized_handler):
		finalized.connect(finalized_handler)

func start(scenario_label: String = "", metadata: Dictionary = {}) -> String:
	if active_run_id != "" or pending_request_ids.has("start"):
		_reject("measurement_already_active", "a measurement run is already active or starting")
		return ""
	if !_can_send_tooling():
		_reject("tooling_transport_unavailable", "measurement tooling transport is not configured")
		return ""

	_pending_start_metadata = metadata.duplicate(true)
	_pending_scenario_label = scenario_label
	var request_id := _report_state.reserve_request("start", "")
	_send(Packets.measurement_start_packet(request_id, scenario_label))
	return request_id


func stop() -> String:
	if active_run_id == "":
		_reject("measurement_not_active", "no active measurement run")
		return ""
	if pending_request_ids.has("stop"):
		_reject("measurement_stop_pending", "a measurement stop request is already pending")
		return ""
	if !_can_send_tooling():
		_reject("tooling_transport_unavailable", "measurement tooling transport is not configured")
		return ""

	var request_id := _report_state.reserve_request("stop", active_run_id)
	_send(Packets.measurement_stop_packet(request_id, active_run_id))
	return request_id


func reset() -> String:
	if active_run_id == "":
		_reset_client_context()
		recording = false
		_emit_state_changed()
		return ""
	if pending_request_ids.has("reset"):
		_reject("measurement_reset_pending", "a measurement reset request is already pending")
		return ""
	if !_can_send_tooling():
		_reject("tooling_transport_unavailable", "measurement tooling transport is not configured")
		return ""

	var request_id := _report_state.reserve_request("reset", active_run_id)
	_send(Packets.measurement_reset_packet(request_id, active_run_id))
	_report_state.remove_pending("reset")
	_report_state.remember_reset_request(request_id)
	_reset_client_context()
	if measurement_context != null and measurement_context.has_method("start"):
		measurement_context.start(_active_metadata)
	recording = true
	_emit_state_changed()
	return request_id


func snapshot() -> String:
	if active_run_id == "":
		_reject("measurement_not_active", "no active measurement run")
		return ""
	if pending_request_ids.has("snapshot"):
		_reject("measurement_snapshot_pending", "a measurement snapshot request is already pending")
		return ""
	if !_can_send_tooling():
		_reject("tooling_transport_unavailable", "measurement tooling transport is not configured")
		return ""

	var request_id := _report_state.reserve_request("snapshot", active_run_id)
	_send(Packets.measurement_snapshot_request_packet(request_id, active_run_id))
	return request_id

func is_recording() -> bool:
	return recording


func get_latest_server_snapshot() -> Dictionary:
	return _report_state.latest_server_snapshot.duplicate(true)


func get_latest_server_report() -> Dictionary:
	return _report_state.latest_server_report.duplicate(true)


func get_latest_combined_report() -> Dictionary:
	return _report_state.latest_combined_report.duplicate(true)


func get_latest_export_result() -> Dictionary:
	return latest_export_result.duplicate(true)


func get_state() -> Dictionary:
	var state := _report_state.get_state(active_run_id, recording)
	state["latest_export_result"] = latest_export_result.duplicate(true)
	return state


func reset_local_state() -> void:
	active_run_id = ""
	recording = false
	_pending_start_metadata.clear()
	_pending_scenario_label = ""
	_active_metadata.clear()
	_report_state.clear_pending_requests()
	_reset_client_context()
	_emit_state_changed()


func _on_measurement_started_received(packet: Dictionary) -> void:
	if !_matches_pending(packet, "start", false):
		return
	var run_id := str(packet.get(Packets.FIELD_RUN_ID, ""))
	if run_id.is_empty():
		_record_error("invalid_measurement_started", "measurement_started did not include a run_id", packet)
		return

	_report_state.remove_pending("start")
	active_run_id = run_id
	recording = true
	latest_export_result.clear()
	_active_metadata = _pending_start_metadata.duplicate(true)
	_active_metadata["scenario_label"] = _pending_scenario_label
	_pending_start_metadata.clear()
	_pending_scenario_label = ""
	_report_state.clear_for_new_run()
	if measurement_context != null and measurement_context.has_method("start"):
		measurement_context.start(_active_metadata)
	run_started.emit(packet.duplicate(true))
	_emit_state_changed()


func _on_measurement_snapshot_received(packet: Dictionary) -> void:
	if !_matches_pending(packet, "snapshot", true):
		return
	_report_state.remove_pending("snapshot")
	var report = packet.get(Packets.FIELD_REPORT, {})
	var snapshot_report := _report_state.set_server_snapshot(report)
	snapshot_received.emit(snapshot_report)
	_emit_state_changed()


func _on_measurement_stopped_received(packet: Dictionary) -> void:
	if !_matches_pending(packet, "stop", true):
		return
	_report_state.remove_pending("stop")
	var run_id := active_run_id
	var server_report = packet.get(Packets.FIELD_REPORT, {})
	_report_state.set_server_report(server_report)
	var partial := bool(packet.get(Packets.FIELD_PARTIAL, false))
	var client_report := {}
	if measurement_context != null and measurement_context.has_method("stop"):
		var stop_reason := "server_partial" if partial else ""
		var result = measurement_context.stop(stop_reason)
		if result is Dictionary:
			client_report = result.duplicate(true)
	_report_state.set_client_report(client_report)
	recording = false
	active_run_id = ""
	_report_state.clear_pending_requests()
	var combined_report := _report_state.build_combined_report(run_id)
	finalized.emit(combined_report)
	_emit_state_changed()


func _on_finalized_report(report: Dictionary) -> void:
	if report_writer == null or !report_writer.has_method("write"):
		latest_export_result = {
			"success": false,
			"path": "",
			"error": "measurement report writer is not configured",
		}
		return
	var result = report_writer.write(report, str(report.get("run_id", "")))
	latest_export_result = result.duplicate(true) if result is Dictionary else {}


func _on_tooling_error_received(packet: Dictionary) -> void:
	var error := _report_state.resolve_tooling_error(packet)
	_record_error(str(error["error_code"]), str(error["message"]), packet)
	if bool(error.get("state_changed", false)):
		_emit_state_changed()


func _matches_pending(packet: Dictionary, operation: String, require_active_run: bool) -> bool:
	var mismatch := _report_state.match_pending(packet, operation, require_active_run, active_run_id)
	if !mismatch.is_empty():
		_record_error(str(mismatch["error_code"]), str(mismatch["message"]), packet)
		return false
	return true


func _send(packet: Dictionary) -> void:
	connection_service.send_tooling_packet(packet)


func _can_send_tooling() -> bool:
	if connection_service == null or !connection_service.has_method("send_tooling_packet"):
		return false
	if connection_service.has_method("is_tooling_ready"):
		return bool(connection_service.is_tooling_ready())
	return true


func _connect_connection_signal(signal_name: StringName, handler: Callable) -> void:
	if connection_service == null or !connection_service.has_signal(signal_name):
		return
	if !connection_service.is_connected(signal_name, handler):
		connection_service.connect(signal_name, handler)


func _reset_client_context() -> void:
	if measurement_context != null and measurement_context.has_method("reset"):
		measurement_context.reset()


func _reject(error_code: String, message: String) -> void:
	_record_error(error_code, message, {})


func _record_error(error_code: String, message: String, packet: Dictionary) -> void:
	error_received.emit(_report_state.record_error(error_code, message, packet))


func _emit_state_changed() -> void:
	state_changed.emit(get_state())
