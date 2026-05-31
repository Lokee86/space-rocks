package devtools

func IsCommandType(packetType string) bool {
	switch packetType {
	case PacketTypeToggleDebugInvincible,
		PacketTypeToggleDebugInfiniteLives,
		PacketTypeToggleDebugFreezeWorld,
		PacketTypeToggleDebugFreezePlayer,
		PacketTypeDebugKillPlayer,
		PacketTypeDebugSpawnEntity,
		PacketTypeDebugRespawnPlayer:
		return true
	default:
		return false
	}
}
