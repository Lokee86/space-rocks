package rooms

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
)

func TestGameplayContextIsCoherentAndMatchesEachDimension(t *testing.T) {
	instance := game.New()
	room := NewRoom("room-1", RoomStateInGame, instance)
	room.match.currentMatchID = "match-1"

	context := room.GameplayContext()
	if context.State != RoomStateInGame || context.Game != instance || context.MatchID != "match-1" {
		t.Fatalf("unexpected gameplay context: %+v", context)
	}
	if !room.GameplayContextMatches(context) {
		t.Fatal("expected original context to match")
	}

	cases := []struct {
		name string
		edit func(*GameplayContext)
	}{
		{"state", func(value *GameplayContext) { value.State = RoomStateLobby }},
		{"game", func(value *GameplayContext) { value.Game = game.New() }},
		{"match", func(value *GameplayContext) { value.MatchID = "match-2" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changed := context
			tc.edit(&changed)
			if room.GameplayContextMatches(changed) {
				t.Fatal("expected changed context not to match")
			}
		})
	}
}

func TestSnapshotForSessionProjectsAndCopiesRoomView(t *testing.T) {
	room := NewRoom("room-1", RoomStateGameOver, game.New())
	room.match.currentMatchID = "match-1"
	member := room.AddMember(NewRoomMember("session-owner"))
	member.Ready = true
	member.AccountID = "account-1"
	room.match.resolvedSummary = &playerdata.MatchResultSummary{
		MatchID: "match-0",
		Mode:    playerdata.MatchModeMultiplayer,
		Players: []playerdata.PlayerMatchSummary{{GamePlayerID: member.PlayerID, Score: 42}},
	}

	projection := room.SnapshotForSession("session-owner")
	if projection.RoomID != "room-1" || projection.State != RoomStateGameOver || projection.CurrentMatchID != "match-1" || projection.LocalPlayerID != member.PlayerID || projection.OwnerID != member.PlayerID || !projection.HasResolvedMatch {
		t.Fatalf("unexpected projection: %+v", projection)
	}
	if len(projection.Members) != 1 || projection.Members[0].SessionID != "session-owner" || !projection.Members[0].Ready {
		t.Fatalf("unexpected members: %+v", projection.Members)
	}
	if projection.ResolvedSummary.MatchID != "match-0" || len(projection.ResolvedSummary.Players) != 1 {
		t.Fatalf("unexpected resolved summary: %+v", projection.ResolvedSummary)
	}

	projection.Members[0].Ready = false
	projection.ResolvedSummary.Players[0].Score = 999
	projection.ResolvedSummary.Players = append(projection.ResolvedSummary.Players, playerdata.PlayerMatchSummary{GamePlayerID: "mutated"})
	later := room.SnapshotForSession("session-owner")
	if !later.Members[0].Ready || later.ResolvedSummary.Players[0].Score != 42 || len(later.ResolvedSummary.Players) != 1 {
		t.Fatalf("projection mutation leaked into room: %+v", later)
	}
}
