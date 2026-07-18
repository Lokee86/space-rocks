class_name ToolingPacketRouter
extends RefCounted

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")

signal telemetry_snapshot_received(packet: Dictionary)
signal telemetry_pong_received(packet: Dictionary)
signal measurement_started_received(packet: Dictionary)
signal measurement_snapshot_received(packet: Dictionary)
signal measurement_stopped_received(packet: Dictionary)
signal tooling_error_received(packet: Dictionary)
signal unknown_packet_received(packet: Dictionary)


func dispatch(packet: Dictionary) -> void:
	match str(packet.get(Packets.FIELD_TYPE, "")):
		Packets.TYPE_TELEMETRY_SNAPSHOT:
			telemetry_snapshot_received.emit(packet)
		Packets.TYPE_TELEMETRY_PONG:
			telemetry_pong_received.emit(packet)
		Packets.TYPE_MEASUREMENT_STARTED:
			measurement_started_received.emit(packet)
		Packets.TYPE_MEASUREMENT_SNAPSHOT:
			measurement_snapshot_received.emit(packet)
		Packets.TYPE_MEASUREMENT_STOPPED:
			measurement_stopped_received.emit(packet)
		Packets.TYPE_TOOLING_ERROR:
			tooling_error_received.emit(packet)
		_:
			unknown_packet_received.emit(packet)
