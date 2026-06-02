package roomrules

import "testing"

func TestDecideStart(t *testing.T) {
	tests := []struct {
		name    string
		input   StartInput
		allowed bool
		code    string
	}{
		{
			name: "owner all connected ready allows",
			input: StartInput{
				State:              "Lobby",
				OwnerID:            "player-1",
				RequestingPlayerID: "player-1",
				Members: []StartMember{
					{PlayerID: "player-1", Ready: true, Connected: true},
					{PlayerID: "player-2", Ready: true, Connected: true},
				},
			},
			allowed: true,
		},
		{
			name: "requesting player not in room rejects",
			input: StartInput{
				State:              "Lobby",
				OwnerID:            "player-1",
				RequestingPlayerID: "player-3",
				Members: []StartMember{
					{PlayerID: "player-1", Ready: true, Connected: true},
				},
			},
			code: "not_in_room",
		},
		{
			name: "non owner rejects",
			input: StartInput{
				State:              "Lobby",
				OwnerID:            "player-1",
				RequestingPlayerID: "player-2",
				Members: []StartMember{
					{PlayerID: "player-1", Ready: true, Connected: true},
					{PlayerID: "player-2", Ready: true, Connected: true},
				},
			},
			code: "not_room_owner",
		},
		{
			name: "connected unready member rejects",
			input: StartInput{
				State:              "Lobby",
				OwnerID:            "player-1",
				RequestingPlayerID: "player-1",
				Members: []StartMember{
					{PlayerID: "player-1", Ready: true, Connected: true},
					{PlayerID: "player-2", Ready: false, Connected: true},
				},
			},
			code: "not_ready",
		},
		{
			name: "disconnected unready member does not block",
			input: StartInput{
				State:              "Lobby",
				OwnerID:            "player-1",
				RequestingPlayerID: "player-1",
				Members: []StartMember{
					{PlayerID: "player-1", Ready: true, Connected: true},
					{PlayerID: "player-2", Ready: false, Connected: false},
				},
			},
			allowed: true,
		},
		{
			name: "starting rejects room in game",
			input: StartInput{
				State:              "Starting",
				OwnerID:            "player-1",
				RequestingPlayerID: "player-1",
				Members: []StartMember{
					{PlayerID: "player-1", Ready: true, Connected: true},
				},
			},
			code: "room_in_game",
		},
		{
			name: "in game rejects room in game",
			input: StartInput{
				State:              "InGame",
				OwnerID:            "player-1",
				RequestingPlayerID: "player-1",
				Members: []StartMember{
					{PlayerID: "player-1", Ready: true, Connected: true},
				},
			},
			code: "room_in_game",
		},
		{
			name: "closed rejects invalid room state",
			input: StartInput{
				State:              "Closed",
				OwnerID:            "player-1",
				RequestingPlayerID: "player-1",
				Members: []StartMember{
					{PlayerID: "player-1", Ready: true, Connected: true},
				},
			},
			code: "invalid_room_state",
		},
		{
			name: "game over rejects invalid room state",
			input: StartInput{
				State:              "GameOver",
				OwnerID:            "player-1",
				RequestingPlayerID: "player-1",
				Members: []StartMember{
					{PlayerID: "player-1", Ready: true, Connected: true},
				},
			},
			code: "invalid_room_state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecideStart(tt.input)
			if got.Allowed != tt.allowed {
				t.Fatalf("Allowed = %v, want %v", got.Allowed, tt.allowed)
			}
			if got.Code != tt.code {
				t.Fatalf("Code = %q, want %q", got.Code, tt.code)
			}
		})
	}
}
