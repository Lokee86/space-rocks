extends RefCounted

const PLAYER_DEV_LABEL_SCENE := preload("res://scenes/devtools/player_dev_label.tscn")
const PlayerDevLabelFormatter := preload("res://scripts/devtools/player_dev_label_formatter.gd")

var remote_player_nodes_provider: Callable
var labels_by_player_id := {}
var mode := ""
var latest_gameplay_state := {}
var latest_network_metrics := {}


func configure(provider: Callable) -> void:
	remote_player_nodes_provider = provider


func apply_gameplay_state(state: Dictionary) -> void:
	latest_gameplay_state = state.duplicate()


func apply_network_metrics(metrics: Dictionary) -> void:
	latest_network_metrics = metrics.duplicate()


func set_mode(next_mode: String) -> void:
	if next_mode != "" and next_mode != "basic" and next_mode != "network":
		next_mode = ""
	mode = next_mode
	if mode == "":
		clear_labels()


func clear_labels() -> void:
	for label in labels_by_player_id.values():
		if is_instance_valid(label):
			label.queue_free()
	labels_by_player_id.clear()


func sync_remote_labels() -> void:
	if mode == "":
		clear_labels()
		return
	if remote_player_nodes_provider.is_null() or !remote_player_nodes_provider.is_valid():
		clear_labels()
		return

	var remote_player_nodes = remote_player_nodes_provider.call()
	if not (remote_player_nodes is Dictionary):
		clear_labels()
		return

	for player_id in labels_by_player_id.keys():
		if remote_player_nodes.has(player_id):
			continue
		var stale_label = labels_by_player_id[player_id]
		if is_instance_valid(stale_label):
			stale_label.queue_free()
		labels_by_player_id.erase(player_id)

	for player_id in remote_player_nodes.keys():
		var player_node = remote_player_nodes[player_id]
		if !is_instance_valid(player_node):
			continue
		if labels_by_player_id.has(player_id):
			var existing_label = labels_by_player_id[player_id]
			if is_instance_valid(existing_label):
				continue

		var label = PLAYER_DEV_LABEL_SCENE.instantiate()
		if label.has_method("configure_as_player_child"):
			label.configure_as_player_child()
		player_node.add_child(label)
		if label.has_method("hide_label"):
			label.hide_label()
		labels_by_player_id[player_id] = label

	if mode == "network":
		var network_text := PlayerDevLabelFormatter.network_text(latest_network_metrics)
		for player_id in labels_by_player_id.keys():
			var network_label = labels_by_player_id[player_id]
			if !is_instance_valid(network_label):
				continue
			if network_label.has_method("show_network"):
				network_label.show_network(network_text)
		return

	if mode != "basic":
		return

	var world: Dictionary = latest_gameplay_state.get("world", {})
	var world_ships: Dictionary = world.get("ships", {})
	var session: Dictionary = latest_gameplay_state.get("session", {})
	var session_players: Dictionary = session.get("players", {})
	for player_id in labels_by_player_id.keys():
		var label = labels_by_player_id[player_id]
		if !is_instance_valid(label):
			continue
		if !world_ships.has(player_id):
			if label.has_method("hide_label"):
				label.hide_label()
			continue

		var state = world_ships[player_id]
		if not (state is Dictionary):
			if label.has_method("hide_label"):
				label.hide_label()
			continue

		var session_state = session_players.get(player_id, {})
		if !(session_state is Dictionary):
			session_state = {}
		var text := PlayerDevLabelFormatter.basic_player_text(player_id, state, session_state)
		if label.has_method("show_basic"):
			label.show_basic(text)
