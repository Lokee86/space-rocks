package awards

import (
	"fmt"
	"sort"
)

type Contribution struct {
	TargetID string
	PlayerID string
	TeamID   string
	Amount   float64
	At       float64
}

type AssistPolicy struct {
	Enabled                 bool
	ThresholdFraction       float64
	WindowSeconds           float64
	RetentionBufferFraction float64
	KillerEligible          bool
}

func DefaultAssistPolicy() AssistPolicy {
	return AssistPolicy{
		Enabled:                 true,
		ThresholdFraction:       0.05,
		WindowSeconds:           5.0,
		RetentionBufferFraction: 0.10,
	}
}

type AssistCredit struct {
	PlayerID string
	TeamID   string
	Amount   float64
	Fraction float64
}

func (runtime *Runtime) RecordContribution(contribution Contribution) error {
	if contribution.TargetID == "" || contribution.PlayerID == "" {
		return fmt.Errorf("contribution target and player are required")
	}
	if contribution.Amount <= 0 {
		return fmt.Errorf("contribution amount must be positive")
	}
	runtime.contributions[contribution.TargetID] = append(runtime.contributions[contribution.TargetID], contribution)
	return nil
}

func (runtime *Runtime) ResolveAssists(targetID string, killerID string, now float64, policy AssistPolicy, eligible map[string]bool) []AssistCredit {
	if !policy.Enabled || policy.WindowSeconds <= 0 || policy.ThresholdFraction < 0 {
		return nil
	}
	byPlayer := make(map[string]AssistCredit)
	total := 0.0
	cutoff := now - policy.WindowSeconds
	for _, contribution := range runtime.contributions[targetID] {
		if contribution.At < cutoff || contribution.At > now {
			continue
		}
		total += contribution.Amount
		credit := byPlayer[contribution.PlayerID]
		credit.PlayerID = contribution.PlayerID
		credit.TeamID = contribution.TeamID
		credit.Amount += contribution.Amount
		byPlayer[contribution.PlayerID] = credit
	}
	if total <= 0 {
		return nil
	}
	result := make([]AssistCredit, 0)
	for playerID, credit := range byPlayer {
		if playerID == killerID && !policy.KillerEligible {
			continue
		}
		if eligible != nil && !eligible[playerID] {
			continue
		}
		credit.Fraction = credit.Amount / total
		if credit.Fraction >= policy.ThresholdFraction {
			result = append(result, credit)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].PlayerID < result[j].PlayerID })
	return result
}

func (runtime *Runtime) PruneContributions(now float64, policy AssistPolicy) {
	retention := policy.WindowSeconds * (1 + policy.RetentionBufferFraction)
	if retention <= 0 {
		runtime.contributions = make(map[string][]Contribution)
		return
	}
	cutoff := now - retention
	for targetID, contributions := range runtime.contributions {
		kept := contributions[:0]
		for _, contribution := range contributions {
			if contribution.At >= cutoff {
				kept = append(kept, contribution)
			}
		}
		if len(kept) == 0 {
			delete(runtime.contributions, targetID)
		} else {
			runtime.contributions[targetID] = kept
		}
	}
}

func (runtime *Runtime) ClearContributions(targetID string) {
	delete(runtime.contributions, targetID)
}

func (runtime *Runtime) ContributionCount(targetID string) int {
	return len(runtime.contributions[targetID])
}
