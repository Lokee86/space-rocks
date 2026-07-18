package lives

import (
	"testing"

	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestRuntimeFiniteLifecycleCooldownAndExhaustion(t *testing.T) {
	policy := NewBaselinePolicy()
	runtime := newRuntimeForTest(t, policy)
	registerForTest(t, runtime, "player-1", teams.NoTeam, 0)

	death := runtime.ApplyDeath(DeathInput{PlayerID: "player-1"})
	if !death.Accepted || death.ResultingStatus != playerstate.StatusPendingRespawn || death.RemainingLives != 2 {
		t.Fatalf("unexpected first death: %+v", death)
	}
	if runtime.CommitRespawn("player-1").Accepted {
		t.Fatal("respawn should remain blocked during cooldown")
	}
	runtime.Step(policy.RespawnDelay)
	if !runtime.CommitRespawn("player-1").Accepted {
		t.Fatal("expected respawn after cooldown")
	}

	for i := 0; i < 2; i++ {
		death = runtime.ApplyDeath(DeathInput{PlayerID: "player-1"})
		if !death.Accepted {
			t.Fatalf("death %d rejected: %+v", i+2, death)
		}
		if i == 0 {
			runtime.Step(policy.RespawnDelay)
			if !runtime.CommitRespawn("player-1").Accepted {
				t.Fatal("expected second respawn")
			}
		}
	}
	state, _ := runtime.ParticipantSnapshot("player-1")
	if state.Status != playerstate.StatusEliminated || state.RemainingLives != 0 || state.DeathCount != 3 {
		t.Fatalf("unexpected finite exhaustion: %+v", state)
	}
	if result := runtime.ApplyDeath(DeathInput{PlayerID: "player-1"}); result.Accepted || result.ReasonCode != "not_active" {
		t.Fatalf("repeated death mutated or accepted: %+v", result)
	}
}

func TestRuntimeInfiniteLivesRemainEffective(t *testing.T) {
	policy := NewBaselinePolicy()
	policy.Model = LifeModelInfinite
	policy.StartingLives = 0
	runtime := newRuntimeForTest(t, policy)
	registerForTest(t, runtime, "player-1", teams.NoTeam, 0)

	death := runtime.ApplyDeath(DeathInput{PlayerID: "player-1"})
	if !death.Accepted || death.RemainingLives != InfiniteLives || death.RespawnDelay != policy.RespawnDelay {
		t.Fatalf("unexpected infinite death: %+v", death)
	}
	state, _ := runtime.ParticipantSnapshot("player-1")
	if state.StartingLives != InfiniteLives || state.EffectiveLives != InfiniteLives {
		t.Fatalf("unexpected infinite state: %+v", state)
	}
	runtime.Step(policy.RespawnDelay)
	if !runtime.CommitRespawn("player-1").Accepted {
		t.Fatal("infinite participant should respawn")
	}
}

func TestRuntimeSharedTeamPoolExhaustsAcrossTeammates(t *testing.T) {
	policy := NewBaselinePolicy()
	policy.Model = LifeModelSharedTeamPool
	policy.StartingLives = 0
	policy.TeamPool = &TeamPoolPolicy{
		PoolID:        "team-pool-1",
		StartingLives: 2,
	}
	runtime := newRuntimeForTest(t, policy)
	registerForTest(t, runtime, "player-1", teams.Team1, 0)
	registerForTest(t, runtime, "player-2", teams.Team1, 0)

	if state, _ := runtime.ParticipantSnapshot("player-2"); state.EffectiveLives != 2 || state.RemainingLives != 0 {
		t.Fatalf("unexpected initial shared state: %+v", state)
	}
	if death := runtime.ApplyDeath(DeathInput{PlayerID: "player-1"}); death.RemainingLives != 1 || death.ResultingStatus != playerstate.StatusPendingRespawn {
		t.Fatalf("unexpected first shared death: %+v", death)
	}
	runtime.Step(policy.RespawnDelay)
	runtime.CommitRespawn("player-1")

	if death := runtime.ApplyDeath(DeathInput{PlayerID: "player-2"}); death.RemainingLives != 0 || death.ResultingStatus != playerstate.StatusEliminated {
		t.Fatalf("unexpected shared exhaustion death: %+v", death)
	}
	pool, ok := runtime.TeamPoolSnapshot(teams.Team1)
	if !ok || pool.RemainingLives != 0 {
		t.Fatalf("unexpected team pool snapshot: %+v, ok=%t", pool, ok)
	}
	if death := runtime.ApplyDeath(DeathInput{PlayerID: "player-1"}); death.ResultingStatus != playerstate.StatusEliminated {
		t.Fatalf("remaining teammate should exhaust with the shared pool: %+v", death)
	}
}

func TestRuntimeSharedPoolExhaustionClosesPendingTeammatesButLeavesActivePlayers(t *testing.T) {
	policy := NewBaselinePolicy()
	policy.Model = LifeModelSharedTeamPool
	policy.StartingLives = 0
	policy.TeamPool = &TeamPoolPolicy{PoolID: "team-pool-1", StartingLives: 2}
	runtime := newRuntimeForTest(t, policy)
	registerForTest(t, runtime, "player-1", teams.Team1, 0)
	registerForTest(t, runtime, "player-2", teams.Team1, 0)
	registerForTest(t, runtime, "player-3", teams.Team1, 0)
	if death := runtime.ApplyDeath(DeathInput{PlayerID: "player-1"}); death.ResultingStatus != playerstate.StatusPendingRespawn {
		t.Fatalf("first teammate death = %+v", death)
	}
	if death := runtime.ApplyDeath(DeathInput{PlayerID: "player-2"}); death.ResultingStatus != playerstate.StatusEliminated {
		t.Fatalf("pool-exhausting death = %+v", death)
	}
	if status, _ := runtime.Status("player-1"); status != playerstate.StatusEliminated {
		t.Fatalf("pending teammate was not closed: %q", status)
	}
	if status, _ := runtime.Status("player-3"); status != playerstate.StatusActive {
		t.Fatalf("active teammate was removed: %q", status)
	}
}

func TestRuntimeSharedLifeGrantRecoversOnlyTargetedParticipant(t *testing.T) {
	policy := NewBaselinePolicy()
	policy.Model = LifeModelSharedTeamPool
	policy.StartingLives = 0
	policy.TeamPool = &TeamPoolPolicy{PoolID: "team-pool-1", StartingLives: 1}
	runtime := newRuntimeForTest(t, policy)
	registerForTest(t, runtime, "player-1", teams.Team1, 0)
	registerForTest(t, runtime, "player-2", teams.Team1, 0)
	runtime.ApplyDeath(DeathInput{PlayerID: "player-1"})
	runtime.ApplyDeath(DeathInput{PlayerID: "player-2"})
	if change := runtime.AddLives("player-1", 1); !change.Accepted || change.ResultingStatus != playerstate.StatusPendingRespawn {
		t.Fatalf("targeted shared recovery = %+v", change)
	}
	if status, _ := runtime.Status("player-2"); status != playerstate.StatusEliminated {
		t.Fatalf("shared recovery fanned out to teammate: %q", status)
	}
	if change := runtime.AddLives("player-2", 1); !change.Accepted || change.ResultingStatus != playerstate.StatusPendingRespawn {
		t.Fatalf("positive shared grant did not recover later target: %+v", change)
	}
}

func TestRuntimeSharedPoolExhaustionDoesNotCloseInfiniteOverrideTeammate(t *testing.T) {
	policy := NewBaselinePolicy()
	policy.Model = LifeModelSharedTeamPool
	policy.StartingLives = 0
	policy.TeamPool = &TeamPoolPolicy{PoolID: "team-pool-1", StartingLives: 1}
	runtime := newRuntimeForTest(t, policy)
	registerForTest(t, runtime, "player-1", teams.Team1, 0)
	registerForTest(t, runtime, "player-2", teams.Team1, 0)
	runtime.SetInfiniteOverride("player-1", true)
	if death := runtime.ApplyDeath(DeathInput{PlayerID: "player-1"}); death.ResultingStatus != playerstate.StatusPendingRespawn {
		t.Fatalf("infinite override death = %+v", death)
	}
	runtime.ApplyDeath(DeathInput{PlayerID: "player-2"})
	if status, _ := runtime.Status("player-1"); status != playerstate.StatusPendingRespawn {
		t.Fatalf("infinite override teammate was closed: %q", status)
	}
}

func TestRuntimeRejectedTransitionsDoNotMutateState(t *testing.T) {
	runtime := newRuntimeForTest(t, NewBaselinePolicy())
	registerForTest(t, runtime, "player-1", teams.NoTeam, 0)

	before, _ := runtime.ParticipantSnapshot("player-1")
	if fact := runtime.EvaluateRespawn("player-1"); fact.Accepted || fact.ReasonCode != "already_active" {
		t.Fatalf("unexpected active respawn evaluation: %+v", fact)
	}
	after, _ := runtime.ParticipantSnapshot("player-1")
	if before != after {
		t.Fatalf("rejected active respawn mutated state: before=%+v after=%+v", before, after)
	}

	if !runtime.ApplyDeath(DeathInput{PlayerID: "player-1"}).Accepted {
		t.Fatal("expected death to be accepted")
	}
	before, _ = runtime.ParticipantSnapshot("player-1")
	if fact := runtime.ApplyDeath(DeathInput{PlayerID: "player-1"}); fact.Accepted || fact.ReasonCode != "not_active" {
		t.Fatalf("unexpected repeated death result: %+v", fact)
	}
	after, _ = runtime.ParticipantSnapshot("player-1")
	if before != after {
		t.Fatalf("rejected repeated death mutated state: before=%+v after=%+v", before, after)
	}
	if fact := runtime.ForceActivateForDevtools("player-1"); !fact.Accepted {
		t.Fatal("devtool force-active should be accepted")
	}
	if fact := runtime.ApplyDeath(DeathInput{PlayerID: "missing"}); fact.Accepted || fact.ReasonCode != "session_missing" {
		t.Fatalf("unexpected missing participant result: %+v", fact)
	}
}

func TestRuntimeLivesQueriesAndFiniteMutations(t *testing.T) {
	runtime := newRuntimeForTest(t, NewBaselinePolicy())
	registerForTest(t, runtime, "player-1", teams.NoTeam, 0)

	if lives, ok := runtime.EffectiveLives("player-1"); !ok || lives != 3 {
		t.Fatalf("unexpected effective lives: %d, ok=%t", lives, ok)
	}
	if status, ok := runtime.Status("player-1"); !ok || status != playerstate.StatusActive {
		t.Fatalf("unexpected status: %q, ok=%t", status, ok)
	}
	if cooldown, ok := runtime.RespawnCooldown("player-1"); !ok || cooldown != 0 {
		t.Fatalf("unexpected initial cooldown: %v, ok=%t", cooldown, ok)
	}
	if deaths, ok := runtime.DeathCount("player-1"); !ok || deaths != 0 {
		t.Fatalf("unexpected initial death count: %d, ok=%t", deaths, ok)
	}
	if respawns, ok := runtime.RespawnCount("player-1"); !ok || respawns != 0 {
		t.Fatalf("unexpected initial respawn count: %d, ok=%t", respawns, ok)
	}

	change := runtime.SetLives("player-1", 1)
	if !change.Accepted || change.PreviousLives != 3 || change.CurrentLives != 1 || change.Delta != -2 || change.ReasonCode != "lives_set" {
		t.Fatalf("unexpected set-lives change: %+v", change)
	}
	change = runtime.AddLives("player-1", -5)
	if !change.Accepted || change.PreviousLives != 1 || change.CurrentLives != 0 || change.Delta != -1 {
		t.Fatalf("unexpected clamped add-lives change: %+v", change)
	}
	state, _ := runtime.ParticipantSnapshot("player-1")
	if state.Status != playerstate.StatusActive || state.RemainingLives != 0 {
		t.Fatalf("numeric mutation changed lifecycle state: %+v", state)
	}

	runtime.ApplyDeath(DeathInput{PlayerID: "player-1"})
	if cooldown, _ := runtime.RespawnCooldown("player-1"); cooldown != 0 {
		t.Fatalf("unexpected exhausted cooldown: %v", cooldown)
	}
	if deaths, _ := runtime.DeathCount("player-1"); deaths != 1 {
		t.Fatalf("unexpected death count after death: %d", deaths)
	}
	change = runtime.SetLives("player-1", 4)
	if !change.Accepted {
		t.Fatalf("set lives after elimination rejected: %+v", change)
	}
	state, _ = runtime.ParticipantSnapshot("player-1")
	if state.Status != playerstate.StatusPendingRespawn || state.RespawnCooldown != 0 {
		t.Fatalf("adding lives did not restore respawn eligibility: %+v", state)
	}
}

func TestRuntimeSharedTeamMutationsUpdateEffectiveLives(t *testing.T) {
	policy := NewBaselinePolicy()
	policy.Model = LifeModelSharedTeamPool
	policy.StartingLives = 0
	policy.TeamPool = &TeamPoolPolicy{PoolID: "team-pool-1", StartingLives: 3}
	runtime := newRuntimeForTest(t, policy)
	registerForTest(t, runtime, "player-1", teams.Team1, 0)
	registerForTest(t, runtime, "player-2", teams.Team1, 0)

	change := runtime.SetLives("player-1", 1)
	if !change.Accepted || change.PreviousLives != 3 || change.CurrentLives != 1 {
		t.Fatalf("unexpected shared set-lives change: %+v", change)
	}
	if lives, _ := runtime.EffectiveLives("player-2"); lives != 1 {
		t.Fatalf("team mutation did not update teammate effective lives: %d", lives)
	}
	change = runtime.AddLives("player-2", 2)
	if !change.Accepted || change.PreviousLives != 1 || change.CurrentLives != 3 {
		t.Fatalf("unexpected shared add-lives change: %+v", change)
	}
	change = runtime.SetLives("player-1", -4)
	if !change.Accepted || change.CurrentLives != 0 {
		t.Fatalf("negative shared set was not clamped: %+v", change)
	}
	if status, _ := runtime.Status("player-1"); status != playerstate.StatusActive {
		t.Fatalf("shared numeric mutation changed status: %q", status)
	}
	if !runtime.SetInfiniteOverride("player-1", true) {
		t.Fatal("setting shared-team infinite override failed")
	}
	if death := runtime.ApplyDeath(DeathInput{PlayerID: "player-1"}); !death.Accepted || death.RemainingLives != InfiniteLives {
		t.Fatalf("shared-team override did not preserve effective lives: %+v", death)
	}
	pool, _ := runtime.TeamPoolSnapshot(teams.Team1)
	if pool.RemainingLives != 0 {
		t.Fatalf("shared-team override consumed the team pool: %+v", pool)
	}
}

func TestRuntimeInfiniteNumericMutationIsRejected(t *testing.T) {
	policy := NewBaselinePolicy()
	policy.Model = LifeModelInfinite
	policy.StartingLives = 0
	runtime := newRuntimeForTest(t, policy)
	registerForTest(t, runtime, "player-1", teams.NoTeam, 0)

	setChange := runtime.SetLives("player-1", 2)
	addChange := runtime.AddLives("player-1", 2)
	for _, change := range []LifeMutation{setChange, addChange} {
		if change.Accepted || change.ReasonCode != "numeric_lives_not_applicable" {
			t.Fatalf("unexpected infinite numeric mutation: %+v", change)
		}
	}
	if lives, _ := runtime.EffectiveLives("player-1"); lives != InfiniteLives {
		t.Fatalf("infinite effective lives changed: %d", lives)
	}
}

func TestRuntimeInfiniteOverrideDoesNotConsumeLivesOrResurrect(t *testing.T) {
	runtime := newRuntimeForTest(t, NewBaselinePolicy())
	registerForTest(t, runtime, "player-1", teams.NoTeam, 0)
	if !runtime.SetInfiniteOverride("player-1", true) {
		t.Fatal("setting infinite override failed")
	}
	if enabled, ok := runtime.InfiniteOverride("player-1"); !ok || !enabled {
		t.Fatalf("unexpected infinite override query: enabled=%t, ok=%t", enabled, ok)
	}
	if lives, _ := runtime.EffectiveLives("player-1"); lives != InfiniteLives {
		t.Fatalf("override did not make effective lives infinite: %d", lives)
	}

	death := runtime.ApplyDeath(DeathInput{PlayerID: "player-1"})
	if !death.Accepted || death.RemainingLives != InfiniteLives || death.RespawnDelay != NewBaselinePolicy().RespawnDelay {
		t.Fatalf("unexpected override death: %+v", death)
	}
	runtime.Step(NewBaselinePolicy().RespawnDelay)
	if !runtime.CommitRespawn("player-1").Accepted {
		t.Fatal("override participant should respawn")
	}
	state, _ := runtime.ParticipantSnapshot("player-1")
	if state.RemainingLives != 3 || state.Status != playerstate.StatusActive {
		t.Fatalf("override consumed or changed numeric lives: %+v", state)
	}

	if !runtime.SetInfiniteOverride("player-1", false) {
		t.Fatal("disabling infinite override failed")
	}
	runtime.ApplyDeath(DeathInput{PlayerID: "player-1"})
	state, _ = runtime.ParticipantSnapshot("player-1")
	if state.RemainingLives != 2 {
		t.Fatalf("disabled override did not restore finite consumption: %+v", state)
	}

	finalPolicy := NewBaselinePolicy()
	finalRuntime := newRuntimeForTest(t, finalPolicy)
	registerForTest(t, finalRuntime, "player-2", teams.NoTeam, 0)
	finalRuntime.SetLives("player-2", 0)
	finalRuntime.ApplyDeath(DeathInput{PlayerID: "player-2"})
	if !finalRuntime.SetInfiniteOverride("player-2", true) {
		t.Fatal("setting override on eliminated participant failed")
	}
	if status, _ := finalRuntime.Status("player-2"); status != playerstate.StatusEliminated {
		t.Fatalf("setting override silently resurrected participant: %q", status)
	}
}

func TestRuntimeDeathHistoryCopiesFactsAndSurvivesRemoval(t *testing.T) {
	runtime := newRuntimeForTest(t, NewBaselinePolicy())
	registerForTest(t, runtime, "player-1", teams.Team1, 0)
	input := DeathInput{
		PlayerID:        "player-1",
		DestroyedShipID: "ship-1",
		TeamID:          teams.Team1,
		MatchID:         "match-1",
		ModeID:          "mode-1",
		CauseCode:       "projectile",
		Attribution:     AttributionPlayerCaused,
		KillerPlayerID:  "player-2",
		AssistPlayerIDs: []string{"player-3"},
	}
	if fact := runtime.ApplyDeath(input); !fact.Accepted || fact.Input.Attribution != AttributionPlayerCaused {
		t.Fatalf("unexpected death fact: %+v", fact)
	}
	history, ok := runtime.DeathHistory("player-1")
	if !ok || len(history) != 1 {
		t.Fatalf("unexpected history: %+v, ok=%t", history, ok)
	}
	history[0].Input.AssistPlayerIDs[0] = "mutated"
	input.AssistPlayerIDs[0] = "mutated-input"
	history, _ = runtime.DeathHistory("player-1")
	if history[0].Input.AssistPlayerIDs[0] != "player-3" {
		t.Fatalf("history was not defensively copied: %+v", history[0])
	}
	if fact := runtime.RemoveParticipant("player-1", "player_removed"); !fact.Accepted {
		t.Fatalf("remove rejected: %+v", fact)
	}
	history, ok = runtime.DeathHistory("player-1")
	if !ok || len(history) != 1 || history[0].Input.MatchID != "match-1" {
		t.Fatalf("removed history was not retained: %+v, ok=%t", history, ok)
	}
}

func newRuntimeForTest(t *testing.T, policy Policy) *Runtime {
	t.Helper()
	runtime, err := NewRuntime(policy)
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	return runtime
}

func registerForTest(t *testing.T, runtime *Runtime, playerID string, teamID teams.ID, allowance float64) {
	t.Helper()
	if err := runtime.RegisterParticipant(ParticipantRegistration{
		PlayerID: playerID,
		TeamID:   teamID,
	}); err != nil {
		t.Fatalf("RegisterParticipant() error = %v", err)
	}
}
