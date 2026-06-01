package devtools

func IsCommandType(packetType string) bool {
	switch packetType {
	case PacketTypeToggleDebugInvincible,
		PacketTypeToggleDebugInfiniteLives,
		PacketTypeToggleDebugFreezeWorld,
		PacketTypeToggleDebugFreezePlayer,
		PacketTypeDebugKillPlayer,
		PacketTypeDebugSpawnEntity,
		PacketTypeDebugBeginContinuousBulletStream,
		PacketTypeDebugRespawnPlayer,
		PacketTypeDebugSetScore,
		PacketTypeDebugAddScore,
		PacketTypeDebugSetLives,
		PacketTypeDebugAddLives,
		PacketTypeDebugClearBullets,
		PacketTypeDebugClearAsteroids:
		return true
	default:
		return false
	}
}
