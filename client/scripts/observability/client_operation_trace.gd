extends RefCounted
class_name ClientOperationTrace

var _trace_id := ""
var _operation := ""


func _init(operation_name: String, uuid_generator: Callable = Callable()) -> void:
	_operation = operation_name
	_trace_id = str(uuid_generator.call()) if uuid_generator.is_valid() else _new_uuid()


static func create(operation_name: String, factory: Callable = Callable()) -> ClientOperationTrace:
	if factory.is_valid():
		var trace = factory.call(operation_name)
		if trace is ClientOperationTrace:
			return trace
	return ClientOperationTrace.new(operation_name)


func trace_id() -> String:
	return _trace_id


func operation() -> String:
	return _operation


func _new_uuid() -> String:
	var bytes := Crypto.new().generate_random_bytes(16)
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	var hex := bytes.hex_encode()
	return "%s-%s-%s-%s-%s" % [hex.substr(0, 8), hex.substr(8, 4), hex.substr(12, 4), hex.substr(16, 4), hex.substr(20, 12)]
