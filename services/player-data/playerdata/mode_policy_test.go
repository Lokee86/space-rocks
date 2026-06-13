package playerdata

import (
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestValidateModeIdentity(t *testing.T) {
	tests := []struct {
		name     string
		playMode string
		identity protocol.PlayerDataIdentity
		wantErr  bool
	}{
		{
			name:     "single player guest allowed",
			playMode: PlayModeSinglePlayer,
			identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest},
		},
		{
			name:     "single player local profile allowed",
			playMode: PlayModeSinglePlayer,
			identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindLocalProfile},
		},
		{
			name:     "single player authenticated account rejected",
			playMode: PlayModeSinglePlayer,
			identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount},
			wantErr:  true,
		},
		{
			name:     "multiplayer authenticated account allowed",
			playMode: PlayModeMultiplayer,
			identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount},
		},
		{
			name:     "multiplayer guest rejected",
			playMode: PlayModeMultiplayer,
			identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest},
			wantErr:  true,
		},
		{
			name:     "multiplayer local profile rejected",
			playMode: PlayModeMultiplayer,
			identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindLocalProfile},
			wantErr:  true,
		},
		{
			name:     "multiplayer simulation authenticated account allowed",
			playMode: PlayModeMultiplayerSimulation,
			identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount},
		},
		{
			name:     "multiplayer simulation guest rejected",
			playMode: PlayModeMultiplayerSimulation,
			identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest},
			wantErr:  true,
		},
		{
			name:     "multiplayer simulation local profile rejected",
			playMode: PlayModeMultiplayerSimulation,
			identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindLocalProfile},
			wantErr:  true,
		},
		{
			name:     "unknown mode rejected",
			playMode: "arcade",
			identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest},
			wantErr:  true,
		},
		{
			name:     "unknown identity kind rejected",
			playMode: PlayModeSinglePlayer,
			identity: protocol.PlayerDataIdentity{IdentityKind: "unknown"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModeIdentity(tt.playMode, tt.identity)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ValidateModeIdentity returned nil error, want rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateModeIdentity returned error: %v", err)
			}
		})
	}
}
