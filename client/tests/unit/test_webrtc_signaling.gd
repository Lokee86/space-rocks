extends GutTest

const ServerPacketRouter := preload("res://scripts/networking/inbound/server_packet_router.gd")
const ServerPacketDispatcher := preload("res://scripts/networking/inbound/server_packet_dispatcher.gd")


func test_server_packet_router_recognizes_webrtc_packet_types() -> void:
	var answer_packet := {"type": "webrtc_answer"}
	var ice_packet := {"type": "webrtc_ice_candidate"}
	var ready_packet := {"type": "webrtc_ready"}
	var smoke_packet := {"type": "webrtc_smoke"}
	var failed_packet := {"type": "webrtc_failed"}

	assert_true(ServerPacketRouter.is_webrtc_answer(answer_packet))
	assert_true(ServerPacketRouter.is_webrtc_ice_candidate(ice_packet))
	assert_true(ServerPacketRouter.is_webrtc_ready(ready_packet))
	assert_true(ServerPacketRouter.is_webrtc_smoke(smoke_packet))
	assert_true(ServerPacketRouter.is_webrtc_failed(failed_packet))


func test_server_packet_dispatcher_exposes_webrtc_signals() -> void:
	var dispatcher: Node = autofree(ServerPacketDispatcher.new())

	assert_true(dispatcher.has_signal("webrtc_answer_received"))
	assert_true(dispatcher.has_signal("webrtc_ice_candidate_received"))
	assert_true(dispatcher.has_signal("webrtc_ready_received"))
	assert_true(dispatcher.has_signal("webrtc_smoke_received"))
	assert_true(dispatcher.has_signal("webrtc_failed_received"))
