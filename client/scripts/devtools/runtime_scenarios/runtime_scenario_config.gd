extends RefCounted
class_name RuntimeScenarioConfig

const ARG_SCENARIO := "--runtime-scenario="
const ARG_ROLE := "--runtime-scenario-role="
const ARG_CLIENT_ID := "--runtime-scenario-client-id="
const ARG_ROOM_CODE := "--runtime-scenario-room-code="
const ARG_STATUS_PATH := "--runtime-scenario-status="
const ARG_SERVER_URL := "--runtime-scenario-server-url="


static func from_command_line() -> Dictionary:
	var arguments: Array = []
	arguments.append_array(OS.get_cmdline_args())
	arguments.append_array(OS.get_cmdline_user_args())
	return from_arguments(arguments)


static func from_arguments(arguments: Array) -> Dictionary:
	var scenario_path := _argument_value(arguments, ARG_SCENARIO)
	if scenario_path.is_empty():
		return {"enabled": false}
	var scenario := _load_scenario(scenario_path)
	if scenario.is_empty():
		return {
			"enabled": true,
			"valid": false,
			"error": "runtime scenario could not be loaded: %s" % scenario_path,
		}
	var role := _argument_value(arguments, ARG_ROLE)
	var client_id := _argument_value(arguments, ARG_CLIENT_ID)
	var status_path := _argument_value(arguments, ARG_STATUS_PATH)
	if role != "coordinator" and role != "participant":
		return _invalid("runtime scenario role must be coordinator or participant")
	if client_id.is_empty():
		return _invalid("runtime scenario client id is required")
	if status_path.is_empty():
		return _invalid("runtime scenario status path is required")
	scenario["enabled"] = true
	scenario["valid"] = true
	scenario["scenario_path"] = scenario_path
	scenario["role"] = role
	scenario["client_id"] = client_id
	scenario["room_code"] = _argument_value(arguments, ARG_ROOM_CODE)
	scenario["status_path"] = status_path
	scenario["server_url"] = _argument_value(arguments, ARG_SERVER_URL)
	return scenario


static func _load_scenario(path: String) -> Dictionary:
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		return {}
	var parsed = JSON.parse_string(file.get_as_text())
	file.close()
	if parsed is Dictionary:
		return parsed.duplicate(true)
	return {}


static func _argument_value(arguments: Array, prefix: String) -> String:
	for raw_argument in arguments:
		var argument := str(raw_argument)
		if argument.begins_with(prefix):
			return argument.trim_prefix(prefix).strip_edges()
	return ""


static func _invalid(message: String) -> Dictionary:
	return {
		"enabled": true,
		"valid": false,
		"error": message,
	}
