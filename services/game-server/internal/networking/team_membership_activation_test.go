package networking

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestActivateRoomPlayersPassesLockedBotAssignmentToGame(t *testing.T) {
	room, err := rooms.NewRoomWithConfig("room", rooms.RoomStateLobby, nil, rooms.RoomCreationConfig{
		TeamConfig: teams.Config{Structure: teams.StructureCustom, AssignmentMode: teams.AssignmentOwnerAssigned},
	})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	owner := room.AddMember(rooms.NewRoomMember("session-owner"))
	bot, roomErr := room.AddBotForOwnerSession(owner.SessionID)
	if roomErr != nil {
		t.Fatalf("add bot: %v", roomErr)
	}
	if err := room.SetTeamAssignment(owner.SessionID, owner.PlayerID, teams.Team1); err != nil {
		t.Fatalf("assign owner: %v", err)
	}
	if err := room.SetTeamAssignment(owner.SessionID, bot.PlayerID, teams.Team2); err != nil {
		t.Fatalf("assign bot: %v", err)
	}
	if err := room.SetReadyForSessionInLobby(owner.SessionID, true); err != nil {
		t.Fatalf("ready owner: %v", err)
	}

	ownerSession := &webSocketSession{sessionID: owner.SessionID, outbound: make(chan []byte, 1)}
	attachRoomSession(room, ownerSession.sessionID, ownerSession)
	ownerSession.bindRoom(room)
	t.Cleanup(func() {
		detachRoomSession(room, ownerSession.sessionID)
		if instance := room.GameInstance(); instance != nil {
			instance.Stop()
		}
	})

	if err := room.StartGameForMember(owner.PlayerID, game.New); err != nil {
		t.Fatalf("start game: %v", err)
	}
	activateRoomPlayers(room)

	botGameID, ok := room.PlayerIDForSession(bot.SessionID)
	if !ok {
		t.Fatal("expected bot game player ID")
	}
	if got := room.GameInstance().PlayerTeam(botGameID); got != teams.Team2 {
		t.Fatalf("bot game team = %q, want %q", got, teams.Team2)
	}
}

func TestActivateRoomPlayersPassesLockedMemberAssignmentsToGame(t *testing.T) {
	tests := []struct {
		name   string
		config teams.Config
		assign []teams.ID
	}{
		{
			name:   "custom",
			config: teams.Config{Structure: teams.StructureCustom, AssignmentMode: teams.AssignmentOwnerAssigned},
			assign: []teams.ID{teams.Team3, teams.Team4},
		},
		{
			name:   "auto balanced",
			config: teams.Config{Structure: teams.StructureAutoBalanced, AutoTeamCount: 2},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			room, err := rooms.NewRoomWithConfig("room", rooms.RoomStateLobby, nil, rooms.RoomCreationConfig{TeamConfig: test.config})
			if err != nil {
				t.Fatalf("create room: %v", err)
			}
			first := room.AddMember(rooms.NewRoomMember("session-1"))
			second := room.AddMember(rooms.NewRoomMember("session-2"))
			if first == nil || second == nil {
				t.Fatal("expected both members to be added")
			}
			if len(test.assign) != 0 {
				if err := room.SetTeamAssignment("session-1", first.PlayerID, test.assign[0]); err != nil {
					t.Fatalf("assign first member: %v", err)
				}
				if err := room.SetTeamAssignment("session-1", second.PlayerID, test.assign[1]); err != nil {
					t.Fatalf("assign second member: %v", err)
				}
			}
			first.SetReady(true)
			second.SetReady(true)

			firstSession := &webSocketSession{sessionID: "session-1", outbound: make(chan []byte, 1)}
			secondSession := &webSocketSession{sessionID: "session-2", outbound: make(chan []byte, 1)}
			attachRoomSession(room, firstSession.sessionID, firstSession)
			attachRoomSession(room, secondSession.sessionID, secondSession)
			firstSession.bindRoom(room)
			secondSession.bindRoom(room)
			t.Cleanup(func() {
				detachRoomSession(room, firstSession.sessionID)
				detachRoomSession(room, secondSession.sessionID)
				if instance := room.GameInstance(); instance != nil {
					instance.Stop()
				}
			})

			if err := room.StartGameForMember(first.PlayerID, game.New); err != nil {
				t.Fatalf("start game: %v", err)
			}
			locked, ok := room.TeamStartSnapshot()
			if !ok {
				t.Fatal("expected locked team snapshot")
			}
			activateRoomPlayers(room)

			facts := room.GameInstance().PlayerMatchFacts()
			gotByGame := make(map[string]teams.ID, len(facts))
			for _, fact := range facts {
				gotByGame[fact.GamePlayerID] = fact.TeamID
			}
			firstGameID, ok := room.PlayerIDForSession("session-1")
			if !ok {
				t.Fatal("expected first session game player ID")
			}
			secondGameID, ok := room.PlayerIDForSession("session-2")
			if !ok {
				t.Fatal("expected second session game player ID")
			}
			wantByGame := map[string]teams.ID{
				firstGameID:  locked.Assignments[first.MemberID],
				secondGameID: locked.Assignments[second.MemberID],
			}
			if len(gotByGame) != len(wantByGame) {
				t.Fatalf("game facts = %#v, want %v", gotByGame, wantByGame)
			}
			for playerID, wantTeam := range wantByGame {
				if gotByGame[playerID] != wantTeam {
					t.Fatalf("%s team = %q, want %q", playerID, gotByGame[playerID], wantTeam)
				}
			}
		})
	}
}
