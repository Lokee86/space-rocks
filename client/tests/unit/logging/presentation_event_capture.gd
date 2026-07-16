extends RefCounted

class_name PresentationEventCapture


var written_lines: Array[String] = []
var failure_count := 0
var last_failure_message := ""


func write_line(line: String) -> void:
	written_lines.append(line)


func close() -> void:
	pass


func last_record() -> Dictionary:
	if written_lines.is_empty():
		return {}
	var parsed = JSON.parse_string(written_lines[-1])
	return parsed if parsed is Dictionary else {}
