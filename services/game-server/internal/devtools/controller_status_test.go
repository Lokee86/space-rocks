package devtools

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/rules"
)

type controllerStatusTestTarget struct {
	Target
	matchDecision rules.MatchDecision
	targetIDs     []string
}

func (target *controllerStatusTestTarget) MatchDecision() rules.MatchDecision { return target.matchDecision }
func (target *controllerStatusTestTarget) TargetPlayerIDs() []string { return target.targetIDs }
func (target *controllerStatusTestTarget) WorldFrozen() bool { return false }
func (target *controllerStatusTestTarget) AsteroidsFrozen() bool { return false }
func (target *controllerStatusTestTarget) BulletsFrozen() bool { return false }
func (target *controllerStatusTestTarget) SpawningFrozen() bool { return false }
func (target *controllerStatusTestTarget) CollisionsFrozen() bool { return false }
func (target *controllerStatusTestTarget) PlayerInvincible(string) (bool, bool) { return false, false }
func (target *controllerStatusTestTarget) InfiniteLives(string) (bool, bool) { return false, false }
func (target *controllerStatusTestTarget) PlayerFrozen(string) (bool, bool) { return false, false }

func TestControllerStatusesForAllPlayersUsesMatchDecisionPlayers(t *testing.T) {
	target := &controllerStatusTestTarget{
		matchDecision: rules.MatchDecision{Players: []rules.PlayerDecision{{ID: "player-1", Status: rules.PlayerActive}, {ID: "player-2", Status: rules.PlayerEliminated}}},
		targetIDs:     []string{"player-1", "player-2", "player-3"},
	}

	statuses := NewController(Dependencies{Target: target}).StatusesForAllPlayers()
	if _, ok := statuses["player-1"]; !ok {
		t.Fatal("expected match/session player-1 in statuses")
	}
	if _, ok := statuses["player-2"]; !ok {
		t.Fatal("expected match/session player-2 in statuses")
	}
	if _, ok := statuses["player-3"]; ok {
		t.Fatal("expected ship-only player-3 to be absent from statuses")
	}

	if got := target.TargetPlayerIDs(); !reflect.DeepEqual(got, []string{"player-1", "player-2", "player-3"}) {
		t.Fatalf("TargetPlayerIDs() = %v, want %v", got, []string{"player-1", "player-2", "player-3"})
	}
}
