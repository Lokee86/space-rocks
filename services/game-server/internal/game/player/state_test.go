package player

import "testing"

func TestBuildWorldState_StatusAndCapabilities(t *testing.T) {
	tests := []struct {
		name               string
		input              BuildWorldStateInput
		wantStatus         Status
		wantHasActiveShip  bool
		wantTargetable     bool
		wantDamageable     bool
		wantCollidable     bool
	}{
		{
			name: "active",
			input: BuildWorldStateInput{
				ID:              "Player-1",
				HasActiveShip:   true,
				X:               10,
				Y:               20,
				Lives:           3,
				RespawnCooldown: 0,
			},
			wantStatus:        StatusActive,
			wantHasActiveShip: true,
			wantTargetable:    true,
			wantDamageable:    true,
			wantCollidable:    true,
		},
		{
			name: "pending respawn",
			input: BuildWorldStateInput{
				ID:              "Player-2",
				HasActiveShip:   false,
				X:               30,
				Y:               40,
				Lives:           2,
				RespawnCooldown: 1.5,
			},
			wantStatus:        StatusPendingRespawn,
			wantHasActiveShip: false,
			wantTargetable:    false,
			wantDamageable:    false,
			wantCollidable:    false,
		},
		{
			name: "eliminated",
			input: BuildWorldStateInput{
				ID:              "Player-3",
				HasActiveShip:   false,
				X:               50,
				Y:               60,
				Lives:           0,
				RespawnCooldown: 0,
			},
			wantStatus:        StatusEliminated,
			wantHasActiveShip: false,
			wantTargetable:    false,
			wantDamageable:    false,
			wantCollidable:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildWorldState(tt.input)
			if got.Status != tt.wantStatus {
				t.Fatalf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.HasActiveShip != tt.wantHasActiveShip {
				t.Fatalf("HasActiveShip = %t, want %t", got.HasActiveShip, tt.wantHasActiveShip)
			}
			if got.Targetable != tt.wantTargetable {
				t.Fatalf("Targetable = %t, want %t", got.Targetable, tt.wantTargetable)
			}
			if got.Damageable != tt.wantDamageable {
				t.Fatalf("Damageable = %t, want %t", got.Damageable, tt.wantDamageable)
			}
			if got.Collidable != tt.wantCollidable {
				t.Fatalf("Collidable = %t, want %t", got.Collidable, tt.wantCollidable)
			}
		})
	}
}
