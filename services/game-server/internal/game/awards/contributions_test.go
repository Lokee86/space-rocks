package awards

import "testing"

func TestAssistResolutionSupportsMultipleRecipientsAndExcludesKiller(t *testing.T) {
	runtime := NewRuntime()
	policy := DefaultAssistPolicy()
	for _, contribution := range []Contribution{
		{TargetID: "target", PlayerID: "killer", Amount: 70, At: 10},
		{TargetID: "target", PlayerID: "assist-a", Amount: 20, At: 10},
		{TargetID: "target", PlayerID: "assist-b", Amount: 10, At: 10},
	} {
		if err := runtime.RecordContribution(contribution); err != nil {
			t.Fatalf("record contribution: %v", err)
		}
	}
	credits := runtime.ResolveAssists("target", "killer", 11, policy, map[string]bool{"killer": true, "assist-a": true, "assist-b": true})
	if len(credits) != 2 || credits[0].PlayerID != "assist-a" || credits[1].PlayerID != "assist-b" {
		t.Fatalf("unexpected assist credits: %+v", credits)
	}
}

func TestContributionWindowAndRetentionBuffer(t *testing.T) {
	runtime := NewRuntime()
	policy := DefaultAssistPolicy()
	_ = runtime.RecordContribution(Contribution{TargetID: "target", PlayerID: "old", Amount: 50, At: 1})
	_ = runtime.RecordContribution(Contribution{TargetID: "target", PlayerID: "recent", Amount: 50, At: 6})

	credits := runtime.ResolveAssists("target", "", 10, policy, nil)
	if len(credits) != 1 || credits[0].PlayerID != "recent" {
		t.Fatalf("unexpected window credits: %+v", credits)
	}
	runtime.PruneContributions(6.4, policy)
	if runtime.ContributionCount("target") != 2 {
		t.Fatalf("buffer should retain both contributions, got %d", runtime.ContributionCount("target"))
	}
	runtime.PruneContributions(6.6, policy)
	if runtime.ContributionCount("target") != 1 {
		t.Fatalf("old contribution should expire after buffer, got %d", runtime.ContributionCount("target"))
	}
}
