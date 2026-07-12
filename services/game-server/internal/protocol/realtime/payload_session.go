package realtime

func (packet SessionFullPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet SessionFullPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindFull
}
func (packet SessionFullPacket) PacketFamily() string { return packet.Type }
func (packet SessionFullPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet SessionFullPacket) WirePacket() map[string]any { return wireSessionFullPacket(packet) }
func (SessionFullPacket) realtimeLanePayload()              {}

func (packet SessionWireFullPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet SessionWireFullPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindFull
}
func (packet SessionWireFullPacket) PacketFamily() string { return packet.Type }
func (packet SessionWireFullPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet SessionWireFullPacket) WirePacket() map[string]any {
	return wireSessionWireFullPacket(packet)
}
func (SessionWireFullPacket) realtimeLanePayload() {}

func (packet SessionLaneDelta) Lane() Lane { return packet.Metadata.Lane }
func (packet SessionLaneDelta) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindDelta
}
func (SessionLaneDelta) PacketFamily() string { return PacketFamilySessionDelta }
func (packet SessionLaneDelta) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet SessionLaneDelta) WirePacket() map[string]any { return wireSessionDeltaPacket(packet) }
func (SessionLaneDelta) realtimeLanePayload()              {}

func (packet SessionWireLaneDelta) Lane() Lane { return packet.Metadata.Lane }
func (packet SessionWireLaneDelta) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindDelta
}
func (SessionWireLaneDelta) PacketFamily() string { return PacketFamilySessionDelta }
func (packet SessionWireLaneDelta) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet SessionWireLaneDelta) WirePacket() map[string]any {
	return wireSessionWireDeltaPacket(packet)
}
func (SessionWireLaneDelta) realtimeLanePayload() {}
