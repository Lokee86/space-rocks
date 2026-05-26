extends Node

var connection_service
var tree: SceneTree


func configure(connection_service_ref, tree_ref: SceneTree) -> void:
	connection_service = connection_service_ref
	tree = tree_ref


func request_shutdown() -> void:
	if connection_service != null && connection_service.is_server_connected():
		connection_service.begin_graceful_close()
	if tree != null:
		tree.quit()
