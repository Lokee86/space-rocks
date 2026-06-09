extends RefCounted
class_name ApiRequestResult

var ok: bool = false
var status_code: int = 0
var body: Dictionary = {}
var error_message: String = ""


static func success(status_code_value: int, body_value: Dictionary):
	var result := ApiRequestResult.new()
	result.ok = true
	result.status_code = status_code_value
	result.body = body_value
	return result


static func failure(status_code_value: int, error_message_value: String):
	var result := ApiRequestResult.new()
	result.status_code = status_code_value
	result.error_message = error_message_value
	return result
