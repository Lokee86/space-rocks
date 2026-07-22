extends Node

var connection_service
var tree: SceneTree
var local_server_process


func configure(connection_service_ref, tree_ref: SceneTree, local_server_process_ref = null) -> void:
	connection_service = connection_service_ref
	tree = tree_ref
	local_server_process = local_server_process_ref


func request_shutdown() -> void:
	if connection_service != null && connection_service.is_server_connected():
		connection_service.begin_graceful_close()
	if local_server_process != null:
		local_server_process.stop()
	if tree != null:
		tree.quit()
