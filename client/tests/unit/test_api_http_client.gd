extends GutTest

const ApiHttpClient := preload("res://scripts/api/api_http_client.gd")


func test_trace_id_is_emitted_as_x_trace_id_header() -> void:
	var client := ApiHttpClient.new()
	var headers := client._build_headers("bearer-token", "00000000-0000-4000-8000-000000000061")

	assert_true(headers.has("Authorization: Bearer bearer-token"))
	assert_true(headers.has("X-Trace-ID: 00000000-0000-4000-8000-000000000061"))


func test_trace_header_is_omitted_when_trace_is_empty() -> void:
	var client := ApiHttpClient.new()
	var headers := client._build_headers("", "")

	assert_false(headers.has("X-Trace-ID: "))