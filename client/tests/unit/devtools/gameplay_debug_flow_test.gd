extends GutTest

const GameplayDebugFlow := preload("res://scripts/devtools/gameplay_debug_flow.gd")
const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")


class FakeConnectionService:
	var sent_packets: Array = []

	func send_packet(packet) -> void:
		sent_packets.append(packet)


func test_toggle_freeze_world_with_target_sends_targeted_packet() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.toggle_freeze_world("collisions")

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "toggle_debug_freeze_world")
	assert_eq(packet["freeze_target"], "collisions")


func test_toggle_freeze_world_without_target_sends_aggregate_packet() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.toggle_freeze_world()

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "toggle_debug_freeze_world")
	assert_false(packet.has("freeze_target"))


func test_toggle_invincible_all_players_sends_target_scope_without_target_player_id() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.toggle_invincible(DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, "")

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "toggle_debug_invincible")
	assert_eq(packet["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS)
	assert_false(packet.has("target_player_id"))


func test_toggle_infinite_lives_single_player_sends_scope_and_target_player_id() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.toggle_infinite_lives(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "player-2")

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "toggle_debug_infinite_lives")
	assert_eq(packet["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER)
	assert_eq(packet["target_player_id"], "player-2")


func test_toggle_freeze_player_hotkey_preserves_single_player_local_behavior_without_target_player_id() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.toggle_freeze_player()

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "toggle_debug_freeze_player")
	assert_eq(packet["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER)
	assert_false(packet.has("target_player_id"))


func test_set_score_all_players_sends_target_scope_without_target_player_id_and_preserves_score() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.set_score(DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, "", 42)

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "debug_set_score")
	assert_eq(packet["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS)
	assert_eq(packet["score"], 42)
	assert_false(packet.has("target_player_id"))


func test_add_score_single_player_sends_scope_target_player_id_and_preserves_amount() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.add_score(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "player-2", 5)

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "debug_add_score")
	assert_eq(packet["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER)
	assert_eq(packet["target_player_id"], "player-2")
	assert_eq(packet["amount"], 5)


func test_set_lives_all_players_sends_target_scope_without_target_player_id_and_preserves_lives() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.set_lives(DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, "", 3)

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "debug_set_lives")
	assert_eq(packet["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS)
	assert_eq(packet["lives"], 3)
	assert_false(packet.has("target_player_id"))


func test_add_lives_single_player_sends_scope_target_player_id_and_preserves_amount() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.add_lives(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "player-2", 2)

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "debug_add_lives")
	assert_eq(packet["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER)
	assert_eq(packet["target_player_id"], "player-2")
	assert_eq(packet["amount"], 2)
