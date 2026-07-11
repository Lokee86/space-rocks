package realtime

func (packet WorldFullPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet WorldFullPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindFull
}
func (packet WorldFullPacket) PacketFamily() string { return packet.Type }
func (packet WorldFullPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet WorldFullPacket) WirePacket() map[string]any { return wireWorldFullPacket(packet) }
func (WorldFullPacket) realtimeLanePayload()              {}

func (packet WorldWireFullPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet WorldWireFullPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindFull
}
func (packet WorldWireFullPacket) PacketFamily() string { return packet.Type }
func (packet WorldWireFullPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet WorldWireFullPacket) WirePacket() map[string]any { return wireWorldWireFullPacket(packet) }
func (WorldWireFullPacket) realtimeLanePayload()              {}

func (packet WorldDeltaPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet WorldDeltaPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindDelta
}
func (packet WorldDeltaPacket) PacketFamily() string { return packet.Type }
func (packet WorldDeltaPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet WorldDeltaPacket) WirePacket() map[string]any { return wireWorldDeltaPacket(packet) }
func (WorldDeltaPacket) realtimeLanePayload()              {}

func (packet WorldWireDeltaPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet WorldWireDeltaPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindDelta
}
func (packet WorldWireDeltaPacket) PacketFamily() string { return packet.Type }
func (packet WorldWireDeltaPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet WorldWireDeltaPacket) WirePacket() map[string]any { return wireWorldWireDeltaPacket(packet) }
func (WorldWireDeltaPacket) realtimeLanePayload()              {}
