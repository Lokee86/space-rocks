extends RefCounted
class_name ClientMeasurementReportState

const COMBINED_REPORT_VERSION := 1
const MAX_RECENT_RESET_REQUESTS := 8

var request_ids: Dictionary = {}
var pending_request_ids: Dictionary = {}
var latest_server_snapshot: Dictionary = {}
var latest_server_report: Dictionary = {}
var latest_client_report: Dictionary = {}
var latest_combined_report: Dictionary = {}
var last_tooling_error: Dictionary = {}

var _request_sequence := 0
var _pending_requests: Dictionary = {}
var _recent_reset_request_ids: Array = []


func clear_for_new_run() -> void:
	latest_server_snapshot.clear()
	latest_server_report.clear()
	latest_client_report.clear()
	latest_combined_report.clear()
	last_tooling_error.clear()


func set_server_snapshot(report) -> Dictionary:
	latest_server_snapshot = report.duplicate(true) if report is Dictionary else {}
	return latest_server_snapshot.duplicate(true)


func set_server_report(report) -> Dictionary:
	latest_server_report = report.duplicate(true) if report is Dictionary else {}
	return latest_server_report.duplicate(true)


func set_client_report(report) -> Dictionary:
	latest_client_report = report.duplicate(true) if report is Dictionary else {}
	return latest_client_report.duplicate(true)


func build_combined_report(run_id: String) -> Dictionary:
	latest_combined_report = {
		"version": COMBINED_REPORT_VERSION,
		"run_id": run_id,
		"client": latest_client_report.duplicate(true),
		"server": latest_server_report.duplicate(true),
	}
	return latest_combined_report.duplicate(true)


func get_state(active_run_id: String, recording: bool) -> Dictionary:
	return {
		"active_run_id": active_run_id,
		"recording": recording,
		"request_ids": request_ids.duplicate(true),
		"pending_request_ids": pending_request_ids.duplicate(true),
		"latest_server_snapshot": latest_server_snapshot.duplicate(true),
		"latest_server_report": latest_server_report.duplicate(true),
		"latest_combined_report": latest_combined_report.duplicate(true),
		"last_tooling_error": last_tooling_error.duplicate(true),
	}


func record_error(error_code: String, message: String, packet: Dictionary) -> Dictionary:
	last_tooling_error = {
		"error_code": error_code,
		"message": message,
		"request_id": packet.get("request_id", ""),
		"run_id": packet.get("run_id", ""),
	}
	return last_tooling_error.duplicate(true)


func reserve_request(operation: String, run_id: String) -> String:
	_request_sequence += 1
	var request_id := "client-measurement-%d" % _request_sequence
	request_ids[operation] = request_id
	pending_request_ids[operation] = request_id
	_pending_requests[request_id] = {"operation": operation, "run_id": run_id}
	return request_id


func pending_request(request_id: String):
	return _pending_requests.get(request_id, null)


func pending_request_id(operation: String) -> String:
	return str(pending_request_ids.get(operation, ""))


func remove_pending(operation: String) -> void:
	var request_id := pending_request_id(operation)
	pending_request_ids.erase(operation)
	if !request_id.is_empty():
		_pending_requests.erase(request_id)


func clear_pending_requests() -> void:
	_pending_requests.clear()
	pending_request_ids.clear()
	_recent_reset_request_ids.clear()


func remember_reset_request(request_id: String) -> void:
	_recent_reset_request_ids.append(request_id)
	while _recent_reset_request_ids.size() > MAX_RECENT_RESET_REQUESTS:
		_recent_reset_request_ids.pop_front()


func consume_recent_reset_error(request_id: String) -> bool:
	if !_recent_reset_request_ids.has(request_id):
		return false
	_recent_reset_request_ids.erase(request_id)
	return true


func match_pending(packet: Dictionary, operation: String, require_active_run: bool, active_run_id: String) -> Dictionary:
	var request_id := str(packet.get("request_id", ""))
	var expected_request_id := pending_request_id(operation)
	if expected_request_id.is_empty() or request_id != expected_request_id:
		return {"error_code": "mismatched_measurement_request_id", "message": "measurement response did not match the pending request"}
	if require_active_run and (active_run_id.is_empty() or str(packet.get("run_id", "")) != active_run_id):
		return {"error_code": "mismatched_measurement_run_id", "message": "measurement response did not match the active run"}
	return {}


func resolve_tooling_error(packet: Dictionary) -> Dictionary:
	var request_id := str(packet.get("request_id", ""))
	var pending = pending_request(request_id)
	if pending == null:
		if consume_recent_reset_error(request_id):
			return {"error_code": packet.get("error_code", "tooling_error"), "message": packet.get("message", ""), "state_changed": true}
		return {"error_code": "unmatched_tooling_error", "message": "tooling error did not match a pending measurement request", "state_changed": false}
	var operation := str(pending.get("operation", ""))
	if str(pending.get("run_id", "")) != str(packet.get("run_id", "")):
		return {"error_code": "mismatched_measurement_run_id", "message": "tooling error run_id did not match the pending request", "state_changed": false}
	remove_pending(operation)
	return {"error_code": packet.get("error_code", "tooling_error"), "message": packet.get("message", ""), "state_changed": true}
