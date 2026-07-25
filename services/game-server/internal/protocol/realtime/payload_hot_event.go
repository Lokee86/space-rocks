package realtime

func (packet ShipWireDeltaPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet ShipWireDeltaPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindDelta
}
func (packet ShipWireDeltaPacket) PacketFamily() string { return packet.Type }
func (packet ShipWireDeltaPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet ShipWireDeltaPacket) WirePacket() map[string]any {
	return wireShipWireDeltaPacket(packet)
}
func (ShipWireDeltaPacket) realtimeLanePayload() {}

func (packet AsteroidWireDeltaPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet AsteroidWireDeltaPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindDelta
}
func (packet AsteroidWireDeltaPacket) PacketFamily() string { return packet.Type }
func (packet AsteroidWireDeltaPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet AsteroidWireDeltaPacket) WirePacket() map[string]any {
	return wireAsteroidWireDeltaPacket(packet)
}
func (AsteroidWireDeltaPacket) realtimeLanePayload() {}

func (packet BulletWireDeltaPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet BulletWireDeltaPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindDelta
}
func (packet BulletWireDeltaPacket) PacketFamily() string { return packet.Type }
func (packet BulletWireDeltaPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet BulletWireDeltaPacket) WirePacket() map[string]any {
	return wireBulletWireDeltaPacket(packet)
}
func (BulletWireDeltaPacket) realtimeLanePayload() {}

func (packet EventBatchPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet EventBatchPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindEventBatch
}
func (packet EventBatchPacket) PacketFamily() string { return packet.Type }
func (packet EventBatchPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet EventBatchPacket) WirePacket() map[string]any { return wireEventBatchPacket(packet) }
func (EventBatchPacket) realtimeLanePayload()              {}
