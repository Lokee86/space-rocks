package realtime

func (packet OverlayFullPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet OverlayFullPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindFull
}
func (packet OverlayFullPacket) PacketFamily() string { return packet.Type }
func (packet OverlayFullPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet OverlayFullPacket) WirePacket() map[string]any { return wireOverlayFullPacket(packet) }
func (OverlayFullPacket) realtimeLanePayload()              {}

func (packet OverlayWireFullPacket) Lane() Lane { return packet.Metadata.Lane }
func (packet OverlayWireFullPacket) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindFull
}
func (packet OverlayWireFullPacket) PacketFamily() string { return packet.Type }
func (packet OverlayWireFullPacket) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet OverlayWireFullPacket) WirePacket() map[string]any {
	return wireOverlayWireFullPacket(packet)
}
func (OverlayWireFullPacket) realtimeLanePayload() {}

func (packet OverlayLaneDelta) Lane() Lane { return packet.Metadata.Lane }
func (packet OverlayLaneDelta) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindDelta
}
func (OverlayLaneDelta) PacketFamily() string { return PacketFamilyOverlayDelta }
func (packet OverlayLaneDelta) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet OverlayLaneDelta) WirePacket() map[string]any { return wireOverlayDeltaPacket(packet) }
func (OverlayLaneDelta) realtimeLanePayload()              {}

func (packet OverlayWireLaneDelta) Lane() Lane { return packet.Metadata.Lane }
func (packet OverlayWireLaneDelta) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindDelta
}
func (OverlayWireLaneDelta) PacketFamily() string { return PacketFamilyOverlayDelta }
func (packet OverlayWireLaneDelta) LaneMetadata() (Metadata, bool) {
	return packet.Metadata, packet.Metadata != (Metadata{})
}
func (packet OverlayWireLaneDelta) WirePacket() map[string]any {
	return wireOverlayWireDeltaPacket(packet)
}
func (OverlayWireLaneDelta) realtimeLanePayload() {}
