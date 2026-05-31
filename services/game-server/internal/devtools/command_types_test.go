package devtools

import "testing"

func TestIsCommandType(t *testing.T) {
	tests := []struct {
		packetType string
		want       bool
	}{
		{packetType: PacketTypeToggleDebugInvincible, want: true},
		{packetType: PacketTypeToggleDebugInfiniteLives, want: true},
		{packetType: PacketTypeToggleDebugFreezeWorld, want: true},
		{packetType: PacketTypeToggleDebugFreezePlayer, want: true},
		{packetType: PacketTypeDebugKillPlayer, want: true},
		{packetType: PacketTypeDebugSpawnEntity, want: true},
		{packetType: PacketTypeDebugRespawnPlayer, want: true},
		{packetType: PacketTypeDebugSetScore, want: true},
		{packetType: PacketTypeDebugAddScore, want: true},
		{packetType: PacketTypeDebugSetLives, want: true},
		{packetType: PacketTypeDebugAddLives, want: true},
		{packetType: PacketTypeDebugClearBullets, want: true},
		{packetType: PacketTypeDebugClearAsteroids, want: true},
		{packetType: "input", want: false},
	}

	for _, tt := range tests {
		if got := IsCommandType(tt.packetType); got != tt.want {
			t.Fatalf("IsCommandType(%q) = %v, want %v", tt.packetType, got, tt.want)
		}
	}
}
