extends RefCounted

const COMPRESSED_ARCHIVE_SUFFIX := ".gz"
const COMPRESSION_TEMP_SUFFIX := ".tmp"

var last_failure_message := ""


func compress(archive_path: String) -> bool:
	last_failure_message = ""
	var compressed_path := "%s%s" % [archive_path, COMPRESSED_ARCHIVE_SUFFIX]
	var temporary_path := "%s%s" % [compressed_path, COMPRESSION_TEMP_SUFFIX]
	if FileAccess.file_exists(compressed_path):
		return _fail("failed to compress archived log file because destination exists: %s" % compressed_path)

	var source = FileAccess.open(archive_path, FileAccess.READ)
	if source == null:
		return _fail("failed to read archived log file for compression: %s" % archive_path)
	var source_bytes = source.get_buffer(source.get_length())
	var source_error := source.get_error()
	source.close()
	if source_error != OK:
		return _fail("failed to read archived log file for compression: %s" % archive_path)

	var compressed = FileAccess.open_compressed(temporary_path, FileAccess.WRITE, FileAccess.COMPRESSION_GZIP)
	if compressed == null:
		return _fail("failed to open compressed archived log file: %s" % temporary_path)
	compressed.store_buffer(source_bytes)
	compressed.flush()
	var compression_error := compressed.get_error()
	compressed.close()
	if compression_error != OK:
		DirAccess.remove_absolute(temporary_path)
		return _fail("failed to write compressed archived log file: %s" % temporary_path)

	if DirAccess.rename_absolute(temporary_path, compressed_path) != OK:
		DirAccess.remove_absolute(temporary_path)
		return _fail("failed to finalize compressed archived log file: %s" % compressed_path)
	if DirAccess.remove_absolute(archive_path) != OK:
		return _fail("failed to remove uncompressed archived log file: %s" % archive_path)
	return true


func _fail(message: String) -> bool:
	last_failure_message = message
	return false