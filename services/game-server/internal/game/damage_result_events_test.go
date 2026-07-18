package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/events"
)

func TestDamageResultEventDistinguishesOutcomeKinds(t *testing.T) {
	tests := []struct {
		name   string
		result damage.DamageResult
		want   events.Type
	}{
		{name: "blocked", result: damage.DamageResult{Kind: damage.DamageResultKindBlocked, Reason: "invulnerable"}, want: events.EventDamageBlocked},
		{name: "healing", result: damage.DamageResult{Kind: damage.DamageResultKindHealing, RestoredToHealth: 3}, want: events.EventHealingApplied},
		{name: "repair", result: damage.DamageResult{Kind: damage.DamageResultKindRepair, RestoredToShield: 2}, want: events.EventRepairApplied},
		{name: "ineffective", result: damage.DamageResult{Kind: damage.DamageResultKindIneffective}, want: events.EventDamageIneffective},
		{name: "discarded", result: damage.DamageResult{Kind: damage.DamageResultKindDiscardedLethalTarget}, want: events.EventDamageDiscarded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, ok := damageResultEvent(test.result, 10, 20)
			if !ok || event.Type != test.want {
				t.Fatalf("event=%+v ok=%v, want type %q", event, ok, test.want)
			}
			state := eventStateForDomainEvent(event)
			if state.Type != string(test.want) {
				t.Fatalf("event state type = %q, want %q", state.Type, test.want)
			}
		})
	}
}
