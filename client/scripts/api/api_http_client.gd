extends Node
class_name ApiHttpClient

const ApiRequestResult := preload("res://scripts/api/api_request_result.gd")


func get_json(url: String, bearer_token: String = "") -> ApiRequestResult:
	return await _request_json(HTTPClient.METHOD_GET, url, {}, bearer_token)


func post_json(url: String, body: Dictionary = {}, bearer_token: String = "") -> ApiRequestResult:
	return await _request_json(HTTPClient.METHOD_POST, url, body, bearer_token)


func delete_json(url: String, body: Dictionary = {}, bearer_token: String = "") -> ApiRequestResult:
	return await _request_json(HTTPClient.METHOD_DELETE, url, body, bearer_token)


func _request_json(method: int, url: String, body: Dictionary, bearer_token: String) -> ApiRequestResult:
	var request := HTTPRequest.new()
	add_child(request)
	request.use_threads = true

	var headers := PackedStringArray([
		"Accept: application/json",
		"Content-Type: application/json"
	])
	if bearer_token != "":
		headers.append("Authorization: Bearer %s" % bearer_token)

	var payload := ""
	if method != HTTPClient.METHOD_GET:
		payload = JSON.stringify(body)

	var request_error := request.request(url, headers, method, payload)
	if request_error != OK:
		request.queue_free()
		return ApiRequestResult.failure(0, "request_failed")

	var completed: Array = await request.request_completed
	request.queue_free()

	var result_code: int = completed[0]
	var status_code: int = completed[1]
	var response_body: PackedByteArray = completed[3]
	var body_text := response_body.get_string_from_utf8()

	if result_code != HTTPRequest.RESULT_SUCCESS:
		return ApiRequestResult.failure(status_code, "network_failure")

	if body_text.is_empty():
		if status_code >= 200 and status_code < 300:
			return ApiRequestResult.success(status_code, {})
		return ApiRequestResult.failure(status_code, "http_%d" % status_code)

	var parser := JSON.new()
	var parse_error := parser.parse(body_text)
	if parse_error != OK:
		return ApiRequestResult.failure(status_code, "invalid_json")

	if typeof(parser.data) != TYPE_DICTIONARY:
		return ApiRequestResult.failure(status_code, "invalid_json")

	var parsed_body: Dictionary = parser.data
	if status_code < 200 or status_code >= 300:
		var error_message: String= parsed_body.get("error", parsed_body.get("message", "http_%d" % status_code))
		return ApiRequestResult.failure(status_code, str(error_message))

	return ApiRequestResult.success(status_code, parsed_body)
