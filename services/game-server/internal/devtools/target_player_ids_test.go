package devtools

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestResolveCommandTargetPlayerIDsReturnsAllPlayersForAllPlayersScope(t *testing.T) {
	target := game.New()
	firstPlayerID := target.AddPlayer()
	secondPlayerID := target.AddPlayer()

	got := resolveCommandTargetPlayerIDs(target, firstPlayerID, DebugCommand{
		TargetScope: targetScopeAllPlayers,
	})
	want := []string{firstPlayerID, secondPlayerID}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolveCommandTargetPlayerIDs() = %v, want %v", got, want)
	}
}

func TestResolveCommandTargetPlayerIDsUsesExplicitTargetPlayerIDForSinglePlayerScope(t *testing.T) {
	target := game.New()
	requestingPlayerID := target.AddPlayer()
	targetPlayerID := target.AddPlayer()

	got := resolveCommandTargetPlayerIDs(target, requestingPlayerID, DebugCommand{
		TargetScope:    targetScopeSinglePlayer,
		TargetPlayerID: targetPlayerID,
	})
	want := []string{targetPlayerID}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolveCommandTargetPlayerIDs() = %v, want %v", got, want)
	}
}

func TestResolveCommandTargetPlayerIDsFallsBackToRequestingPlayerForEmptyScope(t *testing.T) {
	target := game.New()
	requestingPlayerID := target.AddPlayer()

	got := resolveCommandTargetPlayerIDs(target, requestingPlayerID, DebugCommand{})
	want := []string{requestingPlayerID}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolveCommandTargetPlayerIDs() = %v, want %v", got, want)
	}
}

func TestResolveCommandTargetPlayerIDsTreatsUnknownScopeAsSinglePlayer(t *testing.T) {
	target := game.New()
	requestingPlayerID := target.AddPlayer()
	targetPlayerID := target.AddPlayer()

	got := resolveCommandTargetPlayerIDs(target, requestingPlayerID, DebugCommand{
		TargetScope:    "mystery_scope",
		TargetPlayerID: targetPlayerID,
	})
	want := []string{targetPlayerID}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolveCommandTargetPlayerIDs() = %v, want %v", got, want)
	}
}
