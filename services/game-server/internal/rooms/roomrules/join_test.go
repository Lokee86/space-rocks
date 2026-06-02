package roomrules

import "testing"

func TestDecideJoin(t *testing.T) {
	tests := []struct {
		name     string
		input    JoinInput
		allowed  bool
		code     string
	}{
		{
			name: "lobby joinable with space allows",
			input: JoinInput{
				State:       "Lobby",
				Joinable:    true,
				MemberCount: 1,
				MaxMembers:  4,
			},
			allowed: true,
		},
		{
			name: "starting rejects room in game",
			input: JoinInput{
				State:       "Starting",
				Joinable:    true,
				MemberCount: 1,
				MaxMembers:  4,
			},
			code: "room_in_game",
		},
		{
			name: "in game rejects room in game",
			input: JoinInput{
				State:       "InGame",
				Joinable:    true,
				MemberCount: 1,
				MaxMembers:  4,
			},
			code: "room_in_game",
		},
		{
			name: "closed rejects room closed",
			input: JoinInput{
				State:       "Closed",
				Joinable:    true,
				MemberCount: 1,
				MaxMembers:  4,
			},
			code: "room_closed",
		},
		{
			name: "unknown state rejects invalid room state",
			input: JoinInput{
				State:       "mystery",
				Joinable:    true,
				MemberCount: 1,
				MaxMembers:  4,
			},
			code: "invalid_room_state",
		},
		{
			name: "non joinable rejects invalid room state",
			input: JoinInput{
				State:       "Lobby",
				Joinable:    false,
				MemberCount: 1,
				MaxMembers:  4,
			},
			code: "invalid_room_state",
		},
		{
			name: "full room rejects room full",
			input: JoinInput{
				State:       "Lobby",
				Joinable:    true,
				MemberCount: 4,
				MaxMembers:  4,
			},
			code: "room_full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecideJoin(tt.input)
			if got.Allowed != tt.allowed {
				t.Fatalf("Allowed = %v, want %v", got.Allowed, tt.allowed)
			}
			if got.Code != tt.code {
				t.Fatalf("Code = %q, want %q", got.Code, tt.code)
			}
		})
	}
}
